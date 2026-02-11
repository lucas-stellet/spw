package specdir

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var (
	waveNumRe = regexp.MustCompile(`^wave-(\d+)$`)
	runNumRe  = regexp.MustCompile(`^run-(\d+)$`)
)

// Resolve returns the absolute spec directory, ensuring it exists.
// Returns ("", error) if the directory does not exist.
func Resolve(cwd, specName string) (string, error) {
	abs := SpecDirAbs(cwd, specName)
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("spec directory not found: %s", SpecDir(specName))
	}
	return abs, nil
}

// ListWaveDirs returns wave directories sorted by wave number (ascending).
func ListWaveDirs(specDir string) ([]WaveDirEntry, error) {
	wavesPath := filepath.Join(specDir, WavesDir)
	entries, err := os.ReadDir(wavesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var waves []WaveDirEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m := waveNumRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		waves = append(waves, WaveDirEntry{
			Num:  n,
			Name: e.Name(),
			Path: filepath.Join(wavesPath, e.Name()),
		})
	}

	sort.Slice(waves, func(i, j int) bool { return waves[i].Num < waves[j].Num })
	return waves, nil
}

// WaveDirEntry represents a wave directory on disk.
type WaveDirEntry struct {
	Num  int
	Name string
	Path string
}

// ReadStatusJSON reads and unmarshals a status.json file.
func ReadStatusJSON(path string) (StatusDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return StatusDoc{}, err
	}
	var doc StatusDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return StatusDoc{}, fmt.Errorf("invalid status.json at %s: %w", path, err)
	}
	return doc, nil
}

// StatusDoc represents a subagent status.json file.
type StatusDoc struct {
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

// LatestDoc represents a _latest.json file that points to the latest run.
type LatestDoc struct {
	RunID   string `json:"run_id"`
	RunDir  string `json:"run_dir"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

// ReadLatestJSON reads and unmarshals a _latest.json file.
func ReadLatestJSON(path string) (LatestDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LatestDoc{}, err
	}
	var doc LatestDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return LatestDoc{}, fmt.Errorf("invalid _latest.json at %s: %w", path, err)
	}
	return doc, nil
}

// LatestRunDir finds the highest-numbered run-NNN directory in the given path.
func LatestRunDir(dir string) (string, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", 0, err
	}

	maxNum := 0
	maxName := ""
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m := runNumRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		if n > maxNum {
			maxNum = n
			maxName = e.Name()
		}
	}

	if maxName == "" {
		return "", 0, fmt.Errorf("no run directories found in %s", dir)
	}

	return filepath.Join(dir, maxName), maxNum, nil
}

// FileExists checks if a file exists (not a directory).
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
