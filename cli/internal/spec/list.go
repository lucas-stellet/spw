package spec

import (
	"os"
	"path/filepath"
	"sort"
)

// List returns all spec directories found in .spec-workflow/specs/.
func List(cwd string) ([]string, error) {
	specsDir := filepath.Join(cwd, ".spec-workflow", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}

	sort.Strings(names)
	return names, nil
}
