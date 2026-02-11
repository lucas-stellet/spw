package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mkImplLog creates an implementation log file for a task.
func mkImplLog(t *testing.T, specDir, taskID string) {
	t.Helper()
	dir := filepath.Join(specDir, "execution/_implementation-logs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, fmt.Sprintf("task-%s.md", taskID))
	if err := os.WriteFile(logPath, []byte("# Implementation log\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

// writeTasksFile writes content to a temp tasks.md and returns the path.
func writeTasksFile(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

const markTestContent = `---
spec: mark-test
task_ids: [1, 2, 3]
---

# Tasks: mark-test

## Tasks

- [x] 1 First task
  Wave: 1
  Files: ` + "`first.ts`" + `

- [ ] 2 Second task
  Wave: 1
  Depends On: Task 1
  Files: ` + "`second.ts`" + `

- [ ] 3 Third task
  Wave: 2
  Depends On: Task 2
  Files: ` + "`third.ts`" + `
`

// B1: Mark done, impl log exists -> success
func TestMarkTaskInFile_B1_DoneWithImplLog(t *testing.T) {
	dir := t.TempDir()
	specDir := dir
	filePath := writeTasksFile(t, dir, markTestContent)

	// Create impl log for task 2
	mkImplLog(t, specDir, "2")

	err := MarkTaskInFile(filePath, "2", "done", true, specDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the file was updated
	data, _ := os.ReadFile(filePath)
	content := string(data)
	if !strings.Contains(content, "- [x] 2 Second task") {
		t.Errorf("task 2 not marked as done in file:\n%s", content)
	}
}

// B2: Mark done, impl log MISSING + requireImplLog -> error
func TestMarkTaskInFile_B2_DoneMissingImplLog(t *testing.T) {
	dir := t.TempDir()
	specDir := dir
	filePath := writeTasksFile(t, dir, markTestContent)

	// Do NOT create impl log
	err := MarkTaskInFile(filePath, "2", "done", true, specDir)
	if err == nil {
		t.Fatal("expected error for missing impl log, got nil")
	}
	if !strings.Contains(err.Error(), "implementation log missing") {
		t.Errorf("error = %q, want to mention 'implementation log missing'", err.Error())
	}
}

// B3: Mark in-progress -> success
func TestMarkTaskInFile_B3_InProgress(t *testing.T) {
	dir := t.TempDir()
	filePath := writeTasksFile(t, dir, markTestContent)

	err := MarkTaskInFile(filePath, "2", "in_progress", false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filePath)
	content := string(data)
	if !strings.Contains(content, "- [-] 2 Second task") {
		t.Errorf("task 2 not marked as in_progress:\n%s", content)
	}
}

// B4: Verify file formatting preserved -> only checkbox char changed
func TestMarkTaskInFile_B4_PreserveFormatting(t *testing.T) {
	dir := t.TempDir()
	filePath := writeTasksFile(t, dir, markTestContent)

	// Read original content
	origData, _ := os.ReadFile(filePath)
	origLines := strings.Split(string(origData), "\n")

	// Mark task 2 as in-progress
	if err := MarkTaskInFile(filePath, "2", "in_progress", false, ""); err != nil {
		t.Fatal(err)
	}

	// Read updated content
	newData, _ := os.ReadFile(filePath)
	newLines := strings.Split(string(newData), "\n")

	if len(origLines) != len(newLines) {
		t.Fatalf("line count changed: %d -> %d", len(origLines), len(newLines))
	}

	changedCount := 0
	for i := range origLines {
		if origLines[i] != newLines[i] {
			changedCount++
			// The only change should be the checkbox character on the task 2 line
			if !strings.Contains(origLines[i], "- [ ] 2 Second task") {
				t.Errorf("unexpected change on line %d:\n  before: %q\n  after:  %q", i+1, origLines[i], newLines[i])
			}
			if !strings.Contains(newLines[i], "- [-] 2 Second task") {
				t.Errorf("line %d not correctly updated:\n  got: %q", i+1, newLines[i])
			}
		}
	}

	if changedCount != 1 {
		t.Errorf("expected exactly 1 line changed, got %d", changedCount)
	}
}

// B5: Mark already [x] (done->done) -> idempotent, no error
func TestMarkTaskInFile_B5_Idempotent(t *testing.T) {
	dir := t.TempDir()
	filePath := writeTasksFile(t, dir, markTestContent)

	// Task 1 is already [x]
	err := MarkTaskInFile(filePath, "1", "done", false, "")
	if err != nil {
		t.Fatalf("unexpected error for idempotent mark: %v", err)
	}

	// Verify the file is unchanged
	data, _ := os.ReadFile(filePath)
	content := string(data)
	if !strings.Contains(content, "- [x] 1 First task") {
		t.Errorf("task 1 should still be done:\n%s", content)
	}
}

// Test: Task not found returns error
func TestMarkTaskInFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	filePath := writeTasksFile(t, dir, markTestContent)

	err := MarkTaskInFile(filePath, "99", "done", false, "")
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !strings.Contains(err.Error(), "task 99 not found") {
		t.Errorf("error = %q, want 'task 99 not found'", err.Error())
	}
}
