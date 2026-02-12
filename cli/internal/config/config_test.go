package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	// Models
	if cfg.Models.WebResearch != "haiku" {
		t.Errorf("Models.WebResearch = %q, want %q", cfg.Models.WebResearch, "haiku")
	}
	if cfg.Models.ComplexReasoning != "opus" {
		t.Errorf("Models.ComplexReasoning = %q, want %q", cfg.Models.ComplexReasoning, "opus")
	}
	if cfg.Models.Implementation != "sonnet" {
		t.Errorf("Models.Implementation = %q, want %q", cfg.Models.Implementation, "sonnet")
	}

	// Execution
	if cfg.Execution.TDDDefault != false {
		t.Error("Execution.TDDDefault should default to false")
	}
	if cfg.Execution.RequireUserApprovalBetweenWaves != true {
		t.Error("Execution.RequireUserApprovalBetweenWaves should default to true")
	}
	if cfg.Execution.CommitPerTask != "auto" {
		t.Errorf("Execution.CommitPerTask = %q, want %q", cfg.Execution.CommitPerTask, "auto")
	}

	// Planning
	if cfg.Planning.TasksGenerationStrategy != "rolling-wave" {
		t.Errorf("Planning.TasksGenerationStrategy = %q, want %q", cfg.Planning.TasksGenerationStrategy, "rolling-wave")
	}
	if cfg.Planning.MaxWaveSize != 3 {
		t.Errorf("Planning.MaxWaveSize = %d, want %d", cfg.Planning.MaxWaveSize, 3)
	}

	// Hooks
	if cfg.Hooks.Enabled != true {
		t.Error("Hooks.Enabled should default to true")
	}
	if cfg.Hooks.EnforcementMode != "warn" {
		t.Errorf("Hooks.EnforcementMode = %q, want %q", cfg.Hooks.EnforcementMode, "warn")
	}
	if cfg.Hooks.RecentRunWindowMinutes != 30 {
		t.Errorf("Hooks.RecentRunWindowMinutes = %d, want %d", cfg.Hooks.RecentRunWindowMinutes, 30)
	}

	// Statusline
	if cfg.Statusline.CacheTTLSeconds != 10 {
		t.Errorf("Statusline.CacheTTLSeconds = %d, want %d", cfg.Statusline.CacheTTLSeconds, 10)
	}
	if !cfg.Statusline.StickySpec {
		t.Error("Statusline.StickySpec should default to true")
	}

	// Agent Teams
	if cfg.AgentTeams.Enabled {
		t.Error("AgentTeams.Enabled should default to false")
	}
	if cfg.AgentTeams.MaxTeammates != 4 {
		t.Errorf("AgentTeams.MaxTeammates = %d, want %d", cfg.AgentTeams.MaxTeammates, 4)
	}
}

