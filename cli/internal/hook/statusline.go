package hook

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lucas-stellet/spw/internal/config"
	"github.com/lucas-stellet/spw/internal/git"
	"github.com/lucas-stellet/spw/internal/workspace"
)

// HandleStatusline outputs the Claude Code status line.
// Format: Model | Task | Dir | spec:name | 25.3k $0.42 | Context%
func HandleStatusline() error {
	p := workspace.ReadStdinPayload()
	dir := ""
	if p.Workspace != nil {
		dir = p.Workspace.CurrentDir
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}

	modelName := "Claude"
	if p.Model != nil {
		if p.Model.DisplayName != "" {
			modelName = p.Model.DisplayName
		} else if p.Model.Name != "" {
			modelName = p.Model.Name
		}
	}

	// Context window bar
	ctx := formatContextBar(p)

	// Current task from session todos
	task := detectCurrentTask(p.SessionID)

	dirname := filepath.Base(dir)
	spec := detectActiveSpec(dir)

	specLabel := ""
	if spec != "" {
		specLabel = " │ \x1b[2mspec:" + spec + "\x1b[0m"
	}

	// Token cost segment
	showTokenCost := "auto"
	repoRoot := git.RepoRoot(dir)
	if repoRoot != "" {
		if fullCfg, err := loadConfigForRoot(repoRoot); err == nil {
			showTokenCost = fullCfg.Statusline.ShowTokenCost
		}
	}
	tokenCost := formatTokenCost(p, showTokenCost)

	if task != "" {
		fmt.Printf("\x1b[2m%s\x1b[0m │ \x1b[1m%s\x1b[0m │ \x1b[2m%s\x1b[0m%s%s%s", modelName, task, dirname, specLabel, tokenCost, ctx)
	} else {
		fmt.Printf("\x1b[2m%s\x1b[0m │ \x1b[2m%s\x1b[0m%s%s%s", modelName, dirname, specLabel, tokenCost, ctx)
	}

	return nil
}

func formatContextBar(p workspace.Payload) string {
	if p.ContextWindow == nil || p.ContextWindow.RemainingPercentage == nil {
		return ""
	}

	remaining := math.Round(*p.ContextWindow.RemainingPercentage)
	rawUsed := math.Max(0, math.Min(100, 100-remaining))
	used := math.Min(100, math.Round((rawUsed/80)*100))

	filled := int(used / 10)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)

	usedInt := int(used)
	switch {
	case usedInt < 63:
		return fmt.Sprintf(" \x1b[32m%s %d%%\x1b[0m", bar, usedInt)
	case usedInt < 81:
		return fmt.Sprintf(" \x1b[33m%s %d%%\x1b[0m", bar, usedInt)
	case usedInt < 95:
		return fmt.Sprintf(" \x1b[38;5;208m%s %d%%\x1b[0m", bar, usedInt)
	default:
		return fmt.Sprintf(" \x1b[5;31m💀 %s %d%%\x1b[0m", bar, usedInt)
	}
}

func detectCurrentTask(sessionID string) string {
	if sessionID == "" {
		return ""
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	todosDir := filepath.Join(homeDir, ".claude", "todos")

	entries, err := os.ReadDir(todosDir)
	if err != nil {
		return ""
	}

	type fileInfo struct {
		name  string
		mtime int64
	}
	var matches []fileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, sessionID) || !strings.Contains(name, "-agent-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		matches = append(matches, fileInfo{name: name, mtime: info.ModTime().UnixMilli()})
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].mtime > matches[j].mtime
	})

	if len(matches) == 0 {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(todosDir, matches[0].name))
	if err != nil {
		return ""
	}

	var todos []struct {
		Status     string `json:"status"`
		ActiveForm string `json:"activeForm"`
		Content    string `json:"content"`
	}
	if err := json.Unmarshal(data, &todos); err != nil {
		return ""
	}

	for _, t := range todos {
		if t.Status == "in_progress" {
			if t.ActiveForm != "" {
				return t.ActiveForm
			}
			return t.Content
		}
	}

	return ""
}

// formatTokens returns a compact display of token count: 847, 25.3k, 1.2M.
func formatTokens(count int64) string {
	switch {
	case count >= 1_000_000:
		formatted := fmt.Sprintf("%.1f", float64(count)/1_000_000)
		formatted = strings.TrimSuffix(formatted, ".0")
		return formatted + "M"
	case count >= 1_000:
		formatted := fmt.Sprintf("%.1f", float64(count)/1_000)
		formatted = strings.TrimSuffix(formatted, ".0")
		return formatted + "k"
	default:
		return fmt.Sprintf("%d", count)
	}
}

// formatTokenCost builds the token/cost status segment.
// Returns "" when nothing should be shown.
func formatTokenCost(p workspace.Payload, mode string) string {
	if mode == "never" {
		return ""
	}

	var totalTokens int64
	hasTokens := false
	if p.ContextWindow != nil {
		if p.ContextWindow.TotalInputTokens != nil {
			totalTokens += *p.ContextWindow.TotalInputTokens
			hasTokens = true
		}
		if p.ContextWindow.TotalOutputTokens != nil {
			totalTokens += *p.ContextWindow.TotalOutputTokens
			hasTokens = true
		}
	}

	var cost float64
	hasCost := false
	if p.Cost != nil && p.Cost.TotalCostUSD != nil {
		cost = *p.Cost.TotalCostUSD
		hasCost = true
	}

	if mode == "auto" {
		if !hasCost || cost <= 0 {
			return ""
		}
	}

	// "always" mode: show if any data present
	if !hasTokens && !hasCost {
		return ""
	}

	var parts []string
	if hasTokens {
		parts = append(parts, formatTokens(totalTokens))
	}
	if hasCost {
		parts = append(parts, fmt.Sprintf("$%.2f", cost))
	}

	return " │ \x1b[2m" + strings.Join(parts, " ") + "\x1b[0m"
}

