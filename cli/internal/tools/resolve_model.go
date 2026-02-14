package tools

import (
	"github.com/lucas-stellet/oraculo/internal/config"
)

// ResolveModel maps a model alias to the resolved model name from config.
func ResolveModel(cwd, alias string, raw bool) {
	if alias == "" {
		Fail("resolve-model requires <alias>", raw)
	}

	cfg, _ := config.Load(cwd)

	var model string
	switch alias {
	case "web_research":
		model = cfg.Models.WebResearch
	case "complex_reasoning":
		model = cfg.Models.ComplexReasoning
	case "implementation":
		model = cfg.Models.Implementation
	default:
		Fail("unknown model alias: "+alias+"; use web_research, complex_reasoning, or implementation", raw)
	}

	result := map[string]any{
		"ok":     true,
		"alias":  alias,
		"model":  model,
		"source": "config",
	}
	Output(result, model, raw)
}
