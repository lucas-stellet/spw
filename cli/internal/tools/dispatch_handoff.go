package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type subagentStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

// DispatchHandoff generates _handoff.md from subagent status.json files.
func DispatchHandoff(cwd, runDir, command string, raw bool) {
	if runDir == "" {
		Fail("dispatch-handoff requires --run-dir", raw)
	}

	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	info, err := os.Stat(runDir)
	if err != nil || !info.IsDir() {
		Fail("run directory not found: "+runDir, raw)
	}

	entries, err := os.ReadDir(runDir)
	if err != nil {
		Fail("failed to read run dir: "+err.Error(), raw)
	}

	var agents []subagentStatus
	allPass := true

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), "_") || strings.HasPrefix(e.Name(), ".") {
			continue
		}

		statusFile := filepath.Join(runDir, e.Name(), "status.json")
		data, err := os.ReadFile(statusFile)
		if err != nil {
			agents = append(agents, subagentStatus{Name: e.Name(), Status: "missing", Summary: "status.json not found"})
			allPass = false
			continue
		}

		var doc map[string]any
		if err := json.Unmarshal(data, &doc); err != nil {
			agents = append(agents, subagentStatus{Name: e.Name(), Status: "invalid", Summary: "invalid JSON"})
			allPass = false
			continue
		}

		status, _ := doc["status"].(string)
		summary, _ := doc["summary"].(string)
		agents = append(agents, subagentStatus{Name: e.Name(), Status: status, Summary: summary})
		if status != "pass" {
			allPass = false
		}
	}

	// Generate _handoff.md
	var sb strings.Builder
	sb.WriteString("# Handoff Summary\n\n")
	sb.WriteString("| Subagent | Status | Summary |\n")
	sb.WriteString("|----------|--------|---------|\n")
	for _, a := range agents {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", a.Name, a.Status, a.Summary))
	}
	sb.WriteString(fmt.Sprintf("\n**All pass:** %v\n", allPass))

	handoffPath := filepath.Join(runDir, "_handoff.md")
	if err := os.WriteFile(handoffPath, []byte(sb.String()), 0644); err != nil {
		Fail("failed to write _handoff.md: "+err.Error(), raw)
	}

	handoffRel, _ := filepath.Rel(cwd, handoffPath)

	result := map[string]any{
		"ok":           true,
		"handoff_path": handoffRel,
		"subagents":    agents,
		"all_pass":     allPass,
	}

	if command != "" {
		if meta, ok := commandRegistry[command]; ok {
			result["category"] = meta.Category
		}
	}

	Output(result, handoffRel, raw)
}
