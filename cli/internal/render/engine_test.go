package render

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lucas-stellet/oraculo/internal/config"
	"github.com/lucas-stellet/oraculo/internal/embedded"
)

func goldenDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "golden", "workflows")
}

func defaultEngine(t *testing.T) *Engine {
	t.Helper()
	cfg := config.Defaults()
	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return e
}

func TestRenderAllCommands(t *testing.T) {
	e := defaultEngine(t)

	for _, cmd := range AllCommands {
		t.Run(cmd, func(t *testing.T) {
			got, err := e.RenderCommand(cmd)
			if err != nil {
				t.Fatalf("RenderCommand(%q) error: %v", cmd, err)
			}

			goldenPath := filepath.Join(goldenDir(), cmd+".md")
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("reading golden file %s: %v", goldenPath, err)
			}

			if got != string(want) {
				t.Errorf("RenderCommand(%q) does not match golden file.\nFirst diff at line: %d",
					cmd, firstDiffLine(got, string(want)))
			}
		})
	}
}

func TestNoAtReferences(t *testing.T) {
	e := defaultEngine(t)

	for _, cmd := range AllCommands {
		t.Run(cmd, func(t *testing.T) {
			got, err := e.RenderCommand(cmd)
			if err != nil {
				t.Fatalf("RenderCommand(%q) error: %v", cmd, err)
			}

			for i, line := range strings.Split(got, "\n") {
				if strings.Contains(line, "@.claude/workflows/oraculo/") {
					t.Errorf("line %d still has @-reference: %s", i+1, line)
				}
			}
		})
	}
}

func TestNoJSToolReferences(t *testing.T) {
	e := defaultEngine(t)

	for _, cmd := range AllCommands {
		t.Run(cmd, func(t *testing.T) {
			got, err := e.RenderCommand(cmd)
			if err != nil {
				t.Fatalf("RenderCommand(%q) error: %v", cmd, err)
			}

			if strings.Contains(got, "oraculo-tools.js") {
				t.Errorf("rendered output still references oraculo-tools.js")
			}
		})
	}
}

func TestTeamsDisabledNoOverlay(t *testing.T) {
	cfg := config.Defaults()
	cfg.AgentTeams.Enabled = false

	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	got, err := e.RenderCommand("exec")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if strings.Contains(got, "Agent Teams overlay") {
		t.Error("overlay content should not be present when teams disabled")
	}
}

func TestTeamsEnabledOverlayPresent(t *testing.T) {
	cfg := config.Defaults()
	cfg.AgentTeams.Enabled = true

	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	got, err := e.RenderCommand("exec")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if !strings.Contains(got, "Agent Teams overlay for `oraculo:exec`") {
		t.Error("overlay content should be present when teams enabled")
	}
}

func TestTeamsEnabledExcludedPhase(t *testing.T) {
	cfg := config.Defaults()
	cfg.AgentTeams.Enabled = true
	cfg.AgentTeams.ExcludePhases = []string{"exec"}

	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	got, err := e.RenderCommand("exec")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if strings.Contains(got, "Agent Teams overlay") {
		t.Error("overlay content should not be present when phase is excluded")
	}
}

func TestPlanOverlayAppendedWhenTeamsEnabled(t *testing.T) {
	cfg := config.Defaults()
	cfg.AgentTeams.Enabled = true

	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// plan.md has no @-ref for overlay, but should get it appended.
	got, err := e.RenderCommand("plan")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if !strings.Contains(got, "Agent Teams overlay for `oraculo:plan`") {
		t.Error("plan overlay should be appended when teams enabled")
	}
}

func TestRenderAll(t *testing.T) {
	e := defaultEngine(t)

	results, err := e.RenderAll()
	if err != nil {
		t.Fatalf("RenderAll() error: %v", err)
	}

	if len(results) != len(embedded.AllWorkflowNames) {
		t.Errorf("RenderAll() returned %d results, want %d", len(results), len(embedded.AllWorkflowNames))
	}

	for _, cmd := range embedded.AllWorkflowNames {
		if _, ok := results[cmd]; !ok {
			t.Errorf("RenderAll() missing command %q", cmd)
		}
	}
}

func TestSharedPoliciesInlined(t *testing.T) {
	e := defaultEngine(t)

	got, err := e.RenderCommand("exec")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	// All 5 shared policies should be inlined.
	policies := []string{
		"# Config Resolution",
		"File-First Handoff Contract",
		"Resume Policy",
		"Skills Policy",
		"Approval Reconciliation",
	}
	for _, p := range policies {
		if !strings.Contains(got, p) {
			t.Errorf("shared policy %q not found in rendered output", p)
		}
	}
}

func TestDispatchPatternInlined(t *testing.T) {
	e := defaultEngine(t)

	// exec uses wave dispatch.
	got, err := e.RenderCommand("exec")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if !strings.Contains(got, "Wave Execution Dispatch Pattern") {
		t.Error("wave dispatch pattern should be inlined in exec workflow")
	}

	// checkpoint uses audit dispatch.
	got, err = e.RenderCommand("checkpoint")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if !strings.Contains(got, "Audit Dispatch Pattern") {
		t.Error("audit dispatch pattern should be inlined in checkpoint workflow")
	}
}

func TestNoGuidelinesNoInjection(t *testing.T) {
	e := defaultEngine(t)

	got, err := e.RenderCommand("exec")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if strings.Contains(got, "<user_guidelines>") {
		t.Error("no guidelines set, but <user_guidelines> block found in output")
	}
}

func TestGuidelinesInjectedForMatchingPhase(t *testing.T) {
	cfg := config.Defaults()
	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	e.SetGuidelines([]struct {
		Name      string
		Content   string
		AppliesTo []string
	}{
		{Name: "project", Content: "Always use conventional commits.", AppliesTo: nil},
		{Name: "coding", Content: "Follow SOLID principles.", AppliesTo: []string{"exec", "checkpoint"}},
	})

	// exec should have both guidelines.
	got, err := e.RenderCommand("exec")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if !strings.Contains(got, "<user_guidelines>") {
		t.Fatal("expected <user_guidelines> block in exec")
	}
	if !strings.Contains(got, "Always use conventional commits.") {
		t.Error("project guideline should be in exec output")
	}
	if !strings.Contains(got, "Follow SOLID principles.") {
		t.Error("coding guideline should be in exec output")
	}

	// prd should have only the project guideline.
	got, err = e.RenderCommand("prd")
	if err != nil {
		t.Fatalf("RenderCommand error: %v", err)
	}

	if !strings.Contains(got, "Always use conventional commits.") {
		t.Error("project guideline should be in prd output")
	}
	if strings.Contains(got, "Follow SOLID principles.") {
		t.Error("coding guideline should NOT be in prd output")
	}
}

func firstDiffLine(a, b string) int {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	maxLen := len(linesA)
	if len(linesB) > maxLen {
		maxLen = len(linesB)
	}

	for i := 0; i < maxLen; i++ {
		var la, lb string
		if i < len(linesA) {
			la = linesA[i]
		}
		if i < len(linesB) {
			lb = linesB[i]
		}
		if la != lb {
			return i + 1
		}
	}
	return 0
}
