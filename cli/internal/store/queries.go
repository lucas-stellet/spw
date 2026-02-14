package store

import (
	"database/sql"
	"fmt"
	"time"
)

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// --- spec_meta ---

// GetMeta retrieves a value from the spec_meta table.
func (s *SpecStore) GetMeta(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM spec_meta WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("store: get meta %q: %w", key, err)
	}
	return value, nil
}

// SetMeta upserts a key-value pair in the spec_meta table.
func (s *SpecStore) SetMeta(key, value string) error {
	_, err := s.db.Exec(
		"INSERT INTO spec_meta (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("store: set meta %q: %w", key, err)
	}
	return nil
}

// --- artifacts ---

// GetArtifact retrieves a single artifact by phase and relative path.
func (s *SpecStore) GetArtifact(phase, relPath string) (*Artifact, error) {
	a := &Artifact{}
	var metadata sql.NullString
	err := s.db.QueryRow(
		"SELECT id, phase, rel_path, artifact_type, content, content_hash, metadata, created_at, updated_at FROM artifacts WHERE phase = ? AND rel_path = ?",
		phase, relPath,
	).Scan(&a.ID, &a.Phase, &a.RelPath, &a.ArtifactType, &a.Content, &a.ContentHash, &metadata, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get artifact: %w", err)
	}
	a.Metadata = metadata.String
	return a, nil
}

// ListArtifacts returns all artifacts for a given phase.
func (s *SpecStore) ListArtifacts(phase string) ([]Artifact, error) {
	rows, err := s.db.Query(
		"SELECT id, phase, rel_path, artifact_type, content, content_hash, metadata, created_at, updated_at FROM artifacts WHERE phase = ? ORDER BY rel_path",
		phase,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list artifacts: %w", err)
	}
	defer rows.Close()

	var result []Artifact
	for rows.Next() {
		var a Artifact
		var metadata sql.NullString
		if err := rows.Scan(&a.ID, &a.Phase, &a.RelPath, &a.ArtifactType, &a.Content, &a.ContentHash, &metadata, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan artifact: %w", err)
		}
		a.Metadata = metadata.String
		result = append(result, a)
	}
	return result, rows.Err()
}

// --- runs ---

// CreateRun inserts a new run record and returns its ID.
func (s *SpecStore) CreateRun(command string, runNumber int, phase string, waveNum *int, commsPath string) (int64, error) {
	ts := now()
	res, err := s.db.Exec(
		"INSERT INTO runs (command, run_number, phase, wave_number, comms_path, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'in_progress', ?, ?)",
		command, runNumber, phase, waveNum, commsPath, ts, ts,
	)
	if err != nil {
		return 0, fmt.Errorf("store: create run: %w", err)
	}
	return res.LastInsertId()
}

// GetRun retrieves a run by command and run number.
func (s *SpecStore) GetRun(command string, runNum int) (*Run, error) {
	r := &Run{}
	err := s.db.QueryRow(
		"SELECT id, command, run_number, phase, wave_number, comms_path, status, created_at, updated_at FROM runs WHERE command = ? AND run_number = ?",
		command, runNum,
	).Scan(&r.ID, &r.Command, &r.RunNumber, &r.Phase, &r.WaveNumber, &r.CommsPath, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get run: %w", err)
	}
	return r, nil
}

// LatestRun returns the most recent run for a command.
func (s *SpecStore) LatestRun(command string) (*Run, error) {
	r := &Run{}
	err := s.db.QueryRow(
		"SELECT id, command, run_number, phase, wave_number, comms_path, status, created_at, updated_at FROM runs WHERE command = ? ORDER BY run_number DESC LIMIT 1",
		command,
	).Scan(&r.ID, &r.Command, &r.RunNumber, &r.Phase, &r.WaveNumber, &r.CommsPath, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: latest run: %w", err)
	}
	return r, nil
}

// UpdateRunStatus sets the status of a run.
func (s *SpecStore) UpdateRunStatus(runID int64, status string) error {
	_, err := s.db.Exec(
		"UPDATE runs SET status = ?, updated_at = ? WHERE id = ?",
		status, now(), runID,
	)
	if err != nil {
		return fmt.Errorf("store: update run status: %w", err)
	}
	return nil
}

// --- subagents ---

// CreateSubagent inserts a new subagent record and returns its ID.
func (s *SpecStore) CreateSubagent(runID int64, name string) (int64, error) {
	ts := now()
	res, err := s.db.Exec(
		"INSERT INTO subagents (run_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
		runID, name, ts, ts,
	)
	if err != nil {
		return 0, fmt.Errorf("store: create subagent: %w", err)
	}
	return res.LastInsertId()
}

