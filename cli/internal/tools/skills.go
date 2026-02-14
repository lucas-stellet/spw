package tools

import (
	"path/filepath"
	"strings"

	"github.com/lucas-stellet/oraculo/internal/config"
)

// SkillsEffectiveSet resolves the effective skill set for a stage.
func SkillsEffectiveSet(cwd, stage string, raw bool) {
	stage = strings.ToLower(stage)
	if stage != "design" && stage != "implementation" {
		Fail("skills-effective-set requires <design|implementation>", raw)
	}

	configPath := config.ResolveConfigPath(cwd)
	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		cfg = config.Defaults()
	}

	var required, optional []string
	var enforceRequired bool

	switch stage {
	case "design":
		required = append([]string{}, cfg.Skills.Design.Required...)
		optional = append([]string{}, cfg.Skills.Design.Optional...)
		enforceRequired = cfg.Skills.Design.EnforceRequired
	case "implementation":
		required = append([]string{}, cfg.Skills.Implementation.Required...)
		optional = append([]string{}, cfg.Skills.Implementation.Optional...)
		enforceRequired = cfg.Skills.Implementation.EnforceRequired
	}

	// TDD injection: if tdd_default is true and stage is implementation,
	// add test-driven-development to required if not already present.
	if stage == "implementation" && cfg.Execution.TDDDefault {
		found := false
		for _, s := range required {
			if s == "test-driven-development" {
				found = true
				break
			}
		}
		if !found {
			required = append(required, "test-driven-development")
		}
	}

	rel, _ := filepath.Rel(cwd, configPath)
	source := "canonical"
	if strings.Contains(configPath, ".oraculo/") {
		source = "fallback"
	}

	result := map[string]any{
		"ok":               true,
		"stage":            stage,
		"required":         required,
		"optional":         optional,
		"enforce_required": enforceRequired,
		"tdd_default":      cfg.Execution.TDDDefault,
		"config_path":      rel,
		"config_source":    source,
	}

	Output(result, strings.Join(required, ","), raw)
}
