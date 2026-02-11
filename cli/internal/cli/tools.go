package cli

import (
	"os"

	"github.com/lucas-stellet/spw/internal/tools"
	"github.com/spf13/cobra"
)

func newToolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Workflow tools for subagent use",
		Long:  "Provides config-get, spec-resolve, wave-resolve, runs, handoff, skills, approval, dispatch, and resolve-model subcommands.",
	}

	cmd.AddCommand(newToolsConfigGetCmd())
	cmd.AddCommand(newToolsSpecResolveCmd())
	cmd.AddCommand(newToolsWaveResolveCmd())
	cmd.AddCommand(newToolsRunsCmd())
	cmd.AddCommand(newToolsHandoffCmd())
	cmd.AddCommand(newToolsSkillsCmd())
	cmd.AddCommand(newToolsApprovalCmd())
	cmd.AddCommand(newToolsDispatchInitCmd())
	cmd.AddCommand(newToolsDispatchSetupCmd())
	cmd.AddCommand(newToolsDispatchReadStatusCmd())
	cmd.AddCommand(newToolsDispatchHandoffCmd())
	cmd.AddCommand(newToolsResolveModelCmd())
	cmd.AddCommand(newToolsMergeConfigCmd())

	return cmd
}

func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func newToolsConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config-get <section.key>",
		Short: "Read a config value",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			def, _ := cmd.Flags().GetString("default")
			tools.ConfigGet(getCwd(), args[0], def, raw)
		},
	}
	cmd.Flags().String("default", "", "Default value if key is missing")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsSpecResolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec-resolve-dir <spec-name>",
		Short: "Resolve spec directory path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.SpecResolveDir(getCwd(), args[0], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsWaveResolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wave-resolve-current <spec-name>",
		Short: "Resolve current wave number",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.WaveResolveCurrent(getCwd(), args[0], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsRunsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs-latest-unfinished <phase-dir>",
		Short: "Find latest unfinished run directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.RunsLatestUnfinished(getCwd(), args[0], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsHandoffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handoff-validate <run-dir>",
		Short: "Validate file-first handoff completeness",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.HandoffValidate(getCwd(), args[0], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills-effective-set <design|implementation>",
		Short: "List effective skills for a stage",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.SkillsEffectiveSet(getCwd(), args[0], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsApprovalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval-fallback-id <spec-name> <doc-type>",
		Short: "Get fallback approval ID for a document",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.ApprovalFallbackID(getCwd(), args[0], args[1], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsDispatchInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dispatch-init <command> <spec-name>",
		Short: "Initialize a dispatch run directory",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			wave, _ := cmd.Flags().GetString("wave")
			tools.DispatchInit(getCwd(), args[0], args[1], wave, raw)
		},
	}
	cmd.Flags().String("wave", "", "Wave number (required for wave-aware commands)")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsDispatchSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dispatch-setup <subagent-name>",
		Short: "Create subagent directory with brief.md skeleton",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			runDir, _ := cmd.Flags().GetString("run-dir")
			modelAlias, _ := cmd.Flags().GetString("model-alias")
			tools.DispatchSetup(getCwd(), args[0], runDir, modelAlias, raw)
		},
	}
	cmd.Flags().String("run-dir", "", "Run directory path")
	cmd.Flags().String("model-alias", "", "Model alias (web_research, complex_reasoning, implementation)")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("run-dir")
	return cmd
}

func newToolsDispatchReadStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dispatch-read-status <subagent-name>",
		Short: "Read and validate subagent status.json",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			runDir, _ := cmd.Flags().GetString("run-dir")
			tools.DispatchReadStatus(getCwd(), args[0], runDir, raw)
		},
	}
	cmd.Flags().String("run-dir", "", "Run directory path")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("run-dir")
	return cmd
}

func newToolsDispatchHandoffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dispatch-handoff",
		Short: "Generate _handoff.md from subagent status files",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			runDir, _ := cmd.Flags().GetString("run-dir")
			command, _ := cmd.Flags().GetString("command")
			tools.DispatchHandoff(getCwd(), runDir, command, raw)
		},
	}
	cmd.Flags().String("run-dir", "", "Run directory path")
	cmd.Flags().String("command", "", "SPW command name (for category lookup)")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("run-dir")
	return cmd
}

func newToolsResolveModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-model <alias>",
		Short: "Resolve model alias to configured model name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.ResolveModel(getCwd(), args[0], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsMergeConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge-config <template> <user> <output>",
		Short: "Merge template TOML with user TOML, preserving user values",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			tools.MergeConfig(args[0], args[1], args[2])
		},
	}
}
