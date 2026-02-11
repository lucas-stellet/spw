package wave

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- helpers ---

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

// --- Category C: Checkpoint tests ---

func TestC1_WaveSummaryPass(t *testing.T) {
	// C1: _wave-summary.json says "pass" -> CheckpointResult{Status: "pass"}
	specDir := t.TempDir()

	waveDir := filepath.Join(specDir, "execution", "waves", "wave-01")
	mkdirAll(t, waveDir)

	// _wave-summary.json
	writeJSON(t, filepath.Join(waveDir, "_wave-summary.json"), map[string]string{
		"status":  "pass",
		"summary": "all tests passed",
	})

	// checkpoint/run-001/release-gate-decider/status.json
	writeJSON(t, filepath.Join(waveDir, "checkpoint", "run-001", "release-gate-decider", "status.json"), map[string]string{
		"status":  "pass",
		"summary": "ok",
	})

	result := ResolveCheckpoint(specDir, 1)

	if result.Status != "pass" {
		t.Errorf("expected status 'pass', got %q", result.Status)
	}
	if result.StaleFlag {
		t.Error("expected stale_flag false, got true")
	}
	if result.WaveNum != 1 {
		t.Errorf("expected wave_num 1, got %d", result.WaveNum)
	}
}

func TestC2_StaleWaveSummary(t *testing.T) {
	// C2: Stale summary "blocked", _latest.json points to run-003 which has "pass"
	// -> CheckpointResult{Status: "pass", StaleFlag: true}
	specDir := t.TempDir()

	waveDir := filepath.Join(specDir, "execution", "waves", "wave-02")
	mkdirAll(t, waveDir)

	// _wave-summary.json (stale: says "blocked")
	writeJSON(t, filepath.Join(waveDir, "_wave-summary.json"), map[string]string{
		"status":  "blocked",
		"summary": "stale data",
	})

	// _latest.json (authoritative: says "pass", points to run-003)
	writeJSON(t, filepath.Join(waveDir, "_latest.json"), map[string]any{
		"run_id":  "run-003",
		"run_dir": "checkpoint/run-003",
		"status":  "pass",
		"summary": "fixed in third run",
	})

	// checkpoint/run-001 (old blocked run)
	writeJSON(t, filepath.Join(waveDir, "checkpoint", "run-001", "release-gate-decider", "status.json"), map[string]string{
		"status":  "blocked",
		"summary": "first run blocked",
	})

	// checkpoint/run-003 (latest passing run)
	writeJSON(t, filepath.Join(waveDir, "checkpoint", "run-003", "release-gate-decider", "status.json"), map[string]string{
		"status":  "pass",
		"summary": "fixed in third run",
	})

	result := ResolveCheckpoint(specDir, 2)

	if result.Status != "pass" {
		t.Errorf("expected status 'pass', got %q", result.Status)
	}
	if !result.StaleFlag {
		t.Error("expected stale_flag true, got false")
	}
	if result.RunID != "run-003" {
		t.Errorf("expected run_id 'run-003', got %q", result.RunID)
	}
	if result.Source != "latest_json" {
		t.Errorf("expected source 'latest_json', got %q", result.Source)
	}
}

func TestC3_NoCheckpointDir(t *testing.T) {
	// C3: No checkpoint directory -> CheckpointResult{Status: "missing"}
	specDir := t.TempDir()

	// Create wave dir but no checkpoint subdir
	waveDir := filepath.Join(specDir, "execution", "waves", "wave-01")
	mkdirAll(t, waveDir)

	result := ResolveCheckpoint(specDir, 1)

	if result.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", result.Status)
	}
}

