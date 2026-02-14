package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/oraculo/internal/spec"
	"github.com/lucas-stellet/oraculo/internal/specdir"
	"github.com/lucas-stellet/oraculo/internal/summary"
	"github.com/lucas-stellet/oraculo/internal/tasks"
	"github.com/lucas-stellet/oraculo/internal/wave"
	"github.com/spf13/cobra"
)

func newSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary <spec-name>",
		Short: "Generate a progress summary for a spec",
		Long: `Generate a progress summary showing task status, wave progress,
and files changed. Works at any lifecycle stage.`,
		Args: cobra.ExactArgs(1),
		Run:  runSummary,
	}

	cmd.Flags().Bool("export", false, "Export PROGRESS-SUMMARY.md to spec directory")
	cmd.Flags().Bool("vscode", false, "Open in VS Code")
	cmd.Flags().Bool("raw", false, "Output raw Markdown without rendering")

	return cmd
}

func runSummary(cmd *cobra.Command, args []string) {
	raw, _ := cmd.Flags().GetBool("raw")
	vscode, _ := cmd.Flags().GetBool("vscode")
	export, _ := cmd.Flags().GetBool("export")

	cwd := getCwd()
	specName := args[0]

	sd, err := specdir.Resolve(cwd, specName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	stage := spec.ClassifyStage(sd)

	// Parse tasks.md
	tasksPath := specdir.TasksPath(sd)
	var doc *tasks.Document
	if specdir.FileExists(tasksPath) {
		d, err := tasks.ParseFile(tasksPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing tasks.md: %s\n", err)
			os.Exit(1)
		}
		doc = &d
	} else {
		doc = &tasks.Document{}
	}

	// Scan waves if execution dir exists
	var waves []wave.WaveState
	wavesPath := filepath.Join(sd, specdir.WavesDir)
	if specdir.DirExists(wavesPath) {
		waves, err = wave.ScanWaves(sd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not scan waves: %s\n", err)
		}
	}

	// Generate progress summary
	ps, err := summary.GenerateProgress(sd, stage, doc, waves)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating summary: %s\n", err)
		os.Exit(1)
	}

	content, err := summary.RenderFull(ps.Frontmatter, ps.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering summary: %s\n", err)
		os.Exit(1)
	}

	// Export to file if requested
	if export {
		exportPath := filepath.Join(sd, specdir.ProgressSummaryMD)
		if err := os.WriteFile(exportPath, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", exportPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Exported to: %s\n", exportPath)
	}

	renderOutput(content, specName+"-progress-summary.md", vscode, raw)
}
