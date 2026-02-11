package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- helpers ---

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// --- Category E: Prerequisite tests ---

func TestE1_CheckPrereqsExecWithTasksMD(t *testing.T) {
	// E1: CheckPrereqs for exec with tasks.md present -> Ready: true
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "tasks.md"), "# Tasks\n- [ ] 1. Do something\n")

	result := CheckPrereqs(specDir, "exec")

	if !result.Ready {
		t.Errorf("expected Ready true, got false; missing: %v", result.Missing)
	}
}

func TestE2_CheckPrereqsExecWithoutTasksMD(t *testing.T) {
	// E2: CheckPrereqs for exec with tasks.md missing -> Ready: false
	specDir := t.TempDir()

	result := CheckPrereqs(specDir, "exec")

	if result.Ready {
		t.Error("expected Ready false, got true")
	}
	if len(result.Missing) != 1 || result.Missing[0] != "tasks.md" {
		t.Errorf("expected missing [tasks.md], got %v", result.Missing)
	}
}

// --- Category F: Stage classification tests ---

func TestF1_ClassifyStageRequirementsOnly(t *testing.T) {
	// F1: ClassifyStage with only requirements.md -> "requirements"
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")

	stage := ClassifyStage(specDir)

	if stage != "requirements" {
		t.Errorf("expected 'requirements', got %q", stage)
	}
}

func TestF2_ClassifyStageExecution(t *testing.T) {
	// F2: ClassifyStage with execution/waves/ -> "execution"
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")
	writeFile(t, filepath.Join(specDir, "design.md"), "# Design\n")
	writeFile(t, filepath.Join(specDir, "tasks.md"), "# Tasks\n- [ ] 1. Do something\n")
	mkdirAll(t, filepath.Join(specDir, "execution", "waves", "wave-01"))

	stage := ClassifyStage(specDir)

	if stage != "execution" {
		t.Errorf("expected 'execution', got %q", stage)
	}
}

func TestF3_ListSpecs(t *testing.T) {
	// F3: List specs -> returns sorted spec names
	cwd := t.TempDir()
	specsDir := filepath.Join(cwd, ".spec-workflow", "specs")

	// Create spec directories in non-alphabetical order
	mkdirAll(t, filepath.Join(specsDir, "zebra-feature"))
	mkdirAll(t, filepath.Join(specsDir, "alpha-feature"))
	mkdirAll(t, filepath.Join(specsDir, "mid-feature"))

	names, err := List(cwd)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(names))
	}
	if names[0] != "alpha-feature" {
		t.Errorf("expected first spec 'alpha-feature', got %q", names[0])
	}
	if names[1] != "mid-feature" {
		t.Errorf("expected second spec 'mid-feature', got %q", names[1])
	}
	if names[2] != "zebra-feature" {
		t.Errorf("expected third spec 'zebra-feature', got %q", names[2])
	}
}

// --- Additional tests ---

func TestCheckArtifactsDetectsDeviations(t *testing.T) {
	specDir := t.TempDir()

	// Create a known deviation path
	mkdirAll(t, filepath.Join(specDir, "implementation-logs"))

	// Create a canonical artifact
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")

	artifacts, deviations := CheckArtifacts(specDir)

	if !artifacts["requirements.md"] {
		t.Error("expected requirements.md to exist in artifacts")
	}

	if len(deviations) == 0 {
		t.Error("expected at least one deviation, got none")
	}

	found := false
	for _, d := range deviations {
		if d.Found == "implementation-logs" {
			found = true
			if d.Canonical != "execution/_implementation-logs" {
				t.Errorf("expected canonical 'execution/_implementation-logs', got %q", d.Canonical)
			}
		}
	}
	if !found {
		t.Error("expected deviation for 'implementation-logs', not found")
	}
}

func TestCheckApprovalNoFiles(t *testing.T) {
	cwd := t.TempDir()

	result := CheckApproval(cwd, "my-spec", "requirements")

	if result.Found {
		t.Error("expected Found false, got true")
	}
	if result.DocType != "requirements" {
		t.Errorf("expected doc_type 'requirements', got %q", result.DocType)
	}
}

