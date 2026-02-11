package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/spw/internal/config"
)

// DispatchSetup creates a subagent directory with a brief.md skeleton.
func DispatchSetup(cwd, subagentName, runDir, modelAlias string, raw bool) {
	if subagentName == "" || runDir == "" {
		Fail("dispatch-setup requires <subagent-name> --run-dir", raw)
	}

	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	subagentDir := filepath.Join(runDir, subagentName)
	if err := os.MkdirAll(subagentDir, 0755); err != nil {
		Fail("failed to create subagent dir: "+err.Error(), raw)
	}

	// Resolve model from alias
	cfg, _ := config.Load(cwd)
	model := ""
	if modelAlias != "" {
		switch modelAlias {
		case "web_research":
			model = cfg.Models.WebResearch
		case "complex_reasoning":
			model = cfg.Models.ComplexReasoning
		case "implementation":
			model = cfg.Models.Implementation
		default:
			Fail("unknown model alias: "+modelAlias+"; use web_research, complex_reasoning, or implementation", raw)
		}
	}

	subagentRel, _ := filepath.Rel(cwd, subagentDir)
	briefPath := filepath.Join(subagentRel, "brief.md")
	reportPath := filepath.Join(subagentRel, "report.md")
	statusPath := filepath.Join(subagentRel, "status.json")

	brief := fmt.Sprintf(`# Brief: %s

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->

## Task
<!-- Describe what this subagent must do -->

## Output Contract
Write your output to these exact paths:
- Report: %s
- Status: %s

status.json format:
`+"```json"+`
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["skill-name"],
  "skills_missing": [],
  "model_override_reason": null
}
`+"```"+`
`, subagentName, reportPath, statusPath)

	briefFullPath := filepath.Join(subagentDir, "brief.md")
	if err := os.WriteFile(briefFullPath, []byte(brief), 0644); err != nil {
		Fail("failed to write brief.md: "+err.Error(), raw)
	}

	result := map[string]any{
		"ok":           true,
		"subagent_dir": subagentRel,
		"brief_path":   briefPath,
		"report_path":  reportPath,
		"status_path":  statusPath,
		"model":        model,
	}
	Output(result, subagentRel, raw)
}
