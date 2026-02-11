package tasks

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Task line: - [ ] 1 Title here  OR  - [x] 2.1 Title here  OR  - [-] 3 Title
	taskLineRe = regexp.MustCompile(`^- \[([ x\-])\] (\d+(?:\.\d+)?)\s+(.*)$`)

	// Metadata lines (indented under a task)
	waveMeta    = regexp.MustCompile(`(?i)^\s+Wave:\s*(\d+)`)
	dependsMeta = regexp.MustCompile(`(?i)^\s+Depends\s+On:\s*(.+)`)
	filesMeta   = regexp.MustCompile(`(?i)^\s+Files:\s*(.+)`)
	tddMeta     = regexp.MustCompile(`(?i)^\s+TDD:\s*(.+)`)

	// Wave plan line: - Wave 1: Tasks 1, 2, 3
	wavePlanRe = regexp.MustCompile(`(?i)^-\s+Wave\s+(\d+):\s*Tasks?\s+(.+)`)

	// Frontmatter field patterns (regex-based, no YAML library)
	fmSpec       = regexp.MustCompile(`(?i)^spec:\s*(.+)`)
	fmTaskIDs    = regexp.MustCompile(`(?i)^task_ids:\s*\[([^\]]*)\]`)
	fmApproval   = regexp.MustCompile(`(?i)^approval_id:\s*(.+)`)
	fmStrategy   = regexp.MustCompile(`(?i)^generation_strategy:\s*(.+)`)

	// Task ID references: "Task 4", "4", etc.
	taskIDRefRe = regexp.MustCompile(`\d+(?:\.\d+)?`)
)

// ParseFile reads and parses a tasks.md file.
func ParseFile(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("cannot read tasks.md: %w", err)
	}
	return Parse(string(data)), nil
}

// Parse parses the content of a tasks.md file.
// Body scan is authoritative — tasks found in the body but not in
// frontmatter task_ids are flagged as deferred with a mismatch warning.
func Parse(content string) Document {
	lines := strings.Split(content, "\n")
	doc := Document{}

	// Pass 1: Extract frontmatter
	fmStart, fmEnd := findFrontmatter(lines)
	if fmStart >= 0 && fmEnd > fmStart {
		doc.Frontmatter = parseFrontmatter(lines[fmStart+1 : fmEnd])
	}

	// Pass 2: Parse body sections
	var (
		inConstraints bool
		inWavePlan    bool
		inDeferred    bool
		constraintBuf strings.Builder
		currentTask   *Task
	)

	// Build a set from frontmatter task_ids for quick lookup
	fmIDSet := make(map[string]bool, len(doc.Frontmatter.TaskIDs))
	for _, id := range doc.Frontmatter.TaskIDs {
		fmIDSet[id] = true
	}

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Section headers
		if strings.HasPrefix(trimmed, "## ") {
			flushTask(&doc, currentTask)
			currentTask = nil
			inConstraints = false
			inWavePlan = false

			lower := strings.ToLower(trimmed)
			switch {
			case strings.Contains(lower, "execution constraints"):
				inConstraints = true
			case strings.Contains(lower, "wave plan"):
				inWavePlan = true
			case strings.Contains(lower, "deferred"):
				inDeferred = true
			default:
				inDeferred = false
			}
			continue
		}

		// Constraints section: collect raw text
		if inConstraints {
			constraintBuf.WriteString(line)
			constraintBuf.WriteString("\n")
			continue
		}

		// Wave plan section
		if inWavePlan {
			if m := wavePlanRe.FindStringSubmatch(trimmed); m != nil {
				waveNum, _ := strconv.Atoi(m[1])
				ids := parseIDList(m[2])
				doc.WavePlan = append(doc.WavePlan, WavePlanEntry{Wave: waveNum, TaskIDs: ids})
			}
			continue
		}

		// Task lines
		if m := taskLineRe.FindStringSubmatch(line); m != nil {
			flushTask(&doc, currentTask)
			status := charToStatus(m[1])
			task := Task{
				ID:      m[2],
				Title:   strings.TrimSpace(m[3]),
				Status:  status,
				RawLine: lineNum + 1, // 1-based
			}
			if inDeferred {
				task.IsDeferred = true
				doc.HasDeferred = true
			}
			currentTask = &task
			continue
		}

		// Metadata lines (indented, belonging to current task)
		if currentTask != nil && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
			if m := waveMeta.FindStringSubmatch(line); m != nil {
				currentTask.Wave, _ = strconv.Atoi(m[1])
			} else if m := dependsMeta.FindStringSubmatch(line); m != nil {
				currentTask.DependsOn = parseIDList(m[1])
			} else if m := filesMeta.FindStringSubmatch(line); m != nil {
				currentTask.Files = strings.TrimSpace(m[1])
			} else if m := tddMeta.FindStringSubmatch(line); m != nil {
				currentTask.TDD = strings.TrimSpace(m[1])
			}
		}
	}

	// Flush last task
	flushTask(&doc, currentTask)

	doc.Constraints = strings.TrimSpace(constraintBuf.String())

	// Post-processing: detect task_ids mismatch
	if len(doc.Frontmatter.TaskIDs) > 0 {
		bodyIDs := make(map[string]bool, len(doc.Tasks))
		for i := range doc.Tasks {
			bodyIDs[doc.Tasks[i].ID] = true
		}

		// Tasks in body but NOT in frontmatter → deferred + warning
		for i := range doc.Tasks {
			if !fmIDSet[doc.Tasks[i].ID] {
				if !doc.Tasks[i].IsDeferred {
					doc.Tasks[i].IsDeferred = true
					doc.HasDeferred = true
				}
				doc.Warnings = append(doc.Warnings, fmt.Sprintf(
					"task_ids_mismatch: task %s found in body but not in frontmatter task_ids",
					doc.Tasks[i].ID,
				))
			}
		}

		// Frontmatter IDs not in body → warning
		for _, id := range doc.Frontmatter.TaskIDs {
			if !bodyIDs[id] {
				doc.Warnings = append(doc.Warnings, fmt.Sprintf(
					"task_ids_mismatch: frontmatter task_ids contains %s but no matching task in body",
					id,
				))
			}
		}
	}

	return doc
}

