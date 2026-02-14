package tools

import (
	"fmt"
	"os"

	"github.com/lucas-stellet/oraculo/internal/config"
	"github.com/lucas-stellet/oraculo/internal/install"
)

// MergeSettings merges ORACULO hooks into .claude/settings.json at the given root.
func MergeSettings(root string) {
	cfg, _ := config.Load(root)
	if err := install.MergeSettings(root, cfg.AgentTeams); err != nil {
		fmt.Fprintf(os.Stderr, "merge-settings: %v\n", err)
		os.Exit(1)
	}
}
