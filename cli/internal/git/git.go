// Package git provides thin wrappers around git commands.
package git

import (
	"os/exec"
	"strings"
)

// Run executes a git command and returns stdout, or empty string on error.
func Run(args []string, cwd string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// RepoRoot returns the git repo root, or empty string.
func RepoRoot(dir string) string {
	return Run([]string{"rev-parse", "--show-toplevel"}, dir)
}

// DetectBaseRef finds the base branch for diff spec detection.
func DetectBaseRef(repoRoot string, baseBranches []string) string {
	// Try upstream first
	upstream := Run([]string{"rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"}, repoRoot)
	if upstream != "" {
		return upstream
	}

	defaults := []string{"main", "master", "staging", "develop"}
	if len(baseBranches) > 0 {
		defaults = baseBranches
	}

	for _, base := range defaults {
		refs := []string{base, "origin/" + base, "upstream/" + base}
		for _, ref := range refs {
			out := Run([]string{"rev-parse", "--verify", ref}, repoRoot)
			if out != "" {
				return ref
			}
		}
	}

	return ""
}

// DiffNameOnly returns the list of changed files between baseRef...HEAD.
func DiffNameOnly(repoRoot, baseRef string) []string {
	out := Run([]string{"diff", "--name-only", baseRef + "...HEAD"}, repoRoot)
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}
