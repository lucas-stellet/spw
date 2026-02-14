package spec

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-stellet/oraculo/internal/specdir"
)

// ClassifyStage determines the current lifecycle stage of a spec based on
// which artifacts exist and their state.
func ClassifyStage(specDir string) string {
	// Check most-advanced stages first (reverse lifecycle order)

	// Post-mortem
	if specdir.FileExists(filepath.Join(specDir, specdir.PostMortemReport)) {
		return "post-mortem"
	}

	// QA phase
	if specdir.FileExists(filepath.Join(specDir, specdir.QATestPlan)) ||
		specdir.FileExists(filepath.Join(specDir, specdir.QACheckMD)) ||
		specdir.FileExists(filepath.Join(specDir, specdir.QAExecReport)) {
		return "qa"
	}

	// Execution phase
	if specdir.DirExists(filepath.Join(specDir, specdir.WavesDir)) {
		return "execution"
	}

	// Planning phase
	if specdir.FileExists(filepath.Join(specDir, specdir.TasksMD)) {
		// Check if tasks have completed tasks marker — if all done, could be "complete"
		if allTasksComplete(filepath.Join(specDir, specdir.TasksMD)) {
			// If there's no execution/qa/post-mortem phase, we're still in planning
			// (the task completions are tracked in tasks.md but execution hasn't started)
			return "planning"
		}
		return "planning"
	}

	// Design phase
	if specdir.FileExists(filepath.Join(specDir, specdir.DesignMD)) ||
		specdir.FileExists(filepath.Join(specDir, specdir.DesignResearchMD)) {
		return "design"
	}

	// Requirements phase
	if specdir.FileExists(filepath.Join(specDir, specdir.RequirementsMD)) {
		return "requirements"
	}

	return "unknown"
}

// allTasksComplete checks if tasks.md has all tasks marked as complete.
// Returns false if the file can't be read or has no task markers.
func allTasksComplete(tasksPath string) bool {
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return false
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	hasUnchecked := false
	hasChecked := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			hasChecked = true
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			hasUnchecked = true
		}
	}

	// Only return true if there are checked tasks and no unchecked ones
	return hasChecked && !hasUnchecked
}
