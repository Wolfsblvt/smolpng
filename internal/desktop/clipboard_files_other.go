// SPDX-License-Identifier: MIT
//go:build !darwin

package desktop

import (
	"context"

	native "golang.design/x/clipboard"
)

func nativeClipboardFiles(ctx context.Context) ([]string, error) {
	return native.ReadFiles(ctx)
}
