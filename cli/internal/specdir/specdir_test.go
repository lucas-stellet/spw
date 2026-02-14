package specdir

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestImplLogPath(t *testing.T) {
	got := ImplLogPath("/spec", "3")
	want := filepath.Join("/spec", "execution/_implementation-logs/task-3.md")
	if got != want {
		t.Errorf("ImplLogPath = %q, want %q", got, want)
	}
}

func TestWavePath(t *testing.T) {
	got := WavePath("/spec", 2)
	want := filepath.Join("/spec", "execution/waves/wave-02")
	if got != want {
		t.Errorf("WavePath = %q, want %q", got, want)
	}
}

func TestCheckpointRunPath(t *testing.T) {
	got := CheckpointRunPath("/spec", 3, 1)
	want := filepath.Join("/spec", "execution/waves/wave-03/checkpoint/run-001")
	if got != want {
		t.Errorf("CheckpointRunPath = %q, want %q", got, want)
	}
}

func TestCommsPath(t *testing.T) {
	tests := []struct {
		command string
		wave    int
		want    string
	}{
		{"discover", 0, "discover/_comms"},
		{"design-research", 0, "design/_comms/design-research"},
		{"design-draft", 0, "design/_comms/design-draft"},
		{"tasks-plan", 0, "planning/_comms/tasks-plan"},
		{"tasks-check", 0, "planning/_comms/tasks-check"},
		{"qa", 0, "qa/_comms/qa"},
		{"qa-check", 0, "qa/_comms/qa-check"},
		{"post-mortem", 0, "post-mortem/_comms"},
		{"exec", 1, "execution/waves/wave-01/execution"},
		{"checkpoint", 3, "execution/waves/wave-03/checkpoint"},
		{"qa-exec", 12, "qa/_comms/qa-exec/waves/wave-12"},
	}

	for _, tt := range tests {
		got := CommsPath("/spec", tt.command, tt.wave)
		want := filepath.Join("/spec", tt.want)
		if got != want {
			t.Errorf("CommsPath(%q, %d) = %q, want %q", tt.command, tt.wave, got, want)
		}
	}
}

func TestCommsPathUnknownCommand(t *testing.T) {
	got := CommsPath("/spec", "unknown", 0)
	if got != "" {
		t.Errorf("CommsPath(unknown) = %q, want empty", got)
	}
}

func TestListWaveDirs(t *testing.T) {
	tmp := t.TempDir()
	wavesPath := filepath.Join(tmp, WavesDir)
	os.MkdirAll(filepath.Join(wavesPath, "wave-01"), 0755)
	os.MkdirAll(filepath.Join(wavesPath, "wave-03"), 0755)
	os.MkdirAll(filepath.Join(wavesPath, "wave-02"), 0755)
	// Non-wave directory should be ignored
	os.MkdirAll(filepath.Join(wavesPath, "other"), 0755)

	waves, err := ListWaveDirs(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(waves) != 3 {
		t.Fatalf("expected 3 waves, got %d", len(waves))
	}
	// Should be sorted ascending
	if waves[0].Num != 1 || waves[1].Num != 2 || waves[2].Num != 3 {
		t.Errorf("waves not sorted: %v", waves)
	}
}

func TestListWaveDirsEmpty(t *testing.T) {
	tmp := t.TempDir()
	waves, err := ListWaveDirs(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if waves != nil {
		t.Errorf("expected nil, got %v", waves)
	}
}

func TestReadStatusJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "status.json")
	writeJSON(t, path, map[string]any{"status": "pass", "summary": "all good"})

	doc, err := ReadStatusJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Status != "pass" {
		t.Errorf("status = %q, want pass", doc.Status)
	}
	if doc.Summary != "all good" {
		t.Errorf("summary = %q, want 'all good'", doc.Summary)
	}
}

func TestReadStatusJSONMissing(t *testing.T) {
	_, err := ReadStatusJSON("/nonexistent/status.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLatestRunDir(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "run-001"), 0755)
	os.MkdirAll(filepath.Join(tmp, "run-003"), 0755)
	os.MkdirAll(filepath.Join(tmp, "run-002"), 0755)

	path, num, err := LatestRunDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if num != 3 {
		t.Errorf("num = %d, want 3", num)
	}
	if filepath.Base(path) != "run-003" {
		t.Errorf("path = %q, want run-003", path)
	}
}

func TestLatestRunDirEmpty(t *testing.T) {
	tmp := t.TempDir()
	_, _, err := LatestRunDir(tmp)
	if err == nil {
		t.Error("expected error for empty dir")
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")
	os.WriteFile(path, []byte("test"), 0644)

	if !FileExists(path) {
		t.Error("expected file to exist")
	}
	if FileExists(filepath.Join(tmp, "nope.md")) {
		t.Error("expected file not to exist")
	}
	if FileExists(tmp) {
		t.Error("directory should not count as file")
	}
}

func TestResolve(t *testing.T) {
	tmp := t.TempDir()
	specName := "test-spec"
	specDir := filepath.Join(tmp, ".spec-workflow", "specs", specName)
	os.MkdirAll(specDir, 0755)

	got, err := Resolve(tmp, specName)
	if err != nil {
		t.Fatal(err)
	}
	if got != specDir {
		t.Errorf("Resolve = %q, want %q", got, specDir)
	}
}

func TestResolveMissing(t *testing.T) {
	tmp := t.TempDir()
	_, err := Resolve(tmp, "nonexistent")
	if err == nil {
		t.Error("expected error for missing spec")
	}
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
