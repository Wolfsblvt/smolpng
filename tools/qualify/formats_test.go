// SPDX-License-Identifier: MIT
package main

import (
	"reflect"
	"testing"
)

func TestFormatSelection(t *testing.T) {
	for _, input := range []string{"", "png,", "webp", "jpg"} {
		if got, err := parseFormats(input); err == nil {
			t.Fatalf("unsupported selection %q accepted: %v", input, got)
		}
	}
	for _, tc := range []struct {
		input string
		want  []string
	}{{"png", []string{"png"}}, {"jpeg", []string{"jpeg"}}, {"PNG, jpeg,png", []string{"png", "jpeg"}}} {
		got, err := parseFormats(tc.input)
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatal(tc.input, got, err)
		}
	}
}

func TestPreflightRequiresOnlySelectedFormatTools(t *testing.T) {
	jpegOnly := []byte(`{"tools":{"jpegoptim":{"available":true}}}`)
	pngOnly := []byte(`{"tools":{"pngquant":{"available":true},"oxipng":{"available":true}}}`)
	for _, tc := range []struct {
		report   []byte
		formats  []string
		failures int
	}{{jpegOnly, []string{"jpeg"}, 0}, {pngOnly, []string{"png"}, 0},
		{jpegOnly, []string{"png", "jpeg"}, 2}, {pngOnly, []string{"png", "jpeg"}, 1},
		{[]byte(`{}`), []string{"jpeg"}, 1}, {[]byte("not JSON"), []string{"jpeg"}, 1}} {
		if got := dependencyErrors(tc.report, tc.formats); len(got) != tc.failures {
			t.Fatal(string(tc.report), tc.formats, got)
		}
	}
}
