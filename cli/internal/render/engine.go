package render

import (
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/embedded"
)

// refPrefix is the path prefix that Claude Code uses for @ references.
const refPrefix = "@.claude/workflows/spw/"

// sharedRefRe matches lines like "- @.claude/workflows/spw/shared/<name>.md".
var sharedRefRe = regexp.MustCompile(`^-\s+` + regexp.QuoteMeta(refPrefix) + `shared/(.+\.md)\s*$`)

// dispatchPolicyRe matches "policy: @.claude/workflows/spw/shared/dispatch-<cat>.md".
var dispatchPolicyRe = regexp.MustCompile(`^(\s*policy:\s+)` + regexp.QuoteMeta(refPrefix) + `shared/(dispatch-.+\.md)\s*$`)

// overlayRefRe matches standalone "@.claude/workflows/spw/overlays/active/<cmd>.md".
var overlayRefRe = regexp.MustCompile(`^` + regexp.QuoteMeta(refPrefix) + `overlays/active/(.+\.md)\s*$`)

// jsToolRefOld is the legacy tool invocation that should be replaced.
const jsToolRefOld = "node .claude/hooks/spw-tools.js config get"

// jsToolRefNew is the Go binary replacement.
const jsToolRefNew = "spw tools config-get"

// AllCommands lists every SPW workflow command.
var AllCommands = embedded.AllWorkflowNames

// Engine renders composed workflows from embedded sources.
type Engine struct {
	cfg          config.Config
	assets       fs.FS
	shared       map[string]string // name → content
	overlays     map[string]string // command → content
	guidelines   []guideline       // user guidelines per phase
	toolReplacer *strings.Replacer
}

// guideline stores a parsed user guideline with phase scope.
type guideline struct {
	name      string
	content   string
	appliesTo []string // empty means all phases
}

// New creates a rendering engine with the given config.
func New(cfg config.Config) (*Engine, error) {
	e := &Engine{
		cfg:      cfg,
		assets:   embedded.Assets(),
		shared:   make(map[string]string),
		overlays: make(map[string]string),
		toolReplacer: strings.NewReplacer(
			jsToolRefOld, jsToolRefNew,
		),
	}

	if err := e.loadShared(); err != nil {
		return nil, fmt.Errorf("loading shared policies: %w", err)
	}
	if err := e.loadOverlays(); err != nil {
		return nil, fmt.Errorf("loading overlays: %w", err)
	}

	return e, nil
}

// SetGuidelines configures user guidelines for injection into rendered workflows.
// Each guideline has a name, content, and list of phases it applies to.
func (e *Engine) SetGuidelines(gs []struct {
	Name      string
	Content   string
	AppliesTo []string
}) {
	e.guidelines = make([]guideline, len(gs))
	for i, g := range gs {
		e.guidelines[i] = guideline{
			name:      g.Name,
			content:   g.Content,
			appliesTo: g.AppliesTo,
		}
	}
}

// guidelinesForPhase returns the combined content of all guidelines that apply to a phase.
func (e *Engine) guidelinesForPhase(phase string) string {
	var parts []string
	for _, g := range e.guidelines {
		if len(g.appliesTo) == 0 || slices.Contains(g.appliesTo, phase) {
			parts = append(parts, "<!-- guideline: "+g.name+" -->\n"+strings.TrimRight(g.content, "\n"))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n")
}

// RenderCommand renders a single workflow command with all references inlined.
func (e *Engine) RenderCommand(command string) (string, error) {
	base, err := e.readAsset("workflows/" + command + ".md")
	if err != nil {
		return "", fmt.Errorf("reading workflow %s: %w", command, err)
	}

	lines := strings.Split(base, "\n")
	var out []string

	teamsEnabled := e.isTeamsEnabled(command)
	hasOverlayRef := false

	for _, line := range lines {
		// Shared policy reference: inline the content.
		if m := sharedRefRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			content, ok := e.shared[name]
			if !ok {
				return "", fmt.Errorf("shared policy %q not found", name)
			}
			out = append(out, strings.TrimRight(content, "\n"))
			out = append(out, "")
			continue
		}

		// Dispatch policy reference: inline after the policy key.
		if m := dispatchPolicyRe.FindStringSubmatch(line); m != nil {
			name := m[2]
			content, ok := e.shared[name]
			if !ok {
				return "", fmt.Errorf("dispatch policy %q not found", name)
			}
			out = append(out, m[1]+"(inlined below)")
			out = append(out, "")
			out = append(out, strings.TrimRight(content, "\n"))
			out = append(out, "")
			continue
		}

		// Overlay reference: conditionally inline.
		if m := overlayRefRe.FindStringSubmatch(line); m != nil {
			hasOverlayRef = true
			if teamsEnabled {
				name := strings.TrimSuffix(m[1], ".md")
				content, ok := e.overlays[name]
				if ok {
					out = append(out, strings.TrimRight(content, "\n"))
					out = append(out, "")
				}
			}
			// When teams disabled, omit the overlay entirely.
			continue
		}

		out = append(out, line)
	}

	// For workflows that don't have an explicit overlay @-reference (plan, status),
	// append the overlay at the end when teams are enabled.
	if teamsEnabled && !hasOverlayRef {
		content, ok := e.overlays[command]
		if ok {
			out = append(out, "")
			out = append(out, strings.TrimRight(content, "\n"))
		}
	}

	// Inject user guidelines for this phase.
	guidelineContent := e.guidelinesForPhase(command)
	if guidelineContent != "" {
		out = append(out, "")
		out = append(out, "<user_guidelines>")
		out = append(out, guidelineContent)
		out = append(out, "</user_guidelines>")
	}

	result := strings.Join(out, "\n")

	// Apply tool reference replacements across the entire rendered output.
	result = e.toolReplacer.Replace(result)

	return result, nil
}

// RenderAll renders all 13 commands and returns a map of command → content.
func (e *Engine) RenderAll() (map[string]string, error) {
	result := make(map[string]string, len(AllCommands))
	for _, cmd := range AllCommands {
		content, err := e.RenderCommand(cmd)
		if err != nil {
			return nil, err
		}
		result[cmd] = content
	}
	return result, nil
}

// isTeamsEnabled checks whether agent teams should be enabled for a command.
func (e *Engine) isTeamsEnabled(command string) bool {
	if !e.cfg.AgentTeams.Enabled {
		return false
	}
	return !slices.Contains(e.cfg.AgentTeams.ExcludePhases, command)
}

// loadShared reads all shared policy files into memory.
func (e *Engine) loadShared() error {
	entries, err := fs.ReadDir(e.assets, "shared")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := e.readAsset("shared/" + entry.Name())
		if err != nil {
			return err
		}
		e.shared[entry.Name()] = content
	}
	return nil
}

// loadOverlays reads all overlay files into memory.
func (e *Engine) loadOverlays() error {
	entries, err := fs.ReadDir(e.assets, "overlays")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := e.readAsset("overlays/" + entry.Name())
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		e.overlays[name] = content
	}
	return nil
}

// readAsset reads a file from the embedded assets FS.
func (e *Engine) readAsset(path string) (string, error) {
	data, err := fs.ReadFile(e.assets, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
