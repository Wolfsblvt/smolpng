// SPDX-License-Identifier: MIT
package desktop

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/Wolfsblvt/smolpng/internal/process"
)

type Actions interface {
	OpenFolder(context.Context, string) error
	OpenURL(context.Context, string) error
	Notify(context.Context, string, string) error
}
type NativeActions struct{ Runner process.Runner }

func (n NativeActions) run(ctx context.Context, exe string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	r := n.Runner
	if r == nil {
		r = process.OSRunner{}
	}
	result, e := r.Run(ctx, process.Invocation{Executable: exe, Args: args})
	if e != nil {
		return e
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("desktop command %s exited %d", exe, result.ExitCode)
	}
	return nil
}
func (n NativeActions) OpenFolder(ctx context.Context, p string) error {
	abs, e := filepath.Abs(p)
	if e != nil {
		return e
	}
	switch runtime.GOOS {
	case "windows":
		return n.run(ctx, "explorer.exe", abs)
	case "darwin":
		return n.run(ctx, "open", "--", abs)
	default:
		return n.run(ctx, "xdg-open", abs)
	}
}
func (n NativeActions) OpenURL(ctx context.Context, address string) error {
	u, e := url.Parse(address)
	if e != nil || u.Host == "" || u.User != nil || u.Scheme != "https" {
		return fmt.Errorf("only credential-free HTTPS pages may be opened")
	}
	switch runtime.GOOS {
	case "windows":
		return n.run(ctx, "rundll32.exe", "url.dll,FileProtocolHandler", u.String())
	case "darwin":
		return n.run(ctx, "open", "--", u.String())
	default:
		return n.run(ctx, "xdg-open", u.String())
	}
}
func (n NativeActions) Notify(ctx context.Context, title, body string) error {
	switch runtime.GOOS {
	case "windows":
		return n.run(ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-EncodedCommand", notificationScript(title, body))
	case "darwin":
		return n.run(ctx, "osascript", "-e", "display notification "+strconv.Quote(body)+" with title "+strconv.Quote(title))
	default:
		return n.run(ctx, "notify-send", "--app-name=smolpng", "--", title, body)
	}
}
func notificationScript(title, body string) string {
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	script := "$ErrorActionPreference='Stop'; Add-Type -AssemblyName System.Windows.Forms; Add-Type -AssemblyName System.Drawing; $n=New-Object System.Windows.Forms.NotifyIcon; try {$n.Icon=[System.Drawing.SystemIcons]::Information; $n.BalloonTipTitle=" + quote(title) + "; $n.BalloonTipText=" + quote(body) + "; $n.Visible=$true; $n.ShowBalloonTip(5000); Start-Sleep -Milliseconds 1800} finally {$n.Dispose()}"
	words := utf16.Encode([]rune(script))
	b := make([]byte, len(words)*2)
	for i, w := range words {
		binary.LittleEndian.PutUint16(b[i*2:], w)
	}
	return base64.StdEncoding.EncodeToString(b)
}
