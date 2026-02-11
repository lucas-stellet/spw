package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/render"
	"github.com/lucas-stellet/spw/internal/workspace"
)

// HandleSessionStart syncs the tasks template variant based on TDD config
// and triggers a re-render of workflows if the config is newer than rendered files.
func HandleSessionStart() error {
	ctx := newHookContext()

	if err := syncTasksTemplate(ctx.workspaceRoot, ctx.cfg); err != nil {
		logHook(fmt.Sprintf("Template sync error: %v", err))
	}

	if err := reRenderIfStale(ctx.workspaceRoot, ctx.cfg); err != nil {
		logHook(fmt.Sprintf("Re-render error: %v", err))
	}

	return nil
}

// reRenderIfStale checks if config mtime > oldest rendered workflow mtime,
// and re-renders all workflows if so.
func reRenderIfStale(workspaceRoot string, cfg config.Config) error {
	configPath := config.ResolveConfigPath(workspaceRoot)
	configInfo, err := os.Stat(configPath)
	if err != nil {
		return nil // No config, nothing to re-render.
	}

	outDir := filepath.Join(workspaceRoot, ".claude", "workflows", "spw")
	entries, err := os.ReadDir(outDir)
	if err != nil || len(entries) == 0 {
		return nil // No rendered workflows yet.
	}

	// Find oldest rendered workflow.
	configMtime := configInfo.ModTime()
	stale := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if configMtime.After(info.ModTime()) {
			stale = true
			break
		}
	}

	if !stale {
		return nil
	}

	logHook("Config is newer than rendered workflows. Re-rendering...")

	engine, err := render.New(cfg)
	if err != nil {
		return fmt.Errorf("creating render engine: %w", err)
	}

	// Load guidelines.
	if gs := workspace.LoadGuidelines(workspaceRoot); len(gs) > 0 {
		adapted := make([]struct {
			Name      string
			Content   string
			AppliesTo []string
		}, len(gs))
		for i, g := range gs {
			adapted[i].Name = g.Name
			adapted[i].Content = g.Content
			adapted[i].AppliesTo = g.AppliesTo
		}
		engine.SetGuidelines(adapted)
	}

	results, err := engine.RenderAll()
	if err != nil {
		return fmt.Errorf("rendering workflows: %w", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for cmd, content := range results {
		path := filepath.Join(outDir, cmd+".md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", cmd, err)
		}
	}
	logHook(fmt.Sprintf("Re-rendered %d workflows.", len(results)))

	return nil
}

func syncTasksTemplate(workspaceRoot string, cfg config.Config) error {
	if !cfg.Templates.SyncTasksTemplateOnSessionStart {
		logHook("Sync disabled by configuration.")
		return nil
	}

	mode := resolveTemplateMode(cfg)
	if mode != "on" && mode != "off" {
		logHook(fmt.Sprintf("Invalid tasks_template_mode: '%s'. Use auto|on|off.", cfg.Templates.TasksTemplateMode))
		return nil
	}

	variantsDir := filepath.Join(workspaceRoot, ".spec-workflow", "user-templates", "variants")
	targetPath := filepath.Join(workspaceRoot, ".spec-workflow", "user-templates", "tasks-template.md")

	var sourceFile string
	if mode == "on" {
		sourceFile = "tasks-template.tdd-on.md"
	} else {
		sourceFile = "tasks-template.tdd-off.md"
	}
	sourcePath := filepath.Join(variantsDir, sourceFile)

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		logHook("Source template not found: " + sourcePath)
		return nil
	}

	// Check if already in sync
	if filesEqual(sourcePath, targetPath) {
		logHook(fmt.Sprintf("Template is already synchronized (%s).", mode))
		return nil
	}

	// Ensure target dir exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("creating template dir: %w", err)
	}

	// Backup if enabled
	if cfg.Safety.BackupBeforeOverwrite {
		if _, err := os.Stat(targetPath); err == nil {
			backupPath := targetPath + ".bak"
			data, err := os.ReadFile(targetPath)
			if err == nil {
				_ = os.WriteFile(backupPath, data, 0644)
				logHook("Backup created: " + backupPath)
			}
		}
	}

	// Copy source to target
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading source template: %w", err)
	}
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("writing target template: %w", err)
	}
	logHook(fmt.Sprintf("Template synchronized (%s): %s", mode, targetPath))

	return nil
}

func resolveTemplateMode(cfg config.Config) string {
	mode := strings.ToLower(strings.TrimSpace(cfg.Templates.TasksTemplateMode))
	if mode == "auto" {
		if cfg.Execution.TDDDefault {
			return "on"
		}
		return "off"
	}
	return mode
}

func filesEqual(a, b string) bool {
	dataA, errA := os.ReadFile(a)
	dataB, errB := os.ReadFile(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(dataA) == string(dataB)
}

func logHook(msg string) {
	fmt.Fprintf(os.Stderr, "[spw-hook] %s\n", msg)
}
