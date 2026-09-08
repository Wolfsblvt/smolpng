package input

import (
	"context"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTraversalAndPartialInputFailure(t *testing.T) {
	d := t.TempDir()
	for _, name := range []string{"a.png", "a.smol.png", ".hidden.png", "skip.png", "sub/yes.png", "sub/no.png", "out/old.png"} {
		p := filepath.Join(d, name)
		_ = os.MkdirAll(filepath.Dir(p), 0755)
		_ = os.WriteFile(p, []byte("x"), 0600)
	}
	_ = os.WriteFile(filepath.Join(d, ".gitignore"), []byte("skip.png\nsub/no.png\n"), 0600)
	s, e := (Resolver{}).Resolve(context.Background(), []string{d, filepath.Join(d, "missing.png")}, Options{Recursive: true, OutputDir: filepath.Join(d, "out")})
	if e != nil || len(s) != 3 || s[2].Err == nil {
		t.Fatalf("%+v %v", s, e)
	}
}
func TestGlobAndURLPlanningHasNoNetwork(t *testing.T) {
	d := t.TempDir()
	_ = os.WriteFile(filepath.Join(d, "a[1].png"), nil, 0600)
	s, e := (Resolver{}).Resolve(context.Background(), []string{filepath.Join(d, "a[1].png")}, Options{})
	if e != nil || len(s) != 1 {
		t.Fatal(s, e)
	}
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1); _, _ = w.Write([]byte("12345")) }))
	defer srv.Close()
	r := Resolver{}
	s, e = r.Resolve(context.Background(), []string{srv.URL + "/a.png?token=secret"}, Options{})
	if e != nil || hits.Load() != 0 || strings.Contains(s[0].Display, "secret") {
		t.Fatal(s, e)
	}
	if _, e = r.Load(context.Background(), s[0], 4, false); domain.Code(e) != 7 {
		t.Fatal(e)
	}
	if _, e = r.Resolve(context.Background(), []string{srv.URL}, Options{Offline: true}); domain.Code(e) != 2 {
		t.Fatal(e)
	}
}
func TestGitignoreNegation(t *testing.T) {
	d := t.TempDir()
	for _, n := range []string{"yes.png", "no.png"} {
		_ = os.WriteFile(filepath.Join(d, n), nil, 0600)
	}
	_ = os.WriteFile(filepath.Join(d, ".gitignore"), []byte("*.png\n!yes.png\n"), 0600)
	s, e := (Resolver{}).Resolve(context.Background(), []string{d}, Options{})
	if e != nil || len(s) != 1 || s[0].Name != "yes.png" {
		t.Fatal(s, e)
	}
}
