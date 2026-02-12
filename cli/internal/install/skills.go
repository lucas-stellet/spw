package install

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-stellet/spw/internal/config"
)

// ElixirSkills lists Elixir-specific skill names.
var ElixirSkills = []string{
	"using-elixir-skills",
	"elixir-thinking",
	"elixir-anti-patterns",
	"phoenix-thinking",
	"ecto-thinking",
	"otp-thinking",
	"oban-thinking",
}

// GeneralSkills lists technology-agnostic skill names.
var GeneralSkills = []string{
	"mermaid-architecture",
	"qa-validation-planning",
	"conventional-commits",
	"test-driven-development",
}

// DefaultSkills lists all skill names (general + elixir) for backward compatibility.
var DefaultSkills = append(append([]string{}, GeneralSkills...), ElixirSkills...)

// ElixirRequiredSkills are injected into config required lists by --elixir.
// using-elixir-skills acts as a router that helps the agent discover the others.
var ElixirRequiredSkills = []string{
	"using-elixir-skills",
	"elixir-anti-patterns",
}

// InstallDefaultSkills copies all default skills (general + elixir) from known source locations.
func InstallDefaultSkills(root string) {
	installSkillSet(root, DefaultSkills, "all")
}

// InstallGeneralSkills copies only technology-agnostic skills.
func InstallGeneralSkills(root string) {
	installSkillSet(root, GeneralSkills, "general")
}

// InstallElixirSkills copies Elixir-specific skills and patches config required lists.
func InstallElixirSkills(root string) {
	installSkillSet(root, ElixirSkills, "elixir")
	if err := PatchConfigElixirSkills(root); err != nil {
		fmt.Printf("[spw] Warning: could not patch config with Elixir skills: %v\n", err)
	}
}

func installSkillSet(root string, skills []string, label string) {
	targetDir := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("[spw] Failed to create skills dir: %v\n", err)
		return
	}

	var installed, skipped int
	var missing []string

	for _, skill := range skills {
		dest := filepath.Join(targetDir, skill)
		if _, err := os.Stat(dest); err == nil {
			skipped++
			continue
		}

		srcDir := findSkillSource(skill)
		if srcDir == "" {
			missing = append(missing, skill)
			continue
		}

		if err := copyDir(srcDir, dest); err != nil {
			missing = append(missing, skill)
			continue
		}
		installed++
	}

	fmt.Printf("[spw] Skills (%s): installed=%d, existing=%d, missing=%d\n", label, installed, skipped, len(missing))
	if len(missing) > 0 {
		fmt.Printf("[spw] Missing local skill sources (non-blocking): %v\n", missing)
	}
}

// PatchConfigElixirSkills adds using-elixir-skills and elixir-anti-patterns
// to [skills.design].required and [skills.implementation].required in the
// project's spw-config.toml, preserving comments and formatting.
func PatchConfigElixirSkills(root string) error {
	configPath := config.ResolveConfigPath(root)
	if _, err := os.Stat(configPath); err != nil {
		fmt.Println("[spw] No config file found; skipping Elixir config patch.")
		return nil
	}

	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	// Check which sections need patching
	designNeeds := missingSkills(cfg.Skills.Design.Required, ElixirRequiredSkills)
	implNeeds := missingSkills(cfg.Skills.Implementation.Required, ElixirRequiredSkills)

	if len(designNeeds) == 0 && len(implNeeds) == 0 {
		fmt.Println("[spw] Elixir skills already in config required lists.")
		return nil
	}

	// Text-based patch to preserve comments
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	sections := map[string][]string{
		"skills.design":         designNeeds,
		"skills.implementation": implNeeds,
	}

	patched := patchTOMLRequiredArrays(string(data), sections)

	if err := os.WriteFile(configPath, []byte(patched), 0644); err != nil {
		return err
	}

	fmt.Printf("[spw] Patched config: added %v to skills.design.required and skills.implementation.required\n", ElixirRequiredSkills)
	return nil
}

