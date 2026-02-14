package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// openMemory creates a SpecStore backed by a temporary directory with an actual file DB,
// since modernc.org/sqlite requires a file path (not :memory:) for full WAL support.
func openTestStore(t *testing.T) *SpecStore {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open(%q): %v", dir, err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpenClose(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Verify spec.db was created.
	if _, err := os.Stat(filepath.Join(dir, "spec.db")); err != nil {
		t.Fatalf("spec.db not created: %v", err)
	}
}

func TestTryOpen(t *testing.T) {
	dir := t.TempDir()
	s := TryOpen(dir)
	if s == nil {
		t.Fatal("TryOpen returned nil for valid dir")
	}
	s.Close()

	// TryOpen on nonexistent path should return nil.
	s = TryOpen("/nonexistent/path/that/does/not/exist")
	if s != nil {
		t.Fatal("TryOpen should return nil for bad path")
		s.Close()
	}
}

func TestMigrateIdempotent(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir) // calls Migrate() internally
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// Calling Migrate again should be a no-op.
	if err := s.Migrate(); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	if err := s.Migrate(); err != nil {
		t.Fatalf("third Migrate: %v", err)
	}
}

func TestSetGetMeta(t *testing.T) {
	s := openTestStore(t)

	// Get nonexistent key returns empty string.
	val, err := s.GetMeta("nonexistent")
	if err != nil {
		t.Fatalf("GetMeta: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty, got %q", val)
	}

	// Set and get.
	if err := s.SetMeta("name", "my-spec"); err != nil {
		t.Fatalf("SetMeta: %v", err)
	}
	val, err = s.GetMeta("name")
	if err != nil {
		t.Fatalf("GetMeta: %v", err)
	}
	if val != "my-spec" {
		t.Fatalf("expected 'my-spec', got %q", val)
	}

	// Upsert.
	if err := s.SetMeta("name", "updated-spec"); err != nil {
		t.Fatalf("SetMeta upsert: %v", err)
	}
	val, err = s.GetMeta("name")
	if err != nil {
		t.Fatalf("GetMeta after upsert: %v", err)
	}
	if val != "updated-spec" {
		t.Fatalf("expected 'updated-spec', got %q", val)
	}
}

func TestRunCRUD(t *testing.T) {
	s := openTestStore(t)

	// Create run.
	waveNum := 1
	id, err := s.CreateRun("exec", 1, "execution", &waveNum, "execution/waves/wave-01/execution/run-001")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero run ID")
	}

	// Get run.
	r, err := s.GetRun("exec", 1)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if r == nil {
		t.Fatal("GetRun returned nil")
	}
	if r.Command != "exec" || r.RunNumber != 1 || r.Phase != "execution" {
		t.Fatalf("unexpected run: %+v", r)
	}
	if r.WaveNumber == nil || *r.WaveNumber != 1 {
		t.Fatalf("expected wave_number=1, got %v", r.WaveNumber)
	}

	// Latest run.
	latest, err := s.LatestRun("exec")
	if err != nil {
		t.Fatalf("LatestRun: %v", err)
	}
	if latest == nil || latest.ID != id {
		t.Fatal("LatestRun mismatch")
	}

	// Update status.
	if err := s.UpdateRunStatus(id, "pass"); err != nil {
		t.Fatalf("UpdateRunStatus: %v", err)
	}
	r, _ = s.GetRun("exec", 1)
	if r.Status != "pass" {
		t.Fatalf("expected status=pass, got %q", r.Status)
	}

	// Get nonexistent run.
	r, err = s.GetRun("nonexistent", 99)
	if err != nil {
		t.Fatalf("GetRun nonexistent: %v", err)
	}
	if r != nil {
		t.Fatal("expected nil for nonexistent run")
	}
}

func TestSubagentCRUD(t *testing.T) {
	s := openTestStore(t)

	runID, err := s.CreateRun("prd", 1, "prd", nil, "prd/_comms/run-001")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	subID, err := s.CreateSubagent(runID, "researcher")
	if err != nil {
		t.Fatalf("CreateSubagent: %v", err)
	}

	brief := "Research the topic"
	report := "Findings here"
	status := "pass"
	summary := "All good"
	statusJSON := `{"status":"pass","summary":"All good"}`
	if err := s.UpdateSubagent(subID, &brief, &report, &status, &summary, &statusJSON); err != nil {
		t.Fatalf("UpdateSubagent: %v", err)
	}

	subs, err := s.ListSubagents(runID)
	if err != nil {
		t.Fatalf("ListSubagents: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subagent, got %d", len(subs))
	}
	if subs[0].Name != "researcher" || subs[0].Brief != brief || subs[0].Status != status {
		t.Fatalf("unexpected subagent: %+v", subs[0])
	}
}

