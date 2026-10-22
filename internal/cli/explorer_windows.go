// SPDX-License-Identifier: MIT
package cli

import "github.com/spf13/cobra"

func init() {
	// Explorer is an intended parent for both interactive and background
	// launchers. Cobra's default Windows guard exits before either can dispatch.
	// Set this once at startup rather than mutating Cobra during concurrent runs.
	cobra.MousetrapHelpText = ""
}
