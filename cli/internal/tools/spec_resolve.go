package tools

import (
	"os"
	"path/filepath"
)

// SpecResolveDir resolves the spec directory path.
func SpecResolveDir(cwd, specName string, raw bool) {
	if specName == "" {
		Fail("spec-resolve-dir requires <spec-name>", raw)
	}

	rel := filepath.Join(".spec-workflow", "specs", specName)
	abs := filepath.Join(cwd, rel)

	info, err := os.Stat(abs)
	found := err == nil && info.IsDir()

	dir := ""
	if found {
		dir = rel
	}

	result := map[string]any{
		"ok":        true,
		"spec":      specName,
		"found":     found,
		"directory": dir,
	}

	rawVal := ""
	if found {
		rawVal = rel
	}
	Output(result, rawVal, raw)
}
