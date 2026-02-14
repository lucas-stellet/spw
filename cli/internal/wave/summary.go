package wave

import (
	"path/filepath"

	"github.com/lucas-stellet/oraculo/internal/specdir"
)

// GenerateSummary creates a WaveSummary for a specific wave by reading
// subagent status.json files from the latest execution run.
func GenerateSummary(specDir string, waveNum int) WaveSummary {
	wavePath := specdir.WavePath(specDir, waveNum)
	if !specdir.DirExists(wavePath) {
		return WaveSummary{
			Status: "missing",
			Source: "none",
		}
	}

	// Try _wave-summary.json first
	summaryPath := specdir.WaveSummaryPath(specDir, waveNum)
	if specdir.FileExists(summaryPath) {
		doc, err := specdir.ReadStatusJSON(summaryPath)
		if err == nil {
			summary := WaveSummary{
				Status:  doc.Status,
				Summary: doc.Summary,
				Source:  "wave_summary",
			}

			// Cross-check with _latest.json for staleness
			latestPath := specdir.WaveLatestPath(specDir, waveNum)
			if specdir.FileExists(latestPath) {
				latest, err := specdir.ReadLatestJSON(latestPath)
				if err == nil && latest.Status != doc.Status {
					summary.StaleFlag = true
					// Prefer _latest.json status when stale
					summary.Status = latest.Status
					summary.Summary = latest.Summary
					summary.Source = "latest_json"
				}
			}

			return summary
		}
	}

	// Try _latest.json directly
	latestPath := specdir.WaveLatestPath(specDir, waveNum)
	if specdir.FileExists(latestPath) {
		latest, err := specdir.ReadLatestJSON(latestPath)
		if err == nil {
			return WaveSummary{
				Status:  latest.Status,
				Summary: latest.Summary,
				Source:  "latest_json",
			}
		}
	}

	// Fall back to scanning latest checkpoint run's subagent status
	checkDir := specdir.WaveCheckpointPath(specDir, waveNum)
	if specdir.DirExists(checkDir) {
		runDir, _, err := specdir.LatestRunDir(checkDir)
		if err == nil {
			status, summary := scanSubagentStatus(runDir)
			if status != "" {
				return WaveSummary{
					Status:  status,
					Summary: summary,
					Source:  "checkpoint_scan",
				}
			}
		}
	}

	// Fall back to scanning latest execution run's subagent status
	execDir := specdir.WaveExecPath(specDir, waveNum)
	if specdir.DirExists(execDir) {
		runDir, _, err := specdir.LatestRunDir(execDir)
		if err == nil {
			status, summary := scanSubagentStatus(runDir)
			if status != "" {
				return WaveSummary{
					Status:  status,
					Summary: summary,
					Source:  "checkpoint_scan",
				}
			}
		}
	}

	return WaveSummary{
		Status:  "in_progress",
		Summary: "wave exists but no summary data found",
		Source:  "none",
	}
}

// scanSubagentStatus reads all subagent status.json files in a run directory
// and returns an aggregate status. Returns ("", "") if no status files found.
func scanSubagentStatus(runDir string) (string, string) {
	entries, err := readDir(runDir)
	if err != nil {
		return "", ""
	}

	hasBlocked := false
	hasPass := false
	lastSummary := ""

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		statusPath := filepath.Join(runDir, e.Name(), specdir.StatusJSON)
		if !specdir.FileExists(statusPath) {
			continue
		}
		doc, err := specdir.ReadStatusJSON(statusPath)
		if err != nil {
			continue
		}
		lastSummary = doc.Summary
		switch doc.Status {
		case "blocked":
			hasBlocked = true
		case "pass":
			hasPass = true
		}
	}

	if hasBlocked {
		return "blocked", lastSummary
	}
	if hasPass {
		return "pass", lastSummary
	}
	if lastSummary != "" {
		return "in_progress", lastSummary
	}
	return "", ""
}
