// Package tools implements workflow utility subcommands
// (config-get, spec-resolve, wave-resolve, runs, handoff, skills, approval).
package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

// Output writes a result as JSON or raw value.
func Output(result any, rawValue string, raw bool) {
	if raw {
		fmt.Print(rawValue)
		return
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Print("{}")
		return
	}
	os.Stdout.Write(data)
}

// Fail prints an error and exits.
func Fail(message string, raw bool) {
	if raw {
		os.Exit(1)
	}
	result := map[string]any{"ok": false, "error": message}
	data, _ := json.MarshalIndent(result, "", "  ")
	os.Stderr.Write(data)
	os.Stderr.WriteString("\n")
	os.Exit(1)
}
