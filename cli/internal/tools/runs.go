package tools

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunsLatestUnfinished finds the most recent unfinished run directory.
func RunsLatestUnfinished(cwd, phaseDirArg string, raw bool) {
	if phaseDirArg == "" {
		Fail("runs-latest-unfinished requires <phase-dir>", raw)
	}

	phaseDir := phaseDirArg
	if !filepath.IsAbs(phaseDir) {
		phaseDir = filepath.Join(cwd, phaseDir)
	}

	info, err := os.Stat(phaseDir)
	if err != nil || !info.IsDir() {
		result := map[string]any{"ok": true, "phase_dir": phaseDirArg, "found": false, "reason": "phase_dir_missing", "run": nil}
		Output(result, "", raw)
		return
	}

	entries, err := os.ReadDir(phaseDir)
	if err != nil {
		result := map[string]any{"ok": true, "phase_dir": phaseDirArg, "found": false, "reason": "read_error", "run": nil}
		Output(result, "", raw)
		return
	}

	type runInfo struct {
		name  string
		full  string
		mtime int64
	}
	var runs []runInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		full := filepath.Join(phaseDir, e.Name())
		fi, err := os.Stat(full)
		if err != nil {
			continue
		}
		runs = append(runs, runInfo{name: e.Name(), full: full, mtime: fi.ModTime().UnixMilli()})
	}

	sort.Slice(runs, func(i, j int) bool { return runs[i].mtime > runs[j].mtime })

	for _, run := range runs {
		inspection := inspectRunDir(run.full)
		if !inspection.unfinished {
			continue
		}
		rel, _ := filepath.Rel(cwd, run.full)
		result := map[string]any{
			"ok":        true,
			"phase_dir": phaseDirArg,
			"found":     true,
			"run":       rel,
			"issues":    inspection.issues,
			"subagents": inspection.subagents,
		}
		Output(result, rel, raw)
		return
	}

	result := map[string]any{"ok": true, "phase_dir": phaseDirArg, "found": false, "reason": "no_unfinished_run", "run": nil}
	Output(result, "", raw)
}

type runInspection struct {
	unfinished bool
	issues     []string
	subagents  []subagentInfo
}

type subagentInfo struct {
	Name    string   `json:"name"`
	Missing []string `json:"missing"`
	Blocked bool     `json:"blocked"`
}

func inspectRunDir(runDir string) runInspection {
	var issues []string
	var subagents []subagentInfo

	handoff := filepath.Join(runDir, "_handoff.md")
	if _, err := os.Stat(handoff); os.IsNotExist(err) {
		issues = append(issues, "missing:_handoff.md")
	}

	entries, err := os.ReadDir(runDir)
	if err != nil {
		return runInspection{unfinished: len(issues) > 0, issues: issues}
	}

	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), "_") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		full := filepath.Join(runDir, e.Name())
		required := []string{"brief.md", "report.md", "status.json"}
		var missing []string
		for _, f := range required {
			if _, err := os.Stat(filepath.Join(full, f)); os.IsNotExist(err) {
				missing = append(missing, f)
			}
		}
		if len(missing) > 0 {
			issues = append(issues, "missing:"+e.Name()+":"+strings.Join(missing, ","))
		}

		blocked := false
		statusPath := filepath.Join(full, "status.json")
		if data, err := os.ReadFile(statusPath); err == nil {
			if strings.Contains(strings.ToLower(string(data)), `"blocked"`) {
				blocked = true
				issues = append(issues, "blocked:"+e.Name())
			}
		}

		subagents = append(subagents, subagentInfo{Name: e.Name(), Missing: missing, Blocked: blocked})
	}

	return runInspection{unfinished: len(issues) > 0, issues: issues, subagents: subagents}
}
