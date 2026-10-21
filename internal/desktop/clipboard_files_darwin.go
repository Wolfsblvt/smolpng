// SPDX-License-Identifier: MIT
package desktop

import (
	"context"
	"errors"
	"runtime"

	"github.com/ebitengine/purego/objc"
	native "golang.design/x/clipboard"
)

var (
	fileURLObjects = objc.RegisterName("readObjectsForClasses:options:")
	fileURLPath    = objc.RegisterName("path")
	fileURLIsFile  = objc.RegisterName("isFileURL")
	fileURLUTF8    = objc.RegisterName("UTF8String")
	fileURLCount   = objc.RegisterName("count")
	fileURLAtIndex = objc.RegisterName("objectAtIndex:")
)

func nativeClipboardFiles(ctx context.Context) ([]string, error) {
	files, err := native.ReadFiles(ctx)
	if len(files) > 0 || err != nil && !errors.Is(err, native.ErrNoData) && !errors.Is(err, native.ErrUnsupported) {
		return files, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// native.Init has loaded AppKit. Its legacy filename-list reader does not
	// recover modern NSURL pasteboard items, so ask Cocoa to decode them.
	// Keep autoreleased objects and the pool on the same native thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
	defer pool.Send(objc.RegisterName("drain"))
	board := objc.ID(objc.GetClass("NSPasteboard")).Send(objc.RegisterName("generalPasteboard"))
	classes := objc.ID(objc.GetClass("NSArray")).Send(objc.RegisterName("arrayWithObject:"), objc.GetClass("NSURL"))
	options := objc.ID(objc.GetClass("NSDictionary")).Send(objc.RegisterName("dictionary"))
	items := board.Send(fileURLObjects, classes, options)
	for index := uintptr(0); index < uintptr(items.Send(fileURLCount)); index++ {
		item := items.Send(fileURLAtIndex, index)
		if item.Send(fileURLIsFile) == 0 {
			continue
		}
		path := objc.Send[string](item.Send(fileURLPath), fileURLUTF8)
		if path != "" {
			files = append(files, path)
		}
	}
	if len(files) == 0 {
		return nil, native.ErrNoData
	}
	return files, nil
}
