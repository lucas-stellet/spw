package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDispatchInitAuditInlineAudit(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	result, err := dispatchInitAuditCore(tmp, runDir, "inline-audit", 2)
	if err != nil {
		t.Fatal(err)
	}

	if !result.OK {
		t.Error("expected ok=true")
	}
	if result.Type != "inline-audit" {
		t.Errorf("type = %q, want %q", result.Type, "inline-audit")
	}
	if result.Iteration != 2 {
		t.Errorf("iteration = %d, want 2", result.Iteration)
	}

	// Verify directory was created
	expectedDir := filepath.Join(runDir, "_inline-audit", "iteration-2")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestDispatchInitAuditInlineCheckpoint(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	result, err := dispatchInitAuditCore(tmp, runDir, "inline-checkpoint", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Type != "inline-checkpoint" {
		t.Errorf("type = %q, want %q", result.Type, "inline-checkpoint")
	}

	// Verify directory was created (no iteration subdir for checkpoint)
	expectedDir := filepath.Join(runDir, "_inline-checkpoint")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestDispatchInitAuditDefaultIteration(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	result, err := dispatchInitAuditCore(tmp, runDir, "inline-audit", 0)
	if err != nil {
		t.Fatal(err)
	}

	if result.Iteration != 1 {
		t.Errorf("iteration = %d, want 1 (default)", result.Iteration)
	}

	expectedDir := filepath.Join(runDir, "_inline-audit", "iteration-1")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestDispatchInitAuditRelativePath(t *testing.T) {
	tmp := t.TempDir()
	// Create a subdirectory that the relative path will resolve against
	absRunDir := filepath.Join(tmp, "some", "run-001")
	os.MkdirAll(absRunDir, 0755)

	result, err := dispatchInitAuditCore(tmp, "some/run-001", "inline-audit", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result.AuditDir != filepath.Join("some", "run-001", "_inline-audit", "iteration-1") {
		t.Errorf("audit_dir = %q, want relative path", result.AuditDir)
	}

	// Verify the directory was actually created at the absolute path
	expectedDir := filepath.Join(tmp, "some", "run-001", "_inline-audit", "iteration-1")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestDispatchInitAuditValidation(t *testing.T) {
	tmp := t.TempDir()

	// Missing runDir
	_, err := dispatchInitAuditCore(tmp, "", "inline-audit", 1)
	if err == nil {
		t.Error("expected error for empty runDir")
	}

	// Missing auditType
	_, err = dispatchInitAuditCore(tmp, filepath.Join(tmp, "run-001"), "", 1)
	if err == nil {
		t.Error("expected error for empty auditType")
	}

	// Invalid auditType
	_, err = dispatchInitAuditCore(tmp, filepath.Join(tmp, "run-001"), "invalid", 1)
	if err == nil {
		t.Error("expected error for invalid auditType")
	}
}
