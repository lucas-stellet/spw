package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-stellet/oraculo/internal/config"
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
			Command: "oraculo hook statusline",
		},
		Hooks: map[string][]hookMatcher{
			"SessionStart": {
				{
					Matcher: "startup|resume|clear|compact",
					Hooks: []hookEntry{
						{Type: "command", Command: "oraculo hook session-start"},
					},
				},
			},
			"UserPromptSubmit": {
				{
					Matcher: ".*",
					Hooks: []hookEntry{
						{Type: "command", Command: "oraculo hook guard-prompt"},
					},
				},
			},
			"PreToolUse": {
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []hookEntry{
						{Type: "command", Command: "oraculo hook guard-paths"},
					},
				},
			},
			"Stop": {
				{
					Matcher: ".*",
					Hooks: []hookEntry{
						{Type: "command", Command: "oraculo hook guard-stop"},
					},
				},
			},
		},
	}
}

// WriteSettings creates .claude/settings.json if it doesn't exist, or merges
// ORACULO hooks into the existing file. It also manages Agent Teams settings based
// on the provided configuration.
func WriteSettings(root string, agentTeams config.AgentTeamsConfig) error {
	settingsPath := filepath.Join(root, ".claude", "settings.json")

	if _, err := os.Stat(settingsPath); err == nil {
		return MergeSettings(root, agentTeams)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	cfg := NewSettings()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Parse back to map so we can add Agent Teams fields.
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	mergeAgentTeams(m, agentTeams)

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		return err
	}

	fmt.Println("[oraculo] Created .claude/settings.json with hook registrations.")
	return nil
}

// MergeSettings reads an existing .claude/settings.json, merges ORACULO hooks into
// it (preserving all non-ORACULO hooks and other settings), manages Agent Teams
// settings, and writes it back.
func MergeSettings(root string, agentTeams config.AgentTeamsConfig) error {
	settingsPath := filepath.Join(root, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("read settings.json: %w", err)
	}

	var existing map[string]any
	if err := json.Unmarshal(data, &existing); err != nil {
		return fmt.Errorf("parse settings.json: %w", err)
	}

	cfg := NewSettings()

	// Merge statusLine: overwrite only if absent or already an ORACULO statusline.
	mergeStatusLine(existing, cfg)

	// Merge hooks: for each ORACULO event, remove old ORACULO entries, append new ones,
	// preserve all non-ORACULO entries.
	mergeHooks(existing, cfg)

	// Merge Agent Teams settings.
	mergeAgentTeams(existing, agentTeams)

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings.json: %w", err)
	}

	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	fmt.Println("[oraculo] Hooks merged into .claude/settings.json.")
	return nil
}

// isORACULOCommand returns true if the command string is an ORACULO hook command.
func isORACULOCommand(cmd string) bool {
	return strings.HasPrefix(cmd, "oraculo hook ")
}

func mergeStatusLine(existing map[string]any, cfg settingsJSON) {
	shouldSet := false
	if _, ok := existing["statusLine"]; !ok {
		shouldSet = true
	} else if sl, ok := existing["statusLine"].(map[string]any); ok {
		if cmd, ok := sl["command"].(string); ok && isORACULOCommand(cmd) {
			shouldSet = true
		}
	}
	if shouldSet {
		existing["statusLine"] = map[string]any{
			"type":    cfg.StatusLine.Type,
			"command": cfg.StatusLine.Command,
		}
	}
}

func mergeHooks(existing map[string]any, cfg settingsJSON) {
	hooksAny, ok := existing["hooks"]
	if !ok {
		hooksAny = map[string]any{}
	}
	hooks, ok := hooksAny.(map[string]any)
	if !ok {
		hooks = map[string]any{}
	}

	for event, cfgMatchers := range cfg.Hooks {
		var existingMatchers []any
		if raw, ok := hooks[event]; ok {
			if arr, ok := raw.([]any); ok {
				existingMatchers = arr
			}
		}

		// Filter out old ORACULO entries from existing matchers.
		var kept []any
		for _, m := range existingMatchers {
			mMap, ok := m.(map[string]any)
			if !ok {
				kept = append(kept, m)
				continue
			}
			if !matcherHasOnlyORACULOHooks(mMap) {
				kept = append(kept, m)
			}
		}

		// Append ORACULO matchers.
		for _, sm := range cfgMatchers {
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

// matcherHasOnlyORACULOHooks returns true if all hook commands in the matcher are
// ORACULO commands (meaning the whole matcher should be replaced during merge).
func matcherHasOnlyORACULOHooks(m map[string]any) bool {
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
		if !ok || !isORACULOCommand(cmd) {
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
	return containsBytes(data, []byte("node ./.claude/hooks/oraculo-")) ||
		containsBytes(data, []byte("oraculo-statusline.js")) ||
		containsBytes(data, []byte("oraculo-guard-"))
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

// mergeAgentTeams adds or removes Agent Teams keys in settings.json.
func mergeAgentTeams(existing map[string]any, agentTeams config.AgentTeamsConfig) {
	if agentTeams.Enabled {
		// Ensure env map exists and set the flag.
		env, ok := existing["env"].(map[string]any)
		if !ok {
			env = map[string]any{}
		}
		env["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"] = "1"
		existing["env"] = env

		// Set teammateMode (default "in-process").
		mode := agentTeams.TeammateMode
		if mode == "" {
			mode = "in-process"
		}
		existing["teammateMode"] = mode

		fmt.Printf("[oraculo] Enabled Agent Teams in settings.json (teammateMode=%s).\n", mode)
	} else {
		// Remove Agent Teams keys if present.
		if env, ok := existing["env"].(map[string]any); ok {
			delete(env, "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS")
			if len(env) == 0 {
				delete(existing, "env")
			}
		}
		delete(existing, "teammateMode")
	}
}
