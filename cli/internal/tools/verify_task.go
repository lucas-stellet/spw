package tools

import (
	"os/exec"
	"strings"

	"github.com/lucas-stellet/spw/internal/specdir"
)

// verifyTaskResult builds the verification result without Output/Fail side effects.
func verifyTaskResult(cwd, specName, taskID string, checkCommit bool) (map[string]any, error) {
	specDirAbs := specdir.SpecDirAbs(cwd, specName)
	implLogPathAbs := specdir.ImplLogPath(specDirAbs, taskID)
	implLogPathRel := specdir.ImplLogPath(specdir.SpecDir(specName), taskID)

	implLogExists := specdir.FileExists(implLogPathAbs)

	result := map[string]any{
		"ok":      true,
		"task_id": taskID,
		"impl_log": map[string]any{
			"exists": implLogExists,
			"path":   implLogPathRel,
		},
	}

	if checkCommit {
		commitInfo := map[string]any{"exists": false}
		out, err := exec.Command("git", "-C", cwd, "log", "--oneline", "-E", "--grep=task "+taskID+"([^0-9]|$)").Output()
		if err == nil {
			lines := strings.TrimSpace(string(out))
			if lines != "" {
				first := strings.SplitN(lines, "\n", 2)[0]
				parts := strings.SplitN(first, " ", 2)
				commitInfo["exists"] = true
				commitInfo["hash"] = parts[0]
				if len(parts) > 1 {
					commitInfo["message"] = parts[1]
				}
			}
		}
		result["commit"] = commitInfo
	}

	return result, nil
}

// VerifyTask checks whether a task has an implementation log and optionally a commit.
func VerifyTask(cwd, specName, taskID string, checkCommit, raw bool) {
	if specName == "" || taskID == "" {
		Fail("verify-task requires --spec and --task-id", raw)
	}

	result, _ := verifyTaskResult(cwd, specName, taskID, checkCommit)

	// Determine raw value: use path if impl log exists, otherwise task ID
	rawValue := taskID
	implLog, _ := result["impl_log"].(map[string]any)
	if implLog != nil {
		if p, ok := implLog["path"].(string); ok {
			rawValue = p
		}
	}

	Output(result, rawValue, raw)
}
