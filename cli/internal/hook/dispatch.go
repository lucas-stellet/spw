// Package hook implements Claude Code hook handlers.
//
// Each handler reads JSON from stdin, applies validation rules,
// and exits with code 0 (pass/warn) or 2 (block).
package hook

import (
	"fmt"
	"os"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/workspace"
)

// Dispatch routes a hook event name to the appropriate handler.
func Dispatch(event string) error {
	switch event {
	case "statusline":
		return HandleStatusline()
	case "session-start":
		return HandleSessionStart()
	case "guard-prompt":
		return HandleGuardPrompt()
	case "guard-paths":
		return HandleGuardPaths()
	case "guard-stop":
		return HandleGuardStop()
	default:
		return fmt.Errorf("unknown hook event: %s", event)
	}
}

// hookContext holds the shared state for hook execution.
type hookContext struct {
	payload       workspace.Payload
	workspaceRoot string
	cfg           config.Config
}

func newHookContext() hookContext {
	p := workspace.ReadStdinPayload()
	root := workspace.GetWorkspaceRoot(p)
	cfg, _ := config.Load(root) // fail-open on config errors
	return hookContext{
		payload:       p,
		workspaceRoot: root,
		cfg:           cfg,
	}
}

// emitViolation prints a violation message to stderr and exits.
// In block mode, exits 2; in warn mode, exits 0.
func emitViolation(cfg config.HooksConfig, title string, details []string) {
	fmt.Fprintf(os.Stderr, "[spw-hook] %s\n", title)
	for _, d := range details {
		fmt.Fprintf(os.Stderr, "[spw-hook] - %s\n", d)
	}
	if cfg.EnforcementMode == "block" {
		os.Exit(2)
	}
	os.Exit(0)
}

// emitInfo prints a diagnostic message to stderr if verbose is on.
func emitInfo(cfg config.HooksConfig, message string) {
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "[spw-hook] %s\n", message)
	}
}
