package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lucas-stellet/spw/internal/specdir"
	"github.com/lucas-stellet/spw/internal/store"
	"github.com/lucas-stellet/spw/internal/summary"
	"github.com/lucas-stellet/spw/internal/tasks"
	"github.com/lucas-stellet/spw/internal/wave"
	"github.com/spf13/cobra"
)

func newFinalizarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalizar <spec-name>",
		Short: "Mark spec as completed and generate implementation summary",
		Args:  cobra.ExactArgs(1),
		Run:   runFinalizar,
	}

	cmd.Flags().Bool("export", false, "Export COMPLETION-SUMMARY.md to disk")
	cmd.Flags().Bool("force", false, "Skip post-mortem check")
	cmd.Flags().Bool("raw", false, "Output raw JSON")

	return cmd
}

func runFinalizar(cmd *cobra.Command, args []string) {
	export, _ := cmd.Flags().GetBool("export")
	force, _ := cmd.Flags().GetBool("force")
	raw, _ := cmd.Flags().GetBool("raw")

	cwd := getCwd()
	specName := args[0]

	// 1. Resolve spec directory.
	sd, err := specdir.Resolve(cwd, specName)
	if err != nil {
		finalizarFail(fmt.Sprintf("spec not found: %s", specName), raw)
	}

	// 2. Parse tasks.md.
	tasksPath := specdir.TasksPath(sd)
	doc, err := tasks.ParseFile(tasksPath)
	if err != nil {
		finalizarFail(fmt.Sprintf("cannot parse tasks.md: %s", err), raw)
	}

	// 3. Validate all non-deferred tasks are done.
	var incomplete []string
	for _, t := range doc.Tasks {
		if t.IsDeferred {
			continue
		}
		if t.Status != "done" {
			incomplete = append(incomplete, fmt.Sprintf("  - Task %s: %s (%s)", t.ID, t.Title, t.Status))
		}
	}
	if len(incomplete) > 0 {
		finalizarFail(fmt.Sprintf("incomplete tasks:\n%s", strings.Join(incomplete, "\n")), raw)
	}

	// 4. Check post-mortem (unless --force).
	if !force {
		pmPath := filepath.Join(sd, specdir.PostMortemReport)
		if !specdir.FileExists(pmPath) {
			finalizarFail("post-mortem/report.md not found (use --force to skip)", raw)
		}
	}

	// 5. Open store.
	s, err := store.Open(sd)
	if err != nil {
		finalizarFail(fmt.Sprintf("cannot open store: %s", err), raw)
	}
	defer s.Close()

	// 6. Harvest all remaining filesystem artifacts.
	harvested := harvestAllArtifacts(s, sd)

	// 7. Sync all tasks to DB.
	for _, t := range doc.Tasks {
		wavePtr := &t.Wave
		_ = s.SyncTask(store.TaskRecord{
			TaskID:     t.ID,
			Title:      t.Title,
			Status:     t.Status,
			Wave:       wavePtr,
			DependsOn:  strings.Join(t.DependsOn, ","),
			Files:      t.Files,
			TDD:        t.TDD != "",
			IsDeferred: t.IsDeferred,
		})
	}

	// 8. Scan waves.
	waves, _ := wave.ScanWaves(sd)

	// 9. Generate completion summary.
	cs, err := summary.GenerateCompletion(sd, doc, waves, s)
	if err != nil {
		finalizarFail(fmt.Sprintf("cannot generate summary: %s", err), raw)
	}

	// 10. Render and save to DB.
	fmYAML, err := summary.RenderFull(cs.Frontmatter, "")
	if err != nil {
		finalizarFail(fmt.Sprintf("cannot render frontmatter: %s", err), raw)
	}
	// Strip the trailing body separator from frontmatter-only render.
	fmYAML = strings.TrimSuffix(fmYAML, "\n\n")

	if err := s.SaveCompletionSummary(fmYAML, cs.Body); err != nil {
		finalizarFail(fmt.Sprintf("cannot save summary: %s", err), raw)
	}

	// 11. Update spec_meta.
	nowStr := time.Now().UTC().Format(time.RFC3339)
	_ = s.SetMeta("status", "completed")
	_ = s.SetMeta("stage", "complete")
	_ = s.SetMeta("completed_at", nowStr)

	// 12. Index in global .spw-index.db.
	var docsIndexed int
	ix, ixErr := store.OpenIndex(cwd)
	if ixErr == nil {
		defer ix.Close()
		dbPath := filepath.Join(sd, specdir.SpecDB)
		_ = ix.IndexSpec(specName, "complete", dbPath)

		// Index the completion summary.
		fullContent, _ := summary.RenderFull(cs.Frontmatter, cs.Body)
		snippet := truncate(cs.Body, 200)
		if ix.IndexDocument(specName, "completion", "complete", "Completion Summary: "+specName, snippet, fullContent) == nil {
			docsIndexed++
		}
	}

	// 13. Export if requested.
	var summaryPath string
	if export {
		fullContent, _ := summary.RenderFull(cs.Frontmatter, cs.Body)
		summaryPath = filepath.Join(sd, specdir.CompletionSummaryMD)
		if err := os.WriteFile(summaryPath, []byte(fullContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write %s: %s\n", summaryPath, err)
		}
	}

	// 14. Output result.
	result := map[string]any{
		"ok":             true,
		"spec":           specName,
		"status":         "completed",
		"tasks_count":    len(doc.Tasks),
		"waves_count":    len(waves),
		"harvested":      harvested,
		"docs_indexed":   docsIndexed,
		"completed_at":   nowStr,
	}
	if summaryPath != "" {
		result["summary_path"] = summaryPath
	}

	rawValue := fmt.Sprintf("Spec %q finalized: %d tasks, %d waves", specName, len(doc.Tasks), len(waves))
	outputJSON(result, rawValue, raw)
}

// harvestAllArtifacts walks phase directories and harvests files into the DB.
func harvestAllArtifacts(s *store.SpecStore, sd string) int {
	var count int

	phases := []struct {
		name string
		dir  string
	}{
		{specdir.PhasePRD, specdir.PhasePRD},
		{specdir.PhaseDesign, specdir.PhaseDesign},
		{specdir.PhasePlanning, specdir.PhasePlanning},
		{specdir.PhaseExecution, specdir.PhaseExecution},
		{specdir.PhaseQA, specdir.PhaseQA},
		{specdir.PhasePostMortem, specdir.PhasePostMortem},
	}

	for _, p := range phases {
		phaseDir := filepath.Join(sd, p.dir)
		if !specdir.DirExists(phaseDir) {
			continue
		}
		count += harvestPhaseDir(s, sd, p.name, phaseDir)
	}

	// Harvest implementation logs.
	implDir := filepath.Join(sd, specdir.ImplLogsDir)
	if specdir.DirExists(implDir) {
		entries, err := os.ReadDir(implDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				taskID := extractTaskID(e.Name())
				if taskID != "" {
					absPath := filepath.Join(implDir, e.Name())
					if s.HarvestImplLog(taskID, absPath) == nil {
						count++
					}
				}
			}
		}
	}

	return count
}

