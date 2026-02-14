package summary

import (
	"strings"
	"testing"
	"time"

	"github.com/lucas-stellet/spw/internal/tasks"
	"github.com/lucas-stellet/spw/internal/wave"
	"gopkg.in/yaml.v3"
)

func TestInferTechnologies(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  []string
	}{
		{
			name:  "go files",
			files: []string{"main.go", "handler_test.go"},
			want:  []string{"Go", "Testing"},
		},
		{
			name:  "typescript and javascript",
			files: []string{"app.ts", "utils.tsx", "legacy.js"},
			want:  []string{"JavaScript", "TypeScript"},
		},
		{
			name:  "python",
			files: []string{"script.py"},
			want:  []string{"Python"},
		},
		{
			name:  "sql and toml",
			files: []string{"schema.sql", "config.toml"},
			want:  []string{"SQL", "TOML"},
		},
		{
			name:  "dockerfile by name",
			files: []string{"Dockerfile", "Dockerfile.dev"},
			want:  []string{"Docker"},
		},
		{
			name:  "migration paths",
			files: []string{"migrations/001_init.sql"},
			want:  []string{"Database Migrations", "SQL"},
		},
		{
			name:  "empty",
			files: nil,
			want:  nil,
		},
		{
			name:  "deduplicated",
			files: []string{"a.go", "b.go", "c.go"},
			want:  []string{"Go"},
		},
		{
			name:  "sorted output",
			files: []string{"z.py", "a.go", "m.ts"},
			want:  []string{"Go", "Python", "TypeScript"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferTechnologies(tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("InferTechnologies(%v) = %v, want %v", tt.files, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("InferTechnologies(%v)[%d] = %q, want %q", tt.files, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestInferTags(t *testing.T) {
	tests := []struct {
		name   string
		titles []string
		files  []string
		want   []string
	}{
		{
			name:   "auth keyword",
			titles: []string{"Add auth middleware"},
			want:   []string{"authentication", "middleware"},
		},
		{
			name:   "testing keyword",
			titles: []string{"Write test for parser"},
			want:   []string{"testing"},
		},
		{
			name:   "database and api",
			titles: []string{"Create database schema", "Add API endpoint"},
			want:   []string{"api", "database"},
		},
		{
			name:   "no matches",
			titles: []string{"Do something"},
			want:   nil,
		},
		{
			name:   "empty titles",
			titles: nil,
			want:   nil,
		},
		{
			name:   "case insensitive",
			titles: []string{"Add CLI command"},
			want:   []string{"cli"},
		},
		{
			name:   "max tags cap",
			titles: []string{"auth test database api ui cli doc refactor perf config log deploy error"},
			want: []string{
				"api", "authentication", "cli", "configuration",
				"database", "devops", "documentation", "error-handling",
				"frontend", "observability",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferTags(tt.titles, tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("InferTags(%v) = %v (len %d), want %v (len %d)", tt.titles, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("InferTags[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCollectFilesChanged(t *testing.T) {
	tests := []struct {
		name  string
		tasks []tasks.Task
		want  []string
	}{
		{
			name: "only done tasks",
			tasks: []tasks.Task{
				{ID: "1", Status: "done", Files: "a.go, b.go"},
				{ID: "2", Status: "pending", Files: "c.go"},
			},
			want: []string{"a.go", "b.go"},
		},
		{
			name: "deduplicated and sorted",
			tasks: []tasks.Task{
				{ID: "1", Status: "done", Files: "b.go, a.go"},
				{ID: "2", Status: "done", Files: "a.go, c.go"},
			},
			want: []string{"a.go", "b.go", "c.go"},
		},
		{
			name: "strips backticks",
			tasks: []tasks.Task{
				{ID: "1", Status: "done", Files: "`main.go`, `util.go`"},
			},
			want: []string{"main.go", "util.go"},
		},
		{
			name:  "empty task list",
			tasks: nil,
			want:  nil,
		},
		{
			name: "empty files field",
			tasks: []tasks.Task{
				{ID: "1", Status: "done", Files: ""},
			},
			want: nil,
		},
		{
			name: "single file no comma",
			tasks: []tasks.Task{
				{ID: "1", Status: "done", Files: "main.go"},
			},
			want: []string{"main.go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectFilesChanged(tt.tasks)
			if len(got) != len(tt.want) {
				t.Fatalf("CollectFilesChanged = %v (len %d), want %v (len %d)", got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("CollectFilesChanged[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRenderFrontmatter(t *testing.T) {
	fm := ProgressFrontmatter{
		Spec:       "test-spec",
		Status:     "in_progress",
		Stage:      "execution",
		AsOf:       time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		TasksDone:  3,
		TasksTotal: 5,
	}

	result, err := RenderFull(fm, "# Body\n")
	if err != nil {
		t.Fatalf("RenderFull error: %v", err)
	}

	if !strings.HasPrefix(result, "---\n") {
		t.Error("RenderFull should start with ---")
	}
	if !strings.Contains(result, "---\n\n# Body\n") {
		t.Error("RenderFull should have body after closing ---")
	}
	if !strings.Contains(result, "spec: test-spec") {
		t.Error("RenderFull should contain spec field")
	}
	if !strings.Contains(result, "stage: execution") {
		t.Error("RenderFull should contain stage field")
	}

	// Verify YAML between delimiters is parseable.
	parts := strings.SplitN(result, "---", 3)
	if len(parts) < 3 {
		t.Fatal("RenderFull output should have two --- delimiters")
	}
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[1]), &parsed); err != nil {
		t.Fatalf("YAML between delimiters is not valid: %v", err)
	}
	if parsed["spec"] != "test-spec" {
		t.Errorf("parsed spec = %v, want test-spec", parsed["spec"])
	}
}

func TestRenderFrontmatterYAMLEscaping(t *testing.T) {
	fm := CompletionFrontmatter{
		Spec:        "test: special",
		Status:      "completed",
		CompletedAt: time.Now().UTC(),
		Summary:     `Contains "quotes" and: colons`,
		Tags:        []string{"tag-1", "tag with spaces"},
	}

	result, err := RenderFull(fm, "")
	if err != nil {
		t.Fatalf("RenderFull error: %v", err)
	}

	parts := strings.SplitN(result, "---", 3)
	if len(parts) < 3 {
		t.Fatal("Expected two --- delimiters")
	}
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[1]), &parsed); err != nil {
		t.Fatalf("YAML with special chars should parse: %v", err)
	}
	if parsed["spec"] != "test: special" {
		t.Errorf("spec = %v, want 'test: special'", parsed["spec"])
	}
}

func TestGenerateCompletion(t *testing.T) {
	doc := tasks.Document{
		Frontmatter: tasks.Frontmatter{Spec: "my-feature"},
		Tasks: []tasks.Task{
			{ID: "**1.**", Title: "Setup database schema", Status: "done", Wave: 1, Files: "schema.sql, migrate.go"},
			{ID: "**2.**", Title: "Add API endpoint", Status: "done", Wave: 1, Files: "handler.go"},
			{ID: "**3.**", Title: "Write tests", Status: "done", Wave: 2, Files: "handler_test.go"},
		},
	}
	waves := []wave.WaveState{
		{WaveNum: 1, Status: "complete", TaskIDs: []string{"1", "2"}, ExecRuns: 1, CheckRuns: 1},
		{WaveNum: 2, Status: "complete", TaskIDs: []string{"3"}, ExecRuns: 1, CheckRuns: 1},
	}

	cs, err := GenerateCompletion("", doc, waves, nil)
	if err != nil {
		t.Fatalf("GenerateCompletion error: %v", err)
	}

	if cs.Frontmatter.Spec != "my-feature" {
		t.Errorf("spec = %q, want my-feature", cs.Frontmatter.Spec)
	}
	if cs.Frontmatter.Status != "completed" {
		t.Errorf("status = %q, want completed", cs.Frontmatter.Status)
	}
	if cs.Frontmatter.TasksCount != 3 {
		t.Errorf("tasks_count = %d, want 3", cs.Frontmatter.TasksCount)
	}
	if cs.Frontmatter.WavesCount != 2 {
		t.Errorf("waves_count = %d, want 2", cs.Frontmatter.WavesCount)
	}
	if cs.Frontmatter.CheckpointPasses != 2 {
		t.Errorf("checkpoint_passes = %d, want 2", cs.Frontmatter.CheckpointPasses)
	}
	if cs.Frontmatter.CheckpointFailures != 0 {
		t.Errorf("checkpoint_failures = %d, want 0", cs.Frontmatter.CheckpointFailures)
	}

	// Verify body sections.
	if !strings.Contains(cs.Body, "# Completion Summary: my-feature") {
		t.Error("body should contain completion header")
	}
	if !strings.Contains(cs.Body, "## Tasks Completed") {
		t.Error("body should contain Tasks Completed section")
	}
	if !strings.Contains(cs.Body, "## Wave History") {
		t.Error("body should contain Wave History section")
	}
	if !strings.Contains(cs.Body, "## Metrics") {
		t.Error("body should contain Metrics section")
	}

	// Verify frontmatter technologies.
	if len(cs.Frontmatter.Technologies) == 0 {
		t.Error("technologies should be inferred from files")
	}

	// RenderFull should produce valid output.
	rendered, err := RenderFull(cs.Frontmatter, cs.Body)
	if err != nil {
		t.Fatalf("RenderFull error: %v", err)
	}
	if !strings.HasPrefix(rendered, "---\n") {
		t.Error("rendered should start with ---")
	}
}

func TestGenerateProgress(t *testing.T) {
	doc := &tasks.Document{
		Frontmatter: tasks.Frontmatter{Spec: "my-feature"},
		Tasks: []tasks.Task{
			{ID: "**1.**", Title: "Task one", Status: "done", Wave: 1, Files: "a.go"},
			{ID: "**2.**", Title: "Task two", Status: "in_progress", Wave: 2},
			{ID: "**3.**", Title: "Task three", Status: "pending", Wave: 2},
		},
	}
	waves := []wave.WaveState{
		{WaveNum: 1, Status: "complete", TaskIDs: []string{"1"}, ExecRuns: 1, CheckRuns: 1},
		{WaveNum: 2, Status: "in_progress", TaskIDs: []string{"2", "3"}, ExecRuns: 1},
	}

	ps, err := GenerateProgress("", "execution", doc, waves)
	if err != nil {
		t.Fatalf("GenerateProgress error: %v", err)
	}

	if ps.Frontmatter.Spec != "my-feature" {
		t.Errorf("spec = %q, want my-feature", ps.Frontmatter.Spec)
	}
	if ps.Frontmatter.Status != "in_progress" {
		t.Errorf("status = %q, want in_progress", ps.Frontmatter.Status)
	}
	if ps.Frontmatter.Stage != "execution" {
		t.Errorf("stage = %q, want execution", ps.Frontmatter.Stage)
	}
	if ps.Frontmatter.TasksDone != 1 {
		t.Errorf("tasks_done = %d, want 1", ps.Frontmatter.TasksDone)
	}
	if ps.Frontmatter.TasksTotal != 3 {
		t.Errorf("tasks_total = %d, want 3", ps.Frontmatter.TasksTotal)
	}
	if ps.Frontmatter.TasksPending != 1 {
		t.Errorf("tasks_pending = %d, want 1", ps.Frontmatter.TasksPending)
	}
	if ps.Frontmatter.TasksInProgress != 1 {
		t.Errorf("tasks_in_progress = %d, want 1", ps.Frontmatter.TasksInProgress)
	}
	if ps.Frontmatter.CurrentWave != 2 {
		t.Errorf("current_wave = %d, want 2", ps.Frontmatter.CurrentWave)
	}
	if ps.Frontmatter.WavesTotal != 2 {
		t.Errorf("waves_total = %d, want 2", ps.Frontmatter.WavesTotal)
	}

	// Body should contain sections.
	if !strings.Contains(ps.Body, "# Progress Summary: my-feature") {
		t.Error("body should contain progress header")
	}
	if !strings.Contains(ps.Body, "## Task Status") {
		t.Error("body should contain Task Status section")
	}
	if !strings.Contains(ps.Body, "## Wave Status") {
		t.Error("body should contain Wave Status section")
	}

	// RenderFull should produce valid output.
	rendered, err := RenderFull(ps.Frontmatter, ps.Body)
	if err != nil {
		t.Fatalf("RenderFull error: %v", err)
	}
	if !strings.HasPrefix(rendered, "---\n") {
		t.Error("rendered should start with ---")
	}
}

func TestGenerateProgressEmpty(t *testing.T) {
	doc := &tasks.Document{
		Frontmatter: tasks.Frontmatter{Spec: "empty-spec"},
	}

	ps, err := GenerateProgress("", "planning", doc, nil)
	if err != nil {
		t.Fatalf("GenerateProgress error: %v", err)
	}

	if ps.Frontmatter.TasksTotal != 0 {
		t.Errorf("tasks_total = %d, want 0", ps.Frontmatter.TasksTotal)
	}
	if ps.Frontmatter.WavesTotal != 0 {
		t.Errorf("waves_total = %d, want 0", ps.Frontmatter.WavesTotal)
	}
	// Should not contain Wave Status section with no waves.
	if strings.Contains(ps.Body, "## Wave Status") {
		t.Error("body should not contain Wave Status with no waves")
	}
}

func TestGenerateCompletionSingleTask(t *testing.T) {
	doc := tasks.Document{
		Frontmatter: tasks.Frontmatter{Spec: "single"},
		Tasks: []tasks.Task{
			{ID: "**1.**", Title: "Only task", Status: "done", Wave: 1, Files: "main.go"},
		},
	}
	waves := []wave.WaveState{
		{WaveNum: 1, Status: "complete", TaskIDs: []string{"1"}, ExecRuns: 1, CheckRuns: 1},
	}

	cs, err := GenerateCompletion("", doc, waves, nil)
	if err != nil {
		t.Fatalf("GenerateCompletion error: %v", err)
	}
	if cs.Frontmatter.TasksCount != 1 {
		t.Errorf("tasks_count = %d, want 1", cs.Frontmatter.TasksCount)
	}
	if cs.Frontmatter.WavesCount != 1 {
		t.Errorf("waves_count = %d, want 1", cs.Frontmatter.WavesCount)
	}
}

func TestGenerateCompletionNoWaves(t *testing.T) {
	doc := tasks.Document{
		Frontmatter: tasks.Frontmatter{Spec: "no-waves"},
		Tasks: []tasks.Task{
			{ID: "**1.**", Title: "A task", Status: "done", Wave: 0},
		},
	}

	cs, err := GenerateCompletion("", doc, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompletion error: %v", err)
	}
	if cs.Frontmatter.WavesCount != 0 {
		t.Errorf("waves_count = %d, want 0", cs.Frontmatter.WavesCount)
	}
}
