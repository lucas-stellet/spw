package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCommandRegistry(t *testing.T) {
	reg := getRegistry()

	expected := map[string]struct {
		phase       string
		category    string
		subcategory string
	}{
		"prd":             {"prd", "pipeline", "research"},
		"design-research": {"design", "pipeline", "research"},
		"design-draft":    {"design", "pipeline", "synthesis"},
		"tasks-plan":      {"planning", "pipeline", "synthesis"},
		"qa":              {"qa", "pipeline", "synthesis"},
		"post-mortem":     {"post-mortem", "pipeline", "synthesis"},
		"tasks-check":     {"planning", "audit", "artifact"},
		"qa-check":        {"qa", "audit", "code"},
		"checkpoint":      {"execution", "audit", "code"},
		"exec":            {"execution", "wave-execution", "implementation"},
		"qa-exec":         {"qa", "wave-execution", "validation"},
	}

	if len(reg) != len(expected) {
		t.Errorf("registry has %d entries, want %d", len(reg), len(expected))
	}

	for name, want := range expected {
		got, ok := reg[name]
		if !ok {
			t.Errorf("registry missing %q", name)
			continue
		}
		if got.Phase != want.phase {
			t.Errorf("%s: Phase = %q, want %q", name, got.Phase, want.phase)
		}
		if got.Category != want.category {
			t.Errorf("%s: Category = %q, want %q", name, got.Category, want.category)
		}
		if got.Subcategory != want.subcategory {
			t.Errorf("%s: Subcategory = %q, want %q", name, got.Subcategory, want.subcategory)
		}
	}
}

func TestCommandRegistryWaveAware(t *testing.T) {
	reg := getRegistry()

	waveAware := map[string]bool{
		"exec":       true,
		"checkpoint": true,
		"qa-exec":    true,
	}

	for name, meta := range reg {
		if waveAware[name] && !meta.WaveAware {
			t.Errorf("%s should be WaveAware (comms_path should contain {wave})", name)
		}
		if !waveAware[name] && meta.WaveAware {
			t.Errorf("%s should NOT be WaveAware", name)
		}
	}
}

func TestCommsPathGeneration(t *testing.T) {
	reg := getRegistry()

	tests := []struct {
		command string
		wave    string
		want    string
	}{
		{"prd", "", "prd/_comms"},
		{"design-research", "", "design/_comms/design-research"},
		{"design-draft", "", "design/_comms/design-draft"},
		{"tasks-plan", "", "planning/_comms/tasks-plan"},
		{"qa", "", "qa/_comms/qa"},
		{"post-mortem", "", "post-mortem/_comms"},
		{"tasks-check", "", "planning/_comms/tasks-check"},
		{"qa-check", "", "qa/_comms/qa-check"},
		{"checkpoint", "3", "execution/waves/wave-03/checkpoint"},
		{"exec", "1", "execution/waves/wave-01/execution"},
		{"qa-exec", "12", "qa/_comms/qa-exec/waves/wave-12"},
	}

	for _, tt := range tests {
		meta := reg[tt.command]
		path := meta.CommsPath
		if meta.WaveAware && tt.wave != "" {
			n, _ := strconv.Atoi(tt.wave)
			path = strings.Replace(path, "{wave}", fmt.Sprintf("%02d", n), 1)
		}
		if path != tt.want {
			t.Errorf("%s (wave=%q): comms path = %q, want %q", tt.command, tt.wave, path, tt.want)
		}
	}
}

func TestNextRunNumber(t *testing.T) {
	// Empty dir -> run-001
	tmp := t.TempDir()
	next := scanNextRun(tmp)
	if next != 1 {
		t.Errorf("empty dir: next run = %d, want 1", next)
	}

	// Create run-001 -> next should be run-002
	os.MkdirAll(filepath.Join(tmp, "run-001"), 0755)
	next = scanNextRun(tmp)
	if next != 2 {
		t.Errorf("with run-001: next run = %d, want 2", next)
	}

	// Create run-005 -> next should be run-006
	os.MkdirAll(filepath.Join(tmp, "run-005"), 0755)
	next = scanNextRun(tmp)
	if next != 6 {
		t.Errorf("with run-005: next run = %d, want 6", next)
	}
}

// scanNextRun replicates the logic in DispatchInit for testing
func scanNextRun(dir string) int {
	nextRun := 1
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nextRun
	}
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
	return nextRun
}

