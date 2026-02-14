package store

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var runDirRe = regexp.MustCompile(`^run-(\d+)$`)

// HarvestRunDir scans a run directory for subagent dirs, reads their
// brief.md, report.md, and status.json, and stores them in the database.
// The entire operation runs in a single transaction.
func (s *SpecStore) HarvestRunDir(runDir, command string, waveNum *int) error {
	// Determine run number from directory name.
	base := filepath.Base(runDir)
	m := runDirRe.FindStringSubmatch(base)
	if m == nil {
		return fmt.Errorf("store: harvest: %q does not match run-NNN pattern", base)
	}
	runNumber, _ := strconv.Atoi(m[1])

	// Resolve phase from the run directory path.
	phase := inferPhase(runDir)

	// Compute a relative comms path from specDir.
	commsPath := runDir
	if rel, err := filepath.Rel(s.specDir, runDir); err == nil {
		commsPath = rel
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("store: harvest begin tx: %w", err)
	}
	defer tx.Rollback()

	ts := now()

	// Upsert run record.
	res, err := tx.Exec(`
		INSERT INTO runs (command, run_number, phase, wave_number, comms_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'in_progress', ?, ?)
		ON CONFLICT(command, run_number, wave_number) DO UPDATE SET
			status = runs.status,
			updated_at = excluded.updated_at`,
		command, runNumber, phase, waveNum, commsPath, ts, ts,
	)
	if err != nil {
		return fmt.Errorf("store: harvest insert run: %w", err)
	}
	runID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: harvest run id: %w", err)
	}

	// Scan for subagent directories (any subdirectory that is not a special dir).
	entries, err := os.ReadDir(runDir)
	if err != nil {
		return fmt.Errorf("store: harvest readdir %s: %w", runDir, err)
	}

	allPass := true
	hasSubagents := false

	for _, e := range entries {
		if !e.IsDir() || isSpecialDir(e.Name()) {
			continue
		}
		hasSubagents = true
		subDir := filepath.Join(runDir, e.Name())

		brief := readFileOpt(filepath.Join(subDir, "brief.md"))
		report := readFileOpt(filepath.Join(subDir, "report.md"))
		statusJSON := readFileOpt(filepath.Join(subDir, "status.json"))

		var status, summary string
		if statusJSON != "" {
			var doc struct {
				Status  string `json:"status"`
				Summary string `json:"summary"`
			}
			if json.Unmarshal([]byte(statusJSON), &doc) == nil {
				status = doc.Status
				summary = doc.Summary
			}
		}

		if status != "pass" && status != "" {
			allPass = false
		}

		_, err := tx.Exec(`
			INSERT INTO subagents (run_id, name, brief, report, status, summary, status_json, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			runID, e.Name(),
			nullStrPtr(brief), nullStrPtr(report),
			nullStrPtr(status), nullStrPtr(summary), nullStrPtr(statusJSON),
			ts, ts,
		)
		if err != nil {
			return fmt.Errorf("store: harvest insert subagent %s: %w", e.Name(), err)
		}
	}

	// Check for _handoff.md.
	handoffContent := readFileOpt(filepath.Join(runDir, "_handoff.md"))
	if handoffContent != "" {
		pass := 0
		if allPass && hasSubagents {
			pass = 1
		}
		_, err := tx.Exec(
			"INSERT INTO handoffs (run_id, content, all_pass, created_at) VALUES (?, ?, ?, ?)",
			runID, handoffContent, pass, ts,
		)
		if err != nil {
			return fmt.Errorf("store: harvest insert handoff: %w", err)
		}

		// Mark run as completed if handoff exists.
		runStatus := "pass"
		if !allPass {
			runStatus = "blocked"
		}
		_, err = tx.Exec("UPDATE runs SET status = ?, updated_at = ? WHERE id = ?", runStatus, ts, runID)
		if err != nil {
			return fmt.Errorf("store: harvest update run status: %w", err)
		}
	}

	return tx.Commit()
}

// HarvestArtifact reads a file from disk, computes its SHA256 hash,
// and upserts it into the artifacts table.
func (s *SpecStore) HarvestArtifact(phase, relPath, absPath string) error {
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("store: harvest artifact read %s: %w", absPath, err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	artType := inferArtifactType(relPath)
	ts := now()

	_, err = s.db.Exec(`
		INSERT INTO artifacts (phase, rel_path, artifact_type, content, content_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(rel_path) DO UPDATE SET
			content = excluded.content,
			content_hash = excluded.content_hash,
			updated_at = excluded.updated_at
		WHERE content_hash != excluded.content_hash`,
		phase, relPath, artType, string(content), hash, ts, ts,
	)
	if err != nil {
		return fmt.Errorf("store: harvest artifact upsert: %w", err)
	}
	return nil
}

// HarvestImplLog reads a task implementation log file and stores it.
func (s *SpecStore) HarvestImplLog(taskID, absPath string) error {
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("store: harvest impl log read %s: %w", absPath, err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	ts := now()

	_, err = s.db.Exec(`
		INSERT INTO impl_logs (task_id, content, content_hash, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(task_id) DO UPDATE SET
			content = excluded.content,
			content_hash = excluded.content_hash,
			updated_at = excluded.updated_at
		WHERE content_hash != excluded.content_hash`,
		taskID, string(content), hash, ts,
	)
	if err != nil {
		return fmt.Errorf("store: harvest impl log upsert: %w", err)
	}
	return nil
}

// --- helpers ---

func readFileOpt(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func nullStrPtr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func isSpecialDir(name string) bool {
	switch name {
	case "_inline-audit", "_inline-checkpoint":
		return true
	}
	return len(name) > 0 && name[0] == '_'
}

func inferPhase(runDir string) string {
	phases := []string{"prd", "design", "planning", "execution", "qa", "post-mortem"}
	for _, p := range phases {
		if containsPathSegment(runDir, p) {
			return p
		}
	}
	return "unknown"
}

func containsPathSegment(path, segment string) bool {
	for _, part := range filepath.SplitList(path) {
		if part == segment {
			return true
		}
	}
	// SplitList is for PATH-like lists; use manual split instead.
	dir := path
	for dir != "" && dir != "." && dir != "/" {
		if filepath.Base(dir) == segment {
			return true
		}
		dir = filepath.Dir(dir)
	}
	return false
}

func inferArtifactType(relPath string) string {
	ext := filepath.Ext(relPath)
	base := filepath.Base(relPath)
	switch {
	case base == "status.json":
		return "status"
	case base == "brief.md":
		return "brief"
	case base == "report.md":
		return "report"
	case base == "_handoff.md":
		return "handoff"
	case base == "_wave-summary.json" || base == "_latest.json":
		return "wave-state"
	case ext == ".md":
		return "document"
	case ext == ".json":
		return "data"
	default:
		return "other"
	}
}
