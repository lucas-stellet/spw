package cli

import (
	"os"

	"github.com/lucas-stellet/spw/internal/install"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install SPW kit into the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return install.Run(install.Options{WorkspaceRoot: cwd})
		},
	}
}
