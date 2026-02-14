// Package install handles deploying SPW kit files into a project.
package install

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/embedded"
	"github.com/lucas-stellet/spw/internal/render"
	"github.com/lucas-stellet/spw/internal/workspace"
)

// CommandMeta describes an SPW command for stub generation.
type CommandMeta struct {
	Name         string
	Description  string
	ArgumentHint string
}

// AllCommands returns metadata for all 13 SPW commands.
func AllCommands() []CommandMeta {
	return []CommandMeta{
		{"prd", "Zero-to-PRD discovery flow with subagents to generate requirements.md", "<spec-name> [--source <url-or-file.md>]"},
		{"plan", "Design-to-tasks planning gateway — merges design and tasks sub-phases", "<spec-name> [--mode rolling-wave|all-at-once]"},
		{"design-research", "External research subagents gather references for the design doc", "<spec-name>"},
		{"design-draft", "Draft and finalize the design document from research artifacts", "<spec-name>"},
		{"tasks-plan", "Generate executable task waves from the approved design", "<spec-name> [--mode rolling-wave|all-at-once] [--max-wave-size 3]"},
		{"tasks-check", "Audit task quality, deps, estimates, and test plans", "<spec-name>"},
		{"exec", "Subagent-driven task execution in batches with mandatory checkpoints", "<spec-name> [--batch-size 3] [--strict true|false]"},
		{"checkpoint", "Quality gate — audits code against design and tasks after a wave", "<spec-name>"},
		{"post-mortem", "Retrospective analysis after spec completion", "<spec-name>"},
		{"qa", "Generate QA test plan with scenarios and validation strategy", "<spec-name>"},
		{"qa-check", "Audit QA test plan completeness and coverage", "<spec-name>"},
		{"qa-exec", "Execute QA scenarios in waves with defect reporting", "<spec-name>"},
		{"status", "Summarize current spec stage, blockers, and exact next commands", "[<spec-name>] [--all false|true]"},
	}
}

// Options configures an install operation.
type Options struct {
	WorkspaceRoot string
}

// Run performs the full SPW install.
func Run(opts Options) error {
	root := opts.WorkspaceRoot
	fmt.Printf("[spw] Installing into project: %s\n", root)

	// 1. Backup user config before overwrite
	configPath := config.ResolveConfigPath(root)
	var configBackup []byte
	if data, err := os.ReadFile(configPath); err == nil {
		configBackup = data
	}

	// 2. Write default config and templates from embedded defaults
	if err := writeDefaults(root); err != nil {
		return fmt.Errorf("writing defaults: %w", err)
	}

	// 3. Merge config: preserve user values, add new keys
	if configBackup != nil {
		if err := mergeConfig(root, configPath, configBackup); err != nil {
			return fmt.Errorf("merging config: %w", err)
		}
		fmt.Println("[spw] Config merged: user values preserved, new keys added.")
	}

	// 4. Generate command stubs
	if err := writeCommandStubs(root); err != nil {
		return fmt.Errorf("writing command stubs: %w", err)
	}

	// 5. Render and write composed workflows
	cfg, _ := config.Load(root)
	if err := writeRenderedWorkflows(root, cfg); err != nil {
		return fmt.Errorf("rendering workflows: %w", err)
	}

	// 5b. Inject SPW dispatch instructions into CLAUDE.md and AGENTS.md
	if err := injectProjectSnippets(root); err != nil {
		return fmt.Errorf("injecting snippets: %w", err)
	}

	// 6. Generate settings.json
	if err := WriteSettings(root, cfg.AgentTeams); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	// 7. Setup .gitattributes
	SetupGitattributes(root)

	// 8. Install default skills if configured (reload config for merged values)
	cfg, _ = config.Load(root) //nolint:ineffassign
	if cfg.Skills.AutoInstallDefaultsOnSPWInstall {
		InstallDefaultSkills(root)
	} else {
		fmt.Println("[spw] Skipping default skills install (auto_install_defaults_on_spw_install=false).")
	}

	fmt.Println("[spw] Installation complete.")
	fmt.Println("[spw] Next step: adjust .spec-workflow/spw-config.toml")
	return nil
}