func TestParseActualConfig(t *testing.T) {
	// Find the actual spw-config.toml relative to the repo root
	configPath := findRepoFile(t, "config/spw-config.toml")
	if configPath == "" {
		t.Skip("config/spw-config.toml not found")
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath(%q): %v", configPath, err)
	}

	// Verify all fields parsed correctly from the actual config
	if cfg.Models.WebResearch != "haiku" {
		t.Errorf("Models.WebResearch = %q, want %q", cfg.Models.WebResearch, "haiku")
	}
	if cfg.Models.ComplexReasoning != "opus" {
		t.Errorf("Models.ComplexReasoning = %q, want %q", cfg.Models.ComplexReasoning, "opus")
	}
	if cfg.Models.Implementation != "sonnet" {
		t.Errorf("Models.Implementation = %q, want %q", cfg.Models.Implementation, "sonnet")
	}
	if cfg.Execution.CommitPerTask != "auto" {
		t.Errorf("Execution.CommitPerTask = %q, want %q", cfg.Execution.CommitPerTask, "auto")
	}
	if cfg.Execution.RequireCleanWorktreeForWavePass != true {
		t.Error("Execution.RequireCleanWorktreeForWavePass should be true")
	}
	if cfg.Execution.ManualTasksRequireHumanHandoff != true {
		t.Error("Execution.ManualTasksRequireHumanHandoff should be true")
	}
	if cfg.Planning.TasksGenerationStrategy != "rolling-wave" {
		t.Errorf("Planning.TasksGenerationStrategy = %q, want %q", cfg.Planning.TasksGenerationStrategy, "rolling-wave")
	}
	if cfg.QA.MaxScenariosPerWave != 5 {
		t.Errorf("QA.MaxScenariosPerWave = %d, want %d", cfg.QA.MaxScenariosPerWave, 5)
	}
	if cfg.PostMortemMemory.Enabled != true {
		t.Error("PostMortemMemory.Enabled should be true")
	}
	if cfg.PostMortemMemory.MaxEntriesForDesign != 5 {
		t.Errorf("PostMortemMemory.MaxEntriesForDesign = %d, want %d", cfg.PostMortemMemory.MaxEntriesForDesign, 5)
	}
	if cfg.AgentTeams.TeammateMode != "in-process" {
		t.Errorf("AgentTeams.TeammateMode = %q, want %q", cfg.AgentTeams.TeammateMode, "in-process")
	}
	if cfg.AgentTeams.RequireDelegateMode != true {
		t.Error("AgentTeams.RequireDelegateMode should be true")
	}
	if cfg.Skills.Enabled != true {
		t.Error("Skills.Enabled should be true")
	}
	if cfg.Skills.AutoInstallDefaultsOnSPWInstall != true {
		t.Error("Skills.AutoInstallDefaultsOnSPWInstall should be true")
	}
	if cfg.Skills.Design.EnforceRequired != true {
		t.Error("Skills.Design.EnforceRequired should be true")
	}
	if len(cfg.Skills.Design.Required) == 0 {
		t.Error("Skills.Design.Required should have entries")
	}
	if len(cfg.Skills.Design.Optional) == 0 {
		t.Error("Skills.Design.Optional should have entries")
	}
	if len(cfg.Skills.Implementation.Required) == 0 {
		t.Error("Skills.Implementation.Required should have entries")
	}
	if cfg.Templates.SyncTasksTemplateOnSessionStart != true {
		t.Error("Templates.SyncTasksTemplateOnSessionStart should be true")
	}
	if cfg.Templates.TasksTemplateMode != "auto" {
		t.Errorf("Templates.TasksTemplateMode = %q, want %q", cfg.Templates.TasksTemplateMode, "auto")
	}
	if len(cfg.Statusline.BaseBranches) != 4 {
		t.Errorf("Statusline.BaseBranches has %d entries, want 4", len(cfg.Statusline.BaseBranches))
	}
	if cfg.Safety.BackupBeforeOverwrite != true {
		t.Error("Safety.BackupBeforeOverwrite should be true")
	}
	if cfg.Hooks.EnforcementMode != "warn" {
		t.Errorf("Hooks.EnforcementMode = %q, want %q", cfg.Hooks.EnforcementMode, "warn")
	}
	if cfg.Hooks.Verbose != true {
		t.Error("Hooks.Verbose should be true")
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path")
	if err != nil {
		t.Fatalf("Load should not error for missing file: %v", err)
	}
	// Should return defaults
	if cfg.Models.WebResearch != "haiku" {
		t.Error("Missing file should return defaults")
	}
}

func TestLoadLegacyPath(t *testing.T) {
	tmp := t.TempDir()
	legacyDir := filepath.Join(tmp, ".spw")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatal(err)
	}
	legacyConfig := filepath.Join(legacyDir, "spw-config.toml")
	if err := os.WriteFile(legacyConfig, []byte(`[models]
web_research = "custom-haiku"
`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Models.WebResearch != "custom-haiku" {
		t.Errorf("Models.WebResearch = %q, want %q (from legacy path)", cfg.Models.WebResearch, "custom-haiku")
	}
}

func TestMissingSectionsUseDefaults(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, ".spec-workflow")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Only provide models section
	if err := os.WriteFile(filepath.Join(configDir, "spw-config.toml"), []byte(`[models]
web_research = "custom"
`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Models.WebResearch != "custom" {
		t.Errorf("Models.WebResearch = %q, want %q", cfg.Models.WebResearch, "custom")
	}
	// Other sections should have defaults
	if cfg.Hooks.EnforcementMode != "warn" {
		t.Errorf("Missing hooks section should default enforcement_mode to warn, got %q", cfg.Hooks.EnforcementMode)
	}
	if cfg.Planning.MaxWaveSize != 3 {
		t.Errorf("Missing planning section should default max_wave_size to 3, got %d", cfg.Planning.MaxWaveSize)
	}
}

func TestGetValue(t *testing.T) {
	cfg := Defaults()

	tests := []struct {
		key      string
		defVal   string
		expected string
	}{
		{"models.web_research", "", "haiku"},
		{"models.complex_reasoning", "", "opus"},
		{"execution.tdd_default", "", "false"},
		{"execution.commit_per_task", "", "auto"},
		{"planning.max_wave_size", "", "3"},
		{"hooks.enforcement_mode", "", "warn"},
		{"hooks.recent_run_window_minutes", "", "30"},
		{"statusline.cache_ttl_seconds", "", "10"},
		{"agent_teams.enabled", "", "false"},
		{"nonexistent.key", "fallback", "fallback"},
	}

	for _, tt := range tests {
		got := cfg.GetValue(tt.key, tt.defVal)
		if got != tt.expected {
			t.Errorf("GetValue(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		input    string
		fallback bool
		expected bool
	}{
		{"true", false, true},
		{"True", false, true},
		{"TRUE", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"on", false, true},
		{"false", true, false},
		{"0", true, false},
		{"no", true, false},
		{"off", true, false},
		{"", true, true},
		{"", false, false},
		{"garbage", true, true},
	}

	for _, tt := range tests {
		got := ToBool(tt.input, tt.fallback)
		if got != tt.expected {
			t.Errorf("ToBool(%q, %v) = %v, want %v", tt.input, tt.fallback, got, tt.expected)
		}
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input    string
		fallback int
		expected int
	}{
		{"42", 0, 42},
		{"0", 1, 0},
		{"-5", 0, -5},
		{"", 10, 10},
		{"abc", 10, 10},
		{"3.14", 0, 0}, // strconv.Atoi doesn't parse floats
	}

	for _, tt := range tests {
		got := ToInt(tt.input, tt.fallback)
		if got != tt.expected {
			t.Errorf("ToInt(%q, %d) = %d, want %d", tt.input, tt.fallback, got, tt.expected)
		}
	}
}

func TestMerge(t *testing.T) {
	tmp := t.TempDir()
	templatePath := filepath.Join(tmp, "template.toml")
	userPath := filepath.Join(tmp, "user.toml")
	outputPath := filepath.Join(tmp, "output.toml")

	template := `[models]
web_research = "haiku"
complex_reasoning = "opus"
implementation = "sonnet"

[execution]
tdd_default = false
new_key = "new_value"
`
	userConfig := `[models]
web_research = "custom-haiku"

[execution]
tdd_default = true
`

	if err := os.WriteFile(templatePath, []byte(template), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userPath, []byte(userConfig), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Merge(templatePath, userPath, outputPath); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	result, err := LoadFromPath(outputPath)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	// User value preserved
	if result.Models.WebResearch != "custom-haiku" {
		t.Errorf("Merge should preserve user value: got %q, want %q", result.Models.WebResearch, "custom-haiku")
	}
	// Template value used when user doesn't have it
	if result.Models.ComplexReasoning != "opus" {
		t.Errorf("Merge should use template value: got %q, want %q", result.Models.ComplexReasoning, "opus")
	}
	// User value preserved
	if result.Execution.TDDDefault != true {
		t.Error("Merge should preserve user value for tdd_default")
	}
}

func TestMultiLineArrays(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, ".spec-workflow")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	config := `[statusline]
base_branches = ["main", "master", "staging", "develop"]

[skills.design]
enforce_required = true
required = [
  "skill-a",
  "skill-b",
  "skill-c"
]
optional = ["opt-1", "opt-2"]
`
	if err := os.WriteFile(filepath.Join(configDir, "spw-config.toml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Statusline.BaseBranches) != 4 {
		t.Errorf("BaseBranches = %v, want 4 items", cfg.Statusline.BaseBranches)
	}
	if len(cfg.Skills.Design.Required) != 3 {
		t.Errorf("Skills.Design.Required = %v, want 3 items", cfg.Skills.Design.Required)
	}
	if len(cfg.Skills.Design.Optional) != 2 {
		t.Errorf("Skills.Design.Optional = %v, want 2 items", cfg.Skills.Design.Optional)
	}
}

func TestMergePreservesUserMultilineArrays(t *testing.T) {
	tmp := t.TempDir()
	templatePath := filepath.Join(tmp, "template.toml")
	userPath := filepath.Join(tmp, "user.toml")
	outputPath := filepath.Join(tmp, "output.toml")

	template := `[skills.design]
enforce_required = true
required = []
`
	user := `[skills.design]
enforce_required = true
required = [
  "elixir-skill-a",
  "elixir-skill-b"
]
`

	os.WriteFile(templatePath, []byte(template), 0644)
	os.WriteFile(userPath, []byte(user), 0644)

	if err := Merge(templatePath, userPath, outputPath); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	cfg, err := LoadFromPath(outputPath)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if len(cfg.Skills.Design.Required) != 2 {
		t.Errorf("Skills.Design.Required = %v, want [elixir-skill-a, elixir-skill-b]", cfg.Skills.Design.Required)
	}
}

func TestMergeUserMultilineTemplateMultiline(t *testing.T) {
	tmp := t.TempDir()
	templatePath := filepath.Join(tmp, "template.toml")
	userPath := filepath.Join(tmp, "user.toml")
	outputPath := filepath.Join(tmp, "output.toml")

	template := `[skills.implementation]
enforce_required = true
required = [
  "conventional-commits"
]
`
	user := `[skills.implementation]
enforce_required = true
required = [
  "conventional-commits",
  "elixir-skill"
]
`

	os.WriteFile(templatePath, []byte(template), 0644)
	os.WriteFile(userPath, []byte(user), 0644)

	if err := Merge(templatePath, userPath, outputPath); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	cfg, err := LoadFromPath(outputPath)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if len(cfg.Skills.Implementation.Required) != 2 {
		t.Errorf("Skills.Implementation.Required = %v, want [conventional-commits, elixir-skill]", cfg.Skills.Implementation.Required)
	}
}

func TestMergeTemplateMultilineNoUserKey(t *testing.T) {
	tmp := t.TempDir()
	templatePath := filepath.Join(tmp, "template.toml")
	userPath := filepath.Join(tmp, "user.toml")
	outputPath := filepath.Join(tmp, "output.toml")

	template := `[skills.implementation]
enforce_required = true
required = [
  "conventional-commits"
]
`
	user := `[skills.implementation]
enforce_required = true
`

	os.WriteFile(templatePath, []byte(template), 0644)
	os.WriteFile(userPath, []byte(user), 0644)

	if err := Merge(templatePath, userPath, outputPath); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	cfg, err := LoadFromPath(outputPath)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if len(cfg.Skills.Implementation.Required) != 1 || cfg.Skills.Implementation.Required[0] != "conventional-commits" {
		t.Errorf("Skills.Implementation.Required = %v, want [conventional-commits]", cfg.Skills.Implementation.Required)
	}
}

func TestMergeUserSingleLineTemplateMultiline(t *testing.T) {
	tmp := t.TempDir()
	templatePath := filepath.Join(tmp, "template.toml")
	userPath := filepath.Join(tmp, "user.toml")
	outputPath := filepath.Join(tmp, "output.toml")

	template := `[skills.implementation]
enforce_required = true
required = [
  "conventional-commits"
]
`
	user := `[skills.implementation]
enforce_required = true
required = ["user-skill"]
`

	os.WriteFile(templatePath, []byte(template), 0644)
	os.WriteFile(userPath, []byte(user), 0644)

	if err := Merge(templatePath, userPath, outputPath); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	cfg, err := LoadFromPath(outputPath)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if len(cfg.Skills.Implementation.Required) != 1 || cfg.Skills.Implementation.Required[0] != "user-skill" {
		t.Errorf("Skills.Implementation.Required = %v, want [user-skill]", cfg.Skills.Implementation.Required)
	}
}

// findRepoFile searches for a file relative to the repo root
func findRepoFile(t *testing.T, relPath string) string {
	t.Helper()
	// Walk up from the test directory to find the repo root
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(dir, relPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		// Also check if we're inside cli/
		candidate = filepath.Join(dir, "..", "..", relPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
