package wave

import (
	"github.com/lucas-stellet/spw/internal/specdir"
)

// ScanWaves reads all wave directories and returns their states.
// Uses specdir.ListWaveDirs() to enumerate, then inspects each wave's
// execution/ and checkpoint/ subdirs to determine run counts and status.
func ScanWaves(specDir string) ([]WaveState, error) {
	waveDirs, err := specdir.ListWaveDirs(specDir)
	if err != nil {
		return nil, err
	}

	if len(waveDirs) == 0 {
		return nil, nil
	}

	var states []WaveState
	for _, wd := range waveDirs {
		ws := WaveState{
			WaveNum: wd.Num,
		}

		// Count execution runs
		execPath := specdir.WaveExecPath(specDir, wd.Num)
		ws.ExecRuns = countRunDirs(execPath)

		// Count checkpoint runs
		checkPath := specdir.WaveCheckpointPath(specDir, wd.Num)
		ws.CheckRuns = countRunDirs(checkPath)

		// Determine status based on runs and checkpoint result
		ws.Status = classifyWaveStatus(specDir, wd.Num, ws.ExecRuns, ws.CheckRuns)

		states = append(states, ws)
	}

	return states, nil
}

// classifyWaveStatus determines wave status from run counts and checkpoint result.
func classifyWaveStatus(specDir string, waveNum, execRuns, checkRuns int) string {
	if execRuns == 0 && checkRuns == 0 {
		return "pending"
	}

	if checkRuns > 0 {
		cp := ResolveCheckpoint(specDir, waveNum)
		switch cp.Status {
		case "pass":
			return "complete"
		case "blocked":
			return "blocked"
		}
	}

	if execRuns > 0 {
		return "in_progress"
	}

	return "pending"
}

// countRunDirs counts run-NNN directories in a given path.
func countRunDirs(dir string) int {
	if !specdir.DirExists(dir) {
		return 0
	}
	_, _, err := specdir.LatestRunDir(dir)
	if err != nil {
		return 0
	}
	// LatestRunDir returns the max run num; count all by reading the dir
	return countRunEntries(dir)
}

// countRunEntries counts directories matching run-NNN pattern.
func countRunEntries(dir string) int {
	entries, err := readDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() && runNumRe.MatchString(e.Name()) {
			count++
		}
	}
	return count
}
