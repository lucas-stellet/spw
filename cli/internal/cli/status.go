package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/install"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show SPW kit presence and spec summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func newSkillsCmd() *cobra.Command {
	var elixirFlag, allFlag bool

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Install skills (general by default, --elixir for Elixir, --all for everything)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()

			switch {
			case allFlag:
				install.InstallDefaultSkills(cwd)
			case elixirFlag:
				install.InstallElixirSkills(cwd)
			default:
				install.InstallGeneralSkills(cwd)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&elixirFlag, "elixir", false, "Install Elixir-specific skills and patch config")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Install all skills (general + Elixir)")

	return cmd
}

func runStatus() error {
	cwd, _ := os.Getwd()
	fmt.Printf("[spw] Status for project: %s\n", cwd)

	// .claude
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); err == nil {
		fmt.Println("[spw] .claude: present")
	} else {
		fmt.Println("[spw] .claude: missing")
	}

	// .spec-workflow
	if _, err := os.Stat(filepath.Join(cwd, ".spec-workflow")); err == nil {
		fmt.Println("[spw] .spec-workflow: present")
	} else {
		fmt.Println("[spw] .spec-workflow: missing")
	}

	// Config
	configPath := config.ResolveConfigPath(cwd)
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("[spw] config: %s\n", configPath)
	} else {
		fmt.Println("[spw] config: missing")
	}

	// Settings
	if _, err := os.Stat(filepath.Join(cwd, ".claude", "settings.json")); err == nil {
		fmt.Println("[spw] .claude/settings.json: present")
	} else {
		fmt.Println("[spw] .claude/settings.json: missing")
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
		fmt.Printf("[spw] skills: %d installed\n", count)
	} else {
		fmt.Println("[spw] skills: none")
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
			fmt.Printf("[spw] specs: %v\n", names)
		} else {
			fmt.Println("[spw] specs: none")
		}
	} else {
		fmt.Println("[spw] specs: none")
	}

	return nil
}
