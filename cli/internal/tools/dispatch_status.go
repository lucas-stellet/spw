package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DispatchReadStatus reads and validates a subagent's status.json.
func DispatchReadStatus(cwd, subagentName, runDir string, raw bool) {
	if subagentName == "" || runDir == "" {
		Fail("dispatch-read-status requires <subagent-name> --run-dir", raw)
	}

	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	statusFile := filepath.Join(runDir, subagentName, "status.json")

	data, err := os.ReadFile(statusFile)
	if err != nil {
		result := map[string]any{
			"ok":     true,
			"status": "missing",
			"valid":  false,
			"error":  "status.json not found",
		}
		Output(result, "missing", raw)
		return
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		result := map[string]any{
			"ok":     true,
			"status": "invalid",
			"valid":  false,
			"error":  "invalid JSON: " + err.Error(),
		}
		Output(result, "invalid", raw)
		return
	}

	status, _ := doc["status"].(string)
	summary, _ := doc["summary"].(string)

	valid := status == "pass" || status == "blocked"

	result := map[string]any{
		"ok":      true,
		"status":  status,
		"summary": summary,
		"valid":   valid,
	}
	if !valid {
		result["error"] = "status must be \"pass\" or \"blocked\", got \"" + status + "\""
	}
	Output(result, status, raw)
}
