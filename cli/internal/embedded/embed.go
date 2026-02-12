// Package embedded provides go:embed access to all SPW assets.
//
// These files are the canonical source for workflows, shared policies,
// dispatch patterns, team overlays, command stub templates, and default
// user-facing files. They are embedded into the binary at build time.
package embedded

import (
	"embed"
	"io/fs"
)

// Workflows contains the 13 base workflow sources.
//
//go:embed workflows/*.md
var Workflows embed.FS

// Shared contains the 5 shared policy fragments
// (config-resolution, file-handoff, resume-policy, skills-policy, approval-reconciliation).
//
//go:embed shared/*.md
var Shared embed.FS

// Dispatch contains the 3 dispatch pattern policies
// (dispatch-pipeline, dispatch-audit, dispatch-wave).
//
//go:embed dispatch/*.md
var Dispatch embed.FS

// Overlays contains the 13 team overlay files.
//
//go:embed overlays/*.md
var Overlays embed.FS

// Stubs contains the command stub template.
//
//go:embed stubs/command.md.tmpl
var Stubs embed.FS

// Defaults contains default user-facing files
// (spw-config.toml, user-templates/).
//
//go:embed all:defaults
var Defaults embed.FS

// Snippets contains the CLAUDE.md and AGENTS.md injection snippets.
//
//go:embed snippets/*.md
var Snippets embed.FS

// Assets returns a composite FS that mirrors the embedded directory layout.
// It provides access via paths like "workflows/exec.md", "shared/config-resolution.md",
// "overlays/exec.md", etc.
func Assets() *CompositeFS {
	return &CompositeFS{
		fsByPrefix: map[string]fs.FS{
			"workflows/": Workflows,
			"shared/":    Shared,
			"dispatch/":  Dispatch,
			"overlays/":  Overlays,
			"stubs/":     Stubs,
			"defaults/":  Defaults,
			"snippets/":  Snippets,
		},
	}
}

// CompositeFS routes reads to the correct embedded FS based on path prefix.
type CompositeFS struct {
	fsByPrefix map[string]fs.FS
}

// Open implements fs.FS.
func (c *CompositeFS) Open(name string) (fs.File, error) {
	for prefix, fsys := range c.fsByPrefix {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			return fsys.Open(name)
		}
		if name == prefix[:len(prefix)-1] {
			// Directory itself.
			return fsys.Open(name)
		}
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// ReadFile implements fs.ReadFileFS.
func (c *CompositeFS) ReadFile(name string) ([]byte, error) {
	for prefix, fsys := range c.fsByPrefix {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			return fs.ReadFile(fsys, name)
		}
	}
	return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
}

// ReadDir implements fs.ReadDirFS.
func (c *CompositeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	for prefix, fsys := range c.fsByPrefix {
		dirName := prefix[:len(prefix)-1] // remove trailing /
		if name == dirName {
			return fs.ReadDir(fsys, name)
		}
	}
	return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
}

// AllWorkflowNames lists the 13 SPW command names in pipeline order.
var AllWorkflowNames = []string{
	"prd",
	"plan",
	"design-research",
	"design-draft",
	"tasks-plan",
	"tasks-check",
	"exec",
	"checkpoint",
	"post-mortem",
	"qa",
	"qa-check",
	"qa-exec",
	"status",
}
