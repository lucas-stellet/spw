package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setupSpecDir(t *testing.T, specName string) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, ".spec-workflow", "specs", specName)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	return tmp, specDir
}

func writeTasksMD(t *testing.T, specDir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func createWaveDir(t *testing.T, specDir string, waveNum int) string {
	t.Helper()
	waveDir := filepath.Join(specDir, "execution", "waves", fmt.Sprintf("wave-%02d", waveNum))
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatal(err)
	}
	return waveDir
}

func writeWaveSummary(t *testing.T, waveDir string, summary map[string]any) {
	t.Helper()
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(waveDir, "_wave-summary.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func writeWaveLatest(t *testing.T, waveDir string, latest map[string]any) {
	t.Helper()
	data, err := json.MarshalIndent(latest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(waveDir, "_latest.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestWaveStatusNoWaves(t *testing.T) {
	cwd, specDir := setupSpecDir(t, "test-spec")

	writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [ ] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
- [ ] **2.** Implement core logic
  - Files: `+"`src/core.go`"+`
  - Wave: 01
`)

	result, err := waveStatusResult(cwd, "test-spec")
	if err != nil {
		t.Fatal(err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["spec"] != "test-spec" {
		t.Errorf("spec = %v, want test-spec", result["spec"])
	}
	if result["current_wave"] != "01" {
		t.Errorf("current_wave = %v, want 01", result["current_wave"])
	}
	if result["wave_status"] != "not-started" {
		t.Errorf("wave_status = %v, want not-started", result["wave_status"])
	}
	if result["resume_action"] != "start-wave" {
		t.Errorf("resume_action = %v, want start-wave", result["resume_action"])
	}
	if result["total_tasks"] != 2 {
		t.Errorf("total_tasks = %v, want 2", result["total_tasks"])
	}
	if result["completed_tasks"] != 0 {
		t.Errorf("completed_tasks = %v, want 0", result["completed_tasks"])
	}

	waves, ok := result["waves"].([]waveInfo)
	if !ok || len(waves) != 0 {
		t.Errorf("expected empty waves, got %v", result["waves"])
	}

	nextTasks, ok := result["next_tasks"].([]int)
	if !ok {
		t.Fatalf("next_tasks wrong type: %T", result["next_tasks"])
	}
	if len(nextTasks) != 2 || nextTasks[0] != 1 || nextTasks[1] != 2 {
		t.Errorf("next_tasks = %v, want [1 2]", nextTasks)
	}
}

func TestWaveStatusOneCompletedWave(t *testing.T) {
	cwd, specDir := setupSpecDir(t, "test-spec")

	writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [x] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
- [x] **2.** Implement core logic
  - Files: `+"`src/core.go`"+`
  - Wave: 01

## Wave 02

- [ ] **3.** Add validation
  - Files: `+"`src/validate.go`"+`
  - Wave: 02
- [ ] **4.** Error handling
  - Files: `+"`src/errors.go`"+`
  - Wave: 02
`)

	waveDir := createWaveDir(t, specDir, 1)
	writeWaveSummary(t, waveDir, map[string]any{
		"wave":   "01",
		"status": "pass",
		"tasks":  []int{1, 2},
	})
	writeWaveLatest(t, waveDir, map[string]any{
		"status": "pass",
	})

	result, err := waveStatusResult(cwd, "test-spec")
	if err != nil {
		t.Fatal(err)
	}

	if result["current_wave"] != "02" {
		t.Errorf("current_wave = %v, want 02", result["current_wave"])
	}
	if result["wave_status"] != "not-started" {
		t.Errorf("wave_status = %v, want not-started", result["wave_status"])
	}
	if result["resume_action"] != "start-wave" {
		t.Errorf("resume_action = %v, want start-wave", result["resume_action"])
	}
	if result["completed_tasks"] != 2 {
		t.Errorf("completed_tasks = %v, want 2", result["completed_tasks"])
	}
	if result["total_tasks"] != 4 {
		t.Errorf("total_tasks = %v, want 4", result["total_tasks"])
	}

	waves := result["waves"].([]waveInfo)
	if len(waves) != 1 {
		t.Fatalf("expected 1 wave, got %d", len(waves))
	}
	if waves[0].Checkpoint != "pass" {
		t.Errorf("wave 01 checkpoint = %v, want pass", waves[0].Checkpoint)
	}

	nextTasks := result["next_tasks"].([]int)
	if len(nextTasks) != 2 || nextTasks[0] != 3 || nextTasks[1] != 4 {
		t.Errorf("next_tasks = %v, want [3 4]", nextTasks)
	}
}

func TestWaveStatusInProgressWave(t *testing.T) {
	cwd, specDir := setupSpecDir(t, "test-spec")

	writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [x] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
- [x] **2.** Implement core logic
  - Files: `+"`src/core.go`"+`
  - Wave: 01

## Wave 02

- [-] **3.** Add validation
  - Files: `+"`src/validate.go`"+`
  - Wave: 02
- [ ] **4.** Error handling
  - Files: `+"`src/errors.go`"+`
  - Wave: 02
`)

	// Wave 01 completed
	wave1Dir := createWaveDir(t, specDir, 1)
	writeWaveSummary(t, wave1Dir, map[string]any{
		"wave":   "01",
		"status": "pass",
		"tasks":  []int{1, 2},
	})
	writeWaveLatest(t, wave1Dir, map[string]any{
		"status": "pass",
	})

	// Wave 02 in progress (no summary, no latest)
	createWaveDir(t, specDir, 2)

	result, err := waveStatusResult(cwd, "test-spec")
	if err != nil {
		t.Fatal(err)
	}

	if result["current_wave"] != "02" {
		t.Errorf("current_wave = %v, want 02", result["current_wave"])
	}
	if result["wave_status"] != "in-progress" {
		t.Errorf("wave_status = %v, want in-progress", result["wave_status"])
	}
	if result["resume_action"] != "resume-in-progress" {
		t.Errorf("resume_action = %v, want resume-in-progress", result["resume_action"])
	}

	inProgress := result["in_progress_tasks"].([]int)
	if len(inProgress) != 1 || inProgress[0] != 3 {
		t.Errorf("in_progress_tasks = %v, want [3]", inProgress)
	}
}

func TestWaveStatusResumeActions(t *testing.T) {
	t.Run("done", func(t *testing.T) {
		cwd, specDir := setupSpecDir(t, "test-spec")

		writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [x] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
`)

		waveDir := createWaveDir(t, specDir, 1)
		writeWaveSummary(t, waveDir, map[string]any{
			"wave":   "01",
			"status": "pass",
			"tasks":  []int{1},
		})
		writeWaveLatest(t, waveDir, map[string]any{
			"status": "pass",
		})

		result, err := waveStatusResult(cwd, "test-spec")
		if err != nil {
			t.Fatal(err)
		}
		if result["resume_action"] != "done" {
			t.Errorf("resume_action = %v, want done", result["resume_action"])
		}
	})

	t.Run("wait-checkpoint", func(t *testing.T) {
		cwd, specDir := setupSpecDir(t, "test-spec")

		writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [-] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
`)

		waveDir := createWaveDir(t, specDir, 1)
		writeWaveSummary(t, waveDir, map[string]any{
			"wave":   "01",
			"status": "blocked",
			"tasks":  []int{1},
		})
		writeWaveLatest(t, waveDir, map[string]any{
			"status": "blocked",
		})

		result, err := waveStatusResult(cwd, "test-spec")
		if err != nil {
			t.Fatal(err)
		}
		if result["resume_action"] != "wait-checkpoint" {
			t.Errorf("resume_action = %v, want wait-checkpoint", result["resume_action"])
		}
	})

	t.Run("continue-wave", func(t *testing.T) {
		cwd, specDir := setupSpecDir(t, "test-spec")

		writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [ ] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
`)

		// Wave dir exists but no checkpoint and no in-progress tasks
		createWaveDir(t, specDir, 1)

		result, err := waveStatusResult(cwd, "test-spec")
		if err != nil {
			t.Fatal(err)
		}
		if result["resume_action"] != "continue-wave" {
			t.Errorf("resume_action = %v, want continue-wave", result["resume_action"])
		}
	})

	t.Run("start-wave", func(t *testing.T) {
		cwd, specDir := setupSpecDir(t, "test-spec")

		writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [ ] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
`)

		// No wave dirs at all
		result, err := waveStatusResult(cwd, "test-spec")
		if err != nil {
			t.Fatal(err)
		}
		if result["resume_action"] != "start-wave" {
			t.Errorf("resume_action = %v, want start-wave", result["resume_action"])
		}
	})
}

func TestWaveStatusTasksParsing(t *testing.T) {
	cwd, specDir := setupSpecDir(t, "test-spec")

	writeTasksMD(t, specDir, `# Tasks

## Wave 01

- [x] **1.** Setup project structure
  - Files: `+"`src/main.go`"+`
  - Wave: 01
  - TDD: inherit
- [x] **2.** Implement core logic
  - Files: `+"`src/core.go`"+`
  - Wave: 01

## Wave 02

- [-] **3.** Add validation
  - Files: `+"`src/validate.go`"+`
  - Wave: 02
- [ ] **4.** Error handling
  - Files: `+"`src/errors.go`"+`
  - Wave: 02
- [!] **5.** Blocked task
  - Files: `+"`src/blocked.go`"+`
  - Wave: 02
`)

	result, err := waveStatusResult(cwd, "test-spec")
	if err != nil {
		t.Fatal(err)
	}

	if result["total_tasks"] != 5 {
		t.Errorf("total_tasks = %v, want 5", result["total_tasks"])
	}
	if result["completed_tasks"] != 2 {
		t.Errorf("completed_tasks = %v, want 2", result["completed_tasks"])
	}

	inProgress := result["in_progress_tasks"].([]int)
	if len(inProgress) != 1 || inProgress[0] != 3 {
		t.Errorf("in_progress_tasks = %v, want [3]", inProgress)
	}
}

func TestWaveStatusSpecNotFound(t *testing.T) {
	tmp := t.TempDir()
	_, err := waveStatusResult(tmp, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent spec")
	}
}

func TestWaveStatusNoTasksMD(t *testing.T) {
	cwd, _ := setupSpecDir(t, "test-spec")

	// No tasks.md, no waves — should still work
	result, err := waveStatusResult(cwd, "test-spec")
	if err != nil {
		t.Fatal(err)
	}

	if result["total_tasks"] != 0 {
		t.Errorf("total_tasks = %v, want 0", result["total_tasks"])
	}
	if result["current_wave"] != "01" {
		t.Errorf("current_wave = %v, want 01", result["current_wave"])
	}
	if result["resume_action"] != "start-wave" {
		t.Errorf("resume_action = %v, want start-wave", result["resume_action"])
	}
}
