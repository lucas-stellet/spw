package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lucas-stellet/oraculo/internal/workspace"
)

func TestFirstOraculoCommand(t *testing.T) {
	tests := []struct {
		input   string
		wantCmd string
		wantArg string
	}{
		{"/oraculo:exec my-spec", "exec", "my-spec"},
		{"/oraculo:plan my-spec --mode rolling-wave", "plan", "my-spec --mode rolling-wave"},
		{"/oraculo:status", "status", ""},
		{"/oraculo:prd my-spec --source url", "prd", "my-spec --source url"},
		{"hello world", "", ""},
		{"", "", ""},
		{"some text\n/oraculo:exec test-spec\nmore text", "exec", "test-spec"},
	}

	for _, tt := range tests {
		parsed := firstOraculoCommand(tt.input)
		if tt.wantCmd == "" {
			if parsed != nil {
				t.Errorf("firstOraculoCommand(%q) = %+v, want nil", tt.input, parsed)
			}
			continue
		}
		if parsed == nil {
			t.Errorf("firstOraculoCommand(%q) = nil, want command=%q", tt.input, tt.wantCmd)
			continue
		}
		if parsed.command != tt.wantCmd {
			t.Errorf("firstOraculoCommand(%q).command = %q, want %q", tt.input, parsed.command, tt.wantCmd)
		}
		if parsed.argsLine != tt.wantArg {
			t.Errorf("firstOraculoCommand(%q).argsLine = %q, want %q", tt.input, parsed.argsLine, tt.wantArg)
		}
	}
}