// missingSkills returns entries from wanted that are not in current.
func missingSkills(current, wanted []string) []string {
	set := make(map[string]bool, len(current))
	for _, s := range current {
		set[s] = true
	}
	var out []string
	for _, s := range wanted {
		if !set[s] {
			out = append(out, s)
		}
	}
	return out
}

// patchTOMLRequiredArrays inserts skill entries into `required = [...]` arrays
// under the specified TOML sections, preserving file formatting.
func patchTOMLRequiredArrays(content string, sections map[string][]string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	currentSection := ""
	var result []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Track section headers
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentSection = trimmed[1 : len(trimmed)-1]
			result = append(result, line)
			continue
		}

		// Look for `required = [` in a target section
		toAdd, ok := sections[currentSection]
		if ok && len(toAdd) > 0 && strings.HasPrefix(trimmed, "required") && strings.Contains(trimmed, "=") {
			// Detect inline vs multi-line array
			if strings.Contains(trimmed, "]") {
				// Single-line: required = ["a", "b"] or required = []
				result = append(result, expandSingleLineArray(line, toAdd))
			} else {
				// Multi-line: required = [
				result = append(result, line)
				// Collect existing array lines, find the closing ]
				var arrayLines []string
				for i+1 < len(lines) {
					i++
					nextTrimmed := strings.TrimSpace(lines[i])
					if nextTrimmed == "]" {
						// Ensure last existing entry has trailing comma
						if len(arrayLines) > 0 {
							last := arrayLines[len(arrayLines)-1]
							lastTrimmed := strings.TrimSpace(last)
							if lastTrimmed != "" && !strings.HasSuffix(lastTrimmed, ",") {
								arrayLines[len(arrayLines)-1] = last + ","
							}
						}
						result = append(result, arrayLines...)
						// Insert new entries before closing bracket
						for _, skill := range toAdd {
							result = append(result, fmt.Sprintf("  \"%s\",", skill))
						}
						result = append(result, lines[i])
						break
					}
					arrayLines = append(arrayLines, lines[i])
				}
			}
			// Clear so we don't patch again in same section
			sections[currentSection] = nil
			continue
		}

		result = append(result, line)
	}

	out := strings.Join(result, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out
}

// expandSingleLineArray handles `required = []` or `required = ["a"]` on one line.
func expandSingleLineArray(line string, toAdd []string) string {
	eqIdx := strings.Index(line, "=")
	prefix := line[:eqIdx+1] + " "
	afterEq := strings.TrimSpace(line[eqIdx+1:])

	// Extract existing entries
	inner := strings.TrimSuffix(strings.TrimPrefix(afterEq, "["), "]")
	inner = strings.TrimSpace(inner)

	var entries []string
	if inner != "" {
		for _, e := range strings.Split(inner, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				entries = append(entries, e)
			}
		}
	}

	// Add new entries
	for _, s := range toAdd {
		entries = append(entries, fmt.Sprintf("\"%s\"", s))
	}

	if len(entries) <= 2 {
		return prefix + "[" + strings.Join(entries, ", ") + "]"
	}

	// Switch to multi-line for readability
	var result []string
	result = append(result, prefix+"[")
	for _, e := range entries {
		if !strings.HasSuffix(e, ",") {
			e += ","
		}
		result = append(result, "  "+e)
	}
	result = append(result, "]")
	return strings.Join(result, "\n")
}

func findSkillSource(skill string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	candidates := []string{
		filepath.Join(home, ".claude", "skills", skill),
		filepath.Join(home, ".codex", "skills", skill),
		filepath.Join(home, ".codex", "superpowers", "skills", skill),
		filepath.Join(home, ".config", "opencode", "skills", skill),
	}

	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err == nil {
			return dir
		}
	}

	return ""
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

// SetupGitattributes adds the linguist-generated rule if not already present.
func SetupGitattributes(root string) {
	rule := ".spec-workflow/specs/** linguist-generated=true"
	path := filepath.Join(root, ".gitattributes")

	existing, err := os.ReadFile(path)
	if err == nil {
		if strings.Contains(string(existing), rule) {
			return
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintln(f, rule)
	fmt.Println("[spw] Added .gitattributes rule for PR review optimization.")
}

