package cli

import (
	"github.com/lucas-stellet/spw/internal/spec"
	"github.com/lucas-stellet/spw/internal/specdir"
	"github.com/lucas-stellet/spw/internal/tools"
	"github.com/spf13/cobra"
)

func newSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Spec lifecycle inspection commands",
		Long:  "Inspect spec artifacts, lifecycle stage, prerequisites, and approvals.",
	}

	cmd.AddCommand(newSpecArtifactsCmd())
	cmd.AddCommand(newSpecStageCmd())
	cmd.AddCommand(newSpecPrereqsCmd())
	cmd.AddCommand(newSpecApprovalCmd())
	cmd.AddCommand(newSpecListCmd())

	return cmd
}

func newSpecArtifactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifacts <spec-name>",
		Short: "Check which artifacts exist for a spec",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			artifacts, deviations := spec.CheckArtifacts(specDir)
			result := map[string]any{
				"ok":         true,
				"spec":       specName,
				"artifacts":  artifacts,
				"deviations": deviations,
			}
			tools.Output(result, "", raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newSpecStageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stage <spec-name>",
		Short: "Classify the lifecycle stage of a spec",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			stage := spec.ClassifyStage(specDir)
			result := map[string]any{
				"ok":    true,
				"spec":  specName,
				"stage": stage,
			}
			tools.Output(result, stage, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newSpecPrereqsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prereqs <spec-name> <command>",
		Short: "Check prerequisites for an SPW command",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]
			command := args[1]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			pr := spec.CheckPrereqs(specDir, command)
			result := map[string]any{
				"ok":      true,
				"spec":    specName,
				"command": command,
				"ready":   pr.Ready,
				"missing": pr.Missing,
			}
			rawVal := "ready"
			if !pr.Ready {
				rawVal = "not-ready"
			}
			tools.Output(result, rawVal, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newSpecApprovalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval <spec-name> <doc-type>",
		Short: "Check approval status for a document",
		Long:  "Scans filesystem for local approval records. doc-type: requirements, design, tasks.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]
			docType := args[1]

			ar := spec.CheckApproval(cwd, specName, docType)
			result := map[string]any{
				"ok":          true,
				"spec":        specName,
				"doc_type":    docType,
				"found":       ar.Found,
				"approval_id": ar.ApprovalID,
				"source":      ar.Source,
			}
			rawVal := ""
			if ar.Found {
				rawVal = ar.ApprovalID
			}
			tools.Output(result, rawVal, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newSpecListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all specs",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()

			specs, err := spec.List(cwd)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			result := map[string]any{
				"ok":    true,
				"specs": specs,
				"count": len(specs),
			}
			rawVal := ""
			for i, s := range specs {
				if i > 0 {
					rawVal += "\n"
				}
				rawVal += s
			}
			tools.Output(result, rawVal, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}
