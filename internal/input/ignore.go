// SPDX-License-Identifier: MIT
package input

import (
	"github.com/bmatcuk/doublestar/v4"
	"os"
	"path/filepath"
	"strings"
)

type rule struct {
	pattern               string
	negate, dir, basename bool
}
type ignoreFile struct {
	root  string
	rules []rule
}

func readIgnore(dir string) (*ignoreFile, error) {
	b, e := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	f := &ignoreFile{root: dir}
	for _, s := range strings.Split(string(b), "\n") {
		s = strings.TrimSuffix(s, "\r")
		for strings.HasSuffix(s, " ") && !strings.HasSuffix(s, `\ `) {
			s = strings.TrimSuffix(s, " ")
		}
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		r := rule{}
		if s[0] == '!' {
			r.negate = true
			s = s[1:]
		}
		r.dir = strings.HasSuffix(s, "/")
		s = strings.TrimSuffix(s, "/")
		anchored := strings.HasPrefix(s, "/")
		s = strings.TrimPrefix(s, "/")
		r.basename = !anchored && !strings.Contains(s, "/")
		r.pattern = s
		if doublestar.ValidatePattern(s) {
			f.rules = append(f.rules, r)
		}
	}
	return f, nil
}
func ancestors(dir string, disabled bool) ([]ignoreFile, error) {
	if disabled {
		return nil, nil
	}
	ds := []string{}
	for {
		ds = append(ds, dir)
		if _, e := os.Stat(filepath.Join(dir, ".git")); e == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	out := []ignoreFile{}
	for i := len(ds) - 1; i >= 0; i-- {
		f, e := readIgnore(ds[i])
		if e != nil {
			return nil, e
		}
		if f != nil {
			out = append(out, *f)
		}
	}
	return out, nil
}
func ignored(p string, isDir bool, fs []ignoreFile) bool {
	result := false
	for _, f := range fs {
		rel, e := filepath.Rel(f.root, p)
		if e != nil || !filepath.IsLocal(rel) {
			continue
		}
		rel = filepath.ToSlash(rel)
		for _, r := range f.rules {
			if r.dir && !isDir {
				continue
			}
			name := rel
			if r.basename {
				name = filepath.Base(p)
			}
			if yes, _ := doublestar.Match(r.pattern, name); yes {
				result = !r.negate
			}
		}
	}
	return result
}
