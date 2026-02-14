package tools

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var wavePathRe = regexp.MustCompile(`wave-(\d+)`)

// resolveSpecDirFromRunDir walks up from a run directory to find the spec root.
// The spec root is identified by the presence of tasks.md or requirements.md.
func resolveSpecDirFromRunDir(runDir string) string {
	dir := runDir
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "tasks.md")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "requirements.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// extractWaveFromPath extracts wave number from a path containing "wave-NN".
func extractWaveFromPath(path string) int {
	m := wavePathRe.FindStringSubmatch(path)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// extractCommandFromRunDir infers the command name from the run directory path.
// The _comms directory structure is: <phase>/_comms/<command>/run-NNN
// So the command is the parent of the run-NNN directory.
func extractCommandFromRunDir(runDir string) string {
	// runDir = .../specDir/<phase>/_comms/<command>/run-NNN
	// parent of runDir is the command directory
	return filepath.Base(filepath.Dir(runDir))
}
