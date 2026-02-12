package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Merge reads a template TOML file and a user's existing TOML file,
// producing a merged output that preserves user values while adding
// new keys from the template. Comments are preserved from the template.
//
// Strategy:
// - Walk the template file line by line
// - For each key=value line, check if the user file has the same key in the same section
// - If yes, use the user's value; if no, use the template's value
// - New sections/keys in the template are appended
func Merge(templatePath, userPath, outputPath string) error {
	templateLines, err := readLines(templatePath)
	if err != nil {
		return fmt.Errorf("reading template: %w", err)
	}
	templateLines = collapseMultilineArrays(templateLines)

	userValues, err := parseKeyValues(userPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading user config: %w", err)
	}

	var output []string
	currentSection := ""

	for _, line := range templateLines {
		trimmed := strings.TrimSpace(line)

		// Track section headers
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentSection = trimmed[1 : len(trimmed)-1]
			output = append(output, line)
			continue
		}

		// Comments and blank lines pass through
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			output = append(output, line)
			continue
		}

		// Key = value lines
		if idx := strings.Index(trimmed, "="); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			qualifiedKey := currentSection + "." + key

			if userVal, ok := userValues[qualifiedKey]; ok {
				// Preserve leading whitespace from template
				indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				output = append(output, indent+key+" = "+userVal)
			} else {
				output = append(output, line)
			}
			continue
		}

		output = append(output, line)
	}

	content := strings.Join(output, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

// parseKeyValues reads a TOML file, normalizes it (collapsing multiline arrays
// into single lines), and returns a map of "section.key" → raw value string.
func parseKeyValues(filePath string) (map[string]string, error) {
	normalized, err := normalizeToml(filePath)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	currentSection := ""
	scanner := bufio.NewScanner(strings.NewReader(normalized))

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentSection = trimmed[1 : len(trimmed)-1]
			continue
		}

		if idx := strings.Index(trimmed, "="); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			value := strings.TrimSpace(trimmed[idx+1:])
			qualifiedKey := currentSection + "." + key
			result[qualifiedKey] = value
		}
	}

	return result, scanner.Err()
}

// normalizeToml round-trips a TOML file through BurntSushi/toml to produce
// normalized output where all arrays are single-line. This ensures the
// line-by-line parseKeyValues works correctly with multiline arrays.
func normalizeToml(filePath string) (string, error) {
	var data interface{}
	if _, err := toml.DecodeFile(filePath, &data); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// collapseMultilineArrays joins multiline TOML arrays into single lines
// so the template can be walked line-by-line. Comments between array elements
// are discarded; only the array content is preserved.
func collapseMultilineArrays(lines []string) []string {
	var result []string
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// Detect a key = [ without closing ] on the same line
		if idx := strings.Index(trimmed, "="); idx > 0 {
			value := strings.TrimSpace(trimmed[idx+1:])
			if strings.HasPrefix(value, "[") && !strings.Contains(value, "]") {
				// Accumulate until we find the closing ]
				collected := lines[i]
				for i++; i < len(lines); i++ {
					innerTrimmed := strings.TrimSpace(lines[i])
					// Skip comment-only lines inside the array
					if strings.HasPrefix(innerTrimmed, "#") {
						continue
					}
					collected += " " + innerTrimmed
					if strings.Contains(innerTrimmed, "]") {
						break
					}
				}
				// Clean up the collapsed line: normalize whitespace
				result = append(result, collapseSpaces(collected))
				continue
			}
		}

		result = append(result, lines[i])
	}
	return result
}

// collapseSpaces normalizes a collapsed array line by removing excess whitespace.
func collapseSpaces(s string) string {
	// Split around the = to preserve key formatting
	idx := strings.Index(s, "=")
	if idx < 0 {
		return s
	}
	key := s[:idx+1]
	val := strings.TrimSpace(s[idx+1:])

	// Normalize spaces within the array value
	var buf strings.Builder
	prevSpace := false
	for _, r := range val {
		if r == ' ' || r == '\t' {
			if !prevSpace {
				buf.WriteRune(' ')
			}
			prevSpace = true
		} else {
			buf.WriteRune(r)
			prevSpace = false
		}
	}
	return key + " " + buf.String()
}

func readLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