// UpdateSubagent updates optional fields on a subagent record.
func (s *SpecStore) UpdateSubagent(id int64, brief, report, status, summary, statusJSON *string) error {
	// Build dynamic update to only set non-nil fields.
	q := "UPDATE subagents SET updated_at = ?"
	args := []any{now()}

	if brief != nil {
		q += ", brief = ?"
		args = append(args, *brief)
	}
	if report != nil {
		q += ", report = ?"
		args = append(args, *report)
	}
	if status != nil {
		q += ", status = ?"
		args = append(args, *status)
	}
	if summary != nil {
		q += ", summary = ?"
		args = append(args, *summary)
	}
	if statusJSON != nil {
		q += ", status_json = ?"
		args = append(args, *statusJSON)
	}

	q += " WHERE id = ?"
	args = append(args, id)

	_, err := s.db.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("store: update subagent: %w", err)
	}
	return nil
}

// ListSubagents returns all subagents for a given run.
func (s *SpecStore) ListSubagents(runID int64) ([]Subagent, error) {
	rows, err := s.db.Query(
		"SELECT id, run_id, name, brief, report, status, summary, status_json, created_at, updated_at FROM subagents WHERE run_id = ? ORDER BY name",
		runID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list subagents: %w", err)
	}
	defer rows.Close()

	var result []Subagent
	for rows.Next() {
		var a Subagent
		var brief, report, status, summary, statusJSON sql.NullString
		if err := rows.Scan(&a.ID, &a.RunID, &a.Name, &brief, &report, &status, &summary, &statusJSON, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan subagent: %w", err)
		}
		a.Brief = brief.String
		a.Report = report.String
		a.Status = status.String
		a.Summary = summary.String
		a.StatusJSON = statusJSON.String
		result = append(result, a)
	}
	return result, rows.Err()
}

// --- waves ---

