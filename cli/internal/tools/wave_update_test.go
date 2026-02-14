package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupWaveDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, ".spec-workflow", "specs", "test-spec")
	os.MkdirAll(specDir, 0755)
	return tmp
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestWaveUpdateCreateNew(t *testing.T) {
	cwd := setupWaveDir(t)

	result, err := waveUpdateResult(cwd, "test-spec", "02", "pass", "3,4,7", "run-001", "run-001")
	if err != nil {
		t.Fatal(err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["wave"] != "02" {
		t.Errorf("wave = %v, want 02", result["wave"])
	}

	// Verify summary JSON
	summaryPath := filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "execution", "waves", "wave-02", "_wave-summary.json")
	summary := readJSONFile(t, summaryPath)
	if summary["wave"] != "02" {
		t.Errorf("summary wave = %v, want 02", summary["wave"])
	}
	if summary["status"] != "pass" {
		t.Errorf("summary status = %v, want pass", summary["status"])
	}
	tasks, ok := summary["tasks"].([]any)
	if !ok || len(tasks) != 3 {
		t.Errorf("summary tasks = %v, want [3,4,7]", summary["tasks"])
	}
	if summary["updated_at"] == nil {
		t.Error("summary missing updated_at")
	}

	// Verify latest JSON
	latestPath := filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "execution", "waves", "wave-02", "_latest.json")
	latest := readJSONFile(t, latestPath)
	if latest["status"] != "pass" {
		t.Errorf("latest status = %v, want pass", latest["status"])
	}
	if latest["execution"] != "run-001" {
		t.Errorf("latest execution = %v, want run-001", latest["execution"])
	}
	if latest["checkpoint"] != "run-001" {
		t.Errorf("latest checkpoint = %v, want run-001", latest["checkpoint"])
	}
}

func TestWaveUpdateExisting(t *testing.T) {
	cwd := setupWaveDir(t)

	// Create initial wave
	_, err := waveUpdateResult(cwd, "test-spec", "01", "blocked", "1,2", "run-001", "run-001")
	if err != nil {
		t.Fatal(err)
	}

	// Update the wave
	result, err := waveUpdateResult(cwd, "test-spec", "01", "pass", "1,2", "run-002", "run-002")
	if err != nil {
		t.Fatal(err)
	}

	if result["wave"] != "01" {
		t.Errorf("wave = %v, want 01", result["wave"])
	}

	// Verify overwritten summary
	summaryPath := filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "execution", "waves", "wave-01", "_wave-summary.json")
	summary := readJSONFile(t, summaryPath)
	if summary["status"] != "pass" {
		t.Errorf("summary status = %v, want pass after update", summary["status"])
	}

	// Verify overwritten latest
	latestPath := filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "execution", "waves", "wave-01", "_latest.json")
	latest := readJSONFile(t, latestPath)
	if latest["execution"] != "run-002" {
		t.Errorf("latest execution = %v, want run-002", latest["execution"])
	}
}

func TestWaveUpdatePass(t *testing.T) {
	cwd := setupWaveDir(t)

	result, err := waveUpdateResult(cwd, "test-spec", "01", "pass", "1", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}

	// Verify latest has no execution/checkpoint when not provided
	latestPath := filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "execution", "waves", "wave-01", "_latest.json")
	latest := readJSONFile(t, latestPath)
	if _, exists := latest["execution"]; exists {
		t.Error("latest should not have execution field when not provided")
	}
	if _, exists := latest["checkpoint"]; exists {
		t.Error("latest should not have checkpoint field when not provided")
	}
	if latest["status"] != "pass" {
		t.Errorf("latest status = %v, want pass", latest["status"])
	}
}

func TestWaveUpdateBlocked(t *testing.T) {
	cwd := setupWaveDir(t)

	result, err := waveUpdateResult(cwd, "test-spec", "03", "blocked", "5,6", "run-001", "")
	if err != nil {
		t.Fatal(err)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}

	summaryPath := filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "execution", "waves", "wave-03", "_wave-summary.json")
	summary := readJSONFile(t, summaryPath)
	if summary["status"] != "blocked" {
		t.Errorf("summary status = %v, want blocked", summary["status"])
	}
}

func TestWaveUpdateDirCreation(t *testing.T) {
	cwd := setupWaveDir(t)

	waveDir := filepath.Join(cwd, ".spec-workflow", "specs", "test-spec", "execution", "waves", "wave-05")

	// Verify dir does not exist yet
	if _, err := os.Stat(waveDir); !os.IsNotExist(err) {
		t.Fatal("wave dir should not exist before update")
	}

	_, err := waveUpdateResult(cwd, "test-spec", "05", "pass", "10,11", "", "")
	if err != nil {
		t.Fatal(err)
	}

	// Verify dir was created
	info, err := os.Stat(waveDir)
	if err != nil {
		t.Fatalf("wave dir should exist after update: %v", err)
	}
	if !info.IsDir() {
		t.Error("wave path should be a directory")
	}
}
