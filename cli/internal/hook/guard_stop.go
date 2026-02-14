package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lucas-stellet/oraculo/internal/workspace"
)

var runDirRe = regexp.MustCompile(`^run-\d{3}$`)

// HandleGuardStop checks file-first handoff completeness in recent run dirs.
func HandleGuardStop() error {
	ctx := newHookContext()
	if !ctx.cfg.Hooks.Enabled || !ctx.cfg.Hooks.GuardStopHandoff {
		return nil
	}

	now := time.Now()
	windowMs := int64(max(1, ctx.cfg.Hooks.RecentRunWindowMinutes)) * 60 * 1000
	var violations []string

	for _, specDir := range workspace.ListSpecDirs(ctx.workspaceRoot) {
		runDirs := collectRunDirs(specDir)
		for _, runDir := range runDirs {
			if !isRecent(runDir, now, windowMs) {
				continue
			}
			issues := checkRunCompleteness(runDir)
			if len(issues) == 0 {
				continue
			}
			rel, err := filepath.Rel(ctx.workspaceRoot, runDir)
			if err != nil {
				rel = runDir
			}
			violations = append(violations, normalizeSlashes(rel)+" -> "+strings.Join(issues, "; "))
		}
	}

	if len(violations) > 0 {
		details := []string{
			fmt.Sprintf("Window: last %d minute(s)", ctx.cfg.Hooks.RecentRunWindowMinutes),
		}
		limit := 20
		if len(violations) < limit {
			limit = len(violations)
		}
		details = append(details, violations[:limit]...)
		emitViolation(ctx.cfg.Hooks, "Recent run folders are missing required handoff files", details)
	}

	emitInfo(ctx.cfg.Hooks, "Stop guard passed.")
	return nil
}

func isRecent(runDir string, now time.Time, windowMs int64) bool {
	info, err := os.Stat(runDir)
	if err != nil {
		return false
	}
	return now.UnixMilli()-info.ModTime().UnixMilli() <= windowMs
}

func checkRunCompleteness(runDir string) []string {
	var issues []string

	handoffPath := filepath.Join(runDir, "_handoff.md")
	if _, err := os.Stat(handoffPath); os.IsNotExist(err) {
		issues = append(issues, "missing _handoff.md")
	}

	entries := listDirSafe(runDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subagentDir := filepath.Join(runDir, e.Name())
		var missing []string
		for _, f := range []string{"brief.md", "report.md", "status.json"} {
			if _, err := os.Stat(filepath.Join(subagentDir, f)); os.IsNotExist(err) {
				missing = append(missing, f)
			}
		}
		if len(missing) > 0 {
			issues = append(issues, e.Name()+": missing "+strings.Join(missing, ", "))
		}
	}

	return issues
}

func collectRunDirs(specDir string) []string {
	var runs []string

	// Phase dirs with direct run dirs: discover/_comms/run-NNN/, post-mortem/_comms/run-NNN/
	for _, phase := range []string{"discover", "post-mortem"} {
		commsRoot := filepath.Join(specDir, phase, "_comms")
		for _, entry := range listDirSafe(commsRoot) {
			if !entry.IsDir() {
				continue
			}
			if runDirRe.MatchString(entry.Name()) {
				runs = append(runs, filepath.Join(commsRoot, entry.Name()))
			}
		}
		// Nested command dirs (e.g. discover/_comms/discover-revision/run-NNN/)
		for _, entry := range listDirSafe(commsRoot) {
			if !entry.IsDir() || runDirRe.MatchString(entry.Name()) {
				continue
			}
			subDir := filepath.Join(commsRoot, entry.Name())
			for _, runEntry := range listDirSafe(subDir) {
				if runEntry.IsDir() {
					runs = append(runs, filepath.Join(subDir, runEntry.Name()))
				}
			}
		}
	}

	// Command-scoped run dirs
	commandScoped := []struct {
		phase    string
		commands []string
	}{
		{"design", []string{"design-research", "design-draft"}},
		{"planning", []string{"tasks-plan", "tasks-check"}},
		{"qa", []string{"qa", "qa-check"}},
	}
	for _, cs := range commandScoped {
		for _, cmd := range cs.commands {
			cmdRoot := filepath.Join(specDir, cs.phase, "_comms", cmd)
			for _, entry := range listDirSafe(cmdRoot) {
				if entry.IsDir() {
					runs = append(runs, filepath.Join(cmdRoot, entry.Name()))
				}
			}
		}
	}

	// Execution waves: execution/waves/wave-NN/{execution,checkpoint,post-check}/run-NNN/
	execWavesRoot := filepath.Join(specDir, "execution", "waves")
	for _, waveEntry := range listDirSafe(execWavesRoot) {
		if !waveEntry.IsDir() {
			continue
		}
		waveDir := filepath.Join(execWavesRoot, waveEntry.Name())
		for _, stage := range []string{"execution", "checkpoint", "post-check"} {
			stageDir := filepath.Join(waveDir, stage)
			for _, runEntry := range listDirSafe(stageDir) {
				if runEntry.IsDir() {
					runs = append(runs, filepath.Join(stageDir, runEntry.Name()))
				}
			}
		}
	}

	// QA exec waves: qa/_comms/qa-exec/waves/wave-NN/run-NNN/
	qaWavesRoot := filepath.Join(specDir, "qa", "_comms", "qa-exec", "waves")
	for _, waveEntry := range listDirSafe(qaWavesRoot) {
		if !waveEntry.IsDir() {
			continue
		}
		waveDir := filepath.Join(qaWavesRoot, waveEntry.Name())
		for _, runEntry := range listDirSafe(waveDir) {
			if runEntry.IsDir() {
				runs = append(runs, filepath.Join(waveDir, runEntry.Name()))
			}
		}
	}

	return runs
}

func listDirSafe(dir string) []os.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	return entries
}
