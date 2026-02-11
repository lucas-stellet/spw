// Package registry loads dispatch metadata from embedded workflow files,
// making the <dispatch_pattern> section the single source of truth.
package registry

import (
	"bufio"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// CommandMeta holds dispatch metadata parsed from a workflow's <dispatch_pattern>.
type CommandMeta struct {
	Phase       string   // spec directory phase (e.g. "execution", "qa")
	Category    string   // dispatch category (e.g. "pipeline", "audit", "wave-execution")
	Subcategory string   // dispatch subcategory (e.g. "research", "code", "implementation")
	CommsPath   string   // template path with {wave} placeholder for wave-aware commands
	Artifacts   []string // additional dirs to create under spec dir (e.g. "execution/_implementation-logs")
	Policy      string   // policy @-reference (e.g. "@.claude/workflows/spw/shared/dispatch-wave.md")
	WaveAware   bool     // derived: true when CommsPath contains "{wave}"
}

// DispatchPolicy returns the dispatch policy identifier derived from the policy reference.
// e.g. "@.claude/workflows/spw/shared/dispatch-wave.md" → "dispatch-wave"
func (m CommandMeta) DispatchPolicy() string {
	base := filepath.Base(m.Policy)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// Load reads all workflow .md files from the given FS, parses their
// <dispatch_pattern> sections, and returns a map of command name → CommandMeta.
// Workflows without a <dispatch_pattern> (e.g. plan, status) are skipped.
func Load(fsys fs.FS) (map[string]CommandMeta, error) {
	entries, err := fs.ReadDir(fsys, "workflows")
	if err != nil {
		return nil, fmt.Errorf("reading workflows dir: %w", err)
	}

	registry := make(map[string]CommandMeta)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := fs.ReadFile(fsys, "workflows/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		meta, ok := parseDispatchPattern(string(data))
		if !ok {
			continue // no dispatch_pattern — skip (plan.md, status.md)
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		registry[name] = meta
	}

	return registry, nil
}

// Category returns the dispatch category for a command, or empty string if not found.
func Category(registry map[string]CommandMeta, command string) string {
	if meta, ok := registry[command]; ok {
		return meta.Category
	}
	return ""
}

// parseDispatchPattern extracts CommandMeta from the <dispatch_pattern> block.
func parseDispatchPattern(content string) (CommandMeta, bool) {
	const openTag = "<dispatch_pattern>"
	const closeTag = "</dispatch_pattern>"

	start := strings.Index(content, openTag)
	if start == -1 {
		return CommandMeta{}, false
	}
	end := strings.Index(content[start:], closeTag)
	if end == -1 {
		return CommandMeta{}, false
	}

	block := content[start+len(openTag) : start+end]

	var meta CommandMeta
	scanner := bufio.NewScanner(strings.NewReader(block))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		key, value, ok := parseKeyValue(line)
		if !ok {
			continue
		}

		switch key {
		case "category":
			meta.Category = value
		case "subcategory":
			meta.Subcategory = value
		case "phase":
			meta.Phase = value
		case "comms_path":
			meta.CommsPath = value
		case "artifacts":
			for _, a := range strings.Split(value, ",") {
				a = strings.TrimSpace(a)
				if a != "" {
					meta.Artifacts = append(meta.Artifacts, a)
				}
			}
		case "policy":
			meta.Policy = value
		}
	}

	meta.WaveAware = strings.Contains(meta.CommsPath, "{wave}")

	return meta, meta.Category != "" && meta.CommsPath != ""
}

// parseKeyValue splits a "key: value" line.
func parseKeyValue(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx == -1 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	return key, value, key != "" && value != ""
}
