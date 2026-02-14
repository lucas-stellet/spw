// Package summary generates structured completion and progress summaries
// with YAML frontmatter for spec-workflow specs.
package summary

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// CompletionFrontmatter holds metadata for a completed spec summary.
type CompletionFrontmatter struct {
	Spec               string    `yaml:"spec"`
	Status             string    `yaml:"status"`
	CompletedAt        time.Time `yaml:"completed_at"`
	DurationDays       int       `yaml:"duration_days"`
	TasksCount         int       `yaml:"tasks_count"`
	WavesCount         int       `yaml:"waves_count"`
	CheckpointPasses   int       `yaml:"checkpoint_passes"`
	CheckpointFailures int       `yaml:"checkpoint_failures"`
	FilesChanged       []string  `yaml:"files_changed"`
	Technologies       []string  `yaml:"technologies"`
	Tags               []string  `yaml:"tags"`
	Summary            string    `yaml:"summary"`
}

// ProgressFrontmatter holds metadata for an in-progress spec summary.
type ProgressFrontmatter struct {
	Spec            string    `yaml:"spec"`
	Status          string    `yaml:"status"`
	Stage           string    `yaml:"stage"`
	AsOf            time.Time `yaml:"as_of"`
	TasksDone       int       `yaml:"tasks_done"`
	TasksTotal      int       `yaml:"tasks_total"`
	TasksPending    int       `yaml:"tasks_pending"`
	TasksInProgress int       `yaml:"tasks_in_progress,omitempty"`
	CurrentWave     int       `yaml:"current_wave"`
	WavesTotal      int       `yaml:"waves_total"`
	FilesChanged    []string  `yaml:"files_changed,omitempty"`
	Technologies    []string  `yaml:"technologies,omitempty"`
}

// CompletionSummary combines frontmatter and body for a completion summary.
type CompletionSummary struct {
	Frontmatter CompletionFrontmatter
	Body        string
}

// ProgressSummary combines frontmatter and body for a progress summary.
type ProgressSummary struct {
	Frontmatter ProgressFrontmatter
	Body        string
}

// RenderFull serializes frontmatter to YAML between --- delimiters
// and appends the markdown body.
func RenderFull(fm interface{}, body string) (string, error) {
	data, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("summary: marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(data)
	buf.WriteString("---\n\n")
	buf.WriteString(body)
	return buf.String(), nil
}