func TestWaveCRUD(t *testing.T) {
	s := openTestStore(t)

	w := WaveRecord{
		WaveNumber:    1,
		Status:        "in_progress",
		ExecRuns:      2,
		CheckRuns:     1,
		SummaryStatus: "pass",
		SummaryText:   "All tasks completed",
		SummarySource: "wave_summary",
		StaleFlag:     false,
	}
	if err := s.UpsertWave(w); err != nil {
		t.Fatalf("UpsertWave: %v", err)
	}

	got, err := s.GetWave(1)
	if err != nil {
		t.Fatalf("GetWave: %v", err)
	}
	if got == nil {
		t.Fatal("GetWave returned nil")
	}
	if got.Status != "in_progress" || got.ExecRuns != 2 || got.SummaryStatus != "pass" {
		t.Fatalf("unexpected wave: %+v", got)
	}

	// Upsert update.
	w.Status = "complete"
	w.StaleFlag = true
	if err := s.UpsertWave(w); err != nil {
		t.Fatalf("UpsertWave update: %v", err)
	}
	got, _ = s.GetWave(1)
	if got.Status != "complete" || !got.StaleFlag {
		t.Fatalf("upsert did not update: %+v", got)
	}

	// List.
	waves, err := s.ListWaves()
	if err != nil {
		t.Fatalf("ListWaves: %v", err)
	}
	if len(waves) != 1 {
		t.Fatalf("expected 1 wave, got %d", len(waves))
	}

	// Nonexistent.
	got, err = s.GetWave(99)
	if err != nil {
		t.Fatalf("GetWave 99: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for nonexistent wave")
	}
}

func TestTaskCRUD(t *testing.T) {
	s := openTestStore(t)

	wave := 1
	task := TaskRecord{
		TaskID:    "1",
		Title:     "Implement login",
		Status:    "pending",
		Wave:      &wave,
		DependsOn: `["2","3"]`,
		Files:     "src/auth/login.go",
		TDD:       true,
	}
	if err := s.SyncTask(task); err != nil {
		t.Fatalf("SyncTask: %v", err)
	}

	tasks, err := s.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].TaskID != "1" || tasks[0].Title != "Implement login" || !tasks[0].TDD {
		t.Fatalf("unexpected task: %+v", tasks[0])
	}

	// Upsert update.
	task.Status = "done"
	if err := s.SyncTask(task); err != nil {
		t.Fatalf("SyncTask update: %v", err)
	}
	tasks, _ = s.ListTasks()
	if tasks[0].Status != "done" {
		t.Fatalf("expected done, got %q", tasks[0].Status)
	}
}

func TestHandoff(t *testing.T) {
	s := openTestStore(t)

	runID, _ := s.CreateRun("prd", 1, "prd", nil, "prd/_comms/run-001")
	if err := s.CreateHandoff(runID, "All subagents passed.", true); err != nil {
		t.Fatalf("CreateHandoff: %v", err)
	}
}

func TestHarvestArtifact(t *testing.T) {
	s := openTestStore(t)

	// Create a temp file to harvest.
	tmpDir := t.TempDir()
	content := "# Design Research\n\nFindings..."
	filePath := filepath.Join(tmpDir, "DESIGN-RESEARCH.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := s.HarvestArtifact("design", "design/DESIGN-RESEARCH.md", filePath); err != nil {
		t.Fatalf("HarvestArtifact: %v", err)
	}

	a, err := s.GetArtifact("design", "design/DESIGN-RESEARCH.md")
	if err != nil {
		t.Fatalf("GetArtifact: %v", err)
	}
	if a == nil {
		t.Fatal("GetArtifact returned nil")
	}
	if a.Content != content {
		t.Fatalf("content mismatch: got %q", a.Content)
	}
	if a.ContentHash == "" {
		t.Fatal("expected non-empty content_hash")
	}
	if a.ArtifactType != "document" {
		t.Fatalf("expected artifact_type=document, got %q", a.ArtifactType)
	}

	// Harvest again with same content — should not change updated_at significantly.
	if err := s.HarvestArtifact("design", "design/DESIGN-RESEARCH.md", filePath); err != nil {
		t.Fatalf("HarvestArtifact (idempotent): %v", err)
	}

	// List artifacts.
	arts, err := s.ListArtifacts("design")
	if err != nil {
		t.Fatalf("ListArtifacts: %v", err)
	}
	if len(arts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(arts))
	}
}

