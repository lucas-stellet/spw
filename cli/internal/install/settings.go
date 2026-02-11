package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// settingsJSON is the Claude Code settings structure.
type settingsJSON struct {
	StatusLine statusLineEntry          `json:"statusLine"`
	Hooks      map[string][]hookMatcher `json:"hooks"`
}

type statusLineEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type hookMatcher struct {
	Matcher string      `json:"matcher"`
	Hooks   []hookEntry `json:"hooks"`
}

type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// NewSettings returns the default settings.json structure with Go binary hooks.
func NewSettings() settingsJSON {
	return settingsJSON{
		StatusLine: statusLineEntry{
			Type:    "command",
			Command: "spw hook statusline",
		},
		Hooks: map[string][]hookMatcher{
			"SessionStart": {
				{
					Matcher: "startup|resume|clear|compact",
					Hooks: []hookEntry{
						{Type: "command", Command: "spw hook session-start"},
					},
				},
			},
			"UserPromptSubmit": {
				{
					Matcher: ".*",
					Hooks: []hookEntry{
						{Type: "command", Command: "spw hook guard-prompt"},
					},
				},
			},
			"PreToolUse": {
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []hookEntry{
						{Type: "command", Command: "spw hook guard-paths"},
					},
				},
			},
			"Stop": {
				{
					Matcher: ".*",
					Hooks: []hookEntry{
						{Type: "command", Command: "spw hook guard-stop"},
					},
				},
			},
		},
	}
}

// WriteSettings creates .claude/settings.json if it doesn't exist.
func WriteSettings(root string) error {
	settingsPath := filepath.Join(root, ".claude", "settings.json")

	if _, err := os.Stat(settingsPath); err == nil {
		fmt.Println("[spw] .claude/settings.json already exists.")
		fmt.Println("[spw] Verify hooks point to 'spw hook <event>' commands.")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(NewSettings(), "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return err
	}

	fmt.Println("[spw] Created .claude/settings.json with hook registrations.")
	return nil
}

// DetectOldInstall checks if settings.json has old JS-based hook references.
func DetectOldInstall(root string) bool {
	settingsPath := filepath.Join(root, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return false
	}

	// Look for old-style node hook references
	return containsBytes(data, []byte("node ./.claude/hooks/spw-")) ||
		containsBytes(data, []byte("spw-statusline.js")) ||
		containsBytes(data, []byte("spw-guard-"))
}

func containsBytes(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
