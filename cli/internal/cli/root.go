package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd(version, commit, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spw",
		Short: "SPW — spec-workflow command kit",
		Long:  "SPW is a command/template kit for spec-workflow-mcp that provides stricter agent execution patterns with subagent-first orchestration and model routing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newVersionCmd(version, commit, date))
	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newRenderCmd())
	cmd.AddCommand(newHookCmd())
	cmd.AddCommand(newToolsCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newSkillsCmd())
	cmd.AddCommand(newTasksCmd())
	cmd.AddCommand(newWaveCmd())
	cmd.AddCommand(newSpecCmd())

	return cmd
}

func newVersionCmd(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("spw %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}
}
