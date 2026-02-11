// Package config handles parsing and merging of spw-config.toml.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the full spw-config.toml structure.
type Config struct {
	Models           ModelsConfig           `toml:"models"`
	Execution        ExecutionConfig        `toml:"execution"`
	Planning         PlanningConfig         `toml:"planning"`
	QA               QAConfig               `toml:"qa"`
	PostMortemMemory PostMortemMemoryConfig `toml:"post_mortem_memory"`
	AgentTeams       AgentTeamsConfig       `toml:"agent_teams"`
	Skills           SkillsConfig           `toml:"skills"`
	Templates        TemplatesConfig        `toml:"templates"`
	Statusline       StatuslineConfig       `toml:"statusline"`
	Safety           SafetyConfig           `toml:"safety"`
	Hooks            HooksConfig            `toml:"hooks"`
}

type ModelsConfig struct {
	WebResearch      string `toml:"web_research"`
	ComplexReasoning string `toml:"complex_reasoning"`
	Implementation   string `toml:"implementation"`
}

type ExecutionConfig struct {
	TDDDefault                        bool   `toml:"tdd_default"`
	RequireUserApprovalBetweenWaves   bool   `toml:"require_user_approval_between_waves"`
	CommitPerTask                     string `toml:"commit_per_task"`
	RequireCleanWorktreeForWavePass   bool   `toml:"require_clean_worktree_for_wave_pass"`
	ManualTasksRequireHumanHandoff    bool   `toml:"manual_tasks_require_human_handoff"`
}

type PlanningConfig struct {
	TasksGenerationStrategy string `toml:"tasks_generation_strategy"`
	MaxWaveSize             int    `toml:"max_wave_size"`
}

type QAConfig struct {
	MaxScenariosPerWave int `toml:"max_scenarios_per_wave"`
}

type PostMortemMemoryConfig struct {
	Enabled              bool `toml:"enabled"`
	MaxEntriesForDesign  int  `toml:"max_entries_for_design"`
}

type AgentTeamsConfig struct {
	Enabled             bool     `toml:"enabled"`
	TeammateMode        string   `toml:"teammate_mode"`
	RequireDelegateMode bool     `toml:"require_delegate_mode"`
	MaxTeammates        int      `toml:"max_teammates"`
	ExcludePhases       []string `toml:"exclude_phases"`
}

type SkillsStageConfig struct {
	EnforceRequired bool     `toml:"enforce_required"`
	Required        []string `toml:"required"`
	Optional        []string `toml:"optional"`
}

type SkillsConfig struct {
	Enabled                        bool              `toml:"enabled"`
	AutoInstallDefaultsOnSPWInstall bool             `toml:"auto_install_defaults_on_spw_install"`
	Design                         SkillsStageConfig `toml:"design"`
	Implementation                 SkillsStageConfig `toml:"implementation"`
}

type TemplatesConfig struct {
	SyncTasksTemplateOnSessionStart bool   `toml:"sync_tasks_template_on_session_start"`
	TasksTemplateMode               string `toml:"tasks_template_mode"`
}

type StatuslineConfig struct {
	CacheTTLSeconds int      `toml:"cache_ttl_seconds"`
	BaseBranches    []string `toml:"base_branches"`
	StickySpec      bool     `toml:"sticky_spec"`
}

type SafetyConfig struct {
	BackupBeforeOverwrite bool `toml:"backup_before_overwrite"`
}

type HooksConfig struct {
	Enabled                  bool   `toml:"enabled"`
	EnforcementMode          string `toml:"enforcement_mode"`
	Verbose                  bool   `toml:"verbose"`
	RecentRunWindowMinutes   int    `toml:"recent_run_window_minutes"`
	GuardPromptRequireSpec   bool   `toml:"guard_prompt_require_spec"`
	GuardPaths               bool   `toml:"guard_paths"`
	GuardWaveLayout          bool   `toml:"guard_wave_layout"`
	GuardStopHandoff         bool   `toml:"guard_stop_handoff"`
}

