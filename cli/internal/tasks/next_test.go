package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- helpers ---

// mkCheckpointDir creates a wave checkpoint directory with status.json.
func mkCheckpointDir(t *testing.T, specDir string, waveNum int, runID string, status string) {
	t.Helper()
	cpDir := filepath.Join(specDir, fmt.Sprintf("execution/waves/wave-%02d/checkpoint/%s/release-gate-decider", waveNum, runID))
	if err := os.MkdirAll(cpDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(cpDir, "status.json"), map[string]string{"status": status})
}

// mkWaveSummary creates a _wave-summary.json for a wave.
func mkWaveSummary(t *testing.T, specDir string, waveNum int, status string) {
	t.Helper()
	waveDir := filepath.Join(specDir, fmt.Sprintf("execution/waves/wave-%02d", waveNum))
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(waveDir, "_wave-summary.json"), map[string]string{"status": status})
}

// mkLatestJSON creates a _latest.json for a wave.
func mkLatestJSON(t *testing.T, specDir string, waveNum int, runID string) {
	t.Helper()
	waveDir := filepath.Join(specDir, fmt.Sprintf("execution/waves/wave-%02d", waveNum))
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(waveDir, "_latest.json"), map[string]string{"run_id": runID})
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// --- tests ---

// A1: All tasks done, no deferred -> action: "done" (all-at-once)
func TestResolveNextWave_A1_AllDone(t *testing.T) {
	doc := Parse(readFixture(t, "all-done.md"))
	specDir := t.TempDir()

	// Create checkpoint dirs for waves to simulate completed state
	mkCheckpointDir(t, specDir, 1, "run-001", "pass")
	mkWaveSummary(t, specDir, 1, "pass")
	mkCheckpointDir(t, specDir, 2, "run-001", "pass")
	mkWaveSummary(t, specDir, 2, "pass")

	result := ResolveNextWave(doc, specDir)

	if result.Action != "done" {
		t.Errorf("action = %q, want done", result.Action)
	}
	if len(result.DeferredReady) != 0 {
		t.Errorf("DeferredReady = %v, want empty", result.DeferredReady)
	}
}

// A2: Task 5 deferred, deps resolved (THE BUG) -> action: "execute", DeferredReady: ["5"]
func TestResolveNextWave_A2_DeferredReady(t *testing.T) {
	doc := Parse(readFixture(t, "deferred-ready.md"))
	specDir := t.TempDir()

	// All regular tasks (1-4) are done; wave 1+2 checkpoints pass
	mkCheckpointDir(t, specDir, 1, "run-001", "pass")
	mkWaveSummary(t, specDir, 1, "pass")
	mkCheckpointDir(t, specDir, 2, "run-001", "pass")
	mkWaveSummary(t, specDir, 2, "pass")

	result := ResolveNextWave(doc, specDir)

	if result.Action != "execute" {
		t.Errorf("action = %q, want execute", result.Action)
	}
	if !strSliceContains(result.DeferredReady, "5") {
		t.Errorf("DeferredReady = %v, want [5]", result.DeferredReady)
	}
}

// A3: Task 5 deferred, deps NOT resolved -> action: "execute" (wave 2 tasks), no DeferredReady
func TestResolveNextWave_A3_DeferredBlocked(t *testing.T) {
	doc := Parse(readFixture(t, "deferred-blocked.md"))
	specDir := t.TempDir()

	// Wave 1 checkpoint passes, wave 2 not yet started
	mkCheckpointDir(t, specDir, 1, "run-001", "pass")
	mkWaveSummary(t, specDir, 1, "pass")

	result := ResolveNextWave(doc, specDir)

	if result.Action != "execute" {
		t.Errorf("action = %q, want execute", result.Action)
	}
	// Task 3 (wave 2) should be executable; its dep (task 2) is done
	if !strSliceContains(result.TaskIDs, "3") {
		t.Errorf("TaskIDs = %v, want [3]", result.TaskIDs)
	}
	// Deferred task 5 depends on task 3 which is still pending
	if len(result.DeferredReady) != 0 {
		t.Errorf("DeferredReady = %v, want empty (task 5 dep on task 3 not resolved)", result.DeferredReady)
	}
}

// A4: Task 3 in-progress [-] -> action: "continue-wave"
func TestResolveNextWave_A4_InProgress(t *testing.T) {
	doc := Parse(readFixture(t, "in-progress.md"))
	specDir := t.TempDir()

	result := ResolveNextWave(doc, specDir)

	if result.Action != "continue-wave" {
		t.Errorf("action = %q, want continue-wave", result.Action)
	}
	if result.Wave != 2 {
		t.Errorf("wave = %d, want 2", result.Wave)
	}
}

// A5: task_ids mismatch validation -> warnings present, NextWave still works
func TestResolveNextWave_A5_MismatchWarnings(t *testing.T) {
	doc := Parse(readFixture(t, "deferred-ready.md"))
	specDir := t.TempDir()

	mkCheckpointDir(t, specDir, 1, "run-001", "pass")
	mkWaveSummary(t, specDir, 1, "pass")
	mkCheckpointDir(t, specDir, 2, "run-001", "pass")
	mkWaveSummary(t, specDir, 2, "pass")

	result := ResolveNextWave(doc, specDir)

	// Should still produce a valid action despite mismatch
	if result.Action != "execute" {
		t.Errorf("action = %q, want execute (NextWave should work despite mismatch)", result.Action)
	}

	// Warnings should include the mismatch from parser
	hasMismatch := false
	for _, w := range result.Warnings {
		if containsStr(w, "task_ids_mismatch") {
			hasMismatch = true
			break
		}
	}
	if !hasMismatch {
		t.Errorf("warnings = %v, expected task_ids_mismatch warning", result.Warnings)
	}
}

