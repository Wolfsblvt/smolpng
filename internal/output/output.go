// SPDX-License-Identifier: MIT
// Package output plans explicit destinations and protects original artifacts.
package output

import (
	"context"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/input"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Options struct {
	Path, Directory, Detached, Template, Collision string
	InPlace, Overwrite                             bool
}
type Job struct {
	Source input.Source
	Output string
}

var smolTail = regexp.MustCompile(`(?i)\.smol(?:-[0-9]+)?$`)
var reserved = regexp.MustCompile(`(?i)^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(?:\.|$)`)

func SafeName(s string) string {
	s = filepath.Base(strings.ReplaceAll(s, "\\", "/"))
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, s)
	s = strings.Trim(s, " .")
	if s == "" {
		s = "image"
	}
	if reserved.MatchString(s) {
		s = "_" + s
	}
	return s
}
func ValidateTemplate(s string) error {
	if s == "" || strings.ContainsAny(s, "/\\<>:\"|?*\x00\r\n") {
		return fmt.Errorf("template must be a portable filename")
	}
	for _, v := range []string{"{stem}", "{ext}", "{timestamp}"} {
		s = strings.ReplaceAll(s, v, "")
	}
	if strings.ContainsAny(s, "{}") {
		return fmt.Errorf("unknown template token")
	}
	return nil
}
func Name(source, template string, now time.Time) (string, error) {
	if template == "" {
		template = "{stem}.smol{ext}"
	}
	if e := ValidateTemplate(template); e != nil {
		return "", e
	}
	source = SafeName(source)
	ext := strings.ToLower(filepath.Ext(source))
	if ext != ".jpg" && ext != ".jpeg" {
		ext = ".png"
	}
	stem := strings.TrimSuffix(source, filepath.Ext(source))
	stem = smolTail.ReplaceAllString(stem, "")
	name := strings.NewReplacer("{stem}", stem, "{ext}", ext, "{timestamp}", now.UTC().Format("20060102-150405")).Replace(template)
	name = SafeName(name)
	if strings.ToLower(filepath.Ext(name)) != ext {
		return "", fmt.Errorf("output extension must match the input format")
	}
	if len(name) > 230 {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		base = base[:min(220, len(base))]
		for !utf8.ValidString(base) {
			base = base[:len(base)-1]
		}
		name = base + ext
	}
	return name, nil
}
func SameFile(a, b string) bool {
	aa, e := filepath.Abs(a)
	bb, f := filepath.Abs(b)
	if e != nil || f != nil {
		return false
	}
	if aa == bb || runtime.GOOS == "windows" && strings.EqualFold(aa, bb) {
		return true
	}
	ia, e := os.Stat(a)
	ib, f := os.Stat(b)
	return e == nil && f == nil && os.SameFile(ia, ib)
}
func increment(p string, n int) string {
	ext := filepath.Ext(p)
	return strings.TrimSuffix(p, ext) + fmt.Sprintf("-%d", n) + ext
}
func Plan(sources []input.Source, o Options, now time.Time) ([]Job, error) {
	if o.Path != "" && len(sources) != 1 {
		return nil, domain.E(2, "--output needs one input; use --output-dir for a batch", nil)
	}
	if o.Path != "" && o.Directory != "" || o.InPlace && (o.Path != "" || o.Directory != "") {
		return nil, domain.E(2, "conflicting output options", nil)
	}
	if o.Collision != "" && o.Collision != "increment" && o.Collision != "error" {
		return nil, domain.E(2, "collision must be increment or error", nil)
	}
	jobs := []Job{}
	planned := map[string]bool{}
	key := func(s string) string {
		if runtime.GOOS == "windows" {
			return strings.ToLower(s)
		}
		return s
	}
	for _, s := range sources {
		j := Job{Source: s}
		if s.Err != nil {
			jobs = append(jobs, j)
			continue
		}
		p := o.Path
		if o.InPlace {
			if s.Kind != "file" {
				return nil, domain.E(2, "in-place requires local files", nil)
			}
			st, e := os.Lstat(s.Path)
			if e != nil || !st.Mode().IsRegular() {
				return nil, domain.E(2, "in-place refuses symlinks and inaccessible files", e)
			}
			p = s.Path
		} else if p == "" {
			name := s.Name
			if s.Kind == "clipboard" {
				name = "clipboard-" + now.UTC().Format("20060102-150405") + ".png"
			}
			n, e := Name(name, o.Template, now)
			if e != nil {
				return nil, domain.E(2, e.Error(), e)
			}
			dir := o.Detached
			if s.Path != "" {
				dir = filepath.Dir(s.Path)
			}
			if o.Directory != "" {
				dir = o.Directory
				if s.Relative != "" {
					if !filepath.IsLocal(s.Relative) {
						return nil, domain.E(2, "relative output would escape directory", nil)
					}
					dir = filepath.Join(dir, filepath.Dir(s.Relative))
				}
			}
			if dir == "" {
				return nil, domain.E(2, "Downloads folder unavailable; choose --output or --output-dir", nil)
			}
			p = filepath.Join(dir, n)
		}
		if p != "-" {
			var e error
			p, e = filepath.Abs(p)
			if e != nil {
				return nil, e
			}
			if !o.InPlace {
				for _, src := range sources {
					if src.Path != "" && SameFile(p, src.Path) {
						return nil, domain.E(2, "output aliases an original; use a different destination or --in-place", nil)
					}
				}
				base := p
				for n := 2; ; n++ {
					if n > 10000 {
						return nil, domain.E(5, "too many output collisions", nil)
					}
					_, e = os.Lstat(p)
					exists := e == nil || planned[key(p)]
					if e != nil && !os.IsNotExist(e) {
						return nil, e
					}
					if !exists {
						break
					}
					aliases := false
					for _, src := range sources {
						aliases = aliases || src.Path != "" && SameFile(p, src.Path)
					}
					if o.Overwrite && !planned[key(p)] && !aliases {
						break
					}
					if o.Collision == "error" {
						return nil, domain.E(2, "output collision: "+p, os.ErrExist)
					}
					p = increment(base, n)
				}
			}
			planned[key(p)] = true
		}
		j.Output = p
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// A race-time collision refuses rather than rewriting a plan reserved for a batch.
func Write(ctx context.Context, j Job, data []byte, o Options) error {
	mode := os.FileMode(0600)
	if o.InPlace {
		st, e := os.Stat(j.Source.Path)
		if e != nil {
			return e
		}
		mode = st.Mode().Perm()
	}
	return fsutil.AtomicWrite(ctx, j.Output, data, mode, o.InPlace || o.Overwrite)
}