func TestExtractSpecArg(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-spec", "my-spec"},
		{"my-spec --batch-size 3", "my-spec"},
		{"--batch-size 3", ""},
		{"", ""},
		{`"quoted-spec" --flag`, "quoted-spec"},
	}

	for _, tt := range tests {
		got := extractSpecArg(tt.input)
		if got != tt.want {
			t.Errorf("extractSpecArg(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHasSpecArg(t *testing.T) {
	if hasSpecArg("") {
		t.Error("hasSpecArg('') should be false")
	}
	if hasSpecArg("--flag value") {
		t.Error("hasSpecArg('--flag value') should be false")
	}
	if !hasSpecArg("my-spec") {
		t.Error("hasSpecArg('my-spec') should be true")
	}
}

func TestTokenizeArgs(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"spec", []string{"spec"}},
		{"spec --flag val", []string{"spec", "--flag", "val"}},
		{`"quoted spec" --flag`, []string{"quoted spec", "--flag"}},
		{`'single quoted'`, []string{"single quoted"}},
	}

	for _, tt := range tests {
		got := tokenizeArgs(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("tokenizeArgs(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("tokenizeArgs(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestIsManagedArtifactFile(t *testing.T) {
	managed := []string{
		"DESIGN-RESEARCH.md",
		"TASKS-CHECK.md",
		"CHECKPOINT-REPORT.md",
		"STATUS-SUMMARY.md",
		"SKILLS-EXEC.md",
		"SKILLS-DESIGN.md",
		"PRD.md",
		"PRD-SOURCE-NOTES.md",
		"PRD-STRUCTURE.md",
		"PRD-REVISION-PLAN.md",
	}
	for _, f := range managed {
		if !isManagedArtifactFile(f) {
			t.Errorf("isManagedArtifactFile(%q) = false, want true", f)
		}
	}

	notManaged := []string{
		"README.md",
		"tasks.md",
		"design.md",
		"random.md",
		"config.toml",
	}
	for _, f := range notManaged {
		if isManagedArtifactFile(f) {
			t.Errorf("isManagedArtifactFile(%q) = true, want false", f)
		}
	}
}

func TestNormalizeSlashes(t *testing.T) {
	if got := normalizeSlashes(`foo\bar\baz`); got != "foo/bar/baz" {
		t.Errorf("normalizeSlashes = %q, want %q", got, "foo/bar/baz")
	}
}

func TestCheckRunCompleteness(t *testing.T) {
	tmp := t.TempDir()

	// Complete run
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(filepath.Join(runDir, "researcher"), 0755)
	os.WriteFile(filepath.Join(runDir, "_handoff.md"), []byte("done"), 0644)
	os.WriteFile(filepath.Join(runDir, "researcher", "brief.md"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(runDir, "researcher", "report.md"), []byte("r"), 0644)
	os.WriteFile(filepath.Join(runDir, "researcher", "status.json"), []byte("{}"), 0644)

	issues := checkRunCompleteness(runDir)
	if len(issues) != 0 {
		t.Errorf("Complete run should have no issues, got: %v", issues)
	}

	// Incomplete run
	runDir2 := filepath.Join(tmp, "run-002")
	os.MkdirAll(filepath.Join(runDir2, "writer"), 0755)
	os.WriteFile(filepath.Join(runDir2, "writer", "brief.md"), []byte("b"), 0644)

	issues2 := checkRunCompleteness(runDir2)
	if len(issues2) == 0 {
		t.Error("Incomplete run should have issues")
	}
}

func TestIsRecent(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "test")
	os.MkdirAll(dir, 0755)

	now := time.Now()
	windowMs := int64(30) * 60 * 1000

	if !isRecent(dir, now, windowMs) {
		t.Error("Just-created dir should be recent")
	}
}

func TestStatuslineCache(t *testing.T) {
	tmp := t.TempDir()

	// Write cache
	ok := writeStatuslineCache(tmp, "my-spec", map[string]string{"source": "test"})
	if !ok {
		t.Fatal("writeStatuslineCache should succeed")
	}

	// Read with TTL (should succeed since just written)
	spec := readStatuslineCache(tmp, 60, false)
	if spec != "my-spec" {
		t.Errorf("readStatuslineCache = %q, want %q", spec, "my-spec")
	}

	// Read ignoring TTL
	spec2 := readStatuslineCache(tmp, 0, true)
	if spec2 != "my-spec" {
		t.Errorf("readStatuslineCache(ignoreTTL) = %q, want %q", spec2, "my-spec")
	}

	// Clear cache
	clearStatuslineCache(tmp)
	spec3 := readStatuslineCache(tmp, 60, true)
	if spec3 != "" {
		t.Errorf("After clear, readStatuslineCache = %q, want empty", spec3)
	}
}

func TestCollectRunDirs(t *testing.T) {
	tmp := t.TempDir()

	// Create spec structure
	paths := []string{
		"prd/_comms/run-001",
		"design/_comms/design-research/run-001",
		"planning/_comms/tasks-plan/run-001",
		"execution/waves/wave-01/execution/run-001",
		"execution/waves/wave-01/checkpoint/run-001",
		"qa/_comms/qa/run-001",
		"qa/_comms/qa-exec/waves/wave-01/run-001",
		"post-mortem/_comms/run-001",
	}
	for _, p := range paths {
		os.MkdirAll(filepath.Join(tmp, p), 0755)
	}

	runs := collectRunDirs(tmp)
	if len(runs) != len(paths) {
		t.Errorf("collectRunDirs found %d runs, want %d", len(runs), len(paths))
		for _, r := range runs {
			t.Logf("  found: %s", r)
		}
	}
}

func TestDetectSpecByMtime(t *testing.T) {
	tmp := t.TempDir()

	// Create two specs
	specA := filepath.Join(tmp, "spec-a")
	specB := filepath.Join(tmp, "spec-b")
	os.MkdirAll(specA, 0755)
	os.MkdirAll(specB, 0755)

	os.WriteFile(filepath.Join(specA, "requirements.md"), []byte("a"), 0644)
	// Make spec-b newer
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(specB, "tasks.md"), []byte("b"), 0644)

	got := detectSpecByMtime(tmp)
	if got != "spec-b" {
		t.Errorf("detectSpecByMtime = %q, want %q", got, "spec-b")
	}
}

func TestSpecExists(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "my-spec"), 0755)

	if !specExists(tmp, "my-spec") {
		t.Error("specExists should find my-spec")
	}
	if specExists(tmp, "nonexistent") {
		t.Error("specExists should not find nonexistent")
	}
	if specExists(tmp, "") {
		t.Error("specExists should return false for empty name")
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{847, "847"},
		{1000, "1k"},
		{1500, "1.5k"},
		{25300, "25.3k"},
		{25000, "25k"},
		{1000000, "1M"},
		{1200000, "1.2M"},
	}
	for _, tt := range tests {
		got := formatTokens(tt.input)
		if got != tt.want {
			t.Errorf("formatTokens(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatTokenCost(t *testing.T) {
	ptrF64 := func(v float64) *float64 { return &v }
	ptrI64 := func(v int64) *int64 { return &v }

	tests := []struct {
		name string
		p    workspace.Payload
		mode string
		want string
	}{
		{
			name: "never mode",
			p: workspace.Payload{
				Cost: &struct {
					TotalCostUSD *float64 `json:"total_cost_usd"`
				}{TotalCostUSD: ptrF64(1.23)},
			},
			mode: "never",
			want: "",
		},
		{
			name: "auto with zero cost",
			p: workspace.Payload{
				ContextWindow: &struct {
					RemainingPercentage *float64 `json:"remaining_percentage"`
					TotalInputTokens    *int64   `json:"total_input_tokens"`
					TotalOutputTokens   *int64   `json:"total_output_tokens"`
				}{TotalInputTokens: ptrI64(1000), TotalOutputTokens: ptrI64(500)},
				Cost: &struct {
					TotalCostUSD *float64 `json:"total_cost_usd"`
				}{TotalCostUSD: ptrF64(0)},
			},
			mode: "auto",
			want: "",
		},
		{
			name: "auto with positive cost",
			p: workspace.Payload{
				ContextWindow: &struct {
					RemainingPercentage *float64 `json:"remaining_percentage"`
					TotalInputTokens    *int64   `json:"total_input_tokens"`
					TotalOutputTokens   *int64   `json:"total_output_tokens"`
				}{TotalInputTokens: ptrI64(20000), TotalOutputTokens: ptrI64(5300)},
				Cost: &struct {
					TotalCostUSD *float64 `json:"total_cost_usd"`
				}{TotalCostUSD: ptrF64(0.42)},
			},
			mode: "auto",
			want: " │ \x1b[2m25.3k $0.42\x1b[0m",
		},
		{
			name: "auto with nil cost",
			p: workspace.Payload{
				ContextWindow: &struct {
					RemainingPercentage *float64 `json:"remaining_percentage"`
					TotalInputTokens    *int64   `json:"total_input_tokens"`
					TotalOutputTokens   *int64   `json:"total_output_tokens"`
				}{TotalInputTokens: ptrI64(5000)},
			},
			mode: "auto",
			want: "",
		},
		{
			name: "always with nil data",
			p:    workspace.Payload{},
			mode: "always",
			want: "",
		},
		{
			name: "always with tokens only",
			p: workspace.Payload{
				ContextWindow: &struct {
					RemainingPercentage *float64 `json:"remaining_percentage"`
					TotalInputTokens    *int64   `json:"total_input_tokens"`
					TotalOutputTokens   *int64   `json:"total_output_tokens"`
				}{TotalInputTokens: ptrI64(5000), TotalOutputTokens: ptrI64(1000)},
			},
			mode: "always",
			want: " │ \x1b[2m6k\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTokenCost(tt.p, tt.mode)
			if got != tt.want {
				t.Errorf("formatTokenCost() = %q, want %q", got, tt.want)
			}
		})
	}
}
