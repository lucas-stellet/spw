package install

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas-stellet/oraculo/internal/embedded"
)

// InjectSnippet injects or replaces a marker-delimited snippet block in targetFile.
//
// If targetFile exists and contains the start marker, the existing block
// (from start marker through end marker) is replaced with the new snippet.
// If targetFile exists but has no marker, the snippet is appended.
// If targetFile does not exist, it is created with the snippet content.
func InjectSnippet(targetFile string, snippet []byte) error {
	markerStart := []byte("<!-- ORACULO-KIT-START")
	markerEnd := []byte("<!-- ORACULO-KIT-END -->")

	existing, err := os.ReadFile(targetFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", targetFile, err)
	}

	if err == nil && bytes.Contains(existing, markerStart) {
		// Replace existing block.
		lines := bytes.Split(existing, []byte("\n"))
		var out [][]byte
		skip := false
		for _, line := range lines {
			if bytes.Contains(line, markerStart) {
				skip = true
				// Insert new snippet (already includes start+end markers).
				out = append(out, bytes.TrimRight(snippet, "\n"))
				continue
			}
			if bytes.Contains(line, markerEnd) {
				skip = false
				continue
			}
			if !skip {
				out = append(out, line)
			}
		}
		return os.WriteFile(targetFile, append(bytes.Join(out, []byte("\n")), '\n'), 0644)
	}

	// Append or create.
	if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
		return err
	}
	if len(existing) > 0 {
		// Append with leading newline.
		content := append(existing, '\n')
		content = append(content, snippet...)
		return os.WriteFile(targetFile, content, 0644)
	}
	return os.WriteFile(targetFile, snippet, 0644)
}

// injectProjectSnippets reads both embedded snippets and injects them
// into CLAUDE.md and AGENTS.md at the project root.
func injectProjectSnippets(root string) error {
	type target struct {
		file    string
		snippet string
	}
	targets := []target{
		{filepath.Join(root, "CLAUDE.md"), "snippets/claude-md.md"},
		{filepath.Join(root, "AGENTS.md"), "snippets/agents-md.md"},
	}
	for _, t := range targets {
		data, err := embedded.Snippets.ReadFile(t.snippet)
		if err != nil {
			return fmt.Errorf("reading embedded snippet %s: %w", t.snippet, err)
		}
		if err := InjectSnippet(t.file, data); err != nil {
			return fmt.Errorf("injecting snippet into %s: %w", filepath.Base(t.file), err)
		}
	}
	fmt.Println("[oraculo] Updated CLAUDE.md and AGENTS.md with ORACULO dispatch instructions.")
	return nil
}
