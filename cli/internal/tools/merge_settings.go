package tools

import (
	"fmt"
	"os"

	"github.com/lucas-stellet/spw/internal/install"
)

// MergeSettings merges SPW hooks into .claude/settings.json at the given root.
func MergeSettings(root string) {
	if err := install.MergeSettings(root); err != nil {
		fmt.Fprintf(os.Stderr, "merge-settings: %v\n", err)
		os.Exit(1)
	}
}
