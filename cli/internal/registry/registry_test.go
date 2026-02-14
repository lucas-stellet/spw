package registry

import (
	"testing"

	"github.com/lucas-stellet/oraculo/internal/embedded"
)

func TestParseDispatchPattern(t *testing.T) {
	content := `---
name: oraculo:test
description: Test command
---

<dispatch_pattern>
category: pipeline
subcategory: research
phase: discover
comms_path: discover/_comms
policy: @.claude/workflows/oraculo/shared/dispatch-pipeline.md
</dispatch_pattern>

Some other content here.
`

	meta, ok := parseDispatchPattern(content)
	if !ok {
		t.Fatal("parseDispatchPattern returned false")
	}

	if meta.Category != "pipeline" {
		t.Errorf("Category = %q, want %q", meta.Category, "pipeline")
	}
	if meta.Subcategory != "research" {
		t.Errorf("Subcategory = %q, want %q", meta.Subcategory, "research")
	}
	if meta.Phase != "discover" {
		t.Errorf("Phase = %q, want %q", meta.Phase, "discover")
	}
	if meta.CommsPath != "discover/_comms" {
		t.Errorf("CommsPath = %q, want %q", meta.CommsPath, "discover/_comms")
	}
	if meta.Policy != "@.claude/workflows/oraculo/shared/dispatch-pipeline.md" {
		t.Errorf("Policy = %q, want dispatch-pipeline ref", meta.Policy)
	}
	if meta.WaveAware {
		t.Error("WaveAware should be false for non-wave command")
	}
	if len(meta.Artifacts) != 0 {
		t.Errorf("Artifacts = %v, want empty", meta.Artifacts)
	}
}

func TestParseDispatchPatternWaveAware(t *testing.T) {
	content := `<dispatch_pattern>
category: wave-execution
subcategory: implementation
phase: execution
comms_path: execution/waves/wave-{wave}/execution
artifacts: execution/_implementation-logs
policy: @.claude/workflows/oraculo/shared/dispatch-wave.md
</dispatch_pattern>`

	meta, ok := parseDispatchPattern(content)
	if !ok {
		t.Fatal("parseDispatchPattern returned false")
	}

	if !meta.WaveAware {
		t.Error("WaveAware should be true for comms_path containing {wave}")
	}
	if meta.Category != "wave-execution" {
		t.Errorf("Category = %q, want %q", meta.Category, "wave-execution")
	}
	if meta.CommsPath != "execution/waves/wave-{wave}/execution" {
		t.Errorf("CommsPath = %q", meta.CommsPath)
	}
	if len(meta.Artifacts) != 1 || meta.Artifacts[0] != "execution/_implementation-logs" {
		t.Errorf("Artifacts = %v, want [execution/_implementation-logs]", meta.Artifacts)
	}
}

func TestParseDispatchPatternMultipleArtifacts(t *testing.T) {
	content := `<dispatch_pattern>
category: audit
subcategory: code
phase: execution
comms_path: execution/waves/wave-{wave}/checkpoint
artifacts: execution/_implementation-logs, execution/_review-notes
policy: @.claude/workflows/oraculo/shared/dispatch-audit.md
</dispatch_pattern>`

	meta, ok := parseDispatchPattern(content)
	if !ok {
		t.Fatal("parseDispatchPattern returned false")
	}

	if len(meta.Artifacts) != 2 {
		t.Fatalf("Artifacts count = %d, want 2", len(meta.Artifacts))
	}
	if meta.Artifacts[0] != "execution/_implementation-logs" {
		t.Errorf("Artifacts[0] = %q", meta.Artifacts[0])
	}
	if meta.Artifacts[1] != "execution/_review-notes" {
		t.Errorf("Artifacts[1] = %q", meta.Artifacts[1])
	}
}

func TestParseDispatchPatternNoSection(t *testing.T) {
	content := `---
name: oraculo:plan
---

Some content without dispatch_pattern.
`

	_, ok := parseDispatchPattern(content)
	if ok {
		t.Error("parseDispatchPattern should return false for content without dispatch_pattern")
	}
}

func TestParseDispatchPatternMissingCommsPath(t *testing.T) {
	content := `<dispatch_pattern>
category: pipeline
subcategory: research
phase: discover
policy: @.claude/workflows/oraculo/shared/dispatch-pipeline.md
</dispatch_pattern>`

	_, ok := parseDispatchPattern(content)
	if ok {
		t.Error("parseDispatchPattern should return false when comms_path is missing")
	}
}