func TestC4_MultipleRunsLatestMatters(t *testing.T) {
	// C4: Multiple runs, only latest matters -> correct run picked
	specDir := t.TempDir()

	waveDir := filepath.Join(specDir, "execution", "waves", "wave-01")
	mkdirAll(t, waveDir)

	// No _latest.json, so it should scan and pick highest run number

	// run-001 blocked
	writeJSON(t, filepath.Join(waveDir, "checkpoint", "run-001", "release-gate-decider", "status.json"), map[string]string{
		"status":  "blocked",
		"summary": "run 1 blocked",
	})

	// run-002 blocked
	writeJSON(t, filepath.Join(waveDir, "checkpoint", "run-002", "release-gate-decider", "status.json"), map[string]string{
		"status":  "blocked",
		"summary": "run 2 blocked",
	})

	// run-005 pass (highest number)
	writeJSON(t, filepath.Join(waveDir, "checkpoint", "run-005", "release-gate-decider", "status.json"), map[string]string{
		"status":  "pass",
		"summary": "run 5 passed",
	})

	// run-003 blocked (between 002 and 005)
	writeJSON(t, filepath.Join(waveDir, "checkpoint", "run-003", "release-gate-decider", "status.json"), map[string]string{
		"status":  "blocked",
		"summary": "run 3 blocked",
	})

	result := ResolveCheckpoint(specDir, 1)

	if result.Status != "pass" {
		t.Errorf("expected status 'pass', got %q", result.Status)
	}
	if result.RunID != "run-005" {
		t.Errorf("expected run_id 'run-005', got %q", result.RunID)
	}
	if result.Source != "dir_scan" {
		t.Errorf("expected source 'dir_scan', got %q", result.Source)
	}
}

// --- Category D: Wave summary/scanner tests ---

func TestD1_SingleWaveInProgress(t *testing.T) {
	// D1: Single wave with exec runs -> WaveState{Status: "in_progress"}
	specDir := t.TempDir()

	// Create wave-01 with one execution run but no checkpoint
	execRun := filepath.Join(specDir, "execution", "waves", "wave-01", "execution", "run-001", "implementer")
	mkdirAll(t, execRun)
	writeJSON(t, filepath.Join(execRun, "status.json"), map[string]string{
		"status":  "pass",
		"summary": "implemented",
	})

	waves, err := ScanWaves(specDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(waves) != 1 {
		t.Fatalf("expected 1 wave, got %d", len(waves))
	}
	if waves[0].Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got %q", waves[0].Status)
	}
	if waves[0].ExecRuns != 1 {
		t.Errorf("expected 1 exec run, got %d", waves[0].ExecRuns)
	}
	if waves[0].CheckRuns != 0 {
		t.Errorf("expected 0 check runs, got %d", waves[0].CheckRuns)
	}
}

func TestD2_WaveWithCheckpointPass(t *testing.T) {
	// D2: Wave with checkpoint pass -> WaveState{Status: "complete"}
	specDir := t.TempDir()

	waveDir := filepath.Join(specDir, "execution", "waves", "wave-01")

	// Execution run
	execRun := filepath.Join(waveDir, "execution", "run-001", "implementer")
	mkdirAll(t, execRun)
	writeJSON(t, filepath.Join(execRun, "status.json"), map[string]string{
		"status":  "pass",
		"summary": "implemented",
	})

	// Checkpoint run with pass
	checkRun := filepath.Join(waveDir, "checkpoint", "run-001", "release-gate-decider")
	mkdirAll(t, checkRun)
	writeJSON(t, filepath.Join(checkRun, "status.json"), map[string]string{
		"status":  "pass",
		"summary": "all good",
	})

	waves, err := ScanWaves(specDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(waves) != 1 {
		t.Fatalf("expected 1 wave, got %d", len(waves))
	}
	if waves[0].Status != "complete" {
		t.Errorf("expected status 'complete', got %q", waves[0].Status)
	}
	if waves[0].ExecRuns != 1 {
		t.Errorf("expected 1 exec run, got %d", waves[0].ExecRuns)
	}
	if waves[0].CheckRuns != 1 {
		t.Errorf("expected 1 check run, got %d", waves[0].CheckRuns)
	}
}

func TestD3_EmptyWavesDir(t *testing.T) {
	// D3: Empty waves dir -> empty slice
	specDir := t.TempDir()

	// Create the waves dir but leave it empty
	mkdirAll(t, filepath.Join(specDir, "execution", "waves"))

	waves, err := ScanWaves(specDir)
	if err != nil {
		t.Fatal(err)
	}

	if waves != nil {
		t.Errorf("expected nil, got %v", waves)
	}
}
