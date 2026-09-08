// SPDX-License-Identifier: MIT
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/secret"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func (a *app) credentials() secret.Resolver {
	return secret.Resolver{Store: a.env.SecretStore, Lookup: a.env.CredentialLookup}
}
func (a *app) providersCommand() *cobra.Command {
	root := &cobra.Command{Use: "providers", Short: "Optional remote service credentials; never enables upload permission"}
	valid := func(_ *cobra.Command, args []string) error {
		if len(args) > 1 || len(args) == 1 && args[0] != "tinypng" {
			return domain.E(2, "TinyPNG is the only selected remote provider", nil)
		}
		return nil
	}
	root.AddCommand(&cobra.Command{Use: "status [tinypng]", Args: valid, RunE: func(c *cobra.Command, _ []string) error { return a.emit(a.credentials().Status(c.Context())) }})
	var noOpen, stdin bool
	login := &cobra.Command{Use: "login tinypng", Args: func(c *cobra.Command, s []string) error {
		if len(s) != 1 {
			return domain.E(2, "name tinypng explicitly", nil)
		}
		return valid(c, s)
	}, Short: "Store an API key in the native keyring, without a network probe", RunE: func(c *cobra.Command, _ []string) error {
		if !stdin && (a.json || a.jsonl) {
			return domain.E(2, "noninteractive login needs --stdin; keys are never command-line arguments", nil)
		}
		if !noOpen && !stdin {
			if err := a.env.Desktop.OpenURL(c.Context(), "https://tinypng.com/developers"); err != nil {
				return domain.E(5, "cannot open key page; retry with --no-open", err)
			}
		}
		var key string
		var err error
		if stdin {
			b, e := pngimage.ReadBounded(a.env.In, 514)
			if e != nil {
				return domain.E(3, "cannot read key from stdin", e)
			}
			key = strings.TrimSuffix(strings.TrimSuffix(string(b), "\n"), "\r")
		} else if a.env.ReadSecret != nil {
			key, err = a.env.ReadSecret(c.Context())
		} else {
			f, ok := a.env.In.(*os.File)
			if !ok || !term.IsTerminal(int(f.Fd())) {
				return domain.E(2, "login needs a terminal or --stdin", nil)
			}
			fmt.Fprint(a.env.Err, "TinyPNG API key (hidden): ")
			var b []byte
			b, err = term.ReadPassword(int(f.Fd()))
			fmt.Fprintln(a.env.Err)
			key = string(b)
			clear(b)
		}
		if err != nil {
			return domain.E(3, "key entry cancelled or failed", err)
		}
		if err = secret.Validate(key); err != nil {
			return domain.E(2, err.Error(), err)
		}
		store := a.env.SecretStore
		if store == nil {
			store = secret.Native{}
		}
		if err = store.Set(key); err != nil {
			return domain.E(5, "native keyring could not store the key; no plaintext fallback was written", err)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "provider": "tinypng", "stored": "keyring", "verification": "not-probed", "remotePolicyChanged": false, "environmentMayOverride": true})
	}}
	login.Flags().BoolVar(&noOpen, "no-open", false, "Do not open the official API key page")
	login.Flags().BoolVar(&stdin, "stdin", false, "Read a key from standard input; do not echo or store in shell history")
	root.AddCommand(login, &cobra.Command{Use: "logout tinypng", Args: func(c *cobra.Command, s []string) error {
		if len(s) != 1 {
			return domain.E(2, "name tinypng explicitly", nil)
		}
		return valid(c, s)
	}, RunE: func(c *cobra.Command, _ []string) error {
		store := a.env.SecretStore
		if store == nil {
			store = secret.Native{}
		}
		if err := store.Delete(); err != nil {
			return domain.E(5, "could not remove keyring credential", err)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "provider": "tinypng", "keyring": "removed", "environment": "unchanged; unset TINIFY_API_KEY in the owning process if present"})
	}})
	return root
}
