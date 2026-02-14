package install

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GeneralSkills lists technology-agnostic skill names.
var GeneralSkills = []string{
	"mermaid-architecture",
	"qa-validation-planning",
	"conventional-commits",
	"test-driven-development",
}

// SkillStatus represents the installation state of a single skill.
type SkillStatus struct {
	Name      string
	Installed bool // present in .claude/skills
	Available bool // source found (can be installed)
}

// DiagnoseGeneralSkills checks installation status of general skills without modifying anything.
func DiagnoseGeneralSkills(root string) []SkillStatus {
	targetDir := filepath.Join(root, ".claude", "skills")
	var result []SkillStatus

	for _, skill := range GeneralSkills {
		s := SkillStatus{Name: skill}
		if _, err := os.Stat(filepath.Join(targetDir, skill)); err == nil {
			s.Installed = true
		}
		if !s.Installed && findSkillSource(skill) != "" {
			s.Available = true
		}
		result = append(result, s)
	}
	return result
}

// InstallGeneralSkills copies only technology-agnostic skills.
func InstallGeneralSkills(root string) {
	installSkillSet(root, GeneralSkills, "general")
}

func installSkillSet(root string, skills []string, label string) {
	targetDir := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("[oraculo] Failed to create skills dir: %v\n", err)
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

	fmt.Printf("[oraculo] Skills (%s): installed=%d, existing=%d, missing=%d\n", label, installed, skipped, len(missing))
	if len(missing) > 0 {
		fmt.Printf("[oraculo] Missing local skill sources (non-blocking): %v\n", missing)
	}
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
	fmt.Println("[oraculo] Added .gitattributes rule for PR review optimization.")
}

