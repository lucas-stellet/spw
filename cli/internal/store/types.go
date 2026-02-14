// Package store provides SQLite-backed storage for spec-workflow runtime state.
package store

// Artifact represents a stored spec artifact (report, brief, design doc, etc.).
type Artifact struct {
	ID           int64  `json:"id"`
	Phase        string `json:"phase"`
	RelPath      string `json:"rel_path"`
	ArtifactType string `json:"artifact_type"`
	Content      string `json:"content"`
	ContentHash  string `json:"content_hash"`
	Metadata     string `json:"metadata,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Run represents a single command dispatch run (run-NNN directory).
type Run struct {
	ID         int64  `json:"id"`
	Command    string `json:"command"`
	RunNumber  int    `json:"run_number"`
	Phase      string `json:"phase"`
	WaveNumber *int   `json:"wave_number,omitempty"`
	CommsPath  string `json:"comms_path"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// Subagent represents a subagent within a run.
type Subagent struct {
	ID         int64  `json:"id"`
	RunID      int64  `json:"run_id"`
	Name       string `json:"name"`
	Brief      string `json:"brief,omitempty"`
	Report     string `json:"report,omitempty"`
	Status     string `json:"status,omitempty"`
	Summary    string `json:"summary,omitempty"`
	StatusJSON string `json:"status_json,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// WaveRecord represents the state of a single execution wave.
type WaveRecord struct {
	WaveNumber    int    `json:"wave_number"`
	Status        string `json:"status"`
	ExecRuns      int    `json:"exec_runs"`
	CheckRuns     int    `json:"check_runs"`
	SummaryStatus string `json:"summary_status,omitempty"`
	SummaryText   string `json:"summary_text,omitempty"`
	SummarySource string `json:"summary_source,omitempty"`
	StaleFlag     bool   `json:"stale_flag"`
	UpdatedAt     string `json:"updated_at"`
}

// TaskRecord represents a task stored in the database.
type TaskRecord struct {
	TaskID     string `json:"task_id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Wave       *int   `json:"wave,omitempty"`
	DependsOn  string `json:"depends_on,omitempty"`
	Files      string `json:"files,omitempty"`
	TDD        bool   `json:"tdd"`
	IsDeferred bool   `json:"is_deferred"`
	UpdatedAt  string `json:"updated_at"`
}

// ImplLog represents a task's implementation log stored in the database.
type ImplLog struct {
	TaskID      string `json:"task_id"`
	Content     string `json:"content"`
	ContentHash string `json:"content_hash"`
	UpdatedAt   string `json:"updated_at"`
}

// Handoff represents a dispatch handoff record.
type Handoff struct {
	ID        int64  `json:"id"`
	RunID     int64  `json:"run_id"`
	Content   string `json:"content"`
	AllPass   bool   `json:"all_pass"`
	CreatedAt string `json:"created_at"`
}

// Approval represents an MCP approval record.
type Approval struct {
	ID         int64  `json:"id"`
	DocType    string `json:"doc_type"`
	ApprovalID string `json:"approval_id"`
	RawJSON    string `json:"raw_json,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// CompletionRecord represents the singleton completion summary.
type CompletionRecord struct {
	ID          int64  `json:"id"`
	Frontmatter string `json:"frontmatter"`
	Body        string `json:"body"`
	GeneratedAt string `json:"generated_at"`
}

// SearchResult represents a document found via FTS5 search.
type SearchResult struct {
	Spec    string  `json:"spec"`
	DocType string  `json:"doc_type"`
	Phase   string  `json:"phase"`
	Title   string  `json:"title"`
	Snippet string  `json:"snippet"`
	Rank    float64 `json:"rank"`
}
