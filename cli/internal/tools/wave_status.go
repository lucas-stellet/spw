package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucas-stellet/oraculo/internal/specdir"
)

type waveInfo struct {
	Wave       string `json:"wave"`
	Status     string `json:"status"`
	Tasks      []int  `json:"tasks"`
	Checkpoint string `json:"checkpoint"`
}

type taskInfo struct {
	ID     int
	Status string // "pending", "in-progress", "done", "blocked"
	Wave   int    // 0 if not assigned
}

type waveSummaryDoc struct {
	Wave      string `json:"wave"`
	Status    string `json:"status"`
	Tasks     []int  `json:"tasks"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

var taskLineRe = regexp.MustCompile(`(?m)^-\s+\[([ x\-!])\]\s+\*\*(\d+)\.\*\*`)

// waveStatusResult performs the wave status logic and returns the result.
// Extracted for testability — the public WaveStatus function wraps this with Output/Fail.
func waveStatusResult(cwd, specName string) (map[string]any, error) {
	specDirAbs := specdir.SpecDirAbs(cwd, specName)

	if !specdir.DirExists(specDirAbs) {
		return nil, fmt.Errorf("spec directory not found: %s", specdir.SpecDir(specName))
	}

	// List wave directories
	waveDirs, err := specdir.ListWaveDirs(specDirAbs)
	if err != nil {
		return nil, fmt.Errorf("failed to list wave directories: %w", err)
	}

	// Build wave info for each existing wave directory
	var waves []waveInfo
	for _, wd := range waveDirs {
		wi := waveInfo{
			Wave: fmt.Sprintf("%02d", wd.Num),
		}

		// Read wave summary
		summaryPath := specdir.WaveSummaryPath(specDirAbs, wd.Num)
		if specdir.FileExists(summaryPath) {
			data, err := os.ReadFile(summaryPath)
			if err == nil {
				var doc waveSummaryDoc
				if err := json.Unmarshal(data, &doc); err == nil {
					wi.Status = doc.Status
					wi.Tasks = doc.Tasks
				}
			}
		}
		if wi.Status == "" {
			wi.Status = "not-started"
		}
		if wi.Tasks == nil {
			wi.Tasks = []int{}
		}

		// Read latest JSON for checkpoint status
		latestPath := specdir.WaveLatestPath(specDirAbs, wd.Num)
		if specdir.FileExists(latestPath) {
			doc, err := specdir.ReadLatestJSON(latestPath)
			if err == nil {
				wi.Checkpoint = doc.Status
			}
		}
		if wi.Checkpoint == "" {
			wi.Checkpoint = "missing"
		}

		waves = append(waves, wi)
	}

	// Parse tasks.md
	tasks := parseTasks(specDirAbs)
	totalTasks := len(tasks)
	completedTasks := 0
	var inProgressTasks []int
	for _, t := range tasks {
		if t.Status == "done" {
			completedTasks++
		}
		if t.Status == "in-progress" {
			inProgressTasks = append(inProgressTasks, t.ID)
		}
	}
	if inProgressTasks == nil {
		inProgressTasks = []int{}
	}

	// Determine current state
	var currentWave string
	var waveStatus string
	var checkpointStatus string
	var resumeAction string
	var nextTasks []int

	if len(waves) == 0 {
		currentWave = "01"
		waveStatus = "not-started"
		checkpointStatus = "missing"
		resumeAction = "start-wave"
		nextTasks = pendingTasksForWave(tasks, 1)
		if len(nextTasks) == 0 {
			nextTasks = allPendingTasks(tasks)
		}
	} else {
		latest := waves[len(waves)-1]
		currentWave = latest.Wave
		checkpointStatus = latest.Checkpoint

		// Determine wave status and resume action based on latest wave state
		if latest.Checkpoint == "pass" {
			// Wave completed — check if there are still pending tasks
			waveStatus = "completed"

			// Update the status field on the latest wave info
			if waves[len(waves)-1].Status != "pass" {
				waves[len(waves)-1].Status = "completed"
			}

			pendingExist := false
			for _, t := range tasks {
				if t.Status == "pending" {
					pendingExist = true
					break
				}
			}

			if !pendingExist && len(inProgressTasks) == 0 {
				resumeAction = "done"
			} else {
				// Next wave
				latestNum, _ := strconv.Atoi(currentWave)
				nextNum := latestNum + 1
				currentWave = fmt.Sprintf("%02d", nextNum)
				waveStatus = "not-started"
				checkpointStatus = "missing"
				resumeAction = "start-wave"
				nextTasks = pendingTasksForWave(tasks, nextNum)
				if len(nextTasks) == 0 {
					nextTasks = allPendingTasks(tasks)
				}
			}
		} else if latest.Checkpoint == "blocked" {
			waveStatus = "blocked"
			resumeAction = "wait-checkpoint"
			nextTasks = latest.Tasks
		} else {
			// No checkpoint or checkpoint missing — wave is in progress
			waveStatus = "in-progress"
			if len(inProgressTasks) > 0 {
				resumeAction = "resume-in-progress"
			} else {
				resumeAction = "continue-wave"
			}
			nextTasks = pendingTasksForWave(tasks, mustAtoi(currentWave))
			if len(nextTasks) == 0 {
				nextTasks = latest.Tasks
			}
		}
	}

	if nextTasks == nil {
		nextTasks = []int{}
	}

	return map[string]any{
		"ok":                true,
		"spec":              specName,
		"waves":             waves,
		"current_wave":      currentWave,
		"wave_status":       waveStatus,
		"checkpoint_status": checkpointStatus,
		"next_tasks":        nextTasks,
		"in_progress_tasks": inProgressTasks,
		"resume_action":     resumeAction,
		"total_tasks":       totalTasks,
		"completed_tasks":   completedTasks,
	}, nil
}

// parseTasks reads tasks.md and extracts task info with wave assignments.
func parseTasks(specDirAbs string) []taskInfo {
	tasksPath := specdir.TasksPath(specDirAbs)
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var tasks []taskInfo
	for i, line := range lines {
		m := taskLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		marker := m[1]
		id, _ := strconv.Atoi(m[2])

		var status string
		switch marker {
		case " ":
			status = "pending"
		case "x":
			status = "done"
		case "-":
			status = "in-progress"
		case "!":
			status = "blocked"
		default:
			status = "pending"
		}

		wave := 0
		// Scan subsequent indented lines for Wave: NN
		for j := i + 1; j < len(lines); j++ {
			l := lines[j]
			// Stop at next task line or non-indented line (except blank)
			if len(l) > 0 && l[0] == '-' {
				break
			}
			if len(l) > 0 && l[0] != ' ' && l[0] != '\t' && l[0] != '#' {
				// Section header or other non-indented line — could be a wave header
				// Don't break on "## Wave NN" headers since they precede tasks
				if !strings.HasPrefix(l, "##") {
					break
				}
			}
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "- Wave:") || strings.HasPrefix(trimmed, "Wave:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					wn := strings.TrimSpace(parts[1])
					if n, err := strconv.Atoi(wn); err == nil {
						wave = n
					}
				}
				break
			}
		}

		tasks = append(tasks, taskInfo{ID: id, Status: status, Wave: wave})
	}

	return tasks
}

// pendingTasksForWave returns IDs of pending tasks assigned to a specific wave.
func pendingTasksForWave(tasks []taskInfo, waveNum int) []int {
	var result []int
	for _, t := range tasks {
		if t.Wave == waveNum && t.Status == "pending" {
			result = append(result, t.ID)
		}
	}
	return result
}

// allPendingTasks returns IDs of all pending tasks.
func allPendingTasks(tasks []taskInfo) []int {
	var result []int
	for _, t := range tasks {
		if t.Status == "pending" {
			result = append(result, t.ID)
		}
	}
	return result
}

func mustAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// WaveStatus provides comprehensive wave state resolution for a spec.
func WaveStatus(cwd, specName string, raw bool) {
	if specName == "" {
		Fail("wave-status requires <spec-name>", raw)
	}

	result, err := waveStatusResult(cwd, specName)
	if err != nil {
		Fail(err.Error(), raw)
	}

	currentWave, _ := result["current_wave"].(string)
	Output(result, currentWave, raw)
}
