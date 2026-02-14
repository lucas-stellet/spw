package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lucas-stellet/oraculo/internal/config"
	"github.com/lucas-stellet/oraculo/internal/install"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check ORACULO installation health",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd)
		},
	}
}

func runDoctor(cmd *cobra.Command) error {
	version := cmd.Root().Version
	if version == "" {
		version = "dev"
	}

	cwd, _ := os.Getwd()
	fmt.Println("oraculo doctor")
	fmt.Printf("version: %s\n", version)
	fmt.Printf("workspace: %s\n", cwd)

	// Config check
	configPath := config.ResolveConfigPath(cwd)
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("config: %s (found)\n", configPath)
		cfg, err := config.LoadFromPath(configPath)
		if err != nil {
			fmt.Printf("config parse: ERROR — %v\n", err)
		} else {
			fmt.Printf("config parse: OK (models: %s/%s/%s)\n",
				cfg.Models.WebResearch, cfg.Models.ComplexReasoning, cfg.Models.Implementation)
		}
	} else {
		fmt.Printf("config: %s (missing)\n", configPath)
	}

	// Hook registration check
	settingsPath := filepath.Join(cwd, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		if install.DetectOldInstall(cwd) {
			fmt.Println("hooks: .claude/settings.json found (WARNING: old JS-based hooks detected — run 'oraculo install' to migrate)")
		} else {
			fmt.Println("hooks: .claude/settings.json found (OK)")
		}
	} else {
		fmt.Println("hooks: .claude/settings.json missing")
	}

	// Commands check
	cmdsDir := filepath.Join(cwd, ".claude", "commands", "oraculo")
	if entries, err := os.ReadDir(cmdsDir); err == nil {
		fmt.Printf("commands: %d found in .claude/commands/oraculo/\n", len(entries))
	} else {
		fmt.Println("commands: .claude/commands/oraculo/ missing")
	}

	// Workflows check
	wfDir := filepath.Join(cwd, ".claude", "workflows", "oraculo")
	if entries, err := os.ReadDir(wfDir); err == nil {
		fmt.Printf("workflows: %d found in .claude/workflows/oraculo/\n", len(entries))
	} else {
		fmt.Println("workflows: .claude/workflows/oraculo/ missing")
	}

	// Skills check
	skillsDir := filepath.Join(cwd, ".claude", "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				if _, err := os.Stat(filepath.Join(skillsDir, e.Name(), "SKILL.md")); err == nil {
					count++
				}
			}
		}
		fmt.Printf("skills: %d installed in .claude/skills/\n", count)
	} else {
		fmt.Println("skills: .claude/skills/ missing")
	}

	// Spec-workflow directory
	specDir := filepath.Join(cwd, ".spec-workflow", "specs")
	if entries, err := os.ReadDir(specDir); err == nil {
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				count++
			}
		}
		fmt.Printf("specs: %d found in .spec-workflow/specs/\n", count)
	} else {
		fmt.Println("specs: .spec-workflow/specs/ not found")
	}

	// Global install check
	home, _ := os.UserHomeDir()
	if home != "" {
		fmt.Println("")
		fmt.Println("--- global install ---")
		globalCmds := filepath.Join(home, ".claude", "commands", "oraculo")
		globalWfs := filepath.Join(home, ".claude", "workflows", "oraculo")
		globalSettings := filepath.Join(home, ".claude", "settings.json")
		globalSkills := filepath.Join(home, ".claude", "skills")

		if entries, err := os.ReadDir(globalCmds); err == nil {
			fmt.Printf("global commands: %d found in ~/.claude/commands/oraculo/\n", len(entries))
			// Warn on conflict with local
			localCmds := filepath.Join(cwd, ".claude", "commands", "oraculo")
			if _, err := os.Stat(localCmds); err == nil {
				fmt.Println("  (!) local commands also present — local takes precedence")
			}
		} else {
			fmt.Println("global commands: not installed")
		}

		if entries, err := os.ReadDir(globalWfs); err == nil {
			fmt.Printf("global workflows: %d found in ~/.claude/workflows/oraculo/\n", len(entries))
			localWfs := filepath.Join(cwd, ".claude", "workflows", "oraculo")
			if _, err := os.Stat(localWfs); err == nil {
				fmt.Println("  (!) local workflows also present — local takes precedence")
			}
		} else {
			fmt.Println("global workflows: not installed")
		}

		if _, err := os.Stat(globalSettings); err == nil {
			fmt.Println("global settings: ~/.claude/settings.json found")
		} else {
			fmt.Println("global settings: not found")
		}

		if entries, err := os.ReadDir(globalSkills); err == nil {
			count := 0
			for _, e := range entries {
				if e.IsDir() {
					if _, err := os.Stat(filepath.Join(globalSkills, e.Name(), "SKILL.md")); err == nil {
						count++
					}
				}
			}
			fmt.Printf("global skills: %d installed in ~/.claude/skills/\n", count)
		} else {
			fmt.Println("global skills: not installed")
		}
	}

	// oraculo on PATH
	if binPath, err := exec.LookPath("oraculo"); err == nil {
		fmt.Printf("\noraculo binary: %s\n", binPath)
	} else {
		fmt.Printf("\noraculo binary: not found on PATH\n")
	}

	return nil
}
