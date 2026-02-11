package tasks

import (
	"strings"
)

// ScoreComplexity returns a complexity score and model routing hint for a task.
// Factors: number of files, dependency count, TDD requirement, deferred status.
// Low (1-3) -> haiku, Medium (4-6) -> sonnet, High (7+) -> opus
func ScoreComplexity(task Task) ComplexityResult {
	score := 0
	var factors []string

	// Factor 1: Number of files (count backtick-delimited entries)
	fileCount := countFiles(task.Files)
	switch {
	case fileCount >= 5:
		score += 3
		factors = append(factors, "many files (5+)")
	case fileCount >= 3:
		score += 2
		factors = append(factors, "moderate files (3-4)")
	case fileCount >= 1:
		score += 1
		factors = append(factors, "few files (1-2)")
	}

	// Factor 2: Dependency count
	depCount := len(task.DependsOn)
	switch {
	case depCount >= 3:
		score += 3
		factors = append(factors, "many dependencies (3+)")
	case depCount >= 2:
		score += 2
		factors = append(factors, "moderate dependencies (2)")
	case depCount == 1:
		score += 1
		factors = append(factors, "single dependency")
	}

	// Factor 3: TDD requirement
	tddLower := strings.ToLower(strings.TrimSpace(task.TDD))
	if tddLower == "yes" || tddLower == "true" || tddLower == "required" {
		score += 2
		factors = append(factors, "TDD required")
	}

	// Factor 4: Deferred status (deferred tasks are often more complex/risky)
	if task.IsDeferred {
		score += 1
		factors = append(factors, "deferred task")
	}

	// Ensure minimum score of 1
	if score < 1 {
		score = 1
	}

	return ComplexityResult{
		TaskID:    task.ID,
		Score:     score,
		ModelHint: modelHint(score),
		Factors:   factors,
	}
}

// modelHint maps a numeric score to a model routing hint.
func modelHint(score int) string {
	switch {
	case score >= 7:
		return "opus"
	case score >= 4:
		return "sonnet"
	default:
		return "haiku"
	}
}

// countFiles counts backtick-delimited file entries in a Files metadata string.
// E.g., "`a.ts`, `b.ts`" -> 2
func countFiles(files string) int {
	if files == "" {
		return 0
	}
	count := 0
	inBacktick := false
	for _, c := range files {
		if c == '`' {
			if inBacktick {
				count++ // closing backtick = one file counted
			}
			inBacktick = !inBacktick
		}
	}
	// Fallback: if no backticks, count comma-separated entries
	if count == 0 && strings.TrimSpace(files) != "" {
		parts := strings.Split(files, ",")
		for _, p := range parts {
			if strings.TrimSpace(p) != "" {
				count++
			}
		}
	}
	return count
}
