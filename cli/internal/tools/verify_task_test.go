package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas-stellet/spw/internal/specdir"
)

func TestVerifyTaskImplLogExists(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"
	taskID := "1.1"

	// Create impl log file
	implLogPath := specdir.ImplLogPath(specdir.SpecDirAbs(tmp, specName), taskID)
	os.MkdirAll(filepath.Dir(implLogPath), 0755)
	os.WriteFile(implLogPath, []byte("# Task 1.1\n\nImplementation notes."), 0644)

	result, err := verifyTaskResult(tmp, specName, taskID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["task_id"] != taskID {
		t.Errorf("task_id = %q, want %q", result["task_id"], taskID)
	}

	implLog, ok := result["impl_log"].(map[string]any)
	if !ok {
		t.Fatal("impl_log missing or wrong type")
	}
	if implLog["exists"] != true {
		t.Error("expected impl_log.exists=true")
	}
	path, _ := implLog["path"].(string)
	if path == "" {
		t.Error("expected non-empty impl_log.path")
	}
}

func TestVerifyTaskImplLogMissing(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"
	taskID := "2.3"

	// Do NOT create impl log file — spec dir may or may not exist
	result, err := verifyTaskResult(tmp, specName, taskID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}

	implLog, ok := result["impl_log"].(map[string]any)
	if !ok {
		t.Fatal("impl_log missing or wrong type")
	}
	if implLog["exists"] != false {
		t.Error("expected impl_log.exists=false")
	}
}

func TestVerifyTaskWithCommitCheck(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"
	taskID := "3.1"

	// No git repo in tmp, so commit check should return exists=false
	result, err := verifyTaskResult(tmp, specName, taskID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commit, ok := result["commit"].(map[string]any)
	if !ok {
		t.Fatal("commit missing or wrong type when checkCommit=true")
	}
	if commit["exists"] != false {
		t.Error("expected commit.exists=false in non-git dir")
	}
}

func TestVerifyTaskWithoutCommitCheck(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"
	taskID := "4.1"

	result, err := verifyTaskResult(tmp, specName, taskID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result["commit"]; ok {
		t.Error("commit key should not be present when checkCommit=false")
	}
}
