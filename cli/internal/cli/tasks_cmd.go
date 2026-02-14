package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-stellet/oraculo/internal/specdir"
	"github.com/lucas-stellet/oraculo/internal/tasks"
	"github.com/lucas-stellet/oraculo/internal/tools"
	"github.com/spf13/cobra"
)

func newTasksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Task state resolution commands",
		Long:  "Parse tasks.md body-first and resolve task state deterministically.",
	}

	cmd.AddCommand(newTasksStateCmd())
	cmd.AddCommand(newTasksNextCmd())
	cmd.AddCommand(newTasksMarkCmd())
	cmd.AddCommand(newTasksCountCmd())
	cmd.AddCommand(newTasksFilesCmd())
	cmd.AddCommand(newTasksValidateCmd())
	cmd.AddCommand(newTasksComplexityCmd())

	return cmd
}

func newTasksStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state <spec-name>",
		Short: "Show full task state for a spec",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			doc, err := tasks.ParseFile(specdir.TasksPath(specDir))
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			result := map[string]any{
				"ok":       true,
				"spec":     specName,
				"tasks":    doc.Tasks,
				"counts":   doc.Count(),
				"warnings": doc.Warnings,
			}
			tools.Output(result, "", raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newTasksNextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next <spec-name>",
		Short: "Resolve the next executable wave",
		Long:  "Determines which tasks are executable next, including deferred tasks with resolved dependencies.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			doc, err := tasks.ParseFile(specdir.TasksPath(specDir))
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			next := tasks.ResolveNextWave(doc, specDir)
			result := map[string]any{
				"ok":             true,
				"spec":           specName,
				"action":         next.Action,
				"wave":           next.Wave,
				"task_ids":       next.TaskIDs,
				"deferred_ready": next.DeferredReady,
				"reason":         next.Reason,
				"warnings":       next.Warnings,
			}
			tools.Output(result, next.Action, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newTasksMarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark <spec-name> <task-id> <status>",
		Short: "Update a task's checkbox status",
		Long:  "Surgically updates a single task checkbox. Status: done, in_progress, pending.",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			requireImplLog, _ := cmd.Flags().GetBool("require-impl-log")
			cwd := getCwd()
			specName := args[0]
			taskID := args[1]
			newStatus := args[2]

			if newStatus != "done" && newStatus != "in_progress" && newStatus != "pending" {
				tools.Fail("status must be one of: done, in_progress, pending", raw)
			}

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			filePath := specdir.TasksPath(specDir)
			if err := tasks.MarkTaskInFile(filePath, taskID, newStatus, requireImplLog, specDir); err != nil {
				tools.Fail(err.Error(), raw)
			}

			result := map[string]any{
				"ok":      true,
				"spec":    specName,
				"task_id": taskID,
				"status":  newStatus,
			}
			tools.Output(result, "ok", raw)
		},
	}
	cmd.Flags().Bool("require-impl-log", false, "Refuse to mark done unless implementation log exists")
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newTasksCountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count <spec-name>",
		Short: "Count tasks by status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			doc, err := tasks.ParseFile(specdir.TasksPath(specDir))
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			counts := doc.Count()
			result := map[string]any{
				"ok":     true,
				"spec":   specName,
				"counts": counts,
			}
			tools.Output(result, fmt.Sprintf("%d/%d", counts.Done, counts.Total), raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newTasksFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files <spec-name> <task-id>",
		Short: "List files for a specific task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]
			taskID := args[1]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			doc, err := tasks.ParseFile(specdir.TasksPath(specDir))
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			task := doc.TaskByID(taskID)
			if task == nil {
				tools.Fail(fmt.Sprintf("task %s not found", taskID), raw)
			}

			// Parse backtick-delimited files from the Files field
			var files []string
			if task.Files != "" {
				parts := strings.Split(task.Files, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					p = strings.Trim(p, "`")
					if p != "" {
						files = append(files, p)
					}
				}
			}

			result := map[string]any{
				"ok":      true,
				"spec":    specName,
				"task_id": taskID,
				"files":   files,
			}
			rawVal := ""
			if len(files) > 0 {
				rawVal = strings.Join(files, "\n")
			}
			tools.Output(result, rawVal, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newTasksValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <spec-name>",
		Short: "Validate tasks.md against dashboard rules",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			data, err := readFileContent(filepath.Join(specDir, specdir.TasksMD))
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			vr := tasks.Validate(data)
			result := map[string]any{
				"ok":     true,
				"spec":   specName,
				"valid":  vr.Valid,
				"errors": vr.Errors,
			}
			rawVal := "valid"
			if !vr.Valid {
				rawVal = "invalid"
			}
			tools.Output(result, rawVal, raw)
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func newTasksComplexityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complexity <spec-name> [task-id]",
		Short: "Score task complexity for model routing",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := cmd.Flags().GetBool("raw")
			cwd := getCwd()
			specName := args[0]

			specDir, err := specdir.Resolve(cwd, specName)
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			doc, err := tasks.ParseFile(specdir.TasksPath(specDir))
			if err != nil {
				tools.Fail(err.Error(), raw)
			}

			if len(args) == 2 {
				taskID := args[1]
				task := doc.TaskByID(taskID)
				if task == nil {
					tools.Fail(fmt.Sprintf("task %s not found", taskID), raw)
				}
				cr := tasks.ScoreComplexity(*task)
				result := map[string]any{
					"ok":         true,
					"spec":       specName,
					"task_id":    taskID,
					"score":      cr.Score,
					"model_hint": cr.ModelHint,
					"factors":    cr.Factors,
				}
				tools.Output(result, cr.ModelHint, raw)
			} else {
				var scores []tasks.ComplexityResult
				for _, t := range doc.Tasks {
					scores = append(scores, tasks.ScoreComplexity(t))
				}
				result := map[string]any{
					"ok":     true,
					"spec":   specName,
					"scores": scores,
				}
				tools.Output(result, "", raw)
			}
		},
	}
	cmd.Flags().Bool("raw", false, "Output raw value without JSON wrapping")
	return cmd
}

func readFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
