package summary

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lucas-stellet/spw/internal/store"
	"github.com/lucas-stellet/spw/internal/tasks"
	"github.com/lucas-stellet/spw/internal/wave"
)

// GenerateCompletion builds a full completion summary from spec artifacts.
// Reads implementation logs and wave reports from the store.
func GenerateCompletion(specDir string, doc tasks.Document, waves []wave.WaveState, s *store.SpecStore) (*CompletionSummary, error) {
	specName := doc.Frontmatter.Spec
	files := CollectFilesChanged(doc.Tasks)
	techs := InferTechnologies(files)

	titles := make([]string, len(doc.Tasks))
	for i, t := range doc.Tasks {
		titles[i] = t.Title
	}
	tags := InferTags(titles, files)

	// Count checkpoint passes/failures from waves.
	var passes, failures int
	for _, w := range waves {
		if w.Status == "complete" {
			passes++
		} else if w.Status == "blocked" {
			failures++
		}
	}

	// Try to get creation date from store for duration calculation.
	var durationDays int
	if s != nil {
		if created, err := s.GetMeta("created_at"); err == nil && created != "" {
			if t, err := time.Parse(time.RFC3339, created); err == nil {
				durationDays = int(time.Since(t).Hours() / 24)
			}
		}
	}

	// Build summary text from checkpoint summaries.
	summaryText := buildSummaryText(waves, s)

	now := time.Now().UTC()
	fm := CompletionFrontmatter{
		Spec:               specName,
		Status:             "completed",
		CompletedAt:        now,
		DurationDays:       durationDays,
		TasksCount:         len(doc.Tasks),
		WavesCount:         len(waves),
		CheckpointPasses:   passes,
		CheckpointFailures: failures,
		FilesChanged:       files,
		Technologies:       techs,
		Tags:               tags,
		Summary:            summaryText,
	}

	body := renderCompletionBody(specName, doc.Tasks, waves, s)

	return &CompletionSummary{Frontmatter: fm, Body: body}, nil
}

// GenerateProgress builds a progress summary for any stage.
// Works without store (reads only from parsed tasks and scanned waves).
func GenerateProgress(specDir string, stage string, doc *tasks.Document, waves []wave.WaveState) (*ProgressSummary, error) {
	specName := doc.Frontmatter.Spec
	files := CollectFilesChanged(doc.Tasks)
	techs := InferTechnologies(files)

	var done, pending, inProgress int
	for _, t := range doc.Tasks {
		switch t.Status {
		case "done":
			done++
		case "pending":
			pending++
		case "in_progress":
			inProgress++
		}
	}

	var currentWave int
	for _, w := range waves {
		if w.Status == "in_progress" || w.Status == "pending" {
			currentWave = w.WaveNum
			break
		}
		currentWave = w.WaveNum
	}

	fm := ProgressFrontmatter{
		Spec:            specName,
		Status:          "in_progress",
		Stage:           stage,
		AsOf:            time.Now().UTC(),
		TasksDone:       done,
		TasksTotal:      len(doc.Tasks),
		TasksPending:    pending,
		TasksInProgress: inProgress,
		CurrentWave:     currentWave,
		WavesTotal:      len(waves),
		FilesChanged:    files,
		Technologies:    techs,
	}

	body := renderProgressBody(specName, doc.Tasks, waves)

	return &ProgressSummary{Frontmatter: fm, Body: body}, nil
}

