package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/lucas-stellet/spw/internal/specdir"
)

// taskMarkResult performs the task marking logic and returns the result.
// Extracted for testability — the public TaskMark function wraps this with Output/Fail.
func taskMarkResult(cwd, specName, taskID, status string) (map[string]any, error) {
	tasksPath := specdir.TasksPath(specdir.SpecDirAbs(cwd, specName))

	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks.md: %w", err)
	}

	content := string(data)

	// Match task line: `- [ ] **ID.** ...` where checkbox marker is space, x, -, or !
	pattern := fmt.Sprintf(`(?m)^(\s*- \[)([ x\-!])(\] \*\*%s\.\*\*)`, regexp.QuoteMeta(taskID))
	re := regexp.MustCompile(pattern)

	loc := re.FindStringSubmatchIndex(content)
	if loc == nil {
		return nil, fmt.Errorf("task ID %s not found in tasks.md", taskID)
	}

	// Extract previous marker (group 2)
	prevMarker := content[loc[4]:loc[5]]
	var prevStatus string
	switch prevMarker {
	case " ":
		prevStatus = "pending"
	case "-":
		prevStatus = "in-progress"
	case "x":
		prevStatus = "done"
	case "!":
		prevStatus = "blocked"
	default:
		prevStatus = "unknown"
	}

	// Determine new marker
	var newMarker string
	switch status {
	case "in-progress":
		newMarker = "-"
	case "done":
		newMarker = "x"
	case "blocked":
		newMarker = "!"
	}

	// Replace the marker in content
	updated := content[:loc[4]] + newMarker + content[loc[5]:]

	if err := os.WriteFile(tasksPath, []byte(updated), 0644); err != nil {
		return nil, fmt.Errorf("failed to write tasks.md: %w", err)
	}

	tasksRel, _ := filepath.Rel(cwd, tasksPath)

	return map[string]any{
		"ok":              true,
		"task_id":         taskID,
		"previous_status": prevStatus,
		"new_status":      status,
		"tasks_path":      tasksRel,
	}, nil
}

// TaskMark updates a task's checkbox status in tasks.md.
func TaskMark(cwd, specName, taskID, status string, raw bool) {
	if specName == "" || taskID == "" || status == "" {
		Fail("task-mark requires --spec, --task-id, and --status", raw)
	}

	validStatuses := map[string]bool{"in-progress": true, "done": true, "blocked": true}
	if !validStatuses[status] {
		Fail("status must be one of: in-progress, done, blocked", raw)
	}

	result, err := taskMarkResult(cwd, specName, taskID, status)
	if err != nil {
		Fail(err.Error(), raw)
	}

	tasksPath, _ := result["tasks_path"].(string)
	Output(result, tasksPath, raw)
}
