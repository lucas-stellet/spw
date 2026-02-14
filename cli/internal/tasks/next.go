package tasks

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/lucas-stellet/oraculo/internal/specdir"
)

// ResolveNextWave implements the critical next-wave algorithm:
//  1. Check for in-progress tasks [-] -> return "continue-wave"
//  2. Find highest completed wave (all tasks [x])
//  3. Check checkpoint status via _latest.json-first resolution
//  4. If checkpoint blocked -> return "blocked"
//  5. Find next wave's pending tasks with deps resolved
//  6. Scan ALL deferred tasks whose deps are [x] -> include in DeferredReady
//  7. If executable tasks found -> return "execute"
//  8. If nothing left + rolling-wave -> return "plan-next-wave"
//  9. If nothing left + all-at-once -> return "done"
func ResolveNextWave(doc Document, specDir string) NextWaveResult {
	if len(doc.Tasks) == 0 {
		return NextWaveResult{
			Action: "done",
			Reason: "no tasks found",
		}
	}

	var warnings []string
	warnings = append(warnings, doc.Warnings...)

	// Build status lookup map
	statusByID := make(map[string]string, len(doc.Tasks))
	for _, t := range doc.Tasks {
		statusByID[t.ID] = t.Status
	}

	// Step 1: Check for in-progress tasks
	for _, t := range doc.Tasks {
		if t.Status == "in_progress" {
			return NextWaveResult{
				Action:   "continue-wave",
				Wave:     t.Wave,
				Reason:   fmt.Sprintf("task %s is in progress", t.ID),
				Warnings: warnings,
			}
		}
	}

	// Build wave->tasks mapping (non-deferred only)
	waveMap := buildWaveMap(doc.Tasks)

	// Step 2: Find highest completed wave
	highestCompleted := findHighestCompletedWave(waveMap)

	// Step 3-4: Check checkpoint status for highest completed wave
	if highestCompleted > 0 && specDir != "" {
		cpStatus, cpWarnings := resolveCheckpointStatus(specDir, highestCompleted)
		warnings = append(warnings, cpWarnings...)

		if cpStatus == "blocked" {
			return NextWaveResult{
				Action:   "blocked",
				Wave:     highestCompleted,
				Reason:   fmt.Sprintf("wave %d checkpoint is blocked", highestCompleted),
				Warnings: warnings,
			}
		}
	}

	// Step 5: Find next wave's pending tasks with resolved deps
	nextWave, executableIDs := findExecutableTasks(doc.Tasks, statusByID, waveMap, highestCompleted)

	// Step 6: Scan ALL deferred tasks whose deps are resolved
	deferredReady := findDeferredReady(doc.Tasks, statusByID)

	// Step 7: If executable tasks found -> return "execute"
	if len(executableIDs) > 0 || len(deferredReady) > 0 {
		return NextWaveResult{
			Action:        "execute",
			Wave:          nextWave,
			TaskIDs:       executableIDs,
			DeferredReady: deferredReady,
			Warnings:      warnings,
		}
	}

	// Step 8-9: Nothing executable left
	if doc.Frontmatter.GenerationStrategy == "rolling-wave" {
		return NextWaveResult{
			Action:   "plan-next-wave",
			Reason:   "all planned tasks complete, rolling-wave strategy allows planning next wave",
			Warnings: warnings,
		}
	}

	return NextWaveResult{
		Action:   "done",
		Reason:   "all tasks complete",
		Warnings: warnings,
	}
}

// buildWaveMap groups non-deferred tasks by wave number.
func buildWaveMap(tasks []Task) map[int][]Task {
	m := make(map[int][]Task)
	for _, t := range tasks {
		if !t.IsDeferred {
			m[t.Wave] = append(m[t.Wave], t)
		}
	}
	return m
}

// findHighestCompletedWave returns the highest wave number where all
// non-deferred tasks have status == "done". Returns 0 if no wave is complete.
func findHighestCompletedWave(waveMap map[int][]Task) int {
	// Get sorted wave numbers
	waves := make([]int, 0, len(waveMap))
	for w := range waveMap {
		waves = append(waves, w)
	}
	sort.Ints(waves)

	highest := 0
	for _, w := range waves {
		allDone := true
		for _, t := range waveMap[w] {
			if t.Status != "done" {
				allDone = false
				break
			}
		}
		if allDone {
			highest = w
		} else {
			break // waves are sequential; if one isn't done, higher ones aren't either
		}
	}
	return highest
}

