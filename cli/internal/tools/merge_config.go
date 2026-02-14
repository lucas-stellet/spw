package tools

import (
	"fmt"
	"os"

	"github.com/lucas-stellet/oraculo/internal/config"
)

// MergeConfig merges a template TOML with a user TOML, writing the result to outputPath.
func MergeConfig(templatePath, userPath, outputPath string) {
	if err := config.Merge(templatePath, userPath, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "merge-config: %v\n", err)
		os.Exit(1)
	}
}
