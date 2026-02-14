// Package viewer provides terminal and editor rendering for spec artifacts.
package viewer

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/glamour"
)

// RenderTerminal renders markdown with ANSI terminal formatting.
func RenderTerminal(content string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)
	if err != nil {
		return content, err
	}
	return r.Render(content)
}

// OpenInVSCode opens content in VS Code via stdin pipe.
func OpenInVSCode(content, filename string) error {
	cmd := exec.Command("code", "--stdin", "--filename", filename)
	cmd.Stdin = strings.NewReader(content)
	return cmd.Run()
}

// RenderOverview generates an overview markdown document for a spec.
func RenderOverview(specName, stage string, tasksDone, tasksTotal int, waves []WaveInfo) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Spec: %s\n\n", specName)
	fmt.Fprintf(&b, "**Stage:** %s\n\n", stage)

	if tasksTotal > 0 {
		pct := float64(tasksDone) / float64(tasksTotal) * 100
		fmt.Fprintf(&b, "**Tasks:** %d/%d done (%.0f%%)\n\n", tasksDone, tasksTotal, pct)
	}

	if len(waves) > 0 {
		b.WriteString("## Waves\n\n")
		b.WriteString("| Wave | Status | Exec Runs | Checkpoints |\n")
		b.WriteString("|------|--------|-----------|-------------|\n")
		for _, w := range waves {
			fmt.Fprintf(&b, "| %d | %s | %d | %d |\n", w.Num, w.Status, w.ExecRuns, w.CheckRuns)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// WaveInfo holds wave data for overview rendering.
type WaveInfo struct {
	Num       int
	Status    string
	ExecRuns  int
	CheckRuns int
}
