package wave

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/lucas-stellet/spw/internal/specdir"
)

var runNumRe = regexp.MustCompile(`^run-(\d+)$`)

// releaseGateDecider is the canonical subagent name for checkpoint runs.
const releaseGateDecider = "release-gate-decider"

// ResolveCheckpoint determines the checkpoint status for a wave.
// Resolution order:
//  1. Read _latest.json -> get latest checkpoint run-id
//  2. Read checkpoint/{run-id}/release-gate-decider/status.json
//  3. If _latest.json differs from _wave-summary.json -> flag stale_summary
//  4. If _latest.json missing -> scan checkpoint dir, highest run wins
//  5. If no checkpoint dirs -> status: "missing"
func ResolveCheckpoint(specDir string, waveNum int) CheckpointResult {
	checkDir := specdir.WaveCheckpointPath(specDir, waveNum)

	// If checkpoint dir doesn't exist at all
	if !specdir.DirExists(checkDir) {
		return CheckpointResult{
			WaveNum: waveNum,
			Status:  "missing",
			Source:  "dir_scan",
		}
	}

	// Step 1: Try _latest.json first (authoritative source)
	latestPath := specdir.WaveLatestPath(specDir, waveNum)
	if specdir.FileExists(latestPath) {
		latest, err := specdir.ReadLatestJSON(latestPath)
		if err == nil && latest.RunID != "" {
			result := CheckpointResult{
				WaveNum: waveNum,
				RunID:   latest.RunID,
				Source:  "latest_json",
			}

			// Step 2: Read the actual run's status.json for ground truth
			runStatus := readCheckpointRunStatus(checkDir, latest.RunID)
			if runStatus != "" {
				result.Status = runStatus
			} else {
				// Fall back to _latest.json's own status field
				result.Status = latest.Status
			}

			// Step 3: Check staleness against _wave-summary.json
			summaryPath := specdir.WaveSummaryPath(specDir, waveNum)
			if specdir.FileExists(summaryPath) {
				summaryDoc, err := specdir.ReadStatusJSON(summaryPath)
				if err == nil && summaryDoc.Status != result.Status {
					result.StaleFlag = true
					result.Details = "wave-summary says " + summaryDoc.Status + " but latest run says " + result.Status
				}
			}

			return result
		}
	}

	// Step 4: No _latest.json -> scan checkpoint dir, highest run wins
	runDir, _, err := specdir.LatestRunDir(checkDir)
	if err != nil {
		return CheckpointResult{
			WaveNum: waveNum,
			Status:  "no_runs",
			Source:  "dir_scan",
		}
	}

	runID := filepath.Base(runDir)
	status := readRunSubagentStatus(runDir)

	result := CheckpointResult{
		WaveNum: waveNum,
		Status:  status,
		RunID:   runID,
		Source:  "dir_scan",
	}

	if status == "" {
		result.Status = "no_runs"
		result.Details = "checkpoint run exists but no status.json found"
	}

	return result
}

// readCheckpointRunStatus reads the release-gate-decider's status.json for a specific run.
// Falls back to scanning all subagent dirs if release-gate-decider doesn't exist.
func readCheckpointRunStatus(checkDir, runID string) string {
	runPath := filepath.Join(checkDir, runID)
	if !specdir.DirExists(runPath) {
		return ""
	}

	// Try canonical subagent first
	statusPath := filepath.Join(runPath, releaseGateDecider, specdir.StatusJSON)
	if specdir.FileExists(statusPath) {
		doc, err := specdir.ReadStatusJSON(statusPath)
		if err == nil {
			return doc.Status
		}
	}

	// Fall back to scanning all subagent dirs
	return readRunSubagentStatus(runPath)
}

// readRunSubagentStatus scans all subagent directories in a run for any status.json.
func readRunSubagentStatus(runDir string) string {
	entries, err := os.ReadDir(runDir)
	if err != nil {
		return ""
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		statusPath := filepath.Join(runDir, e.Name(), specdir.StatusJSON)
		if specdir.FileExists(statusPath) {
			doc, err := specdir.ReadStatusJSON(statusPath)
			if err == nil {
				return doc.Status
			}
		}
	}

	return ""
}

// readDir is a helper to read directory entries, used by scanner and summary.
func readDir(dir string) ([]os.DirEntry, error) {
	return os.ReadDir(dir)
}
