// SPDX-License-Identifier: MIT
package updates

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"io"
	"net/http"
	"os"
	"time"
)

func Fetch(ctx context.Context, url string, client *http.Client, limit int64) ([]byte, error) {
	if e := HTTPS(url); e != nil {
		return nil, e
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if client == nil {
		client = &http.Client{}
	}
	c := *client
	c.CheckRedirect = func(r *http.Request, via []*http.Request) error {
		if len(via) > 5 {
			return fmt.Errorf("too many redirects")
		}
		return HTTPS(r.URL.String())
	}
	request, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return nil, e
	}
	request.Header.Set("User-Agent", "smolpng-update")
	response, e := c.Do(request)
	if e != nil {
		return nil, fmt.Errorf("release resource download failed")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("release resource returned HTTP %d", response.StatusCode)
	}
	if response.ContentLength > limit {
		return nil, fmt.Errorf("release resource exceeds byte limit")
	}
	return pngimage.ReadBounded(response.Body, limit)
}
func ReadFile(path string, limit int64) ([]byte, error) {
	if e := noLinks(path); e != nil {
		return nil, e
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() || s.Size() > limit {
		return nil, fmt.Errorf("not a bounded regular file")
	}
	return pngimage.ReadBounded(f, limit)
}
func SHA(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func fileHash(path string) (string, int64, error) {
	if e := noLinks(path); e != nil {
		return "", 0, e
	}
	f, e := os.Open(path)
	if os.IsNotExist(e) {
		return "", 0, nil
	}
	if e != nil {
		return "", 0, e
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() || s.Size() > MaxFile {
		return "", 0, fmt.Errorf("not a bounded regular update file")
	}
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, MaxFile+1))
	if e != nil || n > MaxFile {
		return "", 0, fmt.Errorf("update member exceeds limit")
	}
	return fmt.Sprintf("%x", h.Sum(nil)), n, nil
}
