package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/embedded"
	"github.com/lucas-stellet/spw/internal/registry"
)

var runNumRe = regexp.MustCompile(`^run-(\d+)$`)

// loadedRegistry is the lazily-loaded command registry from embedded workflow files.
var loadedRegistry map[string]registry.CommandMeta

// getRegistry returns the command registry, loading it once from embedded files.
func getRegistry() map[string]registry.CommandMeta {
	if loadedRegistry == nil {
		reg, err := registry.Load(embedded.Workflows)
		if err != nil {
			// Should not happen with valid embedded files.
			panic("failed to load command registry: " + err.Error())
		}
		loadedRegistry = reg
	}
	return loadedRegistry
}

// DispatchInit creates a run-NNN directory for a command dispatch.
func DispatchInit(cwd, command, specName, wave string, raw bool) {
	if command == "" || specName == "" {
		Fail("dispatch-init requires <command> <spec-name>", raw)
	}

	specRel := filepath.Join(".spec-workflow", "specs", specName)
	specDir := filepath.Join(cwd, specRel)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		Fail("failed to create spec directory: "+err.Error(), raw)
	}

	reg := getRegistry()
	meta, ok := reg[command]
	if !ok {
		Fail("unknown command: "+command, raw)
	}

	commsPath := meta.CommsPath
	if meta.WaveAware {
		if wave == "" {
			Fail("wave-aware command "+command+" requires --wave", raw)
		}
		n, err := strconv.Atoi(wave)
		if err != nil {
			Fail("wave must be a number: "+wave, raw)
		}
		commsPath = strings.Replace(commsPath, "{wave}", fmt.Sprintf("%02d", n), 1)
	} else {
		if wave != "" {
			Fail("command "+command+" does not accept --wave", raw)
		}
	}

	commsDir := filepath.Join(specDir, commsPath)

	// Scan for next run number
	nextRun := 1
	entries, err := os.ReadDir(commsDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			m := runNumRe.FindStringSubmatch(e.Name())
			if m == nil {
				continue
			}
			n, _ := strconv.Atoi(m[1])
			if n >= nextRun {
				nextRun = n + 1
			}
		}
	}

	runID := fmt.Sprintf("run-%03d", nextRun)
	runDir := filepath.Join(commsDir, runID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		Fail("failed to create run dir: "+err.Error(), raw)
	}

	// Create artifact directories declared in dispatch_pattern.
	for _, artifact := range meta.Artifacts {
		artifactDir := filepath.Join(specDir, artifact)
		if err := os.MkdirAll(artifactDir, 0755); err != nil {
			Fail("failed to create artifact dir: "+err.Error(), raw)
		}
	}

	cfg, _ := config.Load(cwd)
	models := map[string]string{
		"web_research":      cfg.Models.WebResearch,
		"complex_reasoning": cfg.Models.ComplexReasoning,
		"implementation":    cfg.Models.Implementation,
	}

	execution := map[string]any{
		"tdd_default":                          cfg.Execution.TDDDefault,
		"require_user_approval_between_waves":  cfg.Execution.RequireUserApprovalBetweenWaves,
		"commit_per_task":                      cfg.Execution.CommitPerTask,
		"require_clean_worktree_for_wave_pass": cfg.Execution.RequireCleanWorktreeForWavePass,
	}
	planning := map[string]any{
		"tasks_generation_strategy": cfg.Planning.TasksGenerationStrategy,
		"max_wave_size":             cfg.Planning.MaxWaveSize,
	}

	runDirRel, _ := filepath.Rel(cwd, runDir)
	result := map[string]any{
		"ok":              true,
		"run_dir":         runDirRel,
		"run_id":          runID,
		"spec_dir":        specRel,
		"phase":           meta.Phase,
		"command":         command,
		"category":        meta.Category,
		"subcategory":     meta.Subcategory,
		"dispatch_policy": meta.DispatchPolicy(),
		"models":          models,
		"execution":       execution,
		"planning":        planning,
	}
	Output(result, runDirRel, raw)
}
