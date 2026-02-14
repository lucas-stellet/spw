package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/oraculo/internal/config"
	"github.com/lucas-stellet/oraculo/internal/install"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show ORACULO kit presence and spec summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Show skills installation status (use 'skills install' to install)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()

			fmt.Println("[oraculo] Skills diagnosis:")
			printDiagnosis("General", install.DiagnoseGeneralSkills(cwd))
			return nil
		},
	}

	cmd.AddCommand(newSkillsInstallCmd())
	return cmd
}

func newSkillsInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install general skills into .claude/skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			install.InstallGeneralSkills(cwd)
			return nil
		},
	}
}

func printDiagnosis(label string, skills []install.SkillStatus) {
	var installed, available, missing int
	for _, s := range skills {
		switch {
		case s.Installed:
			installed++
		case s.Available:
			available++
		default:
			missing++
		}
	}
	fmt.Printf("  %s: %d installed, %d available, %d missing\n", label, installed, available, missing)
	for _, s := range skills {
		switch {
		case s.Installed:
			fmt.Printf("    ✓ %s\n", s.Name)
		case s.Available:
			fmt.Printf("    ○ %s (available)\n", s.Name)
		default:
			fmt.Printf("    ✗ %s (no source found)\n", s.Name)
		}
	}
}

func runStatus() error {
	cwd, _ := os.Getwd()
	fmt.Printf("[oraculo] Status for project: %s\n", cwd)

	// .claude
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); err == nil {
		fmt.Println("[oraculo] .claude: present")
	} else {
		fmt.Println("[oraculo] .claude: missing")
	}

	// .spec-workflow
	if _, err := os.Stat(filepath.Join(cwd, ".spec-workflow")); err == nil {
		fmt.Println("[oraculo] .spec-workflow: present")
	} else {
		fmt.Println("[oraculo] .spec-workflow: missing")
	}

	// Config
	configPath := config.ResolveConfigPath(cwd)
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("[oraculo] config: %s\n", configPath)
	} else {
		fmt.Println("[oraculo] config: missing")
	}

	// Settings
	if _, err := os.Stat(filepath.Join(cwd, ".claude", "settings.json")); err == nil {
		fmt.Println("[oraculo] .claude/settings.json: present")
	} else {
		fmt.Println("[oraculo] .claude/settings.json: missing")
	}

	// Skills count
	skillsDir := filepath.Join(cwd, ".claude", "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				count++
			}
		}
		fmt.Printf("[oraculo] skills: %d installed\n", count)
	} else {
		fmt.Println("[oraculo] skills: none")
	}

	// Specs
	specsDir := filepath.Join(cwd, ".spec-workflow", "specs")
	if entries, err := os.ReadDir(specsDir); err == nil {
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		if len(names) > 0 {
			fmt.Printf("[oraculo] specs: %v\n", names)
		} else {
			fmt.Println("[oraculo] specs: none")
		}
	} else {
		fmt.Println("[oraculo] specs: none")
	}

	return nil
}