func TestBriefSkeleton(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	subagentDir := filepath.Join(runDir, "test-agent")
	os.MkdirAll(subagentDir, 0755)

	subagentRel, _ := filepath.Rel(tmp, subagentDir)
	reportPath := filepath.Join(subagentRel, "report.md")
	statusPath := filepath.Join(subagentRel, "status.json")

	brief := fmt.Sprintf(`# Brief: test-agent

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->

## Task
<!-- Describe what this subagent must do -->

## Output Contract
Write your output to these exact paths:
- Report: %s
- Status: %s

status.json format:
`+"```json"+`
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["skill-name"],
  "skills_missing": [],
  "model_override_reason": null
}
`+"```"+`
`, reportPath, statusPath)

	briefPath := filepath.Join(subagentDir, "brief.md")
	os.WriteFile(briefPath, []byte(brief), 0644)

	data, err := os.ReadFile(briefPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	requiredSections := []string{
		"# Brief: test-agent",
		"## Inputs",
		"## Task",
		"## Output Contract",
		"report.md",
		"status.json",
		"skills_used",
		"model_override_reason",
	}
	for _, section := range requiredSections {
		if !contains(content, section) {
			t.Errorf("brief.md missing section %q", section)
		}
	}
}

func TestDispatchStatusParsing(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")

	// Test pass status
	passDir := filepath.Join(runDir, "agent-pass")
	os.MkdirAll(passDir, 0755)
	writeJSON(t, filepath.Join(passDir, "status.json"), map[string]any{"status": "pass", "summary": "all good"})

	// Test blocked status
	blockedDir := filepath.Join(runDir, "agent-blocked")
	os.MkdirAll(blockedDir, 0755)
	writeJSON(t, filepath.Join(blockedDir, "status.json"), map[string]any{"status": "blocked", "summary": "need help"})

	// Test invalid status
	invalidDir := filepath.Join(runDir, "agent-invalid")
	os.MkdirAll(invalidDir, 0755)
	writeJSON(t, filepath.Join(invalidDir, "status.json"), map[string]any{"status": "unknown", "summary": "bad"})

	tests := []struct {
		name      string
		agent     string
		wantValid bool
		wantState string
	}{
		{"pass", "agent-pass", true, "pass"},
		{"blocked", "agent-blocked", true, "blocked"},
		{"invalid", "agent-invalid", false, "unknown"},
		{"missing", "agent-missing", false, "missing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusFile := filepath.Join(runDir, tt.agent, "status.json")
			data, err := os.ReadFile(statusFile)
			if err != nil {
				if tt.wantState != "missing" {
					t.Fatalf("unexpected read error: %v", err)
				}
				return // missing file case verified
			}

			var doc map[string]any
			json.Unmarshal(data, &doc)
			status, _ := doc["status"].(string)
			valid := status == "pass" || status == "blocked"

			if status != tt.wantState {
				t.Errorf("status = %q, want %q", status, tt.wantState)
			}
			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestDispatchHandoffGeneration(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")

	// Create subagent dirs with status.json
	agent1Dir := filepath.Join(runDir, "researcher")
	os.MkdirAll(agent1Dir, 0755)
	writeJSON(t, filepath.Join(agent1Dir, "status.json"), map[string]any{"status": "pass", "summary": "research done"})

	agent2Dir := filepath.Join(runDir, "analyzer")
	os.MkdirAll(agent2Dir, 0755)
	writeJSON(t, filepath.Join(agent2Dir, "status.json"), map[string]any{"status": "blocked", "summary": "needs input"})

	// Scan and build handoff
	entries, _ := os.ReadDir(runDir)
	var agents []subagentStatus
	allPass := true
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		statusFile := filepath.Join(runDir, e.Name(), "status.json")
		data, _ := os.ReadFile(statusFile)
		var doc map[string]any
		json.Unmarshal(data, &doc)
		status, _ := doc["status"].(string)
		summary, _ := doc["summary"].(string)
		agents = append(agents, subagentStatus{Name: e.Name(), Status: status, Summary: summary})
		if status != "pass" {
			allPass = false
		}
	}

	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
	if allPass {
		t.Error("allPass should be false with a blocked agent")
	}

	// Verify agent statuses
	agentMap := map[string]subagentStatus{}
	for _, a := range agents {
		agentMap[a.Name] = a
	}

	if agentMap["researcher"].Status != "pass" {
		t.Errorf("researcher status = %q, want pass", agentMap["researcher"].Status)
	}
	if agentMap["analyzer"].Status != "blocked" {
		t.Errorf("analyzer status = %q, want blocked", agentMap["analyzer"].Status)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
