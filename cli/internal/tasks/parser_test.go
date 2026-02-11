package tasks

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func fixturesDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "fixtures", "tasks")
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixturesDir(), name))
	if err != nil {
		t.Fatalf("cannot read fixture %s: %v", name, err)
	}
	return string(data)
}

func TestParseBasic(t *testing.T) {
	doc := Parse(readFixture(t, "basic.md"))

	// Frontmatter
	if doc.Frontmatter.Spec != "test-feature" {
		t.Errorf("spec = %q, want test-feature", doc.Frontmatter.Spec)
	}
	if len(doc.Frontmatter.TaskIDs) != 4 {
		t.Errorf("task_ids count = %d, want 4", len(doc.Frontmatter.TaskIDs))
	}
	if doc.Frontmatter.ApprovalID != "abc-123" {
		t.Errorf("approval_id = %q, want abc-123", doc.Frontmatter.ApprovalID)
	}
	if doc.Frontmatter.GenerationStrategy != "rolling-wave" {
		t.Errorf("generation_strategy = %q, want rolling-wave", doc.Frontmatter.GenerationStrategy)
	}

	// Tasks
	if len(doc.Tasks) != 4 {
		t.Fatalf("tasks count = %d, want 4", len(doc.Tasks))
	}

	// Task 1
	t1 := doc.Tasks[0]
	if t1.ID != "1" || t1.Status != "done" || t1.Wave != 1 {
		t.Errorf("task 1: id=%q status=%q wave=%d", t1.ID, t1.Status, t1.Wave)
	}
	if t1.TDD != "yes" {
		t.Errorf("task 1 TDD = %q, want yes", t1.TDD)
	}

	// Task 3
	t3 := doc.Tasks[2]
	if t3.ID != "3" || t3.Status != "pending" || t3.Wave != 2 {
		t.Errorf("task 3: id=%q status=%q wave=%d", t3.ID, t3.Status, t3.Wave)
	}
	if len(t3.DependsOn) != 1 || t3.DependsOn[0] != "2" {
		t.Errorf("task 3 deps = %v, want [2]", t3.DependsOn)
	}

	// Wave Plan
	if len(doc.WavePlan) != 2 {
		t.Errorf("wave plan entries = %d, want 2", len(doc.WavePlan))
	}

	// Constraints
	if doc.Constraints == "" {
		t.Error("expected non-empty constraints")
	}

	// No warnings expected
	if len(doc.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", doc.Warnings)
	}
}

func TestParseDeferredReady(t *testing.T) {
	doc := Parse(readFixture(t, "deferred-ready.md"))

	if len(doc.Tasks) != 5 {
		t.Fatalf("tasks count = %d, want 5", len(doc.Tasks))
	}

	// Task 5 should be deferred (under Deferred Backlog heading)
	t5 := doc.TaskByID("5")
	if t5 == nil {
		t.Fatal("task 5 not found")
	}
	if !t5.IsDeferred {
		t.Error("task 5 should be deferred (under Deferred Backlog)")
	}
	if t5.Wave != 3 {
		t.Errorf("task 5 wave = %d, want 3", t5.Wave)
	}
	if len(t5.DependsOn) != 1 || t5.DependsOn[0] != "4" {
		t.Errorf("task 5 deps = %v, want [4]", t5.DependsOn)
	}

	// Task 5 should ALSO trigger task_ids_mismatch warning
	// because frontmatter has [1,2,3,4] but body has task 5
	found := false
	for _, w := range doc.Warnings {
		if containsStr(w, "task_ids_mismatch") && containsStr(w, "5") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected task_ids_mismatch warning for task 5, got: %v", doc.Warnings)
	}

	if !doc.HasDeferred {
		t.Error("expected HasDeferred = true")
	}
}

// A5: task_ids mismatch validation — body-scan > frontmatter
func TestParseTaskIDsMismatch(t *testing.T) {
	doc := Parse(readFixture(t, "deferred-ready.md"))

	// Frontmatter says [1,2,3,4] but body has 5 tasks
	if len(doc.Frontmatter.TaskIDs) != 4 {
		t.Errorf("frontmatter task_ids = %d, want 4", len(doc.Frontmatter.TaskIDs))
	}
	if len(doc.Tasks) != 5 {
		t.Errorf("body tasks = %d, want 5 (body scan is authoritative)", len(doc.Tasks))
	}

	// The parser should detect the mismatch
	hasMismatch := false
	for _, w := range doc.Warnings {
		if containsStr(w, "task_ids_mismatch") {
			hasMismatch = true
		}
	}
	if !hasMismatch {
		t.Error("expected task_ids_mismatch warning")
	}
}

func TestParseInProgress(t *testing.T) {
	doc := Parse(readFixture(t, "in-progress.md"))

	t3 := doc.TaskByID("3")
	if t3 == nil {
		t.Fatal("task 3 not found")
	}
	if t3.Status != "in_progress" {
		t.Errorf("task 3 status = %q, want in_progress", t3.Status)
	}
}

func TestParseAllDone(t *testing.T) {
	doc := Parse(readFixture(t, "all-done.md"))

	counts := doc.Count()
	if counts.Done != 3 || counts.Pending != 0 || counts.InProgress != 0 {
		t.Errorf("counts = %+v, want all done", counts)
	}
}

func TestParsePreservesLineNumbers(t *testing.T) {
	doc := Parse(readFixture(t, "basic.md"))

	for _, task := range doc.Tasks {
		if task.RawLine <= 0 {
			t.Errorf("task %s has invalid line number %d", task.ID, task.RawLine)
		}
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	content := `# Tasks: no-frontmatter

## Tasks

- [ ] 1 First task
  Wave: 1
  Files: ` + "`first.ts`" + `

- [ ] 2 Second task
  Wave: 1
  Depends On: Task 1
  Files: ` + "`second.ts`" + `
`
	doc := Parse(content)

	if len(doc.Tasks) != 2 {
		t.Fatalf("tasks = %d, want 2", len(doc.Tasks))
	}
	if doc.Frontmatter.Spec != "" {
		t.Errorf("expected empty spec, got %q", doc.Frontmatter.Spec)
	}
	// No mismatch warnings when no frontmatter task_ids
	if len(doc.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", doc.Warnings)
	}
}

func TestParseMultipleDeps(t *testing.T) {
	doc := Parse(readFixture(t, "mixed-deps.md"))

	t4 := doc.TaskByID("4")
	if t4 == nil {
		t.Fatal("task 4 not found")
	}
	if len(t4.DependsOn) != 2 {
		t.Errorf("task 4 deps = %v, want 2 deps", t4.DependsOn)
	}
}

func TestCount(t *testing.T) {
	doc := Parse(readFixture(t, "in-progress.md"))
	counts := doc.Count()

	if counts.Total != 4 {
		t.Errorf("total = %d, want 4", counts.Total)
	}
	if counts.Done != 2 {
		t.Errorf("done = %d, want 2", counts.Done)
	}
	if counts.InProgress != 1 {
		t.Errorf("in_progress = %d, want 1", counts.InProgress)
	}
	if counts.Pending != 1 {
		t.Errorf("pending = %d, want 1", counts.Pending)
	}
}

func TestTaskByID(t *testing.T) {
	doc := Parse(readFixture(t, "basic.md"))

	if doc.TaskByID("1") == nil {
		t.Error("task 1 should exist")
	}
	if doc.TaskByID("99") != nil {
		t.Error("task 99 should not exist")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
