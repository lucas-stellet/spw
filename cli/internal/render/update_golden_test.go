package render

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUpdateGolden regenerates golden files when UPDATE_GOLDEN=1 is set.
func TestUpdateGolden(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") != "1" {
		t.Skip("set UPDATE_GOLDEN=1 to regenerate golden files")
	}

	e := defaultEngine(t)
	dir := goldenDir()

	for _, cmd := range AllCommands {
		content, err := e.RenderCommand(cmd)
		if err != nil {
			t.Fatalf("RenderCommand(%q) error: %v", cmd, err)
		}
		path := filepath.Join(dir, cmd+".md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
		t.Logf("Updated: %s", cmd+".md")
	}
}
