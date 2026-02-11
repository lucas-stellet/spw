package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGuidelinesEmpty(t *testing.T) {
	dir := t.TempDir()
	gs := LoadGuidelines(dir)
	if gs != nil {
		t.Errorf("expected nil guidelines for missing dir, got %d", len(gs))
	}
}

func TestLoadGuidelinesBuiltinMapping(t *testing.T) {
	dir := t.TempDir()
	gDir := filepath.Join(dir, ".spec-workflow", "guidelines")
	if err := os.MkdirAll(gDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write named guidelines.
	for _, name := range []string{"project", "coding", "quality", "testing"} {
		if err := os.WriteFile(filepath.Join(gDir, name+".md"), []byte("# "+name+" guideline"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	gs := LoadGuidelines(dir)
	if len(gs) != 4 {
		t.Fatalf("expected 4 guidelines, got %d", len(gs))
	}

	byName := make(map[string]Guideline)
	for _, g := range gs {
		byName[g.Name] = g
	}

	// project → all phases (empty AppliesTo).
	if len(byName["project"].AppliesTo) != 0 {
		t.Errorf("project should apply to all phases, got %v", byName["project"].AppliesTo)
	}

	// coding → exec, checkpoint.
	if g := byName["coding"]; len(g.AppliesTo) != 2 || g.AppliesTo[0] != "exec" || g.AppliesTo[1] != "checkpoint" {
		t.Errorf("coding should apply to [exec checkpoint], got %v", g.AppliesTo)
	}

	// quality → checkpoint, qa-check, post-mortem.
	if g := byName["quality"]; len(g.AppliesTo) != 3 {
		t.Errorf("quality should apply to 3 phases, got %v", g.AppliesTo)
	}

	// testing → exec, qa, qa-check, qa-exec.
	if g := byName["testing"]; len(g.AppliesTo) != 4 {
		t.Errorf("testing should apply to 4 phases, got %v", g.AppliesTo)
	}
}

func TestLoadGuidelinesFrontmatter(t *testing.T) {
	dir := t.TempDir()
	gDir := filepath.Join(dir, ".spec-workflow", "guidelines")
	if err := os.MkdirAll(gDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "---\napplies_to: [exec, qa]\n---\n# Custom guideline\nSome content."
	if err := os.WriteFile(filepath.Join(gDir, "custom.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	gs := LoadGuidelines(dir)
	if len(gs) != 1 {
		t.Fatalf("expected 1 guideline, got %d", len(gs))
	}

	g := gs[0]
	if g.Name != "custom" {
		t.Errorf("expected name 'custom', got %q", g.Name)
	}
	if len(g.AppliesTo) != 2 || g.AppliesTo[0] != "exec" || g.AppliesTo[1] != "qa" {
		t.Errorf("expected applies_to [exec qa], got %v", g.AppliesTo)
	}
	if g.Content != "# Custom guideline\nSome content." {
		t.Errorf("content should not include frontmatter, got %q", g.Content)
	}
}

func TestLoadGuidelinesNoFrontmatterCustom(t *testing.T) {
	dir := t.TempDir()
	gDir := filepath.Join(dir, ".spec-workflow", "guidelines")
	if err := os.MkdirAll(gDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Unknown name without frontmatter → all phases.
	if err := os.WriteFile(filepath.Join(gDir, "misc.md"), []byte("# Misc guideline"), 0o644); err != nil {
		t.Fatal(err)
	}

	gs := LoadGuidelines(dir)
	if len(gs) != 1 {
		t.Fatalf("expected 1 guideline, got %d", len(gs))
	}

	if len(gs[0].AppliesTo) != 0 {
		t.Errorf("unknown guideline without frontmatter should apply to all phases, got %v", gs[0].AppliesTo)
	}
}

func TestGuidelinesForPhase(t *testing.T) {
	guidelines := []Guideline{
		{Name: "project", Content: "project rules", AppliesTo: nil},
		{Name: "coding", Content: "coding rules", AppliesTo: []string{"exec", "checkpoint"}},
		{Name: "testing", Content: "testing rules", AppliesTo: []string{"exec", "qa"}},
	}

	// exec → project + coding + testing.
	result := GuidelinesForPhase(guidelines, "exec")
	if len(result) != 3 {
		t.Errorf("exec should match 3 guidelines, got %d", len(result))
	}

	// checkpoint → project + coding.
	result = GuidelinesForPhase(guidelines, "checkpoint")
	if len(result) != 2 {
		t.Errorf("checkpoint should match 2 guidelines, got %d", len(result))
	}

	// prd → project only.
	result = GuidelinesForPhase(guidelines, "prd")
	if len(result) != 1 || result[0].Name != "project" {
		t.Errorf("prd should match only project, got %v", result)
	}
}
