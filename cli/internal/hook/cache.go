package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// statuslineCacheEntry is the JSON structure written to .oraculo-cache/statusline.json.
type statuslineCacheEntry struct {
	Timestamp int64             `json:"ts"`
	Spec      string            `json:"spec"`
	Extra     map[string]string `json:"-"`
}

func (e statuslineCacheEntry) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"ts":   e.Timestamp,
		"spec": e.Spec,
	}
	for k, v := range e.Extra {
		m[k] = v
	}
	return json.MarshalIndent(m, "", "  ")
}

func writeStatuslineCache(workspaceRoot, spec string, meta map[string]string) bool {
	if spec == "" {
		return false
	}
	cacheDir := filepath.Join(workspaceRoot, ".spec-workflow", ".oraculo-cache")
	cacheFile := filepath.Join(cacheDir, "statusline.json")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return false
	}

	entry := statuslineCacheEntry{
		Timestamp: time.Now().UnixMilli(),
		Spec:      spec,
		Extra:     meta,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return false
	}

	return os.WriteFile(cacheFile, data, 0644) == nil
}

func readStatuslineCache(workspaceRoot string, ttlSeconds int, ignoreTTL bool) string {
	cacheFile := filepath.Join(workspaceRoot, ".spec-workflow", ".oraculo-cache", "statusline.json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}

	var entry struct {
		Timestamp int64  `json:"ts"`
		Spec      string `json:"spec"`
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return ""
	}
	if entry.Spec == "" || entry.Timestamp == 0 {
		return ""
	}

	if ignoreTTL {
		return entry.Spec
	}

	age := time.Now().UnixMilli() - entry.Timestamp
	if age <= int64(ttlSeconds)*1000 {
		return entry.Spec
	}

	return ""
}

func clearStatuslineCache(workspaceRoot string) {
	cacheFile := filepath.Join(workspaceRoot, ".spec-workflow", ".oraculo-cache", "statusline.json")
	_ = os.Remove(cacheFile) // fail-open
}
