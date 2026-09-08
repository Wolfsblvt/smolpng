// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/shellmenu"
)

type shellRegistration struct {
	Schema           int    `json:"schema"`
	FullName         string `json:"fullName"`
	PackageSHA256    string `json:"packageSHA256"`
	ExternalLocation string `json:"externalLocation"`
}

func (a *app) shellPackageAction(ctx context.Context, verb, pkg string) (any, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("Windows package registration needs a Windows host")
	}
	_, l, e := a.shellManager()
	if e != nil {
		return nil, e
	}
	if verb != "status" {
		if e = os.MkdirAll(l.State, 0700); e != nil {
			return nil, e
		}
		lock, err := fsutil.TryLock(filepath.Join(l.State, "package-operation.lock"))
		if err != nil {
			return nil, fmt.Errorf("another package operation is active: %w", err)
		}
		defer lock.Close()
	}
	system := os.Getenv("SystemRoot")
	if !filepath.IsAbs(system) {
		return nil, fmt.Errorf("SystemRoot unavailable")
	}
	exe := filepath.Join(system, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	script := "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue'; "
	list := "@(Get-AppxPackage -Name " + quote(shellmenu.PackageName) + " -PackageTypeFilter Main)"
	receiptPath := filepath.Join(l.State, "package-registration.json")
	var record shellRegistration
	var info shellmenu.Package
	switch verb {
	case "status":
		script += "$p=" + list + "; ConvertTo-Json -Compress -InputObject @($p | Select-Object Name,PackageFullName,Publisher,Version,Architecture,Status)"
	case "activate":
		spec, _, e := a.readShell()
		if e != nil {
			return nil, e
		}
		info, e = shellmenu.CheckPackage(pkg, filepath.Dir(spec.Executable), Commit, runtime.GOARCH)
		if e != nil {
			return nil, e
		}
		if _, e = os.Lstat(receiptPath); e == nil {
			if _, e = shellmenu.FileHash(receiptPath); e != nil {
				return nil, e
			}
		} else if !os.IsNotExist(e) {
			return nil, e
		}
		script += "Add-AppxPackage -Path " + quote(pkg) + " -ExternalLocation " + quote(filepath.Dir(spec.Executable)) + "; $p=" + list + " | Where-Object {$_.Publisher -eq " + quote(info.Publisher) + " -and $_.Version.ToString() -eq " + quote(info.Version) + "}; if(@($p).Count -ne 1){throw 'Cannot read back registered package'}; $p.PackageFullName | ConvertTo-Json -Compress"
		record = shellRegistration{Schema: 1, PackageSHA256: info.SHA256, ExternalLocation: filepath.Dir(spec.Executable)}
	case "deactivate":
		if _, e = shellmenu.FileHash(receiptPath); e != nil {
			return nil, fmt.Errorf("no safe app-created package registration receipt: %w", e)
		}
		b, e := os.ReadFile(receiptPath)
		if e != nil || len(b) > 8192 {
			return nil, fmt.Errorf("invalid package registration receipt")
		}
		if e = json.Unmarshal(b, &record); e != nil || record.Schema != 1 || !strings.HasPrefix(record.FullName, shellmenu.PackageName+"_") {
			return nil, fmt.Errorf("invalid owned package identity")
		}
		script += "$p=" + list + " | Where-Object {$_.PackageFullName -eq " + quote(record.FullName) + "}; if($p){Remove-AppxPackage -Package " + quote(record.FullName) + "}; $left=" + list + " | Where-Object {$_.PackageFullName -eq " + quote(record.FullName) + "}; if($left){throw 'Package remains registered'}; 'removed' | ConvertTo-Json -Compress"
	default:
		return nil, fmt.Errorf("unknown native package operation")
	}
	code := utf16.Encode([]rune(script))
	data := make([]byte, len(code)*2)
	for i, c := range code {
		binary.LittleEndian.PutUint16(data[i*2:], c)
	}
	result, e := a.env.Runner.Run(ctx, process.Invocation{Executable: exe, Args: []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(data)}})
	if e != nil || result.ExitCode != 0 {
		return nil, fmt.Errorf("Windows package %s failed (exit %d). A trusted signature, correct architecture and matching payload are required; no certificate trust was changed", verb, result.ExitCode)
	}
	if verb == "status" {
		var state any
		if e = json.Unmarshal([]byte(strings.TrimSpace(string(result.Stdout))), &state); e != nil {
			return nil, e
		}
		return state, nil
	}
	if verb == "activate" {
		if e = json.Unmarshal([]byte(strings.TrimSpace(string(result.Stdout))), &record.FullName); e != nil || !strings.HasPrefix(record.FullName, shellmenu.PackageName+"_") {
			return nil, fmt.Errorf("registration readback was invalid; inspect Get-AppxPackage")
		}
		b, _ := json.MarshalIndent(record, "", "  ")
		if e = fsutil.AtomicWrite(ctx, receiptPath, append(b, '\n'), 0600, true); e != nil {
			return nil, fmt.Errorf("package registered but receipt could not be saved: %w", e)
		}
	} else {
		if e = os.Remove(receiptPath); e != nil {
			return nil, e
		}
	}
	return map[string]any{"schemaVersion": 1, "action": verb, "packageFullName": record.FullName, "completed": true, "explorerRestarted": false, "trustStoreChanged": false}, nil
}
