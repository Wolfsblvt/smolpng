// SPDX-License-Identifier: MIT
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/process"
)

// Configuration is data beside a fixed script, never interpolated AppleScript.
// Each executable/argument/path is quoted separately by the native text facility.
const macDropletScript = `use framework "Foundation"
use scripting additions

on run
	my launchInputs({})
end run

on open droppedItems
	my launchInputs(droppedItems)
end open

on launchInputs(droppedItems)
	set bundleDirectory to POSIX path of (path to me)
	set configData to current application's NSData's dataWithContentsOfFile:(bundleDirectory & "Contents/Resources/smolpng.json")
	if configData is missing value then error "smolpng launcher configuration is missing; reinstall the launcher."
	set settings to current application's NSJSONSerialization's JSONObjectWithData:configData options:0 |error|:(missing value)
	if settings is missing value then error "smolpng launcher configuration is invalid; reinstall the launcher."
	set programPath to (settings's objectForKey:"executable") as text
	set workingDirectory to (settings's objectForKey:"workingDirectory") as text
	set commandText to "cd " & quoted form of workingDirectory & " && " & quoted form of programPath
	repeat with argumentValue in ((settings's objectForKey:"arguments") as list)
		set commandText to commandText & " " & quoted form of (argumentValue as text)
	end repeat
	repeat with droppedItem in droppedItems
		set commandText to commandText & " " & quoted form of (POSIX path of droppedItem)
	end repeat
	set needsTerminal to (settings's objectForKey:"terminal") as boolean
	if not needsTerminal then
		-- Resolve the selected data through Go, rather than parsing TOML here.
		set configCommand to "cd " & quoted form of workingDirectory & " && " & quoted form of programPath
		repeat with argumentValue in ((settings's objectForKey:"configurationArguments") as list)
			set configCommand to configCommand & " " & quoted form of (argumentValue as text)
		end repeat
		set configText to do shell script configCommand & " config show --effective --json"
		set configData to (current application's NSString's stringWithString:configText)'s dataUsingEncoding:(current application's NSUTF8StringEncoding)
		set effectiveSettings to current application's NSJSONSerialization's JSONObjectWithData:configData options:0 |error|:(missing value)
		if effectiveSettings is missing value then error "smolpng could not resolve its drop configuration."
		set needsTerminal to ((effectiveSettings's objectForKey:"dropMode") as text) is "interactive"
	end if
	if needsTerminal then
		-- Opening an owned .command through Launch Services needs no permission
		-- to automate Terminal. The script removes itself before running the CLI.
		set commandPath to (current application's NSTemporaryDirectory() as text) & "smolpng-" & (current application's NSUUID's UUID()'s UUIDString() as text) & ".command"
		set scriptText to "#!/bin/sh" & linefeed & "/bin/rm -f -- " & quote & "$0" & quote & linefeed & commandText & linefeed
		set scriptData to (current application's NSString's stringWithString:scriptText)'s dataUsingEncoding:(current application's NSUTF8StringEncoding)
		if not (scriptData's writeToFile:commandPath atomically:true) then error "smolpng could not prepare its terminal command."
		try
			do shell script "/bin/chmod 700 " & quoted form of commandPath & " && /usr/bin/open -b com.apple.Terminal " & quoted form of commandPath
		on error messageText number errorNumber
			do shell script "/bin/rm -f -- " & quoted form of commandPath
			error messageText number errorNumber
		end try
	else
		do shell script commandText
	end if
end launchInputs
`

// RenderBundle uses the operating system's applet compiler. No Xcode toolchain,
// extra runtime, optimizer, signing identity or user configuration is bundled.
func (n Native) RenderBundle(ctx context.Context, x Entry, destination string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("native app compilation requires macOS")
	}
	if !filepath.IsAbs(x.Executable) || !filepath.IsAbs(x.WorkingDirectory) || !filepath.IsAbs(destination) {
		return fmt.Errorf("native app paths must be absolute")
	}
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		return fmt.Errorf("native app staging destination already exists or is unavailable")
	}
	source := filepath.Join(filepath.Dir(destination), "source.applescript")
	if err := fsutil.AtomicWrite(ctx, source, []byte(macDropletScript), 0600, false); err != nil {
		return err
	}
	defer os.Remove(source)
	runner := n.Runner
	if runner == nil {
		runner = process.OSRunner{}
	}
	result, err := runner.Run(ctx, process.Invocation{Executable: "/usr/bin/osacompile", Args: []string{"-l", "AppleScript", "-o", destination, source}})
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("macOS app compilation failed (exit %d): %s", result.ExitCode, result.Stderr)
	}
	settings := struct {
		Entry
		ConfigurationArguments []string `json:"configurationArguments"`
	}{Entry: x, ConfigurationArguments: macConfigurationArguments(x.Arguments)}
	config, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	if err = fsutil.AtomicWrite(ctx, filepath.Join(destination, "Contents", "Resources", "smolpng.json"), append(config, '\n'), 0644, false); err != nil {
		return err
	}
	// Applets are accessory processes; a selected interactive workflow opens its
	// own Terminal window, while a background drop does not acquire a Dock item.
	plist := filepath.Join(destination, "Contents", "Info.plist")
	result, err = runner.Run(ctx, process.Invocation{Executable: "/usr/libexec/PlistBuddy", Args: []string{"-c", "Add :LSUIElement bool true", plist}})
	if err == nil && result.ExitCode != 0 {
		result, err = runner.Run(ctx, process.Invocation{Executable: "/usr/libexec/PlistBuddy", Args: []string{"-c", "Set :LSUIElement true", plist}})
	}
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("macOS app background metadata could not be set")
	}
	return nil
}

// Only configuration location flags belong in the effective-settings query.
// The drop verb, selected recipe and file arguments belong to the real launch.
func macConfigurationArguments(args []string) []string {
	out := []string{}
	for i := 0; i < len(args) && args[i] != "--"; i++ {
		switch args[i] {
		case "--home", "--config":
			if i+1 < len(args) {
				out = append(out, args[i], args[i+1])
				i++
			}
		case "--no-project":
			out = append(out, args[i])
		}
	}
	return out
}
