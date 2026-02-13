// Package workspace handles workspace root detection, spec directories,
// guidelines, and template loading.
package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/spw/internal/git"
)

// Payload represents the JSON input from Claude Code hooks.
type Payload struct {
	CWD       string `json:"cwd"`
	Prompt    string `json:"prompt"`
	Model     *struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
	} `json:"model"`
	Workspace *struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	SessionID     string `json:"session_id"`
	ContextWindow *struct {
		RemainingPercentage *float64 `json:"remaining_percentage"`
		TotalInputTokens    *int64   `json:"total_input_tokens"`
		TotalOutputTokens   *int64   `json:"total_output_tokens"`
	} `json:"context_window"`
	Cost *struct {
		TotalCostUSD *float64 `json:"total_cost_usd"`
	} `json:"cost"`
	ToolInput json.RawMessage `json:"tool_input"`
	ToolName  string          `json:"tool_name"`
}

// ReadStdinPayload reads JSON from stdin.
func ReadStdinPayload() Payload {
	var p Payload
	dec := json.NewDecoder(os.Stdin)
	_ = dec.Decode(&p) // fail-open
	return p
}

// GetWorkspaceRoot resolves the workspace root from the payload.
func GetWorkspaceRoot(p Payload) string {
	candidates := []string{
		p.CWD,
	}
	if p.Workspace != nil {
		candidates = append(candidates, p.Workspace.CurrentDir)
	}
	if env := os.Getenv("CLAUDE_PROJECT_DIR"); env != "" {
		candidates = append(candidates, env)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}

	for _, c := range candidates {
		if c == "" {
			continue
		}
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}

	cwd, _ := os.Getwd()
	return cwd
}

// GetRepoRoot returns the git repository root, or empty string.
func GetRepoRoot(dir string) string {
	return git.RepoRoot(dir)
}

// SpecsRoot returns the path to .spec-workflow/specs/.
func SpecsRoot(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".spec-workflow", "specs")
}

// ListSpecDirs returns all spec directories under .spec-workflow/specs/.
func ListSpecDirs(workspaceRoot string) []string {
	specsRoot := SpecsRoot(workspaceRoot)
	entries, err := os.ReadDir(specsRoot)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(specsRoot, e.Name()))
		}
	}
	return dirs
}
