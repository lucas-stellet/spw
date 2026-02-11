package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var approvalFileRe = regexp.MustCompile(`(?i)^approval_.*\.json$`)

// CheckApproval checks filesystem for local approval records for a document.
// Scans .spec-workflow/approvals/<spec>/ for approval_*.json files.
func CheckApproval(cwd, specName, docType string) ApprovalResult {
	result := ApprovalResult{
		DocType: docType,
		Found:   false,
	}

	targetRel := targetDocPath(specName, docType)
	if targetRel == "" {
		return result
	}

	approvalsDir := filepath.Join(cwd, ".spec-workflow", "approvals", specName)
	entries, err := os.ReadDir(approvalsDir)
	if err != nil {
		return result
	}

	type fileEntry struct {
		name  string
		full  string
		mtime int64
	}
	var files []fileEntry
	for _, e := range entries {
		if e.IsDir() || !approvalFileRe.MatchString(e.Name()) {
			continue
		}
		full := filepath.Join(approvalsDir, e.Name())
		info, err := os.Stat(full)
		if err != nil {
			continue
		}
		files = append(files, fileEntry{name: e.Name(), full: full, mtime: info.ModTime().UnixMilli()})
	}

	// Sort by modification time, newest first
	sort.Slice(files, func(i, j int) bool { return files[i].mtime > files[j].mtime })

	targetNorm := strings.ReplaceAll(targetRel, "\\", "/")

	for _, f := range files {
		data, err := os.ReadFile(f.full)
		if err != nil {
			continue
		}

		var doc map[string]any
		if err := json.Unmarshal(data, &doc); err != nil {
			continue
		}

		filePath := strings.ReplaceAll(getStr(doc, "filePath", "path"), "\\", "/")
		matches := strings.HasSuffix(filePath, targetNorm) ||
			strings.HasSuffix(filePath, "/"+filepath.Base(targetNorm)) ||
			filePath == targetNorm

		if !matches {
			continue
		}

		approvalID := getStr(doc, "approvalId", "id")
		if approvalID == "" {
			// Try nested approval.id
			if approval, ok := doc["approval"].(map[string]any); ok {
				if id, ok := approval["id"].(string); ok {
					approvalID = id
				}
			}
		}
		if approvalID == "" {
			continue
		}

		rel, _ := filepath.Rel(cwd, f.full)
		result.Found = true
		result.ApprovalID = approvalID
		result.Source = rel
		return result
	}

	return result
}

// targetDocPath returns the relative path for a document type within a spec.
func targetDocPath(specName, docType string) string {
	switch strings.ToLower(docType) {
	case "requirements":
		return filepath.Join(".spec-workflow", "specs", specName, "requirements.md")
	case "design":
		return filepath.Join(".spec-workflow", "specs", specName, "design.md")
	case "tasks":
		return filepath.Join(".spec-workflow", "specs", specName, "tasks.md")
	default:
		return ""
	}
}

// getStr extracts a string value from a map, trying multiple keys.
func getStr(doc map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := doc[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
