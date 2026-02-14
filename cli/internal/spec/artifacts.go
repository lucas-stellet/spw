package spec

import (
	"os"
	"path/filepath"

	"github.com/lucas-stellet/oraculo/internal/specdir"
)

// knownArtifacts lists all canonical artifact paths relative to the spec directory.
var knownArtifacts = []string{
	// Dashboard files
	specdir.RequirementsMD,
	specdir.DesignMD,
	specdir.TasksMD,
	specdir.StatusSummary,

	// Design phase
	specdir.DesignResearchMD,
	specdir.SkillsDesignResearch,
	specdir.SkillsDesignDraft,

	// Planning phase
	specdir.TasksCheckMD,
	specdir.SkillsTasksPlan,
	specdir.SkillsExec,
	specdir.SkillsTasksCheck,

	// Execution phase
	specdir.CheckpointReport,
	specdir.SkillsCheckpoint,

	// QA phase
	specdir.QATestPlan,
	specdir.QACheckMD,
	specdir.QAExecReport,
	specdir.QADefectReport,

	// Post-mortem phase
	specdir.PostMortemReport,
}

// CheckArtifacts scans a spec directory for expected artifacts and returns
// which ones exist. Also detects known deviations (wrong paths).
func CheckArtifacts(specDir string) (map[string]bool, []Deviation) {
	artifacts := make(map[string]bool, len(knownArtifacts))

	for _, rel := range knownArtifacts {
		full := filepath.Join(specDir, rel)
		artifacts[rel] = specdir.FileExists(full)
	}

	deviations := detectDeviations(specDir)

	return artifacts, deviations
}

// detectDeviations checks for files at known non-canonical paths.
func detectDeviations(specDir string) []Deviation {
	var deviations []Deviation

	for deviated, canonical := range specdir.KnownDeviations {
		// Check if the deviated path exists as a file or directory
		full := filepath.Join(specDir, deviated)
		if fileOrDirExists(full) {
			deviations = append(deviations, Deviation{
				Found:     deviated,
				Canonical: canonical,
			})
		}
	}

	return deviations
}

// fileOrDirExists checks if a path exists (file or directory).
func fileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
