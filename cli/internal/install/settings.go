package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// WriteSettings creates .claude/settings.json if it doesn't exist, or merges
// SPW hooks into the existing file.
func WriteSettings(root string) error {
	settingsPath := filepath.Join(root, ".claude", "settings.json")

	if _, err := os.Stat(settingsPath); err == nil {
		return MergeSettings(root)
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

// MergeSettings reads an existing .claude/settings.json, merges SPW hooks into
// it (preserving all non-SPW hooks and other settings), and writes it back.
func MergeSettings(root string) error {
	settingsPath := filepath.Join(root, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("read settings.json: %w", err)
	}

	var existing map[string]any
	if err := json.Unmarshal(data, &existing); err != nil {
		return fmt.Errorf("parse settings.json: %w", err)
	}

	spw := NewSettings()

	// Merge statusLine: overwrite only if absent or already an SPW statusline.
	mergeStatusLine(existing, spw)

	// Merge hooks: for each SPW event, remove old SPW entries, append new ones,
	// preserve all non-SPW entries.
	mergeHooks(existing, spw)

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings.json: %w", err)
	}

	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	fmt.Println("[spw] Hooks merged into .claude/settings.json.")
	return nil
}

// isSPWCommand returns true if the command string is an SPW hook command.
func isSPWCommand(cmd string) bool {
	return strings.HasPrefix(cmd, "spw hook ")
}

func mergeStatusLine(existing map[string]any, spw settingsJSON) {
	shouldSet := false
	if _, ok := existing["statusLine"]; !ok {
		shouldSet = true
	} else if sl, ok := existing["statusLine"].(map[string]any); ok {
		if cmd, ok := sl["command"].(string); ok && isSPWCommand(cmd) {
			shouldSet = true
		}
	}
	if shouldSet {
		existing["statusLine"] = map[string]any{
			"type":    spw.StatusLine.Type,
			"command": spw.StatusLine.Command,
		}
	}
}

func mergeHooks(existing map[string]any, spw settingsJSON) {
	hooksAny, ok := existing["hooks"]
	if !ok {
		hooksAny = map[string]any{}
	}
	hooks, ok := hooksAny.(map[string]any)
	if !ok {
		hooks = map[string]any{}
	}

	for event, spwMatchers := range spw.Hooks {
		var existingMatchers []any
		if raw, ok := hooks[event]; ok {
			if arr, ok := raw.([]any); ok {
				existingMatchers = arr
			}
		}

		// Filter out old SPW entries from existing matchers.
		var kept []any
		for _, m := range existingMatchers {
			mMap, ok := m.(map[string]any)
			if !ok {
				kept = append(kept, m)
				continue
			}
			if !matcherHasOnlySPWHooks(mMap) {
				kept = append(kept, m)
			}
		}

		// Append SPW matchers.
		for _, sm := range spwMatchers {
			entry := map[string]any{
				"matcher": sm.Matcher,
				"hooks":   hookEntriesToAny(sm.Hooks),
			}
			kept = append(kept, entry)
		}

		hooks[event] = kept
	}

	existing["hooks"] = hooks
}

// matcherHasOnlySPWHooks returns true if all hook commands in the matcher are
// SPW commands (meaning the whole matcher should be replaced during merge).
func matcherHasOnlySPWHooks(m map[string]any) bool {
	hooksRaw, ok := m["hooks"]
	if !ok {
		return false
	}
	hooksArr, ok := hooksRaw.([]any)
	if !ok {
		return false
	}
	if len(hooksArr) == 0 {
		return false
	}
	for _, h := range hooksArr {
		hMap, ok := h.(map[string]any)
		if !ok {
			return false
		}
		cmd, ok := hMap["command"].(string)
		if !ok || !isSPWCommand(cmd) {
			return false
		}
	}
	return true
}

func hookEntriesToAny(entries []hookEntry) []any {
	result := make([]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"type":    e.Type,
			"command": e.Command,
		}
	}
	return result
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
