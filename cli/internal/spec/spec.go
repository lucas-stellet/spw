// Package spec provides spec-level state resolution for spec-workflow.
// It inspects spec directories to determine lifecycle stage, check artifacts,
// verify prerequisites, and scan approval records.
package spec

// SpecInfo captures the resolved state of a spec directory.
type SpecInfo struct {
	Name       string          `json:"name"`
	Dir        string          `json:"dir"`
	Stage      string          `json:"stage"`                  // "requirements", "design", "planning", "execution", "qa", "post-mortem", "complete"
	Artifacts  map[string]bool `json:"artifacts"`              // artifact path -> exists
	Deviations []Deviation     `json:"deviations,omitempty"`
}

// Deviation records a file found at a non-canonical path.
type Deviation struct {
	Found     string `json:"found"`
	Canonical string `json:"canonical"`
}

// PrereqResult indicates whether prerequisites are met for a command.
type PrereqResult struct {
	Ready   bool     `json:"ready"`
	Missing []string `json:"missing,omitempty"`
}

// ApprovalResult captures the result of scanning for a local approval record.
type ApprovalResult struct {
	DocType    string `json:"doc_type"`
	ApprovalID string `json:"approval_id,omitempty"`
	Source     string `json:"source,omitempty"` // file path
	Found      bool   `json:"found"`
}
