// SPDX-License-Identifier: MIT
package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func parseFormats(value string) ([]string, error) {
	formats := []string{}
	for _, name := range strings.Split(value, ",") {
		name = strings.ToLower(strings.TrimSpace(name))
		if name != "png" && name != "jpeg" {
			return nil, fmt.Errorf("unsupported format %q; select png, jpeg, or png,jpeg", name)
		}
		if !slices.Contains(formats, name) {
			formats = append(formats, name)
		}
	}
	return formats, nil
}

func dependencyErrors(report []byte, formats []string) []string {
	var health struct {
		Tools map[string]struct {
			Available bool `json:"available"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(report, &health); err != nil {
		return []string{"dependency report is not JSON"}
	}
	failures := []string{}
	for _, format := range formats {
		tools := []string{"jpegoptim"}
		if format == "png" {
			tools = []string{"pngquant", "oxipng"}
		}
		for _, name := range tools {
			if !health.Tools[name].Available {
				failures = append(failures, fmt.Sprintf("%s requires real compatible %s; install separately or supply its explicit path", format, name))
			}
		}
	}
	return failures
}
