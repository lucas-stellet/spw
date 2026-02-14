package cli

import (
	"os"

	"github.com/lucas-stellet/oraculo/internal/install"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install ORACULO kit into the current project (or globally with --global)",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			if global {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				return install.RunGlobal(install.Options{WorkspaceRoot: home})
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return install.Run(install.Options{WorkspaceRoot: cwd})
		},
	}
	cmd.Flags().Bool("global", false, "Install to ~/.claude/ for all projects")
	return cmd
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize project-specific config (use with global install)",
		Long:  "Creates .spec-workflow/ config and templates, injects CLAUDE.md/AGENTS.md snippets, and sets up .gitattributes. Commands and workflows are expected from a global install.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return install.RunInit(install.Options{WorkspaceRoot: cwd})
		},
	}
}
