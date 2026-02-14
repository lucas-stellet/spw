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
	cmd.AddCommand(newToolsMergeSettingsCmd())
	cmd.AddCommand(newToolsVerifyTaskCmd())
	cmd.AddCommand(newToolsImplLogCmd())
	cmd.AddCommand(newToolsTaskMarkCmd())
	cmd.AddCommand(newToolsWaveUpdateCmd())
	cmd.AddCommand(newToolsWaveStatusCmd())
	cmd.AddCommand(newToolsDispatchInitAuditCmd())
	cmd.AddCommand(newToolsAuditIterationCmd())

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

func newToolsMergeSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge-settings",
		Short: "Merge SPW hooks into .claude/settings.json, preserving non-SPW entries",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			global, _ := cmd.Flags().GetBool("global")
			root := getCwd()
			if global {
				home, err := os.UserHomeDir()
				if err != nil {
					root = getCwd()
				} else {
					root = home
				}
			}
			tools.MergeSettings(root)
		},
	}
	cmd.Flags().Bool("global", false, "Target ~/.claude/settings.json")
	return cmd
}

func newToolsVerifyTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-task <spec-name>",
		Short: "Verify a task has implementation log and optionally a commit",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			taskID, _ := cmd.Flags().GetString("task-id")
			checkCommit, _ := cmd.Flags().GetBool("check-commit")
			tools.VerifyTask(getCwd(), args[0], taskID, checkCommit, raw)
		},
	}
	cmd.Flags().String("task-id", "", "Task ID to verify")
	cmd.Flags().Bool("check-commit", false, "Also check for a git commit mentioning the task")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("task-id")
	return cmd
}

func newToolsImplLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "impl-log",
		Short: "Implementation log commands (register, check)",
	}
	cmd.AddCommand(newToolsImplLogRegisterCmd())
	cmd.AddCommand(newToolsImplLogCheckCmd())
	return cmd
}

func newToolsImplLogRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <spec-name>",
		Short: "Register an implementation log for a completed task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			taskID, _ := cmd.Flags().GetString("task-id")
			wave, _ := cmd.Flags().GetString("wave")
			title, _ := cmd.Flags().GetString("title")
			files, _ := cmd.Flags().GetString("files")
			changes, _ := cmd.Flags().GetString("changes")
			tests, _ := cmd.Flags().GetString("tests")
			tools.ImplLogRegister(getCwd(), args[0], taskID, wave, title, files, changes, tests, raw)
		},
	}
	cmd.Flags().String("task-id", "", "Task ID")
	cmd.Flags().String("wave", "", "Wave number (e.g. 01)")
	cmd.Flags().String("title", "", "Task title")
	cmd.Flags().String("files", "", "Comma-separated list of changed files")
	cmd.Flags().String("changes", "", "Description of changes made")
	cmd.Flags().String("tests", "", "Description of tests added (optional)")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("task-id")
	_ = cmd.MarkFlagRequired("wave")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("files")
	_ = cmd.MarkFlagRequired("changes")
	return cmd
}

func newToolsImplLogCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <spec-name>",
		Short: "Check if implementation logs exist for given task IDs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			taskIDs, _ := cmd.Flags().GetString("task-ids")
			tools.ImplLogCheck(getCwd(), args[0], taskIDs, raw)
		},
	}
	cmd.Flags().String("task-ids", "", "Comma-separated task IDs to check")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("task-ids")
	return cmd
}

func newToolsTaskMarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task-mark <spec-name>",
		Short: "Update a task checkbox status in tasks.md",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			taskID, _ := cmd.Flags().GetString("task-id")
			status, _ := cmd.Flags().GetString("status")
			tools.TaskMark(getCwd(), args[0], taskID, status, raw)
		},
	}
	cmd.Flags().String("task-id", "", "Task ID to mark")
	cmd.Flags().String("status", "", "New status: in-progress, done, blocked")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("task-id")
	_ = cmd.MarkFlagRequired("status")
	return cmd
}

func newToolsWaveUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wave-update <spec-name>",
		Short: "Write wave summary and latest JSON files",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			wave, _ := cmd.Flags().GetString("wave")
			status, _ := cmd.Flags().GetString("status")
			tasks, _ := cmd.Flags().GetString("tasks")
			checkpointRun, _ := cmd.Flags().GetString("checkpoint-run")
			executionRun, _ := cmd.Flags().GetString("execution-run")
			tools.WaveUpdate(getCwd(), args[0], wave, status, tasks, checkpointRun, executionRun, raw)
		},
	}
	cmd.Flags().String("wave", "", "Wave number (e.g. 02)")
	cmd.Flags().String("status", "", "Wave status: pass, blocked")
	cmd.Flags().String("tasks", "", "Comma-separated task numbers in this wave")
	cmd.Flags().String("checkpoint-run", "", "Checkpoint run ID (optional)")
	cmd.Flags().String("execution-run", "", "Execution run ID (optional)")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("wave")
	_ = cmd.MarkFlagRequired("status")
	_ = cmd.MarkFlagRequired("tasks")
	return cmd
}

func newToolsWaveStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wave-status <spec-name>",
		Short: "Resolve comprehensive wave state for a spec",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			tools.WaveStatus(getCwd(), args[0], raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newToolsDispatchInitAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dispatch-init-audit",
		Short: "Create an audit subdirectory within a run directory",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			runDir, _ := cmd.Flags().GetString("run-dir")
			auditType, _ := cmd.Flags().GetString("type")
			iteration, _ := cmd.Flags().GetInt("iteration")
			tools.DispatchInitAudit(getCwd(), runDir, auditType, iteration, raw)
		},
	}
	cmd.Flags().String("run-dir", "", "Run directory path")
	cmd.Flags().String("type", "", "Audit type: inline-audit, inline-checkpoint")
	cmd.Flags().Int("iteration", 1, "Iteration number (default 1)")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("run-dir")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func newToolsAuditIterationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit-iteration",
		Short: "Audit iteration tracking (start, check, advance)",
	}
	cmd.AddCommand(newToolsAuditIterationStartCmd())
	cmd.AddCommand(newToolsAuditIterationCheckCmd())
	cmd.AddCommand(newToolsAuditIterationAdvanceCmd())
	return cmd
}

func newToolsAuditIterationStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Initialize iteration tracking for an audit",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			runDir, _ := cmd.Flags().GetString("run-dir")
			auditType, _ := cmd.Flags().GetString("type")
			max, _ := cmd.Flags().GetInt("max")
			tools.AuditIterationStart(getCwd(), runDir, auditType, max, raw)
		},
	}
	cmd.Flags().String("run-dir", "", "Run directory path")
	cmd.Flags().String("type", "", "Audit type: inline-audit, inline-checkpoint")
	cmd.Flags().Int("max", 3, "Maximum iterations allowed")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("run-dir")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func newToolsAuditIterationCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check whether another iteration is allowed",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			runDir, _ := cmd.Flags().GetString("run-dir")
			auditType, _ := cmd.Flags().GetString("type")
			tools.AuditIterationCheck(getCwd(), runDir, auditType, raw)
		},
	}
	cmd.Flags().String("run-dir", "", "Run directory path")
	cmd.Flags().String("type", "", "Audit type: inline-audit, inline-checkpoint")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("run-dir")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func newToolsAuditIterationAdvanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advance",
		Short: "Increment iteration counter and record result",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			runDir, _ := cmd.Flags().GetString("run-dir")
			auditType, _ := cmd.Flags().GetString("type")
			result, _ := cmd.Flags().GetString("result")
			tools.AuditIterationAdvance(getCwd(), runDir, auditType, result, raw)
		},
	}
	cmd.Flags().String("run-dir", "", "Run directory path")
	cmd.Flags().String("type", "", "Audit type: inline-audit, inline-checkpoint")
	cmd.Flags().String("result", "", "Result of the current iteration (pass, blocked, etc.)")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	_ = cmd.MarkFlagRequired("run-dir")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("result")
	return cmd
}