// buildSummaryText concatenates checkpoint pass summaries from wave records.
func buildSummaryText(waves []wave.WaveState, s *store.SpecStore) string {
	if s == nil {
		return ""
	}

	var parts []string
	dbWaves, err := s.ListWaves()
	if err != nil {
		return ""
	}
	for _, w := range dbWaves {
		if w.SummaryStatus == "pass" && w.SummaryText != "" {
			parts = append(parts, strings.TrimSpace(w.SummaryText))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

// renderCompletionBody generates the markdown body for a completion summary.
func renderCompletionBody(specName string, taskList []tasks.Task, waves []wave.WaveState, s *store.SpecStore) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Completion Summary: %s\n\n", specName))

	// Tasks Completed table.
	b.WriteString("## Tasks Completed\n\n")
	b.WriteString("| # | Task | Wave | Files |\n")
	b.WriteString("|---|------|------|-------|\n")
	for _, t := range taskList {
		files := strings.ReplaceAll(t.Files, "`", "")
		b.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n", t.ID, t.Title, t.Wave, files))
	}
	b.WriteString("\n")

	// Wave History table.
	b.WriteString("## Wave History\n\n")
	b.WriteString("| Wave | Status | Tasks | Execution Runs | Checkpoint |\n")
	b.WriteString("|------|--------|-------|----------------|------------|\n")
	for _, w := range waves {
		taskIDs := strings.Join(w.TaskIDs, ", ")
		checkpoint := "---"
		if s != nil {
			if wr, err := s.GetWave(w.WaveNum); err == nil && wr != nil && wr.SummaryStatus != "" {
				checkpoint = wr.SummaryStatus
			}
		}
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %d | %s |\n",
			w.WaveNum, w.Status, taskIDs, w.ExecRuns, checkpoint))
	}
	b.WriteString("\n")

	// Implementation Logs section (if available from store).
	if s != nil {
		logs := queryImplLogs(s)
		if len(logs) > 0 {
			b.WriteString("## Implementation Logs\n\n")
			for _, log := range logs {
				b.WriteString(fmt.Sprintf("### Task %s\n\n", log.TaskID))
				b.WriteString(log.Content)
				b.WriteString("\n\n")
			}
		}
	}

	// Metrics table.
	fileCount := 0
	for _, t := range taskList {
		for _, f := range parseFilesList(t.Files) {
			if strings.TrimSpace(f) != "" {
				fileCount++
			}
		}
	}
	b.WriteString("## Metrics\n\n")
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| Tasks | %d |\n", len(taskList)))
	b.WriteString(fmt.Sprintf("| Waves | %d |\n", len(waves)))
	b.WriteString(fmt.Sprintf("| Files Changed | %d |\n", fileCount))
	var checkpointTotal int
	for _, w := range waves {
		checkpointTotal += w.CheckRuns
	}
	b.WriteString(fmt.Sprintf("| Checkpoints | %d |\n", checkpointTotal))

	return b.String()
}

// renderProgressBody generates the markdown body for a progress summary.
func renderProgressBody(specName string, taskList []tasks.Task, waves []wave.WaveState) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Progress Summary: %s\n\n", specName))

	// Task Status overview.
	var done, pending, inProgress int
	for _, t := range taskList {
		switch t.Status {
		case "done":
			done++
		case "pending":
			pending++
		case "in_progress":
			inProgress++
		}
	}
	total := len(taskList)
	var pct float64
	if total > 0 {
		pct = float64(done) / float64(total) * 100
	}
	b.WriteString("## Task Status\n\n")
	b.WriteString(fmt.Sprintf("  Done: %d/%d (%.1f%%)\n", done, total, pct))
	if inProgress > 0 {
		b.WriteString(fmt.Sprintf("  In Progress: %d\n", inProgress))
	}
	if pending > 0 {
		b.WriteString(fmt.Sprintf("  Pending: %d\n", pending))
	}
	b.WriteString("\n")

	// Task table.
	b.WriteString("| # | Task | Status | Wave |\n")
	b.WriteString("|---|------|--------|------|\n")
	for _, t := range taskList {
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n", t.ID, t.Title, t.Status, t.Wave))
	}
	b.WriteString("\n")

	// Wave Status table (if waves exist).
	if len(waves) > 0 {
		b.WriteString("## Wave Status\n\n")
		b.WriteString("| Wave | Status | Tasks | Checkpoint |\n")
		b.WriteString("|------|--------|-------|------------|\n")
		for _, w := range waves {
			taskIDs := strings.Join(w.TaskIDs, ", ")
			checkpoint := "---"
			if w.Status == "complete" {
				checkpoint = "passed"
			} else if w.Status == "blocked" {
				checkpoint = "failed"
			}
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
				w.WaveNum, w.Status, taskIDs, checkpoint))
		}
	}

	return b.String()
}

// queryImplLogs reads implementation logs from the store database.
func queryImplLogs(s *store.SpecStore) []store.ImplLog {
	rows, err := s.DB().Query("SELECT task_id, content, content_hash, updated_at FROM impl_logs ORDER BY task_id")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var logs []store.ImplLog
	for rows.Next() {
		var l store.ImplLog
		var hash sql.NullString
		if err := rows.Scan(&l.TaskID, &l.Content, &hash, &l.UpdatedAt); err != nil {
			continue
		}
		l.ContentHash = hash.String
		logs = append(logs, l)
	}
	return logs
}
