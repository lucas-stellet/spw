// Package wave provides wave-level state resolution for spec-workflow execution phases.
// It scans wave directories, resolves checkpoint status (fixing stale _wave-summary.json bugs),
// generates summaries, and computes resume actions.
package wave

// WaveState represents the current state of a single wave directory.
type WaveState struct {
	WaveNum   int      `json:"wave_num"`
	Status    string   `json:"status"`     // "pending", "in_progress", "complete", "blocked"
	TaskIDs   []string `json:"task_ids"`
	ExecRuns  int      `json:"exec_runs"`
	CheckRuns int      `json:"check_runs"`
}

// WaveSummary captures the resolved summary for a wave, indicating the source
// of the resolution and whether stale data was detected.
type WaveSummary struct {
	Status    string `json:"status"`     // "pass", "blocked", "in_progress", "missing"
	Summary   string `json:"summary"`
	Source    string `json:"source"`     // "wave_summary", "latest_json", "checkpoint_scan", "none"
	StaleFlag bool   `json:"stale_flag"` // true if _wave-summary.json disagrees with latest run
}

// CheckpointResult captures the resolved checkpoint status for a wave.
// Resolution is _latest.json-first to avoid stale _wave-summary.json bugs.
type CheckpointResult struct {
	WaveNum   int    `json:"wave_num"`
	Status    string `json:"status"`    // "pass", "blocked", "missing", "no_runs"
	RunID     string `json:"run_id"`
	Source    string `json:"source"`    // "latest_json", "dir_scan"
	StaleFlag bool   `json:"stale_flag"`
	Details   string `json:"details,omitempty"`
}

// ResumeState indicates what action to take when resuming a spec's execution phase.
type ResumeState struct {
	Action  string `json:"action"`           // "continue-wave", "next-wave", "blocked", "done"
	WaveNum int    `json:"wave_num,omitempty"`
	Reason  string `json:"reason,omitempty"`
}
