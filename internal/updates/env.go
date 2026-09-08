// SPDX-License-Identifier: MIT
package updates

import (
	"os"
	"strings"
)

func cleanEnv() []string {
	e := []string{}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(v), "TINIFY_API_KEY=") {
			e = append(e, v)
		}
	}
	return e
}
