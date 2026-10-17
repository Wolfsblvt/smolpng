// SPDX-License-Identifier: MIT
package desktop

import (
	"errors"
	"testing"

	native "golang.design/x/clipboard"
)

func TestSensitiveClipboardReadPolicy(t *testing.T) {
	if e := sensitiveReadError(false, nil); e != nil {
		t.Fatal(e)
	}
	if e := sensitiveReadError(false, native.ErrUnsupported); e != nil {
		t.Fatal("unsupported marker detection changed existing clipboard behavior", e)
	}
	if e := sensitiveReadError(true, nil); !errors.Is(e, errSensitiveClipboard) {
		t.Fatal("marked clipboard content was not refused", e)
	}
	backend := errors.New("clipboard backend failed")
	if e := sensitiveReadError(false, backend); !errors.Is(e, backend) {
		t.Fatal("backend failure was hidden", e)
	}
}
