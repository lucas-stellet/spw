package cli

import (
	"fmt"

	"github.com/lucas-stellet/oraculo/internal/specdir"
	"github.com/lucas-stellet/oraculo/internal/tools"
	"github.com/lucas-stellet/oraculo/internal/wave"
	"github.com/spf13/cobra"
)

func newWaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wave",
		Short: "Wave state inspection commands",
		Long:  "Inspect wave execution state, checkpoint status, and summaries.",
	}

	cmd.AddCommand(newWaveStateCmd())
	cmd.AddCommand(newWaveSummaryCmd())
	cmd.AddCommand(newWaveCheckpointCmd())
	cmd.AddCommand(newWaveResumeCmd())

	return cmd
}

func newWaveStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state <spec-name>",
		Short: "Show state of all waves",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			waves, err := wave.ScanWaves(specDir)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			result := map[string]any{
				"ok":    true,
				"spec":  specName,
				"waves": waves,
				"count": len(waves),
			}
			tools.Output(result, fmt.Sprintf("%d waves", len(waves)), raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newWaveSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary <spec-name> <wave-num>",
		Short: "Generate summary for a specific wave",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]
			waveNum := parseWaveNum(args[1], raw)

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			summary := wave.GenerateSummary(specDir, waveNum)
			result := map[string]any{
				"ok":         true,
				"spec":       specName,
				"wave":       waveNum,
				"status":     summary.Status,
				"summary":    summary.Summary,
				"source":     summary.Source,
				"stale_flag": summary.StaleFlag,
			}
			tools.Output(result, summary.Status, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newWaveCheckpointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint <spec-name> <wave-num>",
		Short: "Resolve checkpoint status for a wave",
		Long:  "Uses _latest.json-first resolution to avoid stale _wave-summary.json bugs.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]
			waveNum := parseWaveNum(args[1], raw)

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			cp := wave.ResolveCheckpoint(specDir, waveNum)
			result := map[string]any{
				"ok":         true,
				"spec":       specName,
				"wave_num":   cp.WaveNum,
				"status":     cp.Status,
				"run_id":     cp.RunID,
				"source":     cp.Source,
				"stale_flag": cp.StaleFlag,
				"details":    cp.Details,
			}
			tools.Output(result, cp.Status, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newWaveResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <spec-name>",
		Short: "Compute resume state for a spec",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			rs := wave.ComputeResume(specDir)
			result := map[string]any{
				"ok":       true,
				"spec":     specName,
				"action":   rs.Action,
				"wave_num": rs.WaveNum,
				"reason":   rs.Reason,
			}
			tools.Output(result, rs.Action, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func parseWaveNum(s string, raw bool) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil || n < 1 {
		tools.Fail("wave number must be a positive integer", raw)
	}
	return n
}