// harvestPhaseDir walks a phase directory and harvests markdown/json files and run dirs.
func harvestPhaseDir(s *store.SpecStore, sd, phase, phaseDir string) int {
	var count int

	_ = filepath.Walk(phaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Harvest run directories via HarvestRunDir.
		if info.IsDir() && runDirReFC.MatchString(info.Name()) {
			// Determine wave number if this is inside a wave dir.
			var waveNum *int
			wn := extractWaveNum(path)
			if wn > 0 {
				waveNum = &wn
			}
			command := inferCommandFromPath(path)
			if s.HarvestRunDir(path, command, waveNum) == nil {
				count++
			}
			return filepath.SkipDir
		}

		// Harvest individual markdown and json files as artifacts.
		if !info.IsDir() && (filepath.Ext(path) == ".md" || filepath.Ext(path) == ".json") {
			relPath, _ := filepath.Rel(sd, path)
			if s.HarvestArtifact(phase, relPath, path) == nil {
				count++
			}
		}

		return nil
	})

	return count
}

var (
	taskIDRe   = regexp.MustCompile(`^task-(.+)\.md$`)
	waveNumRe  = regexp.MustCompile(`wave-(\d+)`)
	runDirReFC = regexp.MustCompile(`^run-\d+$`)
)

// extractTaskID extracts a task ID from a filename like "task-3.md".
func extractTaskID(filename string) string {
	m := taskIDRe.FindStringSubmatch(filename)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

// extractWaveNum extracts a wave number from a path containing "wave-NN".
func extractWaveNum(path string) int {
	m := waveNumRe.FindStringSubmatch(path)
	if len(m) == 2 {
		var n int
		fmt.Sscanf(m[1], "%d", &n)
		return n
	}
	return 0
}

// inferCommandFromPath guesses the command name from the run directory path.
func inferCommandFromPath(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, p := range parts {
		if p == "checkpoint" {
			return "checkpoint"
		}
		if p == "execution" && i+1 < len(parts) && parts[i+1] == "waves" {
			return "exec"
		}
		if p == "_comms" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	// Fallback: infer from phase.
	for _, p := range parts {
		switch p {
		case "prd":
			return "prd"
		case "design":
			return "design-research"
		case "planning":
			return "tasks-plan"
		case "qa":
			return "qa"
		case "post-mortem":
			return "post-mortem"
		}
	}
	return "unknown"
}

// truncate returns the first n characters of s, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// outputJSON is a local helper to write JSON output (avoids importing tools package
// which may have compilation issues from other WIP changes).
func outputJSON(result map[string]any, rawValue string, raw bool) {
	if raw {
		fmt.Print(rawValue)
		return
	}
	// Use the same pattern as tools.Output.
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Print("{}")
		return
	}
	os.Stdout.Write(data)
}

func finalizarFail(message string, raw bool) {
	if raw {
		fmt.Fprint(os.Stderr, message)
		os.Exit(1)
	}
	result := map[string]any{"ok": false, "error": message}
	data, _ := json.MarshalIndent(result, "", "  ")
	os.Stderr.Write(data)
	os.Stderr.WriteString("\n")
	os.Exit(1)
}
