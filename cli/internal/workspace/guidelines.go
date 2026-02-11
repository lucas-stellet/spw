package workspace

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Guideline represents a user guideline file with its content and phase mapping.
type Guideline struct {
	Name      string   // Filename without extension (e.g., "project", "coding").
	Content   string   // Raw markdown content (without frontmatter).
	AppliesTo []string // Phase names this guideline applies to. Empty means all phases.
}

// builtinPhaseMap maps well-known guideline filenames to their target phases.
var builtinPhaseMap = map[string][]string{
	"project": {}, // Empty means all phases.
	"coding":  {"exec", "checkpoint"},
	"quality": {"checkpoint", "qa-check", "post-mortem"},
	"testing": {"exec", "qa", "qa-check", "qa-exec"},
}

// LoadGuidelines reads all .md files from .spec-workflow/guidelines/ and resolves
// their phase mappings. Returns nil if the directory doesn't exist.
func LoadGuidelines(workspaceRoot string) []Guideline {
	dir := filepath.Join(workspaceRoot, ".spec-workflow", "guidelines")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var guidelines []Guideline
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		content, appliesTo := parseFrontmatter(string(data))

		// If no frontmatter applies_to, use builtin mapping.
		if appliesTo == nil {
			if phases, ok := builtinPhaseMap[name]; ok {
				appliesTo = phases
			}
			// Unknown names without frontmatter → all phases (empty slice).
		}

		guidelines = append(guidelines, Guideline{
			Name:      name,
			Content:   content,
			AppliesTo: appliesTo,
		})
	}

	return guidelines
}

// GuidelinesForPhase filters guidelines that apply to a given phase.
func GuidelinesForPhase(guidelines []Guideline, phase string) []Guideline {
	var result []Guideline
	for _, g := range guidelines {
		if len(g.AppliesTo) == 0 {
			result = append(result, g)
			continue
		}
		if slices.Contains(g.AppliesTo, phase) {
			result = append(result, g)
		}
	}
	return result
}

// parseFrontmatter extracts applies_to from YAML-like frontmatter and returns
// the content without frontmatter. Returns nil appliesTo if no frontmatter found.
func parseFrontmatter(raw string) (content string, appliesTo []string) {
	if !strings.HasPrefix(raw, "---\n") {
		return raw, nil
	}

	end := strings.Index(raw[4:], "\n---")
	if end == -1 {
		return raw, nil
	}

	frontmatter := raw[4 : 4+end]
	content = strings.TrimLeft(raw[4+end+4:], "\n")

	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "applies_to:") {
			val := strings.TrimPrefix(line, "applies_to:")
			val = strings.TrimSpace(val)
			// Parse [phase1, phase2] or phase1, phase2.
			val = strings.Trim(val, "[]")
			for _, item := range strings.Split(val, ",") {
				item = strings.TrimSpace(item)
				item = strings.Trim(item, `"'`)
				if item != "" {
					appliesTo = append(appliesTo, item)
				}
			}
			break
		}
	}

	return content, appliesTo
}
