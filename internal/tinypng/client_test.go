package tinypng

import (
	"context"
	"errors"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/secret"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func fake(t *testing.T, f transport) Client {
	t.Helper()
	return Client{Credentials: secret.Resolver{Lookup: func(string) (string, bool) { return "fake-test-key", true }}, HTTP: &http.Client{Transport: f}}
}
func response(status int, b string, h http.Header) *http.Response {
	return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(strings.NewReader(b))}
}
func TestAuthenticatedRoundTripAndQuota(t *testing.T) {
	calls := 0
	c := fake(t, func(r *http.Request) (*http.Response, error) {
		calls++
		u, p, ok := r.BasicAuth()
		if !ok || u != "api" || p != "fake-test-key" || r.URL.Host != "api.tinify.com" {
			t.Fatal("credential boundary")
		}
		if calls == 1 {
			b, _ := io.ReadAll(r.Body)
			if string(b) != "original" || r.Method != "POST" {
				t.Fatal("upload changed")
			}
			return response(201, "{}", http.Header{"Location": []string{"https://api.tinify.com/output/test"}, "Compression-Count": []string{"9"}}), nil
		}
		if r.Method != "GET" {
			t.Fatal(r.Method)
		}
		return response(200, "compressed", http.Header{"Compression-Count": []string{"10"}}), nil
	})
	out, e := c.Optimize(context.Background(), Request{[]byte("original"), 100})
	if e != nil || string(out.Data) != "compressed" || !out.Usage.Attempted || *out.Usage.CompressionCount != 10 || calls != 2 {
		t.Fatal(out, e)
	}
}
func TestNoAuthForwardingToUntrustedLocation(t *testing.T) {
	for _, location := range []string{"https://evil.invalid/output/x", "http://api.tinify.com/output/x", "https://api.tinify.com:444/output/x", "https://api.tinify.com/output/x?secret=bad", "https://api.tinify.com/output/../x"} {
		calls := 0
		c := fake(t, func(*http.Request) (*http.Response, error) {
			calls++
			return response(201, "", http.Header{"Location": []string{location}}), nil
		})
		_, e := c.Optimize(context.Background(), Request{[]byte("x"), 100})
		if domain.Code(e) != 7 || calls != 1 {
			t.Fatal(location, calls, e)
		}
	}
}
func TestFailuresNeverRetryOrExposeResponse(t *testing.T) {
	for _, code := range []int{301, 401, 429, 500} {
		calls := 0
		c := fake(t, func(*http.Request) (*http.Response, error) {
			calls++
			return response(code, "echo fake-test-key", http.Header{}), nil
		})
		_, e := c.Optimize(context.Background(), Request{[]byte("x"), 100})
		if domain.Code(e) != 7 || calls != 1 || strings.Contains(e.Error(), "fake-test-key") {
			t.Fatal(code, calls, e)
		}
	}
	c := fake(t, func(*http.Request) (*http.Response, error) { return nil, errors.New("fake-test-key transport leak") })
	_, e := c.Optimize(context.Background(), Request{[]byte("x"), 100})
	if strings.Contains(e.Error(), "fake-test-key") {
		t.Fatal(e)
	}
}
func TestOutputReadIsBounded(t *testing.T) {
	calls := 0
	c := fake(t, func(*http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return response(201, "", http.Header{"Location": []string{"https://api.tinify.com/output/test"}}), nil
		}
		return response(200, strings.Repeat("x", 101), http.Header{}), nil
	})
	_, e := c.Optimize(context.Background(), Request{[]byte("x"), 100})
	if domain.Code(e) != 7 {
		t.Fatal(e)
	}
}
