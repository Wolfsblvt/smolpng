// SPDX-License-Identifier: MIT
// Package tinypng owns authenticated Tinify HTTP, with no endpoint override in product configuration.
package tinypng

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/secret"
)

const Endpoint = "https://api.tinify.com/shrink"

type Request struct {
	Data     []byte
	MaxBytes int64
}
type Response struct {
	Data  []byte
	Usage domain.RemoteUsage
}
type Optimizer interface {
	Optimize(context.Context, Request) (Response, error)
}
type Client struct {
	Credentials secret.Resolver
	HTTP        *http.Client
}

func (c Client) Optimize(ctx context.Context, in Request) (out Response, err error) {
	out.Usage.Provider = "tinypng"
	key, source, e := c.Credentials.Resolve(ctx)
	out.Usage.CredentialSource = source
	if e != nil {
		return out, domain.E(7, e.Error(), e)
	}
	if in.MaxBytes < 1 || int64(len(in.Data)) > in.MaxBytes {
		return out, domain.E(3, "remote input exceeds configured bound", nil)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	if c.HTTP != nil {
		copy := *c.HTTP
		client = &copy
	}
	// Neither redirects nor provider-supplied output URLs may forward authorization to another origin.
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, Endpoint, bytes.NewReader(in.Data))
	request.SetBasicAuth("api", key)
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("User-Agent", "smolpng")
	// A POST may have been accepted even if its response is lost. Never retry it automatically.
	out.Usage.Attempted = true
	res, e := client.Do(request)
	if e != nil {
		return out, safeError(ctx, "TinyPNG upload failed; the request is not retried automatically because it may have consumed quota", e)
	}
	readCount(&out.Usage, res.Header)
	location := res.Header.Get("Location")
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 16<<10))
	res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return out, statusError(res.StatusCode)
	}
	u, e := url.Parse(location)
	if e != nil || u.Scheme != "https" || u.Host != "api.tinify.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !strings.HasPrefix(u.EscapedPath(), "/output/") || strings.Contains(u.EscapedPath(), "%") || strings.Contains(u.Path, "..") {
		return out, domain.E(7, "TinyPNG returned an untrusted output location; download refused", nil)
	}
	request, _ = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	request.SetBasicAuth("api", key)
	request.Header.Set("User-Agent", "smolpng")
	res, e = client.Do(request)
	if e != nil {
		return out, safeError(ctx, "TinyPNG output download failed", e)
	}
	defer res.Body.Close()
	readCount(&out.Usage, res.Header)
	if res.StatusCode != http.StatusOK {
		return out, statusError(res.StatusCode)
	}
	if res.ContentLength > in.MaxBytes {
		return out, domain.E(7, "TinyPNG output exceeds configured byte limit", nil)
	}
	out.Data, e = pngimage.ReadBounded(res.Body, in.MaxBytes)
	if e != nil {
		return out, safeError(ctx, "TinyPNG output could not be read within configured bounds", e)
	}
	return out, nil
}
func readCount(u *domain.RemoteUsage, h http.Header) {
	if v, e := strconv.Atoi(h.Get("Compression-Count")); e == nil && v >= 0 {
		u.CompressionCount = &v
	}
}
func safeError(ctx context.Context, message string, cause error) error {
	if ctx.Err() != nil {
		return domain.E(domain.Code(ctx.Err()), "TinyPNG operation cancelled or timed out", ctx.Err())
	}
	return domain.E(7, message, cause)
}
func statusError(code int) error {
	message := fmt.Sprintf("TinyPNG returned HTTP %d", code)
	switch code {
	case 401:
		message = "TinyPNG rejected the API key (HTTP 401)"
	case 429:
		message = "TinyPNG quota or rate limit reached (HTTP 429); retry explicitly later"
	}
	return domain.E(7, message, nil)
}