// resolveCheckpointStatus resolves checkpoint status using _latest.json-first resolution:
//  1. Read _latest.json -> get latest checkpoint run-id
//  2. Read checkpoint/{run-id}/release-gate-decider/status.json
//  3. If differs from _wave-summary.json -> flag stale_summary
//  4. If _latest.json missing -> scan checkpoint dirs, highest run wins
//  5. If no checkpoint dirs -> return "missing"
func resolveCheckpointStatus(specDir string, waveNum int) (string, []string) {
	var warnings []string

	cpDir := specdir.WaveCheckpointPath(specDir, waveNum)
	latestPath := specdir.WaveLatestPath(specDir, waveNum)
	summaryPath := specdir.WaveSummaryPath(specDir, waveNum)

	var runDir string
	var resolvedFromLatest bool

	// Step 1: Try _latest.json first
	latestDoc, err := specdir.ReadLatestJSON(latestPath)
	if err == nil && latestDoc.RunID != "" {
		// Build the run path from the run ID
		candidate := filepath.Join(cpDir, latestDoc.RunID)
		if specdir.DirExists(candidate) {
			runDir = candidate
			resolvedFromLatest = true
		}
	}

	// Step 4: Fallback — scan checkpoint dir for highest run
	if runDir == "" {
		if specdir.DirExists(cpDir) {
			found, _, scanErr := specdir.LatestRunDir(cpDir)
			if scanErr == nil {
				runDir = found
			}
		}
	}

	// Step 5: No checkpoint at all
	if runDir == "" {
		return "missing", warnings
	}

	// Step 2: Read release-gate-decider status.json
	statusPath := filepath.Join(runDir, "release-gate-decider", specdir.StatusJSON)
	statusDoc, err := specdir.ReadStatusJSON(statusPath)
	if err != nil {
		return "missing", append(warnings, fmt.Sprintf("cannot read checkpoint status: %v", err))
	}

	// Step 3: Compare with _wave-summary.json if resolved from _latest.json
	if resolvedFromLatest {
		summaryDoc, summaryErr := specdir.ReadStatusJSON(summaryPath)
		if summaryErr == nil && summaryDoc.Status != statusDoc.Status {
			warnings = append(warnings, fmt.Sprintf(
				"stale_summary: wave %d _wave-summary.json says %q but latest checkpoint says %q",
				waveNum, summaryDoc.Status, statusDoc.Status,
			))
		}
	}

	return statusDoc.Status, warnings
}

// findExecutableTasks finds the next wave after highestCompleted and returns
// tasks within it whose dependencies are all resolved (status == "done").
func findExecutableTasks(tasks []Task, statusByID map[string]string, waveMap map[int][]Task, highestCompleted int) (int, []string) {
	// Find the lowest wave number higher than highestCompleted that has pending tasks
	waves := make([]int, 0, len(waveMap))
	for w := range waveMap {
		waves = append(waves, w)
	}
	sort.Ints(waves)

	for _, w := range waves {
		if w <= highestCompleted {
			continue
		}

		// Collect tasks in this wave whose deps are resolved
		var ids []string
		for _, t := range waveMap[w] {
			if t.Status != "pending" {
				continue
			}
			if depsResolved(t, statusByID) {
				ids = append(ids, t.ID)
			}
		}

		if len(ids) > 0 {
			return w, ids
		}
	}

	return 0, nil
}

// findDeferredReady returns IDs of deferred tasks whose deps are all resolved.
func findDeferredReady(tasks []Task, statusByID map[string]string) []string {
	var ready []string
	for _, t := range tasks {
		if t.IsDeferred && t.Status == "pending" && depsResolved(t, statusByID) {
			ready = append(ready, t.ID)
		}
	}
	return ready
}

// depsResolved returns true if all of a task's dependencies have status == "done".
func depsResolved(t Task, statusByID map[string]string) bool {
	for _, dep := range t.DependsOn {
		if statusByID[dep] != "done" {
			return false
		}
	}
	return true
}
