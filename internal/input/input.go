// SPDX-License-Identifier: MIT
// Package input separates effect-free source planning from bounded byte acquisition.
package input

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/bmatcuk/doublestar/v4"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Snapshot struct {
	Image []byte
	Files []string
	Text  string
}
type Clipboard interface {
	Read(context.Context) (Snapshot, error)
}
type Source struct {
	Kind     string `json:"kind"`
	Display  string `json:"input"`
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	Relative string `json:"relative,omitempty"`
	URL      string `json:"-"`
	Data     []byte `json:"-"`
	Err      error  `json:"-"`
}
type Options struct {
	CWD                                                             string
	Clipboard, Recursive, Hidden, NoIgnore, FollowSymlinks, Offline bool
	Include, Exclude, Globs                                         []string
	OutputDir, StdinName                                            string
}
type Resolver struct {
	Clipboard Clipboard
	Stdin     io.Reader
	HTTP      *http.Client
}

var generated = regexp.MustCompile(`(?i)\.smol(?:-[0-9]+)?\.(?:png|jpe?g)$`)

func key(p string) string {
	p = filepath.Clean(p)
	if runtime.GOOS == "windows" {
		return strings.ToLower(p)
	}
	return p
}
func (r Resolver) Resolve(ctx context.Context, args []string, o Options) ([]Source, error) {
	if o.CWD == "" {
		var e error
		o.CWD, e = os.Getwd()
		if e != nil {
			return nil, e
		}
	}
	for _, p := range append(append([]string{}, o.Include...), o.Exclude...) {
		if !doublestar.ValidatePattern(filepath.ToSlash(p)) {
			return nil, domain.E(2, "invalid include/exclude pattern", nil)
		}
	}
	out := []Source{}
	seen := map[string]bool{}
	stdinSeen := false
	add := func(s Source) {
		k := s.Kind + ":" + s.Display
		if s.URL != "" {
			k = s.URL
		}
		if s.Path != "" {
			p := s.Path
			if real, e := filepath.EvalSymlinks(p); e == nil {
				p = real
			}
			k = key(p)
		}
		if !seen[k] {
			seen[k] = true
			out = append(out, s)
		}
	}
	failure := func(p, message string) {
		add(Source{Kind: "file", Display: p, Name: filepath.Base(p), Path: p, Err: domain.E(3, message, nil)})
	}
	var visit func(string, string, bool, []ignoreFile, map[string]bool) error
	visit = func(p, base string, explicit bool, ignores []ignoreFile, visited map[string]bool) error {
		if e := ctx.Err(); e != nil {
			return e
		}
		p = filepath.Clean(p)
		st, e := os.Lstat(p)
		if e != nil {
			failure(p, "input is not accessible")
			return nil
		}
		link := st.Mode()&os.ModeSymlink != 0
		if link {
			st, e = os.Stat(p)
			if e != nil {
				failure(p, "broken input symlink")
				return nil
			}
			if st.IsDir() && !o.FollowSymlinks {
				if explicit {
					failure(p, "directory symlink needs --follow-symlinks")
				}
				return nil
			}
		}
		rel, e := filepath.Rel(base, p)
		if e != nil {
			rel = filepath.Base(p)
		}
		if !explicit {
			if filepath.Base(p) == ".git" || !o.Hidden && strings.HasPrefix(filepath.Base(p), ".") {
				return nil
			}
			if !o.NoIgnore && ignored(p, st.IsDir(), ignores) || matches(o.Exclude, rel) {
				return nil
			}
			if o.OutputDir != "" {
				dest := o.OutputDir
				if !filepath.IsAbs(dest) {
					dest = filepath.Join(o.CWD, dest)
				}
				if within(dest, p) {
					return nil
				}
			}
		}
		if st.IsDir() {
			real, e := filepath.EvalSymlinks(p)
			if e != nil {
				return e
			}
			if visited[key(real)] {
				return nil
			}
			visited[key(real)] = true
			if !o.NoIgnore {
				f, e := readIgnore(p)
				if e != nil {
					return e
				}
				if f != nil {
					ignores = append(append([]ignoreFile{}, ignores...), *f)
				}
			}
			entries, e := os.ReadDir(p)
			if e != nil {
				failure(p, "cannot list input directory")
				return nil
			}
			for _, entry := range entries {
				child := filepath.Join(p, entry.Name())
				s, e := os.Stat(child)
				if e == nil && s.IsDir() && !o.Recursive {
					continue
				}
				if e = visit(child, base, false, ignores, visited); e != nil {
					return e
				}
			}
			return nil
		}
		if !st.Mode().IsRegular() {
			if explicit {
				failure(p, "input is not a regular file")
			}
			return nil
		}
		if !explicit && (!strings.EqualFold(filepath.Ext(p), ".png") || generated.MatchString(p) || len(o.Include) > 0 && !matches(o.Include, rel)) {
			return nil
		}
		add(Source{Kind: "file", Display: p, Name: filepath.Base(p), Path: p, Relative: rel})
		return nil
	}
	var resolve func(string, bool) error
	resolve = func(arg string, forceGlob bool) error {
		if arg == "-" && !forceGlob {
			if stdinSeen {
				return domain.E(2, "stdin may only be consumed once", nil)
			}
			stdinSeen = true
			name := o.StdinName
			if name == "" {
				name = "stdin.png"
			}
			add(Source{Kind: "stdin", Display: "<stdin>", Name: filepath.Base(name)})
			return nil
		}
		if strings.HasPrefix(strings.ToLower(arg), "http://") || strings.HasPrefix(strings.ToLower(arg), "https://") {
			if o.Offline {
				return domain.E(2, "offline mode rejects URL input", nil)
			}
			u, e := url.Parse(arg)
			if e != nil || u.Host == "" || u.User != nil {
				return domain.E(3, "invalid HTTP(S) URL or embedded credentials", nil)
			}
			public := *u
			public.RawQuery = ""
			public.Fragment = ""
			name := filepath.Base(u.Path)
			if name == "." || name == "/" || name == "" {
				name = "download.png"
			}
			add(Source{Kind: "url", Display: public.String(), Name: name, URL: arg})
			return nil
		}
		if strings.HasPrefix(arg, "file://") {
			u, e := url.Parse(arg)
			if e != nil || u.Host != "" && u.Host != "localhost" {
				return domain.E(3, "unsupported file URI", nil)
			}
			arg = u.Path
			if runtime.GOOS == "windows" && len(arg) > 2 && arg[0] == '/' && arg[2] == ':' {
				arg = arg[1:]
			}
			arg = filepath.FromSlash(arg)
		}
		p := arg
		if !filepath.IsAbs(p) {
			p = filepath.Join(o.CWD, p)
		}
		p = filepath.Clean(p)
		st, se := os.Stat(p)
		if forceGlob || se != nil && strings.ContainsAny(arg, "*?[") {
			ms, e := doublestar.FilepathGlob(p, doublestar.WithFailOnIOErrors(), doublestar.WithNoFollow())
			if e != nil {
				return domain.E(2, "invalid glob", e)
			}
			if len(ms) == 0 {
				failure(p, "glob matched no files")
				return nil
			}
			base, _ := doublestar.SplitPattern(filepath.ToSlash(p))
			for _, m := range ms {
				ign, e := ancestors(filepath.Dir(m), o.NoIgnore)
				if e != nil {
					return e
				}
				if e = visit(m, filepath.FromSlash(base), false, ign, map[string]bool{}); e != nil {
					return e
				}
			}
			return nil
		}
		base := filepath.Dir(p)
		if se == nil && st.IsDir() {
			base = p
		}
		ign, e := ancestors(filepath.Dir(base), o.NoIgnore)
		if e != nil {
			return e
		}
		before := len(out)
		if e = visit(p, base, true, ign, map[string]bool{}); e != nil {
			return e
		}
		if len(out) == before && se == nil && st.IsDir() {
			failure(p, "directory contains no selected PNG images")
		}
		return nil
	}
	args = append([]string{}, args...)
	if o.Clipboard {
		if len(args) > 0 || len(o.Globs) > 0 {
			return nil, domain.E(2, "choose explicit input or --clipboard", nil)
		}
		if r.Clipboard == nil {
			return nil, domain.E(3, "clipboard unavailable on this host", nil)
		}
		s, e := r.Clipboard.Read(ctx)
		if e != nil {
			return nil, domain.E(3, "cannot read clipboard", e)
		}
		switch {
		case len(s.Image) > 0:
			add(Source{Kind: "clipboard", Display: "<clipboard image>", Name: "clipboard.png", Data: s.Image})
		case len(s.Files) > 0:
			args = s.Files
		case strings.TrimSpace(s.Text) != "":
			if strings.Contains(s.Text, "[presets.") {
				return nil, domain.E(3, "clipboard contains a preset; use preset import --clipboard", nil)
			}
			for _, line := range strings.Split(s.Text, "\n") {
				line = strings.Trim(strings.TrimSpace(line), "\"")
				if line != "" {
					args = append(args, line)
				}
			}
		default:
			return nil, domain.E(3, "clipboard has no usable image, file or text source", nil)
		}
	}
	for _, arg := range args {
		if e := resolve(arg, false); e != nil {
			return nil, e
		}
	}
	for _, g := range o.Globs {
		if e := resolve(g, true); e != nil {
			return nil, e
		}
	}
	if len(out) == 0 {
		return nil, domain.E(2, "provide input, --glob, stdin (-), or --clipboard", nil)
	}
	return out, nil
}
func (r Resolver) Load(ctx context.Context, s Source, limit int64, offline bool) ([]byte, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	var reader io.Reader
	code := domain.InvalidInput
	switch s.Kind {
	case "clipboard":
		reader = bytes.NewReader(s.Data)
	case "stdin":
		if r.Stdin == nil {
			return nil, domain.E(3, "stdin unavailable", nil)
		}
		reader = r.Stdin
	case "file":
		f, e := os.Open(s.Path)
		if e != nil {
			return nil, domain.E(3, "cannot open input", e)
		}
		defer f.Close()
		st, e := f.Stat()
		if e != nil || !st.Mode().IsRegular() {
			return nil, domain.E(3, "input is no longer a regular file", e)
		}
		if st.Size() > limit {
			return nil, domain.E(3, "input exceeds byte limit", nil)
		}
		reader = f
	case "url":
		if offline {
			return nil, domain.E(2, "offline mode rejects URL input", nil)
		}
		client := r.HTTP
		if client == nil {
			client = &http.Client{Timeout: 45 * time.Second}
		}
		copyClient := *client
		check := client.CheckRedirect
		copyClient.CheckRedirect = func(q *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if q.URL.User != nil || q.URL.Scheme != "https" && q.URL.Scheme != "http" {
				return fmt.Errorf("unsupported redirect")
			}
			if len(via) > 0 && via[0].URL.Scheme == "https" && q.URL.Scheme != "https" {
				return fmt.Errorf("HTTPS downgrade rejected")
			}
			if check != nil {
				return check(q, via)
			}
			return nil
		}
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
		if e != nil {
			return nil, domain.E(3, "invalid image URL", e)
		}
		req.Header.Set("User-Agent", "smolpng")
		res, e := copyClient.Do(req)
		if e != nil {
			return nil, domain.E(7, "image download failed (query omitted)", e)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return nil, domain.E(7, fmt.Sprintf("image server returned HTTP %d", res.StatusCode), nil)
		}
		if res.ContentLength > limit {
			return nil, domain.E(7, "image download exceeds byte limit", nil)
		}
		reader = res.Body
		code = domain.NetworkFailed
	default:
		return nil, domain.E(3, "unknown input kind", nil)
	}
	if c, ok := reader.(io.Closer); ok {
		stop := context.AfterFunc(ctx, func() { _ = c.Close() })
		defer stop()
	}
	b, e := pngimage.ReadBounded(reader, limit)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if e != nil {
		return nil, domain.E(code, "cannot read input within the byte limit", e)
	}
	return b, nil
}
func matches(ps []string, name string) bool {
	name = filepath.ToSlash(name)
	for _, p := range ps {
		p = filepath.ToSlash(p)
		n := name
		if !strings.Contains(p, "/") {
			n = filepath.Base(name)
		}
		if ok, _ := doublestar.Match(p, n); ok {
			return true
		}
	}
	return false
}
func within(root, p string) bool {
	r, e := filepath.Rel(root, p)
	return e == nil && (r == "." || r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator)))
}
