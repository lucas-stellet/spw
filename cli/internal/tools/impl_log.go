package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lucas-stellet/oraculo/internal/specdir"
)

// implLogRegisterResult builds the registration result without Output/Fail side effects.
func implLogRegisterResult(cwd, specName, taskID, wave, title, files, changes, tests string) (map[string]any, string, error) {
	specDirAbs := specdir.SpecDirAbs(cwd, specName)
	logPath := specdir.ImplLogPath(specDirAbs, taskID)

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create impl-logs dir: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Task %s: %s\n\n", taskID, title))
	sb.WriteString(fmt.Sprintf("- **Wave**: %s\n", wave))
	sb.WriteString(fmt.Sprintf("- **Timestamp**: %s\n\n", timestamp))
	sb.WriteString("## Files Changed\n\n")
	for _, f := range strings.Split(files, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			sb.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
	}
	sb.WriteString("\n## Changes\n\n")
	sb.WriteString(changes)
	sb.WriteString("\n")
	if tests != "" {
		sb.WriteString("\n## Tests\n\n")
		sb.WriteString(tests)
		sb.WriteString("\n")
	}

	if err := os.WriteFile(logPath, []byte(sb.String()), 0644); err != nil {
		return nil, "", fmt.Errorf("failed to write impl log: %w", err)
	}

	relPath := specdir.ImplLogPath(specdir.SpecDir(specName), taskID)

	result := map[string]any{
		"ok":      true,
		"task_id": taskID,
		"path":    relPath,
		"created": true,
	}
	return result, relPath, nil
}

// ImplLogRegister creates an implementation log file for a completed task.
func ImplLogRegister(cwd, specName, taskID, wave, title, files, changes, tests string, raw bool) {
	if specName == "" || taskID == "" || wave == "" || title == "" || files == "" || changes == "" {
		Fail("impl-log register requires --spec, --task-id, --wave, --title, --files, --changes", raw)
	}

	result, relPath, err := implLogRegisterResult(cwd, specName, taskID, wave, title, files, changes, tests)
	if err != nil {
		Fail(err.Error(), raw)
	}

	Output(result, relPath, raw)
}

// implLogCheckResult builds the check result without Output/Fail side effects.
func implLogCheckResult(cwd, specName, taskIDs string) (map[string]any, error) {
	ids := strings.Split(taskIDs, ",")

	allPresent := true
	var missing []string
	tasks := map[string]any{}

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		specDirAbs := specdir.SpecDirAbs(cwd, specName)
		logPath := specdir.ImplLogPath(specDirAbs, id)
		exists := specdir.FileExists(logPath)
		relPath := specdir.ImplLogPath(specdir.SpecDir(specName), id)

		tasks[id] = map[string]any{
			"exists": exists,
			"path":   relPath,
		}

		if !exists {
			allPresent = false
			missing = append(missing, id)
		}
	}

	result := map[string]any{
		"ok":          true,
		"all_present": allPresent,
		"tasks":       tasks,
	}
	if len(missing) > 0 {
		result["missing"] = missing
	}

	return result, nil
}

// ImplLogCheck verifies that implementation logs exist for a set of task IDs.
func ImplLogCheck(cwd, specName, taskIDs string, raw bool) {
	if specName == "" || taskIDs == "" {
		Fail("impl-log check requires --spec and --task-ids", raw)
	}

	result, err := implLogCheckResult(cwd, specName, taskIDs)
	if err != nil {
		Fail(err.Error(), raw)
	}

	allPresent, _ := result["all_present"].(bool)
	rawValue := "false"
	if allPresent {
		rawValue = "true"
	}

	Output(result, rawValue, raw)
}