func TestHarvestImplLog(t *testing.T) {
	s := openTestStore(t)

	tmpDir := t.TempDir()
	content := "# Task 1 Implementation\n\n## Changes..."
	filePath := filepath.Join(tmpDir, "task-1.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := s.HarvestImplLog("1", filePath); err != nil {
		t.Fatalf("HarvestImplLog: %v", err)
	}

	// Harvest again — should be idempotent.
	if err := s.HarvestImplLog("1", filePath); err != nil {
		t.Fatalf("HarvestImplLog (idempotent): %v", err)
	}
}

func TestHarvestRunDir(t *testing.T) {
	s := openTestStore(t)

	// Create a fake run directory with subagents.
	runDir := filepath.Join(t.TempDir(), "run-001")
	researcherDir := filepath.Join(runDir, "researcher")
	analystDir := filepath.Join(runDir, "analyst")
	os.MkdirAll(researcherDir, 0755)
	os.MkdirAll(analystDir, 0755)

	// Write subagent files.
	os.WriteFile(filepath.Join(researcherDir, "brief.md"), []byte("Research brief"), 0644)
	os.WriteFile(filepath.Join(researcherDir, "report.md"), []byte("Research report"), 0644)
	statusDoc := map[string]string{"status": "pass", "summary": "Done"}
	statusBytes, _ := json.Marshal(statusDoc)
	os.WriteFile(filepath.Join(researcherDir, "status.json"), statusBytes, 0644)

	os.WriteFile(filepath.Join(analystDir, "brief.md"), []byte("Analysis brief"), 0644)
	os.WriteFile(filepath.Join(analystDir, "status.json"), []byte(`{"status":"pass","summary":"Analysis complete"}`), 0644)

	// Write handoff.
	os.WriteFile(filepath.Join(runDir, "_handoff.md"), []byte("All agents done."), 0644)

	if err := s.HarvestRunDir(runDir, "prd", nil); err != nil {
		t.Fatalf("HarvestRunDir: %v", err)
	}

	// Verify run was created.
	r, err := s.LatestRun("prd")
	if err != nil {
		t.Fatalf("LatestRun: %v", err)
	}
	if r == nil {
		t.Fatal("no run found after harvest")
	}
	if r.RunNumber != 1 || r.Status != "pass" {
		t.Fatalf("unexpected run: number=%d status=%s", r.RunNumber, r.Status)
	}

	// Verify subagents.
	subs, err := s.ListSubagents(r.ID)
	if err != nil {
		t.Fatalf("ListSubagents: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 subagents, got %d", len(subs))
	}
}

func TestIndexStore(t *testing.T) {
	dir := t.TempDir()
	specWorkflow := filepath.Join(dir, ".spec-workflow")
	os.MkdirAll(specWorkflow, 0755)

	ix, err := OpenIndex(dir)
	if err != nil {
		t.Fatalf("OpenIndex: %v", err)
	}
	defer ix.Close()

	// Index a spec.
	if err := ix.IndexSpec("auth-feature", "execution", "/path/to/spec.db"); err != nil {
		t.Fatalf("IndexSpec: %v", err)
	}

	// Index documents.
	if err := ix.IndexDocument("auth-feature", "report", "prd", "PRD Report", "User auth requirements", "Full content of the PRD report about user authentication..."); err != nil {
		t.Fatalf("IndexDocument: %v", err)
	}
	if err := ix.IndexDocument("auth-feature", "checkpoint", "execution", "Wave 1 Checkpoint", "All tests pass", "Checkpoint report showing all tests passing..."); err != nil {
		t.Fatalf("IndexDocument 2: %v", err)
	}

	// Search.
	results, err := ix.Search("authentication", "", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
	if results[0].Spec != "auth-feature" {
		t.Fatalf("unexpected spec: %q", results[0].Spec)
	}

	// Search with spec filter.
	results, err = ix.Search("checkpoint", "auth-feature", 5)
	if err != nil {
		t.Fatalf("Search filtered: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 filtered result, got %d", len(results))
	}

	// Search with no results.
	results, err = ix.Search("nonexistent-xyz-term", "", 5)
	if err != nil {
		t.Fatalf("Search no results: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
