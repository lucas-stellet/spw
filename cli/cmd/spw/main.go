package main

import (
	"os"

	"github.com/lucas-stellet/spw/internal/cli"
)

// Set by goreleaser via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := cli.NewRootCmd(version, commit, date)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
