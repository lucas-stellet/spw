package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

// parseKeyValues reads a TOML file and returns a map of "section.key" → raw value string.
func parseKeyValues(filePath string) (map[string]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	currentSection := ""
	scanner := bufio.NewScanner(f)

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
