package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lucas-stellet/oraculo/internal/specdir"
)

// waveUpdateResult performs the wave update logic and returns the result.
// Extracted for testability — the public WaveUpdate function wraps this with Output/Fail.
func waveUpdateResult(cwd, specName, wave, status, tasks, checkpointRun, executionRun string) (map[string]any, error) {
	specDirAbs := specdir.SpecDirAbs(cwd, specName)

	waveNum, err := strconv.Atoi(wave)
	if err != nil {
		return nil, fmt.Errorf("wave must be a number: %s", wave)
	}

	// Parse tasks as comma-separated list of ints
	taskParts := strings.Split(tasks, ",")
	taskNums := make([]int, 0, len(taskParts))
	for _, part := range taskParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid task number %q: %w", part, err)
		}
		taskNums = append(taskNums, n)
	}

	// Create wave directory
	waveDir := specdir.WavePath(specDirAbs, waveNum)
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create wave directory: %w", err)
	}

	// Build and write summary JSON
	summaryPath := specdir.WaveSummaryPath(specDirAbs, waveNum)
	summary := map[string]any{
		"wave":       fmt.Sprintf("%02d", waveNum),
		"status":     status,
		"tasks":      taskNums,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeJSONFile(summaryPath, summary); err != nil {
		return nil, fmt.Errorf("failed to write wave summary: %w", err)
	}

	// Build and write latest JSON
	latestPath := specdir.WaveLatestPath(specDirAbs, waveNum)
	latest := map[string]any{
		"status": status,
	}
	if executionRun != "" {
		latest["execution"] = executionRun
	}
	if checkpointRun != "" {
		latest["checkpoint"] = checkpointRun
	}
	if err := writeJSONFile(latestPath, latest); err != nil {
		return nil, fmt.Errorf("failed to write wave latest: %w", err)
	}

	summaryRel, _ := filepath.Rel(cwd, summaryPath)
	latestRel, _ := filepath.Rel(cwd, latestPath)

	return map[string]any{
		"ok":           true,
		"wave":         fmt.Sprintf("%02d", waveNum),
		"summary_path": summaryRel,
		"latest_path":  latestRel,
	}, nil
}

// writeJSONFile marshals v as indented JSON and writes it to path.
func writeJSONFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// WaveUpdate writes wave summary and latest JSON files for a spec wave.
func WaveUpdate(cwd, specName, wave, status, tasks, checkpointRun, executionRun string, raw bool) {
	if specName == "" || wave == "" || status == "" || tasks == "" {
		Fail("wave-update requires --spec, --wave, --status, and --tasks", raw)
	}

	validStatuses := map[string]bool{"pass": true, "blocked": true}
	if !validStatuses[status] {
		Fail("status must be one of: pass, blocked", raw)
	}

	result, err := waveUpdateResult(cwd, specName, wave, status, tasks, checkpointRun, executionRun)
	if err != nil {
		Fail(err.Error(), raw)
	}

	summaryPath, _ := result["summary_path"].(string)
	Output(result, summaryPath, raw)
}
