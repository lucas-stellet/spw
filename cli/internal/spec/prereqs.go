package spec

import (
	"path/filepath"

	"github.com/lucas-stellet/spw/internal/specdir"
)

// commandPrereqs maps SPW commands to their prerequisite artifact paths
// (relative to spec directory).
var commandPrereqs = map[string][]string{
	"prd":             {},
	"design-research": {specdir.RequirementsMD},
	"design-draft":    {specdir.RequirementsMD, specdir.DesignResearchMD},
	"tasks-plan":      {specdir.DesignMD},
	"tasks-check":     {specdir.TasksMD},
	"exec":            {specdir.TasksMD},
	"checkpoint":      {specdir.TasksMD, specdir.WavesDir},
	"qa":              {specdir.TasksMD},
	"qa-check":        {specdir.QATestPlan},
	"qa-exec":         {specdir.QATestPlan},
	"post-mortem":     {specdir.TasksMD},
}

// CheckPrereqs verifies prerequisites are met for a given SPW command.
func CheckPrereqs(specDir, command string) PrereqResult {
	reqs, ok := commandPrereqs[command]
	if !ok {
		// Unknown command — no prerequisites defined
		return PrereqResult{Ready: true}
	}

	if len(reqs) == 0 {
		return PrereqResult{Ready: true}
	}

	var missing []string
	for _, rel := range reqs {
		full := filepath.Join(specDir, rel)
		if !specdir.FileExists(full) && !specdir.DirExists(full) {
			missing = append(missing, rel)
		}
	}

	if len(missing) > 0 {
		return PrereqResult{
			Ready:   false,
			Missing: missing,
		}
	}

	return PrereqResult{Ready: true}
}
