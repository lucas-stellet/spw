package hook

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/workspace"
)

var (
	managedArtifactRe = regexp.MustCompile(`(?i)^(DESIGN-RESEARCH|TASKS-CHECK|CHECKPOINT-REPORT|STATUS-SUMMARY|SKILLS-[A-Z0-9-]+|PRD(-[A-Z0-9-]+)?|PRD-SOURCE-NOTES|PRD-STRUCTURE|PRD-REVISION-(PLAN|QUESTIONS|NOTES))\.md$`)
	waveIDRe          = regexp.MustCompile(`^wave-\d{2}$`)
)

// HandleGuardPaths validates that Write/Edit operations target valid paths.
func HandleGuardPaths() error {
	ctx := newHookContext()
	if !ctx.cfg.Hooks.Enabled || (!ctx.cfg.Hooks.GuardPaths && !ctx.cfg.Hooks.GuardWaveLayout) {
		return nil
	}

	resolved := resolveTargetPath(ctx.payload, ctx.workspaceRoot)
	if resolved == nil {
		return nil
	}

	relPath := normalizeSlashes(resolved.relPath)
	baseName := filepath.Base(resolved.absPath)

	if ctx.cfg.Hooks.GuardPaths {
		isSpecLocal := strings.Contains(relPath, ".spec-workflow/specs/")
		if isManagedArtifactFile(baseName) && !isSpecLocal {
			emitViolation(ctx.cfg.Hooks, "SPW artifact path violation", []string{
				"File: " + relPath,
				"Managed SPW artifacts must stay under .spec-workflow/specs/<spec-name>/",
			})
		}
	}

	if ctx.cfg.Hooks.GuardWaveLayout {
		// Block legacy _agent-comms/ paths
		if strings.Contains(relPath, "_agent-comms/") {
			emitViolation(ctx.cfg.Hooks, "Legacy _agent-comms/ path is not allowed", []string{
				"File: " + relPath,
				"Use phase-based _comms/ directories instead (e.g. execution/waves/, qa/_comms/)",
			})
		}

		// Validate execution wave format
		if strings.Contains(relPath, "execution/waves/") {
			validateExecutionWave(ctx.cfg.Hooks, relPath)
		}

		// Validate QA exec wave format
		if strings.Contains(relPath, "qa/_comms/qa-exec/waves/") {
			validateQAWave(ctx.cfg.Hooks, relPath)
		}
	}

	return nil
}

func validateExecutionWave(hooks config.HooksConfig, relPath string) {
	waveRe := regexp.MustCompile(`execution/waves/([^/]+)`)
	if m := waveRe.FindStringSubmatch(relPath); m != nil {
		if !waveIDRe.MatchString(m[1]) {
			emitViolation(hooks, "Wave folder must use zero-padded format", []string{
				"Found wave folder: " + m[1],
				"Expected format: wave-01, wave-02, ...",
			})
		}
	}

	stageRe := regexp.MustCompile(`execution/waves/wave-\d{2}/([^/]+)`)
	if m := stageRe.FindStringSubmatch(relPath); m != nil {
		stage := m[1]
		allowed := map[string]bool{
			"execution":          true,
			"checkpoint":         true,
			"post-check":         true,
			"_wave-summary.json": true,
			"_latest.json":       true,
		}
		if !allowed[stage] {
			emitViolation(hooks, "Invalid wave stage folder", []string{
				"File: " + relPath,
				"Allowed wave entries: execution, checkpoint, post-check, _wave-summary.json, _latest.json",
			})
		}
	}
}

func validateQAWave(hooks config.HooksConfig, relPath string) {
	qaWaveRe := regexp.MustCompile(`qa-exec/waves/([^/]+)`)
	if m := qaWaveRe.FindStringSubmatch(relPath); m != nil {
		if !waveIDRe.MatchString(m[1]) {
			emitViolation(hooks, "QA exec wave folder must use zero-padded format", []string{
				"Found wave folder: " + m[1],
				"Expected format: wave-01, wave-02, ...",
			})
		}
	}
}

func isManagedArtifactFile(baseName string) bool {
	return managedArtifactRe.MatchString(baseName)
}

// resolvedPath holds normalized paths for a tool target.
type resolvedPath struct {
	raw     string
	absPath string
	relPath string
}

func resolveTargetPath(p workspace.Payload, workspaceRoot string) *resolvedPath {
	var input map[string]any
	if p.ToolInput != nil {
		if err := json.Unmarshal(p.ToolInput, &input); err != nil {
			return nil
		}
	}

	var filePath string
	for _, key := range []string{"file_path", "path", "target_path", "filename"} {
		if v, ok := input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				filePath = s
				break
			}
		}
	}
	if filePath == "" {
		return nil
	}

	cwd := p.CWD
	if cwd == "" && p.Workspace != nil {
		cwd = p.Workspace.CurrentDir
	}
	if cwd == "" {
		cwd = workspaceRoot
	}

	var absPath string
	if filepath.IsAbs(filePath) {
		absPath = filePath
	} else {
		absPath = filepath.Join(cwd, filePath)
	}

	rel, err := filepath.Rel(workspaceRoot, absPath)
	if err != nil {
		rel = filePath
	}

	return &resolvedPath{
		raw:     filePath,
		absPath: absPath,
		relPath: rel,
	}
}

func normalizeSlashes(s string) string {
	return strings.ReplaceAll(s, "\\", "/")
}
