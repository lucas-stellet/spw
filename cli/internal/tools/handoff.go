package tools

import (
	"os"
	"path/filepath"
)

// HandoffValidate checks file-first handoff completeness for a run directory.
func HandoffValidate(cwd, runDirArg string, raw bool) {
	if runDirArg == "" {
		Fail("handoff-validate requires <run-dir>", raw)
	}

	runDir := runDirArg
	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	info, err := os.Stat(runDir)
	if err != nil || !info.IsDir() {
		Fail("run directory not found: "+runDirArg, raw)
	}

	inspection := inspectRunDir(runDir)

	rawVal := "valid"
	if inspection.unfinished {
		rawVal = "invalid"
	}

	result := map[string]any{
		"ok":        true,
		"run_dir":   runDirArg,
		"valid":     !inspection.unfinished,
		"issues":    inspection.issues,
		"subagents": inspection.subagents,
	}
	Output(result, rawVal, raw)
}
