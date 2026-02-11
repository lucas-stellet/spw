// Package tasks implements tasks.md parsing and state resolution.
// The parser uses body scan as authoritative source, not frontmatter task_ids.
package tasks

// Document represents a fully parsed tasks.md file.
type Document struct {
	Frontmatter  Frontmatter
	Tasks        []Task
	WavePlan     []WavePlanEntry
	Constraints  string
	HasDeferred  bool
	Warnings     []string
}

// Frontmatter represents the YAML frontmatter of tasks.md.
type Frontmatter struct {
	Spec                string   `json:"spec"`
	TaskIDs             []string `json:"task_ids"`
	ApprovalID          string   `json:"approval_id"`
	GenerationStrategy  string   `json:"generation_strategy"`
}

// Task represents a single task parsed from the body of tasks.md.
type Task struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Status     string   `json:"status"` // "pending", "in_progress", "done"
	Wave       int      `json:"wave"`
	DependsOn  []string `json:"depends_on,omitempty"`
	Files      string   `json:"files,omitempty"`
	TDD        string   `json:"tdd,omitempty"`
	IsDeferred bool     `json:"is_deferred"`
	RawLine    int      `json:"raw_line"` // line number in original file (1-based)
}

// WavePlanEntry represents one line from the Wave Plan section.
type WavePlanEntry struct {
	Wave    int      `json:"wave"`
	TaskIDs []string `json:"task_ids"`
}

// NextWaveResult is returned by ResolveNextWave.
type NextWaveResult struct {
	Action        string   `json:"action"`         // "continue-wave", "execute", "blocked", "plan-next-wave", "done", "error"
	Wave          int      `json:"wave,omitempty"`
	TaskIDs       []string `json:"task_ids,omitempty"`
	DeferredReady []string `json:"deferred_ready,omitempty"`
	Reason        string   `json:"reason,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
}

// CountResult holds task count statistics.
type CountResult struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	InProgress int `json:"in_progress"`
	Done       int `json:"done"`
	Deferred   int `json:"deferred"`
}

// StateResult holds the full state of all tasks.
type StateResult struct {
	Spec   string      `json:"spec"`
	Tasks  []Task      `json:"tasks"`
	Counts CountResult `json:"counts"`
}

// FilesResult lists files for a specific task.
type FilesResult struct {
	TaskID string   `json:"task_id"`
	Files  []string `json:"files"`
}

// ValidateResult holds validation findings.
type ValidateResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ComplexityResult holds per-task complexity scoring.
type ComplexityResult struct {
	TaskID     string `json:"task_id"`
	Score      int    `json:"score"`
	ModelHint  string `json:"model_hint"` // "haiku", "sonnet", "opus"
	Factors    []string `json:"factors,omitempty"`
}
