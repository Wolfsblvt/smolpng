// SPDX-License-Identifier: MIT
package desktop

import "io"

// Console is explicitly requested by the windowless launcher only when the user
// selected an interactive drop workflow. Ordinary automation never allocates one.
type Console struct {
	In       io.Reader
	Out, Err io.Writer
	Close    func()
}