func writeDefaults(root string) error {
	return fs.WalkDir(embedded.Defaults, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// Map embedded path to target path
		// defaults/spw-config.toml → .spec-workflow/spw-config.toml
		// defaults/user-templates/... → .spec-workflow/user-templates/...
		rel, _ := filepath.Rel("defaults", path)
		target := filepath.Join(root, ".spec-workflow", rel)

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		data, err := embedded.Defaults.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

func mergeConfig(_, configPath string, backup []byte) error {
	// Write backup to temp file for merge
	tmp, err := os.CreateTemp("", "spw-config-backup-*.toml")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(backup); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	return config.Merge(configPath, tmp.Name(), configPath)
}

func writeCommandStubs(root string) error {
	tmplData, err := embedded.Stubs.ReadFile("stubs/command.md.tmpl")
	if err != nil {
		return fmt.Errorf("reading stub template: %w", err)
	}

	tmpl, err := template.New("command").Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("parsing stub template: %w", err)
	}

	cmdsDir := filepath.Join(root, ".claude", "commands", "spw")
	if err := os.MkdirAll(cmdsDir, 0755); err != nil {
		return err
	}

	for _, cmd := range AllCommands() {
		path := filepath.Join(cmdsDir, cmd.Name+".md")
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := tmpl.Execute(f, cmd); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	fmt.Printf("[spw] Generated %d command stubs.\n", len(AllCommands()))
	return nil
}

// RunGlobal performs a global SPW install to ~/.claude/.
// Commands, workflows, hooks, skills, and overlay symlinks are installed globally.
// No project-specific config, templates, or snippets are touched.
func RunGlobal(opts Options) error {
	home := opts.WorkspaceRoot
	fmt.Printf("[spw] Installing globally into: %s\n", home)

	// 1. Generate command stubs → ~/.claude/commands/spw/
	if err := writeCommandStubs(home); err != nil {
		return fmt.Errorf("writing command stubs: %w", err)
	}

	// 2. Render workflows with default config (no project guidelines)
	cfg := config.Defaults()
	if err := writeRenderedWorkflows(home, cfg); err != nil {
		return fmt.Errorf("rendering workflows: %w", err)
	}

	// 3. Write/merge settings.json → ~/.claude/settings.json
	// Agent Teams disabled by default for global install
	if err := WriteSettings(home, config.AgentTeamsConfig{Enabled: false}); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	// 4. Install default skills → ~/.claude/skills/
	InstallDefaultSkills(home)

	// 5. Overlay symlinks → noop by default (project controls activation)
	WriteOverlaySymlinks(home, false)

	fmt.Println("[spw] Global installation complete.")
	fmt.Println("[spw] Use 'spw init' in each project to set up project-specific config.")
	return nil
}

// RunInit performs a lightweight project initialization.
// Only project-specific assets are created: config, templates, snippets, .gitattributes.
// Commands and workflows are expected to come from a global install.
func RunInit(opts Options) error {
	root := opts.WorkspaceRoot
	fmt.Printf("[spw] Initializing project: %s\n", root)

	// 1. Write default config and templates
	if err := writeDefaults(root); err != nil {
		return fmt.Errorf("writing defaults: %w", err)
	}

	// 2. Inject snippets (CLAUDE.md, AGENTS.md)
	if err := injectProjectSnippets(root); err != nil {
		return fmt.Errorf("injecting snippets: %w", err)
	}

	// 3. Setup .gitattributes
	SetupGitattributes(root)

	// 4. Diagnose global install presence
	home, err := os.UserHomeDir()
	if err == nil {
		globalCmds := filepath.Join(home, ".claude", "commands", "spw")
		if entries, err := os.ReadDir(globalCmds); err == nil && len(entries) > 0 {
			fmt.Printf("[spw] Global install detected: %d commands in ~/.claude/commands/spw/\n", len(entries))
		} else {
			fmt.Println("[spw] No global install detected. Run 'spw install --global' or 'spw install' for a full local install.")
		}
	}

	fmt.Println("[spw] Project initialized.")
	fmt.Println("[spw] Next step: adjust .spec-workflow/spw-config.toml")
	return nil
}

// WriteOverlaySymlinks creates overlay symlinks in the target directory.
// When teamsEnabled is true, symlinks point to ../teams/<cmd>.md;
// when false, they point to ../noop.md.
func WriteOverlaySymlinks(root string, teamsEnabled bool) {
	activeDir := filepath.Join(root, ".claude", "workflows", "spw", "overlays", "active")
	if err := os.MkdirAll(activeDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[spw] Failed to create overlay active dir: %v\n", err)
		return
	}

	// Also ensure noop.md exists
	noopPath := filepath.Join(root, ".claude", "workflows", "spw", "overlays", "noop.md")
	if _, err := os.Stat(noopPath); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(noopPath), 0755)
		os.WriteFile(noopPath, []byte("<!-- noop overlay -->\n"), 0644)
	}

	for _, cmd := range AllCommands() {
		linkPath := filepath.Join(activeDir, cmd.Name+".md")
		os.Remove(linkPath) // remove existing symlink/file

		var target string
		if teamsEnabled {
			target = "../teams/" + cmd.Name + ".md"
		} else {
			target = "../noop.md"
		}
		if err := os.Symlink(target, linkPath); err != nil {
			fmt.Fprintf(os.Stderr, "[spw] Failed to create symlink %s: %v\n", linkPath, err)
		}
	}

	if teamsEnabled {
		fmt.Println("[spw] Activated team overlays via symlinks.")
	} else {
		fmt.Println("[spw] Overlay symlinks set to noop (teams disabled).")
	}
}

func writeRenderedWorkflows(root string, cfg config.Config) error {
	engine, err := render.New(cfg)
	if err != nil {
		return fmt.Errorf("creating render engine: %w", err)
	}

	// Load user guidelines if available.
	if gs := workspace.LoadGuidelines(root); len(gs) > 0 {
		adapted := make([]struct {
			Name      string
			Content   string
			AppliesTo []string
		}, len(gs))
		for i, g := range gs {
			adapted[i].Name = g.Name
			adapted[i].Content = g.Content
			adapted[i].AppliesTo = g.AppliesTo
		}
		engine.SetGuidelines(adapted)
	}

	results, err := engine.RenderAll()
	if err != nil {
		return err
	}

	wfDir := filepath.Join(root, ".claude", "workflows", "spw")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		return err
	}

	for cmd, content := range results {
		target := filepath.Join(wfDir, cmd+".md")
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return err
		}
	}

	fmt.Printf("[spw] Rendered %d workflows.\n", len(results))
	return nil
}
