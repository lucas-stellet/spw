package tasks

import (
	"regexp"
	"strings"
)

var (
	// checkboxLineRe matches any line with a checkbox marker
	checkboxLineRe = regexp.MustCompile(`^- \[([ x\-])\] `)

	// taskCheckboxRe matches a valid task line: checkbox + numeric ID
	taskCheckboxRe = regexp.MustCompile(`^- \[([ x\-])\] \d+(?:\.\d+)?\s`)

	// starListRe matches lines using * as list marker (forbidden)
	starListRe = regexp.MustCompile(`^\*\s`)

	// nestedCheckboxRe matches indented checkbox lines (nested checkboxes in metadata)
	nestedCheckboxRe = regexp.MustCompile(`^\s+- \[([ x\-])\] `)

	// multiLineFilesRe detects if a Files entry spans multiple lines
	// (we check if a line after "Files:" is indented and looks like a continuation)
	filesMetaRe = regexp.MustCompile(`(?i)^\s+Files:\s*$`)
)

// Validate checks tasks.md content against dashboard compatibility rules:
//   - Checkbox markers only on task lines with numeric IDs
//   - "-" as list marker (never "*")
//   - No nested checkboxes in metadata
//   - "Files" in single line
func Validate(content string) ValidateResult {
	lines := strings.Split(content, "\n")
	var errors []string

	for i, line := range lines {
		lineNum := i + 1

		// Rule 1: Checkbox markers only on task lines with numeric IDs
		if checkboxLineRe.MatchString(line) && !taskCheckboxRe.MatchString(line) {
			errors = append(errors, formatError(lineNum,
				"checkbox marker on non-task line (must have numeric ID after checkbox)"))
		}

		// Rule 2: "*" as list marker is forbidden
		trimmed := strings.TrimSpace(line)
		if starListRe.MatchString(trimmed) {
			errors = append(errors, formatError(lineNum,
				"use '-' as list marker, not '*'"))
		}

		// Rule 3: No nested checkboxes in metadata
		if nestedCheckboxRe.MatchString(line) {
			errors = append(errors, formatError(lineNum,
				"nested checkbox found in metadata (checkboxes only allowed on top-level task lines)"))
		}

		// Rule 4: "Files" must be on a single line (not empty/multiline)
		if filesMetaRe.MatchString(line) {
			errors = append(errors, formatError(lineNum,
				"Files metadata must be on a single line (found empty Files: with no value)"))
		}
	}

	return ValidateResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func formatError(lineNum int, msg string) string {
	return "line " + itoa(lineNum) + ": " + msg
}

// itoa is a simple int-to-string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
