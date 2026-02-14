package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/spw/internal/spec"
	"github.com/lucas-stellet/spw/internal/specdir"
	"github.com/lucas-stellet/spw/internal/store"
	"github.com/lucas-stellet/spw/internal/tasks"
	"github.com/lucas-stellet/spw/internal/viewer"
	"github.com/lucas-stellet/spw/internal/wave"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <spec-name> [artifact-type]",
		Short: "View spec artifacts in the terminal or VS Code",
		Long: `View spec artifacts with terminal rendering (glamour) or pipe to VS Code.

Artifact types:
  overview              Overview of the spec (default)
  report                Subagent report
  brief                 Subagent brief
  checkpoint            Checkpoint report
  implementation-log    Task implementation log
  wave-summary          Wave summary
  completion-summary    Completion summary`,
		Args: cobra.RangeArgs(1, 2),
		Run:  runView,
	}

	cmd.Flags().Int("wave", 0, "Wave number (for checkpoint, wave-summary)")
	cmd.Flags().Int("run", 0, "Run number")
	cmd.Flags().String("task", "", "Task ID (for implementation-log)")
	cmd.Flags().Bool("vscode", false, "Open in VS Code")
	cmd.Flags().Bool("raw", false, "Output raw Markdown without rendering")

	return cmd
}

