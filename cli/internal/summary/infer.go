package summary

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/lucas-stellet/spw/internal/tasks"
)

// extToTech maps file extensions to technology names.
var extToTech = map[string]string{
	".go":         "Go",
	".ts":         "TypeScript",
	".tsx":        "TypeScript",
	".js":         "JavaScript",
	".jsx":        "JavaScript",
	".py":         "Python",
	".ex":         "Elixir",
	".exs":        "Elixir",
	".rs":         "Rust",
	".java":       "Java",
	".kt":         "Kotlin",
	".rb":         "Ruby",
	".swift":      "Swift",
	".sql":        "SQL",
	".sh":         "Shell",
	".bash":       "Shell",
	".css":        "CSS",
	".scss":       "CSS",
	".html":       "HTML",
	".yaml":       "YAML",
	".yml":        "YAML",
	".toml":       "TOML",
	".json":       "JSON",
	".md":         "Markdown",
	".proto":      "Protocol Buffers",
	".graphql":    "GraphQL",
	".gql":        "GraphQL",
	".dockerfile": "Docker",
	".tf":         "Terraform",
}

// tagKeywords maps keyword patterns (case-insensitive) to tag names.
var tagKeywords = []struct {
	keywords []string
	tag      string
}{
	{[]string{"auth", "login", "jwt", "oauth", "session"}, "authentication"},
	{[]string{"test", "spec", "assert", "tdd", "coverage"}, "testing"},
	{[]string{"database", "db", "sql", "migration", "schema", "query"}, "database"},
	{[]string{"api", "endpoint", "route", "handler", "rest", "graphql"}, "api"},
	{[]string{"ui", "frontend", "component", "layout", "form", "button", "page"}, "frontend"},
	{[]string{"cli", "command", "flag", "terminal"}, "cli"},
	{[]string{"doc", "readme", "guide"}, "documentation"},
	{[]string{"refactor", "cleanup", "rename"}, "refactoring"},
	{[]string{"perf", "optimize", "cache", "speed"}, "performance"},
	{[]string{"config", "setting", "env"}, "configuration"},
	{[]string{"log", "monitor", "metric", "trace"}, "observability"},
	{[]string{"deploy", "ci", "cd", "pipeline"}, "devops"},
	{[]string{"error", "exception", "retry", "fallback"}, "error-handling"},
	{[]string{"file", "upload", "download", "storage"}, "storage"},
	{[]string{"queue", "worker", "job", "async"}, "background-jobs"},
	{[]string{"security", "encrypt", "hash", "permission"}, "security"},
	{[]string{"fix", "bug", "patch", "hotfix"}, "bugfix"},
	{[]string{"cache", "redis", "memcache"}, "caching"},
	{[]string{"middleware"}, "middleware"},
	{[]string{"webhook", "event", "notification"}, "events"},
}

const maxTags = 10

// InferTechnologies maps file extensions to technology names.
// Returns a sorted, deduplicated list.
func InferTechnologies(files []string) []string {
	seen := make(map[string]bool)
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if tech, ok := extToTech[ext]; ok {
			seen[tech] = true
		}
		// Check for Dockerfile by basename.
		base := strings.ToLower(filepath.Base(f))
		if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") {
			seen["Docker"] = true
		}
		// Additional rules from doc.
		if strings.Contains(f, "_test.go") || strings.HasSuffix(f, ".test.ts") || strings.HasSuffix(f, ".test.js") {
			seen["Testing"] = true
		}
		if strings.Contains(f, "migrations/") || strings.Contains(f, "migration/") {
			seen["Database Migrations"] = true
		}
	}

	result := make([]string, 0, len(seen))
	for tech := range seen {
		result = append(result, tech)
	}
	sort.Strings(result)
	return result
}

// InferTags extracts topic tags from task titles via keyword matching.
// Returns a sorted, deduplicated list limited to maxTags entries.
func InferTags(titles []string, files []string) []string {
	seen := make(map[string]bool)

	// Build combined text from titles for matching.
	for _, title := range titles {
		lower := strings.ToLower(title)
		words := strings.FieldsFunc(lower, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_')
		})
		wordSet := make(map[string]bool, len(words))
		for _, w := range words {
			wordSet[w] = true
		}

		for _, entry := range tagKeywords {
			for _, kw := range entry.keywords {
				if wordSet[kw] {
					seen[entry.tag] = true
					break
				}
			}
		}
	}

	result := make([]string, 0, len(seen))
	for tag := range seen {
		result = append(result, tag)
	}
	sort.Strings(result)
	if len(result) > maxTags {
		result = result[:maxTags]
	}
	return result
}

// CollectFilesChanged aggregates unique file paths from completed tasks.
// Returns a sorted, deduplicated list.
func CollectFilesChanged(taskList []tasks.Task) []string {
	seen := make(map[string]bool)
	var files []string
	for _, t := range taskList {
		if t.Status != "done" {
			continue
		}
		for _, f := range parseFilesList(t.Files) {
			f = strings.TrimSpace(f)
			if f != "" && !seen[f] {
				seen[f] = true
				files = append(files, f)
			}
		}
	}
	sort.Strings(files)
	return files
}

// parseFilesList splits a files field that may be comma-separated or contain backticks.
func parseFilesList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Remove backticks used for code formatting.
	s = strings.ReplaceAll(s, "`", "")
	// Split on commas.
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