// Defaults returns a Config populated with all default values.
func Defaults() Config {
	return Config{
		Models: ModelsConfig{
			WebResearch:      "haiku",
			ComplexReasoning: "opus",
			Implementation:   "sonnet",
		},
		Execution: ExecutionConfig{
			TDDDefault:                      false,
			RequireUserApprovalBetweenWaves: true,
			CommitPerTask:                   "auto",
			RequireCleanWorktreeForWavePass: true,
			ManualTasksRequireHumanHandoff:  true,
		},
		Planning: PlanningConfig{
			TasksGenerationStrategy: "rolling-wave",
			MaxWaveSize:             3,
		},
		QA: QAConfig{
			MaxScenariosPerWave: 5,
		},
		PostMortemMemory: PostMortemMemoryConfig{
			Enabled:             true,
			MaxEntriesForDesign: 5,
		},
		AgentTeams: AgentTeamsConfig{
			Enabled:             false,
			TeammateMode:        "in-process",
			RequireDelegateMode: true,
			MaxTeammates:        4,
			ExcludePhases:       []string{},
		},
		Skills: SkillsConfig{
			Enabled:                         true,
			AutoInstallDefaultsOnSPWInstall: true,
			Design: SkillsStageConfig{
				EnforceRequired: true,
				Required:        []string{},
				Optional:        []string{},
			},
			Implementation: SkillsStageConfig{
				EnforceRequired: true,
				Required:        []string{},
				Optional:        []string{},
			},
		},
		Templates: TemplatesConfig{
			SyncTasksTemplateOnSessionStart: true,
			TasksTemplateMode:               "auto",
		},
		Statusline: StatuslineConfig{
			CacheTTLSeconds: 10,
			BaseBranches:    []string{"main", "master", "staging", "develop"},
			StickySpec:      true,
		},
		Safety: SafetyConfig{
			BackupBeforeOverwrite: true,
		},
		Hooks: HooksConfig{
			Enabled:                true,
			EnforcementMode:        "warn",
			Verbose:                true,
			RecentRunWindowMinutes: 30,
			GuardPromptRequireSpec: true,
			GuardPaths:             true,
			GuardWaveLayout:        true,
			GuardStopHandoff:       true,
		},
	}
}

// ResolveConfigPath finds the spw-config.toml file path.
// Checks canonical path first, then legacy fallback.
func ResolveConfigPath(workspaceRoot string) string {
	canonical := filepath.Join(workspaceRoot, ".spec-workflow", "spw-config.toml")
	if _, err := os.Stat(canonical); err == nil {
		return canonical
	}
	legacy := filepath.Join(workspaceRoot, ".spw", "spw-config.toml")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return canonical
}

// Load reads spw-config.toml from the workspace root.
// Returns defaults if the file doesn't exist.
func Load(workspaceRoot string) (Config, error) {
	cfg := Defaults()
	configPath := ResolveConfigPath(workspaceRoot)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config %s: %w", configPath, err)
	}

	// Normalize enforcement_mode
	cfg.Hooks.EnforcementMode = normalizeEnforcementMode(cfg.Hooks.EnforcementMode)

	return cfg, nil
}

// LoadFromPath reads a config from a specific file path.
func LoadFromPath(configPath string) (Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config %s: %w", configPath, err)
	}

	cfg.Hooks.EnforcementMode = normalizeEnforcementMode(cfg.Hooks.EnforcementMode)

	return cfg, nil
}

// GetValue retrieves a config value by dot-separated key path (e.g., "models.web_research").
// Returns the value as a string, or the defaultValue if not found.
func (c *Config) GetValue(key string, defaultValue string) string {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) < 2 {
		return defaultValue
	}

	section := parts[0]
	field := parts[1]

	v := reflect.ValueOf(*c)
	t := v.Type()

	// Find the section struct field
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("toml")
		if tag == section {
			sectionVal := v.Field(i)
			return getFieldValue(sectionVal, field, defaultValue)
		}
	}

	// Try nested sections like "skills.design"
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("toml")
		sectionVal := v.Field(i)
		if sectionVal.Kind() != reflect.Struct {
			continue
		}
		st := sectionVal.Type()
		for j := 0; j < st.NumField(); j++ {
			subTag := st.Field(j).Tag.Get("toml")
			if tag+"."+subTag == section {
				nestedVal := sectionVal.Field(j)
				return getFieldValue(nestedVal, field, defaultValue)
			}
		}
	}

	return defaultValue
}

func getFieldValue(structVal reflect.Value, fieldName string, defaultValue string) string {
	if structVal.Kind() != reflect.Struct {
		return defaultValue
	}

	t := structVal.Type()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("toml")
		if tag == fieldName {
			return formatValue(structVal.Field(i))
		}
	}
	return defaultValue
}

func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Slice:
		items := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = formatValue(v.Index(i))
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

func normalizeEnforcementMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "block" {
		return "block"
	}
	return "warn"
}

// ToBool converts a string value to bool, matching JS toBool behavior.
func ToBool(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return fallback
	}
}

// ToInt converts a string value to int, matching JS toInt behavior.
func ToInt(value string, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}
