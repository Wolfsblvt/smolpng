// SPDX-License-Identifier: MIT
// Package state stores bounded, recoverable human-workflow state, never secrets.
package state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
)

type Last struct {
	SchemaVersion int           `json:"schemaVersion"`
	AppVersion    string        `json:"appVersion"`
	Completed     time.Time     `json:"completed"`
	Report        domain.Report `json:"report"`
	Warnings      []string      `json:"warnings"`
}
type Setup struct {
	SchemaVersion int  `json:"schemaVersion"`
	Acknowledged  bool `json:"acknowledged"`
}

func Write(ctx context.Context, path string, value any) error {
	if st, e := os.Lstat(path); e == nil && !st.Mode().IsRegular() {
		return fmt.Errorf("state path is not a regular file")
	}
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	if len(b) > 4<<20 {
		return fmt.Errorf("state exceeds the 4MiB retention bound")
	}
	return fsutil.AtomicWrite(ctx, path, append(b, '\n'), 0600, true)
}
func Read(path string, dest any) error {
	if st, e := os.Lstat(path); e != nil {
		return e
	} else if !st.Mode().IsRegular() {
		return fmt.Errorf("state is not a regular file")
	}
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() {
		return fmt.Errorf("state is not a regular file")
	}
	b, e := io.ReadAll(io.LimitReader(f, (4<<20)+1))
	if e != nil {
		return e
	}
	if len(b) > 4<<20 {
		return fmt.Errorf("state exceeds 4MiB")
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if e = dec.Decode(dest); e != nil {
		return fmt.Errorf("invalid state: %w", e)
	}
	if dec.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing state content")
	}
	return nil
}
func ReadLast(path string) (Last, error) {
	var l Last
	e := Read(path, &l)
	if e == nil && (l.SchemaVersion != 1 || l.Report.SchemaVersion != 1) {
		e = fmt.Errorf("unsupported last-result state version")
	}
	return l, e
}
func Outputs(r domain.Report) []string {
	files := []string{}
	for _, v := range r.Results {
		if v.Output != "" && v.Output != "-" && v.Status != "planned" {
			files = append(files, v.Output)
		}
	}
	return files
}
