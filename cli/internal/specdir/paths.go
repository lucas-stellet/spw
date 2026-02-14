// Package specdir provides canonical path constants and resolver functions
// for spec directory structure. This is the single source of truth for all
// path resolution — no command ever constructs paths ad-hoc.
package specdir

import (
	"fmt"
	"path/filepath"
)

// Dashboard files (spec root).
const (
	RequirementsMD = "requirements.md"
	DesignMD       = "design.md"
	TasksMD        = "tasks.md"
	StatusSummary  = "STATUS-SUMMARY.md"
)

// Phase directories.
const (
	PhaseDiscover   = "discover"
	PhaseDesign     = "design"
	PhasePlanning   = "planning"
	PhaseExecution  = "execution"
	PhaseQA         = "qa"
	PhasePostMortem = "post-mortem"
)

// Execution phase structure.
const (
	WavesDir         = "execution/waves"
	ImplLogsDir      = "execution/_implementation-logs"
	CheckpointReport = "execution/CHECKPOINT-REPORT.md"
	SkillsCheckpoint = "execution/SKILLS-CHECKPOINT.md"
)

// Wave internal structure.
const (
	WaveDirFmt        = "execution/waves/wave-%02d"
	WaveExecDir       = "execution"
	WaveCheckpointDir = "checkpoint"
	WaveSummaryJSON   = "_wave-summary.json"
	LatestJSON        = "_latest.json"
)

// Subagent handoff structure (inside any run-NNN/).
const (
	BriefMD    = "brief.md"
	ReportMD   = "report.md"
	StatusJSON = "status.json"
	HandoffMD  = "_handoff.md"
)

// Run directory format.
const RunDirFmt = "run-%03d"

// Implementation log file format.
const ImplLogFmt = "task-%s.md"

// Design phase.
const (
	DesignResearchMD     = "design/DESIGN-RESEARCH.md"
	SkillsDesignResearch = "design/SKILLS-DESIGN-RESEARCH.md"
	SkillsDesignDraft    = "design/SKILLS-DESIGN-DRAFT.md"
)

// Planning phase.
const (
	TasksCheckMD     = "planning/TASKS-CHECK.md"
	SkillsTasksPlan  = "planning/SKILLS-TASKS-PLAN.md"
	SkillsExec       = "planning/SKILLS-EXEC.md"
	SkillsTasksCheck = "planning/SKILLS-TASKS-CHECK.md"
)

// QA phase.
const (
	QATestPlan     = "qa/QA-TEST-PLAN.md"
	QACheckMD      = "qa/QA-CHECK.md"
	QAExecReport   = "qa/QA-EXECUTION-REPORT.md"
	QADefectReport = "qa/QA-DEFECT-REPORT.md"
	QAArtifactsDir = "qa/qa-artifacts"
)

// Post-mortem phase.
const PostMortemReport = "post-mortem/report.md"

// Store files.
const (
	SpecDB              = "spec.db"
	CompletionSummaryMD = "COMPLETION-SUMMARY.md"
	ProgressSummaryMD   = "PROGRESS-SUMMARY.md"
)

// _comms path patterns per command.
const (
	CommsDiscover       = "discover/_comms"
	CommsDesignResearch = "design/_comms/design-research"
	CommsDesignDraft    = "design/_comms/design-draft"
	CommsTasksPlan      = "planning/_comms/tasks-plan"
	CommsTasksCheck     = "planning/_comms/tasks-check"
	CommsQA             = "qa/_comms/qa"
	CommsQACheck        = "qa/_comms/qa-check"
	CommsQAExecFmt      = "qa/_comms/qa-exec/waves/wave-%02d"
	CommsPostMortem     = "post-mortem/_comms"
)

// ImplLogPath returns the canonical path for a task's implementation log.
func ImplLogPath(specDir, taskID string) string {
	return filepath.Join(specDir, ImplLogsDir, fmt.Sprintf(ImplLogFmt, taskID))
}

// WavePath returns the canonical path for a wave directory.
func WavePath(specDir string, waveNum int) string {
	return filepath.Join(specDir, fmt.Sprintf(WaveDirFmt, waveNum))
}

// WaveExecPath returns the path to a wave's execution subdir.
func WaveExecPath(specDir string, waveNum int) string {
	return filepath.Join(WavePath(specDir, waveNum), WaveExecDir)
}

// WaveCheckpointPath returns the path to a wave's checkpoint subdir.
func WaveCheckpointPath(specDir string, waveNum int) string {
	return filepath.Join(WavePath(specDir, waveNum), WaveCheckpointDir)
}

// CheckpointRunPath returns the path to a specific checkpoint run.
func CheckpointRunPath(specDir string, waveNum, runNum int) string {
	return filepath.Join(WaveCheckpointPath(specDir, waveNum), fmt.Sprintf(RunDirFmt, runNum))
}

// WaveSummaryPath returns the path to a wave's summary JSON.
func WaveSummaryPath(specDir string, waveNum int) string {
	return filepath.Join(WavePath(specDir, waveNum), WaveSummaryJSON)
}

// WaveLatestPath returns the path to a wave's _latest.json.
func WaveLatestPath(specDir string, waveNum int) string {
	return filepath.Join(WavePath(specDir, waveNum), LatestJSON)
}

// CommsPath returns the canonical _comms path for a given ORACULO command.
// For wave-aware commands, waveNum must be provided (>0).
func CommsPath(specDir, command string, waveNum int) string {
	switch command {
	case "discover":
		return filepath.Join(specDir, CommsDiscover)
	case "design-research":
		return filepath.Join(specDir, CommsDesignResearch)
	case "design-draft":
		return filepath.Join(specDir, CommsDesignDraft)
	case "tasks-plan":
		return filepath.Join(specDir, CommsTasksPlan)
	case "tasks-check":
		return filepath.Join(specDir, CommsTasksCheck)
	case "qa":
		return filepath.Join(specDir, CommsQA)
	case "qa-check":
		return filepath.Join(specDir, CommsQACheck)
	case "qa-exec":
		return filepath.Join(specDir, fmt.Sprintf(CommsQAExecFmt, waveNum))
	case "post-mortem":
		return filepath.Join(specDir, CommsPostMortem)
	case "exec":
		return filepath.Join(WaveExecPath(specDir, waveNum))
	case "checkpoint":
		return filepath.Join(WaveCheckpointPath(specDir, waveNum))
	default:
		return ""
	}
}

// TasksPath returns the canonical path to tasks.md for a spec.
func TasksPath(specDir string) string {
	return filepath.Join(specDir, TasksMD)
}

// SpecDir returns the relative spec directory path.
func SpecDir(specName string) string {
	return filepath.Join(".spec-workflow", "specs", specName)
}

// SpecDirAbs returns the absolute spec directory path.
func SpecDirAbs(cwd, specName string) string {
	return filepath.Join(cwd, SpecDir(specName))
}

// KnownDeviations maps common path mistakes to their canonical forms.
var KnownDeviations = map[string]string{
	"Implementation Logs":           ImplLogsDir,
	"implementation_logs":           ImplLogsDir,
	"implementation-logs":           ImplLogsDir,
	"research":                      "(removed from standard)",
	"_agent-comms":                  "(use phase-based _comms/ layout)",
	"execution/SKILLS-EXEC.md":      SkillsExec,
	"execution/SKILLS-CHECKPOINT.md": SkillsCheckpoint,
}