// A6: Previous wave checkpoint "blocked" -> action: "blocked"
func TestResolveNextWave_A6_CheckpointBlocked(t *testing.T) {
	// Use basic.md: tasks 1,2 done (wave 1), tasks 3,4 pending (wave 2)
	doc := Parse(readFixture(t, "basic.md"))
	specDir := t.TempDir()

	// Wave 1 checkpoint is blocked
	mkCheckpointDir(t, specDir, 1, "run-001", "blocked")
	mkWaveSummary(t, specDir, 1, "blocked")

	result := ResolveNextWave(doc, specDir)

	if result.Action != "blocked" {
		t.Errorf("action = %q, want blocked", result.Action)
	}
	if result.Wave != 1 {
		t.Errorf("wave = %d, want 1 (the blocked wave)", result.Wave)
	}
}

// A7: Previous wave checkpoint "pass" -> action: "execute" (next wave tasks)
func TestResolveNextWave_A7_CheckpointPass(t *testing.T) {
	doc := Parse(readFixture(t, "basic.md"))
	specDir := t.TempDir()

	// Wave 1 checkpoint passes
	mkCheckpointDir(t, specDir, 1, "run-001", "pass")
	mkWaveSummary(t, specDir, 1, "pass")

	result := ResolveNextWave(doc, specDir)

	if result.Action != "execute" {
		t.Errorf("action = %q, want execute", result.Action)
	}
	if result.Wave != 2 {
		t.Errorf("wave = %d, want 2", result.Wave)
	}
	// Task 3 depends on task 2 (done), so it should be executable
	if !strSliceContains(result.TaskIDs, "3") {
		t.Errorf("TaskIDs = %v, want to include 3", result.TaskIDs)
	}
	// Task 4 depends on task 3 (still pending), so NOT executable
	if strSliceContains(result.TaskIDs, "4") {
		t.Errorf("TaskIDs = %v, should NOT include 4 (depends on pending task 3)", result.TaskIDs)
	}
}

// A7 variant: Stale summary detection — _latest.json overrides _wave-summary.json
func TestResolveNextWave_A7_StaleSummary(t *testing.T) {
	doc := Parse(readFixture(t, "basic.md"))
	specDir := t.TempDir()

	// Wave 1: _wave-summary says "blocked" (stale!)
	// but _latest.json points to run-003 which says "pass"
	mkWaveSummary(t, specDir, 1, "blocked")
	mkLatestJSON(t, specDir, 1, "run-003")
	mkCheckpointDir(t, specDir, 1, "run-003", "pass")

	result := ResolveNextWave(doc, specDir)

	// Should proceed because latest checkpoint says "pass"
	if result.Action != "execute" {
		t.Errorf("action = %q, want execute (latest run overrides stale summary)", result.Action)
	}

	// Should have stale_summary warning
	hasStale := false
	for _, w := range result.Warnings {
		if containsStr(w, "stale_summary") {
			hasStale = true
			break
		}
	}
	if !hasStale {
		t.Errorf("warnings = %v, expected stale_summary warning", result.Warnings)
	}
}

// A8: No tasks (empty doc) -> action: "done"
func TestResolveNextWave_A8_EmptyDoc(t *testing.T) {
	doc := Parse("")
	result := ResolveNextWave(doc, "")

	if result.Action != "done" {
		t.Errorf("action = %q, want done", result.Action)
	}
}

// A9: All done, rolling-wave, no deferred -> action: "plan-next-wave"
func TestResolveNextWave_A9_RollingWavePlanNext(t *testing.T) {
	doc := Parse(readFixture(t, "all-done-rolling.md"))
	specDir := t.TempDir()

	// Wave 1 checkpoint passes
	mkCheckpointDir(t, specDir, 1, "run-001", "pass")
	mkWaveSummary(t, specDir, 1, "pass")

	result := ResolveNextWave(doc, specDir)

	if result.Action != "plan-next-wave" {
		t.Errorf("action = %q, want plan-next-wave", result.Action)
	}
}

// A10: Multiple tasks in wave, mixed deps -> only tasks with resolved deps included
func TestResolveNextWave_A10_MixedDeps(t *testing.T) {
	doc := Parse(readFixture(t, "mixed-deps.md"))
	specDir := t.TempDir()

	// Wave 1 checkpoint passes (tasks 1, 2 done)
	mkCheckpointDir(t, specDir, 1, "run-001", "pass")
	mkWaveSummary(t, specDir, 1, "pass")

	result := ResolveNextWave(doc, specDir)

	if result.Action != "execute" {
		t.Errorf("action = %q, want execute", result.Action)
	}
	if result.Wave != 2 {
		t.Errorf("wave = %d, want 2", result.Wave)
	}

	// Task 3: depends on Task 1 (done) -> INCLUDED
	if !strSliceContains(result.TaskIDs, "3") {
		t.Errorf("TaskIDs = %v, want to include 3 (dep on done task 1)", result.TaskIDs)
	}

	// Task 4: depends on Task 1 + Task 2 (both done) -> INCLUDED
	if !strSliceContains(result.TaskIDs, "4") {
		t.Errorf("TaskIDs = %v, want to include 4 (deps on done tasks 1,2)", result.TaskIDs)
	}

	// Task 5: depends on Task 3 (pending) -> NOT INCLUDED
	if strSliceContains(result.TaskIDs, "5") {
		t.Errorf("TaskIDs = %v, should NOT include 5 (dep on pending task 3)", result.TaskIDs)
	}
}
