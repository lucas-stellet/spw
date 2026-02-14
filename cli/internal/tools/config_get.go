package tools

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-stellet/oraculo/internal/config"
)

// ConfigGet reads a config value by section.key path.
func ConfigGet(cwd string, key string, defaultValue string, raw bool) {
	if key == "" {
		Fail("config-get requires <section.key>", raw)
	}

	configPath := config.ResolveConfigPath(cwd)
	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		cfg = config.Defaults()
	}

	value := cfg.GetValue(key, defaultValue)

	rel, _ := filepath.Rel(cwd, configPath)
	source := "canonical"
	if strings.Contains(configPath, ".oraculo/") {
		source = "fallback"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		source = "missing"
	}

	result := map[string]any{
		"ok":            true,
		"key":           key,
		"value":         value,
		"config_path":   rel,
		"config_source": source,
	}

	Output(result, value, raw)
}