var specPathRe = regexp.MustCompile(`\.spec-workflow/specs/([^/]+)/`)

func detectActiveSpec(dir string) string {
	repoRoot := git.RepoRoot(dir)
	specsRoot := filepath.Join(dir, ".spec-workflow", "specs")
	if repoRoot != "" {
		specsRoot = filepath.Join(repoRoot, ".spec-workflow", "specs")
	}

	if _, err := os.Stat(specsRoot); os.IsNotExist(err) {
		return ""
	}

	if repoRoot == "" {
		return detectSpecByMtime(specsRoot)
	}

	cfg, _ := loadStatuslineConfig(repoRoot)

	cacheFile := filepath.Join(repoRoot, ".spec-workflow", ".spw-cache", "statusline.json")
	if cfg.stickySpec {
		cached := readStatuslineCache(repoRoot, cfg.cacheTTLSeconds, true)
		if cached != "" && specExists(specsRoot, cached) {
			return cached
		}
		if cached != "" {
			clearStatuslineCache(repoRoot)
		}
	} else {
		_ = cacheFile // suppress unused
		cached := readStatuslineCache(repoRoot, cfg.cacheTTLSeconds, false)
		if cached != "" && specExists(specsRoot, cached) {
			return cached
		}
		if cached != "" {
			clearStatuslineCache(repoRoot)
		}
	}

	specFromGit := detectSpecFromGit(repoRoot, cfg.baseBranches)
	if specFromGit != "" {
		writeStatuslineCache(repoRoot, specFromGit, map[string]string{
			"source": "git",
		})
		return specFromGit
	}

	specByMtime := detectSpecByMtime(specsRoot)
	if specByMtime != "" {
		writeStatuslineCache(repoRoot, specByMtime, map[string]string{
			"source": "mtime",
		})
	}
	return specByMtime
}

type statuslineConfig struct {
	baseBranches   []string
	cacheTTLSeconds int
	stickySpec     bool
}

func loadStatuslineConfig(repoRoot string) (statuslineConfig, error) {
	cfg := statuslineConfig{
		baseBranches:   []string{"main", "master", "staging", "develop"},
		cacheTTLSeconds: 10,
		stickySpec:     false,
	}

	// Use the full config loader
	fullCfg, err := loadConfigForRoot(repoRoot)
	if err != nil {
		return cfg, err
	}

	cfg.baseBranches = fullCfg.Statusline.BaseBranches
	cfg.cacheTTLSeconds = fullCfg.Statusline.CacheTTLSeconds
	cfg.stickySpec = fullCfg.Statusline.StickySpec

	return cfg, nil
}

func loadConfigForRoot(root string) (configResult, error) {
	from, err := configLoad(root)
	return from, err
}

// configResult wraps the config for statusline use.
type configResult = config.Config

var configLoad = config.Load

func specExists(specsRoot, specName string) bool {
	if specName == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(specsRoot, specName))
	return err == nil && info.IsDir()
}

func detectSpecFromGit(repoRoot string, baseBranches []string) string {
	baseRef := git.DetectBaseRef(repoRoot, baseBranches)
	if baseRef == "" {
		return ""
	}

	files := git.DiffNameOnly(repoRoot, baseRef)
	if len(files) == 0 {
		return ""
	}

	type candidate struct {
		score int
		idx   int
	}
	candidates := make(map[string]candidate)

	dashboardRe := regexp.MustCompile(`/(requirements|design|tasks)\.md$`)
	artifactRe := regexp.MustCompile(`/(DESIGN-RESEARCH|TASKS-CHECK|PRD)\.md$`)

	for idx, file := range files {
		m := specPathRe.FindStringSubmatch(file)
		if m == nil {
			continue
		}

		name := m[1]
		score := 1
		if dashboardRe.MatchString(file) {
			score = 3
		} else if artifactRe.MatchString(file) {
			score = 2
		}

		prev, ok := candidates[name]
		if !ok || score > prev.score || (score == prev.score && idx < prev.idx) {
			candidates[name] = candidate{score: score, idx: idx}
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	var bestName string
	var best candidate
	for name, c := range candidates {
		if bestName == "" || c.score > best.score || (c.score == best.score && c.idx < best.idx) {
			bestName = name
			best = c
		}
	}

	return bestName
}

func detectSpecByMtime(specsRoot string) string {
	entries, err := os.ReadDir(specsRoot)
	if err != nil {
		return ""
	}

	type specEntry struct {
		name  string
		mtime int64
	}
	var latest *specEntry

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		specDir := filepath.Join(specsRoot, e.Name())
		var maxMtime int64
		for _, f := range []string{"requirements.md", "design.md", "tasks.md"} {
			info, err := os.Stat(filepath.Join(specDir, f))
			if err != nil {
				continue
			}
			mt := info.ModTime().UnixMilli()
			if mt > maxMtime {
				maxMtime = mt
			}
		}
		if maxMtime == 0 {
			continue
		}
		if latest == nil || maxMtime > latest.mtime {
			latest = &specEntry{name: e.Name(), mtime: maxMtime}
		}
	}

	if latest != nil {
		return latest.name
	}
	return ""
}
