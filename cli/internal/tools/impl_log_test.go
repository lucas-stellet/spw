package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucas-stellet/spw/internal/specdir"
)

func TestImplLogRegister(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"
	taskID := "1.1"
	wave := "01"
	title := "Add user authentication"
	files := "src/auth.go, src/auth_test.go"
	changes := "Implemented JWT-based authentication flow."
	tests := "Added unit tests for token generation and validation."

	result, relPath, err := implLogRegisterResult(tmp, specName, taskID, wave, title, files, changes, tests)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["task_id"] != taskID {
		t.Errorf("task_id = %q, want %q", result["task_id"], taskID)
	}
	if result["created"] != true {
		t.Error("expected created=true")
	}
	if relPath == "" {
		t.Error("expected non-empty relPath")
	}

	// Verify the file was actually created
	absPath := specdir.ImplLogPath(specdir.SpecDirAbs(tmp, specName), taskID)
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("impl log file not created: %v", err)
	}

	content := string(data)
	expectations := []string{
		"# Task 1.1: Add user authentication",
		"**Wave**: 01",
		"**Timestamp**:",
		"`src/auth.go`",
		"`src/auth_test.go`",
		"Implemented JWT-based authentication flow.",
		"Added unit tests for token generation and validation.",
	}
	for _, exp := range expectations {
		if !strings.Contains(content, exp) {
			t.Errorf("impl log missing expected content: %q", exp)
		}
	}
}

func TestImplLogRegisterWithoutTests(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"
	taskID := "2.1"

	result, _, err := implLogRegisterResult(tmp, specName, taskID, "02", "Fix bug", "src/fix.go", "Fixed null pointer.", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}

	// Verify no Tests section when tests is empty
	absPath := specdir.ImplLogPath(specdir.SpecDirAbs(tmp, specName), taskID)
	data, _ := os.ReadFile(absPath)
	if strings.Contains(string(data), "## Tests") {
		t.Error("should not contain Tests section when tests is empty")
	}
}

func TestImplLogCheckAllPresent(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"

	// Create impl logs for tasks 1.1 and 1.2
	for _, id := range []string{"1.1", "1.2"} {
		logPath := specdir.ImplLogPath(specdir.SpecDirAbs(tmp, specName), id)
		os.MkdirAll(filepath.Dir(logPath), 0755)
		os.WriteFile(logPath, []byte("# Task "+id), 0644)
	}

	result, err := implLogCheckResult(tmp, specName, "1.1,1.2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["all_present"] != true {
		t.Error("expected all_present=true")
	}
	if _, hasMissing := result["missing"]; hasMissing {
		t.Error("should not have missing key when all present")
	}
}

func TestImplLogCheckSomeMissing(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"

	// Create impl log only for task 1.1
	logPath := specdir.ImplLogPath(specdir.SpecDirAbs(tmp, specName), "1.1")
	os.MkdirAll(filepath.Dir(logPath), 0755)
	os.WriteFile(logPath, []byte("# Task 1.1"), 0644)

	result, err := implLogCheckResult(tmp, specName, "1.1,2.1,3.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["all_present"] != false {
		t.Error("expected all_present=false")
	}

	missing, ok := result["missing"].([]string)
	if !ok {
		t.Fatal("missing key should be a string slice")
	}
	if len(missing) != 2 {
		t.Errorf("expected 2 missing tasks, got %d", len(missing))
	}

	// Verify tasks map has correct entries
	tasks, _ := result["tasks"].(map[string]any)
	task11, _ := tasks["1.1"].(map[string]any)
	if task11["exists"] != true {
		t.Error("task 1.1 should exist")
	}
	task21, _ := tasks["2.1"].(map[string]any)
	if task21["exists"] != false {
		t.Error("task 2.1 should not exist")
	}
}

func TestImplLogCheckAllMissing(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"

	result, err := implLogCheckResult(tmp, specName, "5.1,5.2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["all_present"] != false {
		t.Error("expected all_present=false")
	}

	missing, ok := result["missing"].([]string)
	if !ok {
		t.Fatal("missing key should be a string slice")
	}
	if len(missing) != 2 {
		t.Errorf("expected 2 missing tasks, got %d", len(missing))
	}
}
