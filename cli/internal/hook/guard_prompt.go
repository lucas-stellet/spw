package hook

import (
	"regexp"
	"strings"

	"github.com/lucas-stellet/oraculo/internal/workspace"
)

// commandsRequiringSpec is the set of ORACULO commands that need a <spec-name> argument.
var commandsRequiringSpec = map[string]bool{
	"discover":        true,
	"plan":            true,
	"design-research": true,
	"design-draft":    true,
	"tasks-plan":      true,
	"tasks-check":     true,
	"exec":            true,
	"checkpoint":      true,
	"qa":              true,
	"qa-check":        true,
	"qa-exec":         true,
}

var oraculoCommandRe = regexp.MustCompile(`(?i)^/oraculo:([a-z-]+)(?:\s+(.*))?$`)

// parsedCommand represents a parsed /oraculo: command line.
type parsedCommand struct {
	command  string
	argsLine string
}

// HandleGuardPrompt validates that ORACULO commands include a spec-name argument.
func HandleGuardPrompt() error {
	ctx := newHookContext()
	if !ctx.cfg.Hooks.Enabled {
		return nil
	}

	prompt := extractPrompt(ctx.payload)
	if prompt == "" {
		return nil
	}

	parsed := firstOraculoCommand(prompt)
	if parsed == nil {
		return nil
	}

	if !commandsRequiringSpec[parsed.command] {
		return nil
	}

	specName := extractSpecArg(parsed.argsLine)
	if specName != "" {
		writeStatuslineCache(ctx.workspaceRoot, specName, map[string]string{
			"source": "oraculo-command",
			"sticky": "true",
		})
	}

	if ctx.cfg.Hooks.GuardPromptRequireSpec && !hasSpecArg(parsed.argsLine) {
		emitViolation(ctx.cfg.Hooks, "Missing <spec-name> for /oraculo:"+parsed.command, []string{
			"Expected usage: /oraculo:" + parsed.command + " <spec-name>",
			"Tip: use /oraculo:status if you need help discovering the current stage.",
		})
	}

	return nil
}

func extractPrompt(p workspace.Payload) string {
	if p.Prompt != "" {
		return p.Prompt
	}
	return ""
}

func firstOraculoCommand(prompt string) *parsedCommand {
	for _, line := range strings.Split(prompt, "\n") {
		trimmed := strings.TrimSpace(line)
		m := oraculoCommandRe.FindStringSubmatch(trimmed)
		if m != nil {
			return &parsedCommand{
				command:  strings.ToLower(m[1]),
				argsLine: strings.TrimSpace(m[2]),
			}
		}
	}
	return nil
}

func extractSpecArg(argsLine string) string {
	tokens := tokenizeArgs(argsLine)
	if len(tokens) == 0 {
		return ""
	}
	first := tokens[0]
	if strings.HasPrefix(first, "--") {
		return ""
	}
	return first
}

func hasSpecArg(argsLine string) bool {
	return extractSpecArg(argsLine) != ""
}

var tokenRe = regexp.MustCompile(`"[^"]*"|'[^']*'|\S+`)

func tokenizeArgs(argsLine string) []string {
	if argsLine == "" {
		return nil
	}
	matches := tokenRe.FindAllString(argsLine, -1)
	var result []string
	for _, m := range matches {
		if len(m) >= 2 && ((m[0] == '"' && m[len(m)-1] == '"') || (m[0] == '\'' && m[len(m)-1] == '\'')) {
			m = m[1 : len(m)-1]
		}
		if m != "" {
			result = append(result, m)
		}
	}
	return result
}