func runView(cmd *cobra.Command, args []string) {
	raw, _ := cmd.Flags().GetBool("raw")
	vscode, _ := cmd.Flags().GetBool("vscode")
	waveNum, _ := cmd.Flags().GetInt("wave")
	runNum, _ := cmd.Flags().GetInt("run")
	taskID, _ := cmd.Flags().GetString("task")

	cwd := getCwd()
	specName := args[0]
	artifactType := "overview"
	if len(args) > 1 {
		artifactType = args[1]
	}

	sd, err := specdir.Resolve(cwd, specName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	var content string

	switch artifactType {
	case "overview":
		content = buildOverview(sd, specName)
	case "implementation-log":
		content = getImplLog(sd, taskID)
	case "completion-summary":
		content = getCompletionSummary(sd)
	case "checkpoint":
		content = getCheckpointReport(sd, waveNum, runNum)
	case "wave-summary":
		content = getWaveSummary(sd, waveNum)
	case "report":
		content = getSubagentArtifact(sd, waveNum, runNum, specdir.ReportMD)
	case "brief":
		content = getSubagentArtifact(sd, waveNum, runNum, specdir.BriefMD)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown artifact type %q\n", artifactType)
		os.Exit(1)
	}

	renderOutput(content, specName+"-"+artifactType+".md", vscode, raw)
}

func buildOverview(sd, specName string) string {
	stage := spec.ClassifyStage(sd)

	var tasksDone, tasksTotal int
	tasksPath := specdir.TasksPath(sd)
	if specdir.FileExists(tasksPath) {
		doc, err := tasks.ParseFile(tasksPath)
		if err == nil {
			counts := doc.Count()
			tasksDone = counts.Done
			tasksTotal = counts.Total
		}
	}

	var waveInfos []viewer.WaveInfo
	wavesPath := filepath.Join(sd, specdir.WavesDir)
	if specdir.DirExists(wavesPath) {
		states, err := wave.ScanWaves(sd)
		if err == nil {
			for _, ws := range states {
				waveInfos = append(waveInfos, viewer.WaveInfo{
					Num:       ws.WaveNum,
					Status:    ws.Status,
					ExecRuns:  ws.ExecRuns,
					CheckRuns: ws.CheckRuns,
				})
			}
		}
	}

	return viewer.RenderOverview(specName, stage, tasksDone, tasksTotal, waveInfos)
}

func getImplLog(sd, taskID string) string {
	if taskID == "" {
		fmt.Fprintln(os.Stderr, "Error: --task flag is required for implementation-log")
		os.Exit(1)
	}

	// Try DB first
	s := store.TryOpen(sd)
	if s != nil {
		defer s.Close()
		log, err := s.GetImplLog(taskID)
		if err == nil && log != nil {
			return log.Content
		}
	}

	// Fallback to filesystem
	path := specdir.ImplLogPath(sd, taskID)
	return readFileOrFail(path, "implementation log")
}

func getCompletionSummary(sd string) string {
	// Try DB first
	s := store.TryOpen(sd)
	if s != nil {
		defer s.Close()
		cs, err := s.GetCompletionSummary()
		if err == nil && cs != nil {
			return cs.Frontmatter + "\n" + cs.Body
		}
	}

	// Fallback to filesystem
	path := filepath.Join(sd, specdir.CompletionSummaryMD)
	return readFileOrFail(path, "completion summary")
}

func getCheckpointReport(sd string, waveNum, runNum int) string {
	if waveNum == 0 {
		// Try the top-level checkpoint report
		path := filepath.Join(sd, specdir.CheckpointReport)
		if specdir.FileExists(path) {
			return readFileOrFail(path, "checkpoint report")
		}
		fmt.Fprintln(os.Stderr, "Error: --wave flag is required for checkpoint (or no top-level checkpoint report found)")
		os.Exit(1)
	}

	// Try DB first
	s := store.TryOpen(sd)
	if s != nil {
		defer s.Close()
		wavePtr := waveNum
		command := "checkpoint"
		if runNum > 0 {
			r, err := s.GetRun(command, runNum)
			if err == nil && r != nil && r.WaveNumber != nil && *r.WaveNumber == wavePtr {
				subs, err := s.ListSubagents(r.ID)
				if err == nil && len(subs) > 0 {
					return subs[0].Report
				}
			}
		}
	}

	// Fallback to filesystem: find latest checkpoint run dir
	checkDir := specdir.WaveCheckpointPath(sd, waveNum)
	if runNum > 0 {
		runDir := filepath.Join(checkDir, fmt.Sprintf(specdir.RunDirFmt, runNum))
		return findReportInDir(runDir)
	}
	latestDir, _, err := specdir.LatestRunDir(checkDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: no checkpoint runs found for wave %d\n", waveNum)
		os.Exit(1)
	}
	return findReportInDir(latestDir)
}

func getWaveSummary(sd string, waveNum int) string {
	if waveNum == 0 {
		fmt.Fprintln(os.Stderr, "Error: --wave flag is required for wave-summary")
		os.Exit(1)
	}

	// Try DB first
	s := store.TryOpen(sd)
	if s != nil {
		defer s.Close()
		w, err := s.GetWave(waveNum)
		if err == nil && w != nil && w.SummaryText != "" {
			return w.SummaryText
		}
	}

	// Fallback to filesystem
	path := specdir.WaveSummaryPath(sd, waveNum)
	return readFileOrFail(path, "wave summary")
}

func getSubagentArtifact(sd string, waveNum, runNum int, filename string) string {
	// Try DB if we have wave+run context
	if waveNum > 0 && runNum > 0 {
		s := store.TryOpen(sd)
		if s != nil {
			defer s.Close()
			command := "exec"
			r, err := s.GetRun(command, runNum)
			if err == nil && r != nil {
				subs, err := s.ListSubagents(r.ID)
				if err == nil && len(subs) > 0 {
					for _, sub := range subs {
						if filename == specdir.ReportMD && sub.Report != "" {
							return sub.Report
						}
						if filename == specdir.BriefMD && sub.Brief != "" {
							return sub.Brief
						}
					}
				}
			}
		}
	}

	// Fallback: need wave+run to find filesystem path
	if waveNum == 0 || runNum == 0 {
		fmt.Fprintln(os.Stderr, "Error: --wave and --run flags are required for report/brief")
		os.Exit(1)
	}

	execDir := specdir.WaveExecPath(sd, waveNum)
	runDir := filepath.Join(execDir, fmt.Sprintf(specdir.RunDirFmt, runNum))

	// Search for the file in subdirs of the run dir
	entries, err := os.ReadDir(runDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read run directory: %s\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(runDir, e.Name(), filename)
		if specdir.FileExists(path) {
			return readFileOrFail(path, filename)
		}
	}

	// Try directly in run dir
	path := filepath.Join(runDir, filename)
	return readFileOrFail(path, filename)
}

func findReportInDir(dir string) string {
	// Look for report.md in subdirectories (subagent dirs)
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read directory: %s\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name(), specdir.ReportMD)
		if specdir.FileExists(path) {
			data, err := os.ReadFile(path)
			if err == nil {
				return string(data)
			}
		}
	}

	// Try report.md directly in the dir
	path := filepath.Join(dir, specdir.ReportMD)
	return readFileOrFail(path, "report")
}

func readFileOrFail(path, label string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s not found at %s\n", label, path)
		os.Exit(1)
	}
	return string(data)
}

func renderOutput(content, filename string, vscode, raw bool) {
	if raw {
		fmt.Print(content)
		return
	}

	if vscode {
		if err := viewer.OpenInVSCode(content, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening VS Code: %s\n", err)
			os.Exit(1)
		}
		return
	}

	rendered, err := viewer.RenderTerminal(content)
	if err != nil {
		fmt.Print(content)
		return
	}
	fmt.Print(rendered)
}
