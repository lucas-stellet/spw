package tools

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var waveNumRe = regexp.MustCompile(`^wave-(\d+)$`)

// WaveResolveCurrent finds the highest wave directory for a spec.
func WaveResolveCurrent(cwd, specName string, raw bool) {
	if specName == "" {
		Fail("wave-resolve-current requires <spec-name>", raw)
	}

	// Check new phase-based layout first
	wavesDir := filepath.Join(cwd, ".spec-workflow", "specs", specName, "execution", "waves")
	if _, err := os.Stat(wavesDir); os.IsNotExist(err) {
		// Legacy layout
		wavesDir = filepath.Join(cwd, ".spec-workflow", "specs", specName, "_agent-comms", "waves")
	}

	if _, err := os.Stat(wavesDir); os.IsNotExist(err) {
		result := map[string]any{"ok": true, "spec": specName, "found": false, "wave": nil, "directory": nil}
		Output(result, "none", raw)
		return
	}

	entries, err := os.ReadDir(wavesDir)
	if err != nil {
		result := map[string]any{"ok": true, "spec": specName, "found": false, "wave": nil, "directory": nil}
		Output(result, "none", raw)
		return
	}

	type wave struct {
		name string
		num  int
		full string
	}
	var waves []wave
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m := waveNumRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		waves = append(waves, wave{name: e.Name(), num: n, full: filepath.Join(wavesDir, e.Name())})
	}

	sort.Slice(waves, func(i, j int) bool { return waves[i].num > waves[j].num })

	if len(waves) == 0 {
		result := map[string]any{"ok": true, "spec": specName, "found": false, "wave": nil, "directory": nil}
		Output(result, "none", raw)
		return
	}

	current := waves[0]
	rel, _ := filepath.Rel(cwd, current.full)
	result := map[string]any{"ok": true, "spec": specName, "found": true, "wave": current.name, "directory": rel}
	Output(result, current.name, raw)
}