// UpsertWave inserts or updates a wave record.
func (s *SpecStore) UpsertWave(w WaveRecord) error {
	stale := 0
	if w.StaleFlag {
		stale = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO waves (wave_number, status, exec_runs, check_runs, summary_status, summary_text, summary_source, stale_flag, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(wave_number) DO UPDATE SET
			status = excluded.status,
			exec_runs = excluded.exec_runs,
			check_runs = excluded.check_runs,
			summary_status = excluded.summary_status,
			summary_text = excluded.summary_text,
			summary_source = excluded.summary_source,
			stale_flag = excluded.stale_flag,
			updated_at = excluded.updated_at`,
		w.WaveNumber, w.Status, w.ExecRuns, w.CheckRuns,
		nullStr(w.SummaryStatus), nullStr(w.SummaryText), nullStr(w.SummarySource),
		stale, now(),
	)
	if err != nil {
		return fmt.Errorf("store: upsert wave: %w", err)
	}
	return nil
}

// GetWave retrieves a wave by number.
func (s *SpecStore) GetWave(num int) (*WaveRecord, error) {
	w := &WaveRecord{}
	var summaryStatus, summaryText, summarySource sql.NullString
	var stale int
	err := s.db.QueryRow(
		"SELECT wave_number, status, exec_runs, check_runs, summary_status, summary_text, summary_source, stale_flag, updated_at FROM waves WHERE wave_number = ?",
		num,
	).Scan(&w.WaveNumber, &w.Status, &w.ExecRuns, &w.CheckRuns, &summaryStatus, &summaryText, &summarySource, &stale, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get wave: %w", err)
	}
	w.SummaryStatus = summaryStatus.String
	w.SummaryText = summaryText.String
	w.SummarySource = summarySource.String
	w.StaleFlag = stale != 0
	return w, nil
}

// ListWaves returns all waves ordered by wave number.
func (s *SpecStore) ListWaves() ([]WaveRecord, error) {
	rows, err := s.db.Query(
		"SELECT wave_number, status, exec_runs, check_runs, summary_status, summary_text, summary_source, stale_flag, updated_at FROM waves ORDER BY wave_number",
	)
	if err != nil {
		return nil, fmt.Errorf("store: list waves: %w", err)
	}
	defer rows.Close()

	var result []WaveRecord
	for rows.Next() {
		var w WaveRecord
		var summaryStatus, summaryText, summarySource sql.NullString
		var stale int
		if err := rows.Scan(&w.WaveNumber, &w.Status, &w.ExecRuns, &w.CheckRuns, &summaryStatus, &summaryText, &summarySource, &stale, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan wave: %w", err)
		}
		w.SummaryStatus = summaryStatus.String
		w.SummaryText = summaryText.String
		w.SummarySource = summarySource.String
		w.StaleFlag = stale != 0
		result = append(result, w)
	}
	return result, rows.Err()
}

// --- tasks ---

// SyncTask upserts a task record.
func (s *SpecStore) SyncTask(t TaskRecord) error {
	tdd := 0
	if t.TDD {
		tdd = 1
	}
	deferred := 0
	if t.IsDeferred {
		deferred = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO tasks (task_id, title, status, wave, depends_on, files, tdd, is_deferred, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(task_id) DO UPDATE SET
			title = excluded.title,
			status = excluded.status,
			wave = excluded.wave,
			depends_on = excluded.depends_on,
			files = excluded.files,
			tdd = excluded.tdd,
			is_deferred = excluded.is_deferred,
			updated_at = excluded.updated_at`,
		t.TaskID, t.Title, t.Status, t.Wave,
		nullStr(t.DependsOn), nullStr(t.Files),
		tdd, deferred, now(),
	)
	if err != nil {
		return fmt.Errorf("store: sync task: %w", err)
	}
	return nil
}

// ListTasks returns all tasks ordered by task_id.
func (s *SpecStore) ListTasks() ([]TaskRecord, error) {
	rows, err := s.db.Query(
		"SELECT task_id, title, status, wave, depends_on, files, tdd, is_deferred, updated_at FROM tasks ORDER BY task_id",
	)
	if err != nil {
		return nil, fmt.Errorf("store: list tasks: %w", err)
	}
	defer rows.Close()

	var result []TaskRecord
	for rows.Next() {
		var t TaskRecord
		var dependsOn, files sql.NullString
		var tdd, deferred int
		if err := rows.Scan(&t.TaskID, &t.Title, &t.Status, &t.Wave, &dependsOn, &files, &tdd, &deferred, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan task: %w", err)
		}
		t.DependsOn = dependsOn.String
		t.Files = files.String
		t.TDD = tdd != 0
		t.IsDeferred = deferred != 0
		result = append(result, t)
	}
	return result, rows.Err()
}

// --- handoffs ---

// CreateHandoff inserts a new handoff record.
func (s *SpecStore) CreateHandoff(runID int64, content string, allPass bool) error {
	pass := 0
	if allPass {
		pass = 1
	}
	_, err := s.db.Exec(
		"INSERT INTO handoffs (run_id, content, all_pass, created_at) VALUES (?, ?, ?, ?)",
		runID, content, pass, now(),
	)
	if err != nil {
		return fmt.Errorf("store: create handoff: %w", err)
	}
	return nil
}

// --- impl_logs ---

// GetImplLog retrieves an implementation log by task ID.
func (s *SpecStore) GetImplLog(taskID string) (*ImplLog, error) {
	l := &ImplLog{}
	err := s.db.QueryRow(
		"SELECT task_id, content, content_hash, updated_at FROM impl_logs WHERE task_id = ?",
		taskID,
	).Scan(&l.TaskID, &l.Content, &l.ContentHash, &l.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get impl log: %w", err)
	}
	return l, nil
}

// --- completion_summary ---

// GetCompletionSummary retrieves the singleton completion summary.
func (s *SpecStore) GetCompletionSummary() (*CompletionRecord, error) {
	c := &CompletionRecord{}
	err := s.db.QueryRow(
		"SELECT id, frontmatter, body, generated_at FROM completion_summary WHERE id = 1",
	).Scan(&c.ID, &c.Frontmatter, &c.Body, &c.GeneratedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get completion summary: %w", err)
	}
	return c, nil
}

// SaveCompletionSummary upserts the singleton completion summary.
func (s *SpecStore) SaveCompletionSummary(frontmatter, body string) error {
	_, err := s.db.Exec(`
		INSERT INTO completion_summary (id, frontmatter, body, generated_at)
		VALUES (1, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			frontmatter = excluded.frontmatter,
			body = excluded.body,
			generated_at = excluded.generated_at`,
		frontmatter, body, now(),
	)
	if err != nil {
		return fmt.Errorf("store: save completion summary: %w", err)
	}
	return nil
}

// --- helpers ---

// nullStr returns a sql.NullString that is null when the string is empty.
func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
