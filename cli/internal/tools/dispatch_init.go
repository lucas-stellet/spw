package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/lucas-stellet/spw/internal/config"
)

type commandMeta struct {
	Phase       string
	Category    string
	Subcategory string
	CommsPath   string // template, %s for wave-NN in wave-aware
	WaveAware   bool
}

var commandRegistry = map[string]commandMeta{
	"prd":             {Phase: "prd", Category: "pipeline", Subcategory: "research", CommsPath: "prd/_comms"},
	"design-research": {Phase: "design", Category: "pipeline", Subcategory: "research", CommsPath: "design/_comms/design-research"},
	"design-draft":    {Phase: "design", Category: "pipeline", Subcategory: "synthesis", CommsPath: "design/_comms/design-draft"},
	"tasks-plan":      {Phase: "planning", Category: "pipeline", Subcategory: "synthesis", CommsPath: "planning/_comms/tasks-plan"},
	"qa":              {Phase: "qa", Category: "pipeline", Subcategory: "synthesis", CommsPath: "qa/_comms/qa"},
	"post-mortem":     {Phase: "post-mortem", Category: "pipeline", Subcategory: "synthesis", CommsPath: "post-mortem/_comms"},
	"tasks-check":     {Phase: "planning", Category: "audit", Subcategory: "artifact", CommsPath: "planning/_comms/tasks-check"},
	"qa-check":        {Phase: "qa", Category: "audit", Subcategory: "code", CommsPath: "qa/_comms/qa-check"},
	"checkpoint":      {Phase: "execution", Category: "audit", Subcategory: "code", CommsPath: "execution/waves/wave-%s/checkpoint", WaveAware: true},
	"exec":            {Phase: "execution", Category: "wave", Subcategory: "implementation", CommsPath: "execution/waves/wave-%s/execution", WaveAware: true},
	"qa-exec":         {Phase: "qa", Category: "wave", Subcategory: "validation", CommsPath: "qa/_comms/qa-exec/waves/wave-%s", WaveAware: true},
}

var runNumRe = regexp.MustCompile(`^run-(\d+)$`)

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

	meta, ok := commandRegistry[command]
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
		commsPath = fmt.Sprintf(commsPath, fmt.Sprintf("%02d", n))
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

	cfg, _ := config.Load(cwd)
	models := map[string]string{
		"web_research":      cfg.Models.WebResearch,
		"complex_reasoning": cfg.Models.ComplexReasoning,
		"implementation":    cfg.Models.Implementation,
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
		"dispatch_policy": "dispatch-" + meta.Category,
		"models":          models,
	}
	Output(result, runDirRel, raw)
}
