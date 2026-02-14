package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lucas-stellet/spw/internal/specdir"
	"github.com/lucas-stellet/spw/internal/store"
)

// statusToChar maps status strings to their checkbox characters.
var statusToChar = map[string]string{
	"done":        "x",
	"in_progress": "-",
	"pending":     " ",
}

// implLogRe matches implementation log filenames: task-<id>.md or task_<id>.md
var implLogRe = regexp.MustCompile(`^task[_\-]?(\S+)\.md$`)

// MarkTaskInFile atomically updates a single task's checkbox in tasks.md.
// The update is surgical: only the checkbox character changes.
// If requireImplLog is true, the task cannot be marked "done" unless an
// implementation log exists at execution/_implementation-logs/task-<id>.md.
func MarkTaskInFile(filePath, taskID, newStatus string, requireImplLog bool, specDir string) error {
	char, ok := statusToChar[newStatus]
	if !ok {
		return fmt.Errorf("invalid status %q: must be done, in_progress, or pending", newStatus)
	}

	// If marking done, check for implementation log
	if newStatus == "done" && requireImplLog {
		if err := checkImplLog(specDir, taskID); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	found := false

	// Build a regex to match the specific task line
	taskPattern := regexp.MustCompile(`^(- \[)([ x\-])(\] ` + regexp.QuoteMeta(taskID) + `\s)`)

	for i, line := range lines {
		if m := taskPattern.FindStringSubmatchIndex(line); m != nil {
			// Replace only the checkbox character (group 2)
			lines[i] = line[:m[4]] + char + line[m[5]:]
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task %s not found in %s", taskID, filePath)
	}

	// Write back preserving original line endings
	output := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return err
	}

	// Dual-write: sync task status to spec.db
	if specDir != "" {
		if s := store.TryOpen(specDir); s != nil {
			defer s.Close()
			s.SyncTask(store.TaskRecord{
				TaskID: taskID,
				Status: newStatus,
			})
		}
	}

	return nil
}

// checkImplLog verifies that an implementation log exists for the given task.
// Primary check: specdir.ImplLogPath (task-<id>.md).
// Fallback: scan the impl logs directory for flexible naming (task_<id>.md, task<id>.md).
func checkImplLog(specDir, taskID string) error {
	if specDir == "" {
		return fmt.Errorf("specDir required when requireImplLog is true")
	}

	// Primary: canonical path
	primary := specdir.ImplLogPath(specDir, taskID)
	if specdir.FileExists(primary) {
		return nil
	}

	// Fallback: scan directory for flexible naming
	logsDir := filepath.Join(specDir, specdir.ImplLogsDir)
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return fmt.Errorf("implementation log missing for task %s (expected at %s)", taskID, primary)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := implLogRe.FindStringSubmatch(e.Name())
		if m != nil && m[1] == taskID {
			return nil
		}
	}

	return fmt.Errorf("implementation log missing for task %s (expected at %s)", taskID, primary)
}
