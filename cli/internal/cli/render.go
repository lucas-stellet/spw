package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/oraculo/internal/config"
	"github.com/lucas-stellet/oraculo/internal/render"
	"github.com/lucas-stellet/oraculo/internal/workspace"
	"github.com/spf13/cobra"
)

func newRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render [command]",
		Short: "Render a composed workflow to stdout or disk",
		Long:  "Renders workflow markdown with all shared policies, dispatch patterns, and overlays inlined.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			if !all && len(args) == 0 {
				return fmt.Errorf("provide a command name or use --all")
			}

			cwd, _ := os.Getwd()
			cfg, _ := config.Load(cwd)

			engine, err := render.New(cfg)
			if err != nil {
				return fmt.Errorf("initializing render engine: %w", err)
			}

			// Load user guidelines and inject into engine.
			if gs := workspace.LoadGuidelines(cwd); len(gs) > 0 {
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

			if all {
				return renderAll(engine, cwd)
			}
			return renderOne(engine, args[0])
		},
	}

	cmd.Flags().Bool("all", false, "Render all 13 commands to .claude/workflows/oraculo/")

	return cmd
}

func renderOne(engine *render.Engine, command string) error {
	content, err := engine.RenderCommand(command)
	if err != nil {
		return err
	}
	fmt.Print(content)
	return nil
}

func renderAll(engine *render.Engine, cwd string) error {
	results, err := engine.RenderAll()
	if err != nil {
		return err
	}

	outDir := filepath.Join(cwd, ".claude", "workflows", "oraculo")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for cmd, content := range results {
		path := filepath.Join(outDir, cmd+".md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", cmd, err)
		}
		fmt.Printf("[oraculo] rendered %s\n", path)
	}
	return nil
}
