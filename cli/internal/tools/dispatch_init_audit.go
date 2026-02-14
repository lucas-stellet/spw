package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// dispatchInitAuditResult holds the result for testability.
type dispatchInitAuditResult struct {
	OK        bool   `json:"ok"`
	AuditDir  string `json:"audit_dir"`
	Type      string `json:"type"`
	Iteration int    `json:"iteration"`
}

// dispatchInitAuditCore contains the testable logic for DispatchInitAudit.
func dispatchInitAuditCore(cwd, runDir, auditType string, iteration int) (*dispatchInitAuditResult, error) {
	if runDir == "" {
		return nil, fmt.Errorf("dispatch-init-audit requires <run-dir>")
	}
	if auditType == "" {
		return nil, fmt.Errorf("dispatch-init-audit requires <audit-type>")
	}
	if auditType != "inline-audit" && auditType != "inline-checkpoint" {
		return nil, fmt.Errorf("audit-type must be 'inline-audit' or 'inline-checkpoint', got %q", auditType)
	}

	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	if iteration <= 0 {
		iteration = 1
	}

	var targetDir string
	if auditType == "inline-audit" {
		targetDir = filepath.Join(runDir, "_inline-audit", fmt.Sprintf("iteration-%d", iteration))
	} else {
		targetDir = filepath.Join(runDir, "_inline-checkpoint")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit dir: %w", err)
	}

	rel, _ := filepath.Rel(cwd, targetDir)

	return &dispatchInitAuditResult{
		OK:        true,
		AuditDir:  rel,
		Type:      auditType,
		Iteration: iteration,
	}, nil
}

// DispatchInitAudit creates an audit subdirectory within a run directory.
func DispatchInitAudit(cwd, runDir, auditType string, iteration int, raw bool) {
	result, err := dispatchInitAuditCore(cwd, runDir, auditType, iteration)
	if err != nil {
		Fail(err.Error(), raw)
	}
	Output(result, result.AuditDir, raw)
}
