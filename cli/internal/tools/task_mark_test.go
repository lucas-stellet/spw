package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testTasksMD = `# Tasks

## Wave 01

- [ ] **1.** Setup project structure
  - Files: ` + "`src/main.go`" + `
  - Wave: 01
- [-] **2.** Implement core logic
  - Files: ` + "`src/core.go`" + `
  - Wave: 01
- [x] **3.** Write tests
  - Files: ` + "`src/core_test.go`" + `
  - Wave: 01
`

func setupTasksDir(t *testing.T, content string) string {
	t.Helper()
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, ".spec-workflow", "specs", "test-spec")
	os.MkdirAll(specDir, 0755)
	os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(content), 0644)
	return tmp
}

func readTasksMD(t *testing.T, cwd string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "tasks.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestTaskMarkPendingToInProgress(t *testing.T) {
	cwd := setupTasksDir(t, testTasksMD)

	result, err := taskMarkResult(cwd, "test-spec", "1", "in-progress")
	if err != nil {
		t.Fatal(err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["task_id"] != "1" {
		t.Errorf("task_id = %v, want 1", result["task_id"])
	}
	if result["previous_status"] != "pending" {
		t.Errorf("previous_status = %v, want pending", result["previous_status"])
	}
	if result["new_status"] != "in-progress" {
		t.Errorf("new_status = %v, want in-progress", result["new_status"])
	}

	content := readTasksMD(t, cwd)
	if !strings.Contains(content, "- [-] **1.** Setup project structure") {
		t.Error("tasks.md should have [-] marker for task 1")
	}
}

func TestTaskMarkInProgressToDone(t *testing.T) {
	cwd := setupTasksDir(t, testTasksMD)

	result, err := taskMarkResult(cwd, "test-spec", "2", "done")
	if err != nil {
		t.Fatal(err)
	}

	if result["previous_status"] != "in-progress" {
		t.Errorf("previous_status = %v, want in-progress", result["previous_status"])
	}
	if result["new_status"] != "done" {
		t.Errorf("new_status = %v, want done", result["new_status"])
	}

	content := readTasksMD(t, cwd)
	if !strings.Contains(content, "- [x] **2.** Implement core logic") {
		t.Error("tasks.md should have [x] marker for task 2")
	}
}

func TestTaskMarkToBlocked(t *testing.T) {
	cwd := setupTasksDir(t, testTasksMD)

	result, err := taskMarkResult(cwd, "test-spec", "1", "blocked")
	if err != nil {
		t.Fatal(err)
	}

	if result["previous_status"] != "pending" {
		t.Errorf("previous_status = %v, want pending", result["previous_status"])
	}
	if result["new_status"] != "blocked" {
		t.Errorf("new_status = %v, want blocked", result["new_status"])
	}

	content := readTasksMD(t, cwd)
	if !strings.Contains(content, "- [!] **1.** Setup project structure") {
		t.Error("tasks.md should have [!] marker for task 1")
	}
}

func TestTaskMarkNotFound(t *testing.T) {
	cwd := setupTasksDir(t, testTasksMD)

	_, err := taskMarkResult(cwd, "test-spec", "99", "done")
	if err == nil {
		t.Fatal("expected error for non-existent task ID")
	}
	if !strings.Contains(err.Error(), "task ID 99 not found") {
		t.Errorf("error = %v, want 'task ID 99 not found'", err)
	}
}

const testTasksMDMultiDigit = `# Tasks

## Wave 01

- [ ] **1.** Setup project structure
  - Files: ` + "`src/main.go`" + `
  - Wave: 01
- [ ] **10.** Configure deployment
  - Files: ` + "`deploy.yaml`" + `
  - Wave: 01
- [ ] **11.** Add monitoring
  - Files: ` + "`src/monitor.go`" + `
  - Wave: 01
`

func TestTaskMarkMultiDigitNoSubstringMatch(t *testing.T) {
	cwd := setupTasksDir(t, testTasksMDMultiDigit)

	// Mark task 1 as done — must NOT accidentally match task 10 or 11
	_, err := taskMarkResult(cwd, "test-spec", "1", "done")
	if err != nil {
		t.Fatal(err)
	}

	content := readTasksMD(t, cwd)
	if !strings.Contains(content, "- [x] **1.** Setup project structure") {
		t.Error("task 1 should be marked done")
	}
	if !strings.Contains(content, "- [ ] **10.** Configure deployment") {
		t.Error("task 10 should remain pending (substring must not match)")
	}
	if !strings.Contains(content, "- [ ] **11.** Add monitoring") {
		t.Error("task 11 should remain pending (substring must not match)")
	}

	// Now mark task 10 as in-progress
	_, err = taskMarkResult(cwd, "test-spec", "10", "in-progress")
	if err != nil {
		t.Fatal(err)
	}

	content = readTasksMD(t, cwd)
	if !strings.Contains(content, "- [-] **10.** Configure deployment") {
		t.Error("task 10 should be marked in-progress")
	}
	if !strings.Contains(content, "- [ ] **11.** Add monitoring") {
		t.Error("task 11 should still remain pending")
	}
}
