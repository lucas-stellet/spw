package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var approvalFileRe = regexp.MustCompile(`(?i)^approval_.*\.json$`)

// ApprovalFallbackID finds the latest local approval ID for a spec document.
func ApprovalFallbackID(cwd, specName, docType string, raw bool) {
	if specName == "" || docType == "" {
		Fail("approval-fallback-id requires <spec-name> <doc-type>", raw)
	}

	targetRel := targetDocPath(specName, docType)
	if targetRel == "" {
		Fail("doc-type must be one of: requirements|design|tasks", raw)
	}

	approvalsDir := filepath.Join(cwd, ".spec-workflow", "approvals", specName)
	entries, err := os.ReadDir(approvalsDir)
	if err != nil {
		result := map[string]any{"ok": true, "spec": specName, "doc_type": docType, "approval_id": nil, "source": nil}
		Output(result, "", raw)
		return
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
		result := map[string]any{
			"ok":          true,
			"spec":        specName,
			"doc_type":    docType,
			"approval_id": approvalID,
			"source":      rel,
		}
		Output(result, approvalID, raw)
		return
	}

	result := map[string]any{"ok": true, "spec": specName, "doc_type": docType, "approval_id": nil, "source": nil}
	Output(result, "", raw)
}

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