func TestDispatchPolicy(t *testing.T) {
	tests := []struct {
		policy string
		want   string
	}{
		{"@.claude/workflows/oraculo/shared/dispatch-pipeline.md", "dispatch-pipeline"},
		{"@.claude/workflows/oraculo/shared/dispatch-audit.md", "dispatch-audit"},
		{"@.claude/workflows/oraculo/shared/dispatch-wave.md", "dispatch-wave"},
	}

	for _, tt := range tests {
		meta := CommandMeta{Policy: tt.policy}
		got := meta.DispatchPolicy()
		if got != tt.want {
			t.Errorf("DispatchPolicy(%q) = %q, want %q", tt.policy, got, tt.want)
		}
	}
}

func TestCategoryHelper(t *testing.T) {
	reg := map[string]CommandMeta{
		"discover": {Category: "pipeline"},
	}

	if got := Category(reg, "discover"); got != "pipeline" {
		t.Errorf("Category(discover) = %q, want %q", got, "pipeline")
	}
	if got := Category(reg, "unknown"); got != "" {
		t.Errorf("Category(unknown) = %q, want empty", got)
	}
}

// TestRegistryFromEmbedded loads the registry from real embedded workflow files
// and validates all 11 dispatch-capable commands are present with correct metadata.
func TestRegistryFromEmbedded(t *testing.T) {
	reg, err := Load(embedded.Workflows)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	expected := map[string]struct {
		phase       string
		category    string
		subcategory string
		commsPath   string
		waveAware   bool
		artifacts   []string
	}{
		"discover":        {"discover", "pipeline", "research", "discover/_comms", false, nil},
		"design-research": {"design", "pipeline", "research", "design/_comms/design-research", false, nil},
		"design-draft":    {"design", "pipeline", "synthesis", "design/_comms/design-draft", false, nil},
		"tasks-plan":      {"planning", "pipeline", "synthesis", "planning/_comms/tasks-plan", false, nil},
		"qa":              {"qa", "pipeline", "synthesis", "qa/_comms/qa", false, nil},
		"post-mortem":     {"post-mortem", "pipeline", "synthesis", "post-mortem/_comms", false, nil},
		"tasks-check":     {"planning", "audit", "artifact", "planning/_comms/tasks-check", false, nil},
		"qa-check":        {"qa", "audit", "code", "qa/_comms/qa-check", false, nil},
		"checkpoint":      {"execution", "audit", "code", "execution/waves/wave-{wave}/checkpoint", true, nil},
		"exec":            {"execution", "wave-execution", "implementation", "execution/waves/wave-{wave}/execution", true, []string{"execution/_implementation-logs"}},
		"qa-exec":         {"qa", "wave-execution", "validation", "qa/_comms/qa-exec/waves/wave-{wave}", true, nil},
	}

	if len(reg) != len(expected) {
		t.Errorf("registry has %d entries, want %d", len(reg), len(expected))
		for name := range reg {
			if _, ok := expected[name]; !ok {
				t.Errorf("unexpected command in registry: %q", name)
			}
		}
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
		if got.CommsPath != want.commsPath {
			t.Errorf("%s: CommsPath = %q, want %q", name, got.CommsPath, want.commsPath)
		}
		if got.WaveAware != want.waveAware {
			t.Errorf("%s: WaveAware = %v, want %v", name, got.WaveAware, want.waveAware)
		}
		if len(got.Artifacts) != len(want.artifacts) {
			t.Errorf("%s: Artifacts = %v, want %v", name, got.Artifacts, want.artifacts)
		} else {
			for i := range want.artifacts {
				if got.Artifacts[i] != want.artifacts[i] {
					t.Errorf("%s: Artifacts[%d] = %q, want %q", name, i, got.Artifacts[i], want.artifacts[i])
				}
			}
		}

		// Verify dispatch policy is derivable
		dp := got.DispatchPolicy()
		if dp == "" {
			t.Errorf("%s: DispatchPolicy() is empty", name)
		}
	}

	// plan and status should NOT be in the registry
	for _, skip := range []string{"plan", "status"} {
		if _, ok := reg[skip]; ok {
			t.Errorf("%q should not be in registry (no dispatch_pattern)", skip)
		}
	}
}