func TestCheckApprovalWithFile(t *testing.T) {
	cwd := t.TempDir()
	specName := "my-spec"

	approvalsDir := filepath.Join(cwd, ".spec-workflow", "approvals", specName)
	writeJSON(t, filepath.Join(approvalsDir, "approval_001.json"), map[string]any{
		"approvalId": "appr-123",
		"filePath":   ".spec-workflow/specs/my-spec/requirements.md",
	})

	result := CheckApproval(cwd, specName, "requirements")

	if !result.Found {
		t.Error("expected Found true, got false")
	}
	if result.ApprovalID != "appr-123" {
		t.Errorf("expected approval_id 'appr-123', got %q", result.ApprovalID)
	}
}

func TestClassifyStageDesign(t *testing.T) {
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")
	writeFile(t, filepath.Join(specDir, "design.md"), "# Design\n")

	stage := ClassifyStage(specDir)

	if stage != "design" {
		t.Errorf("expected 'design', got %q", stage)
	}
}

func TestClassifyStageQA(t *testing.T) {
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")
	writeFile(t, filepath.Join(specDir, "design.md"), "# Design\n")
	writeFile(t, filepath.Join(specDir, "tasks.md"), "# Tasks\n- [ ] 1. Do something\n")
	writeFile(t, filepath.Join(specDir, "qa", "QA-TEST-PLAN.md"), "# QA Test Plan\n")

	stage := ClassifyStage(specDir)

	if stage != "qa" {
		t.Errorf("expected 'qa', got %q", stage)
	}
}

func TestClassifyStagePostMortem(t *testing.T) {
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")
	writeFile(t, filepath.Join(specDir, "post-mortem", "report.md"), "# Post-Mortem\n")

	stage := ClassifyStage(specDir)

	if stage != "post-mortem" {
		t.Errorf("expected 'post-mortem', got %q", stage)
	}
}

func TestClassifyStageUnknown(t *testing.T) {
	specDir := t.TempDir()

	stage := ClassifyStage(specDir)

	if stage != "unknown" {
		t.Errorf("expected 'unknown', got %q", stage)
	}
}

func TestCheckPrereqsDesignResearch(t *testing.T) {
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")

	result := CheckPrereqs(specDir, "design-research")

	if !result.Ready {
		t.Errorf("expected Ready true, got false; missing: %v", result.Missing)
	}
}

func TestCheckPrereqsDesignDraftMissing(t *testing.T) {
	specDir := t.TempDir()
	// Only requirements.md, missing design/DESIGN-RESEARCH.md
	writeFile(t, filepath.Join(specDir, "requirements.md"), "# Requirements\n")

	result := CheckPrereqs(specDir, "design-draft")

	if result.Ready {
		t.Error("expected Ready false, got true")
	}
	if len(result.Missing) != 1 || result.Missing[0] != "design/DESIGN-RESEARCH.md" {
		t.Errorf("expected missing [design/DESIGN-RESEARCH.md], got %v", result.Missing)
	}
}

func TestCheckPrereqsCheckpointNeedsWaves(t *testing.T) {
	specDir := t.TempDir()
	writeFile(t, filepath.Join(specDir, "tasks.md"), "# Tasks\n")
	// No execution/waves/ dir

	result := CheckPrereqs(specDir, "checkpoint")

	if result.Ready {
		t.Error("expected Ready false, got true")
	}
	found := false
	for _, m := range result.Missing {
		if m == "execution/waves" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'execution/waves' in missing list, got %v", result.Missing)
	}
}

func TestListSpecsEmptyDir(t *testing.T) {
	cwd := t.TempDir()

	names, err := List(cwd)
	if err != nil {
		t.Fatal(err)
	}
	if names != nil {
		t.Errorf("expected nil, got %v", names)
	}
}
