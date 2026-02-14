package cli

import (
	"github.com/lucas-stellet/oraculo/internal/hook"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook <event>",
		Short: "Handle Claude Code hook events",
		Long:  "Dispatches hook events: statusline, session-start, guard-prompt, guard-paths, guard-stop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return hook.Dispatch(args[0])
		},
	}
}