func findFrontmatter(lines []string) (start, end int) {
	start = -1
	end = -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if start < 0 {
				start = i
			} else {
				end = i
				return
			}
		}
	}
	return -1, -1
}

func parseFrontmatter(lines []string) Frontmatter {
	var fm Frontmatter
	for _, line := range lines {
		if m := fmSpec.FindStringSubmatch(line); m != nil {
			fm.Spec = strings.TrimSpace(m[1])
		} else if m := fmTaskIDs.FindStringSubmatch(line); m != nil {
			fm.TaskIDs = parseIDList(m[1])
		} else if m := fmApproval.FindStringSubmatch(line); m != nil {
			fm.ApprovalID = strings.TrimSpace(m[1])
		} else if m := fmStrategy.FindStringSubmatch(line); m != nil {
			fm.GenerationStrategy = strings.TrimSpace(m[1])
		}
	}
	return fm
}

func parseIDList(s string) []string {
	matches := taskIDRefRe.FindAllString(s, -1)
	var ids []string
	for _, m := range matches {
		ids = append(ids, strings.TrimSpace(m))
	}
	return ids
}

func charToStatus(c string) string {
	switch c {
	case "x":
		return "done"
	case "-":
		return "in_progress"
	default:
		return "pending"
	}
}

func flushTask(doc *Document, task *Task) {
	if task == nil {
		return
	}
	doc.Tasks = append(doc.Tasks, *task)
}

// TaskByID returns the task with the given ID, or nil if not found.
func (d *Document) TaskByID(id string) *Task {
	for i := range d.Tasks {
		if d.Tasks[i].ID == id {
			return &d.Tasks[i]
		}
	}
	return nil
}

// Count returns task count statistics.
func (d *Document) Count() CountResult {
	r := CountResult{Total: len(d.Tasks)}
	for _, t := range d.Tasks {
		switch t.Status {
		case "done":
			r.Done++
		case "in_progress":
			r.InProgress++
		default:
			r.Pending++
		}
		if t.IsDeferred {
			r.Deferred++
		}
	}
	return r
}
