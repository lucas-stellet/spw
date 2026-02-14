package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// iterationState is the persistent state for audit iteration tracking.
type iterationState struct {
	Type             string           `json:"type"`
	CurrentIteration int              `json:"current_iteration"`
	MaxIterations    int              `json:"max_iterations"`
	History          []iterationEntry `json:"history"`
}

// iterationEntry records a single iteration attempt.
type iterationEntry struct {
	Iteration int    `json:"iteration"`
	StartedAt string `json:"started_at"`
	Result    string `json:"result,omitempty"`
}

// iterationStatePath returns the canonical path to the iteration state file.
func iterationStatePath(runDir, auditType string) string {
	dir := "_inline-audit"
	if auditType == "inline-checkpoint" {
		dir = "_inline-checkpoint"
	}
	return filepath.Join(runDir, dir, "_iteration-state.json")
}

// --- Start ---

type auditIterationStartResult struct {
	OK        bool   `json:"ok"`
	Iteration int    `json:"iteration"`
	Max       int    `json:"max"`
	Remaining int    `json:"remaining"`
	Allowed   bool   `json:"allowed"`
	StatePath string `json:"state_path"`
}

func auditIterationStartCore(cwd, runDir, auditType string, max int) (*auditIterationStartResult, error) {
	if runDir == "" {
		return nil, fmt.Errorf("audit-iteration start requires <run-dir>")
	}
	if auditType == "" {
		return nil, fmt.Errorf("audit-iteration start requires <audit-type>")
	}
	if auditType != "inline-audit" && auditType != "inline-checkpoint" {
		return nil, fmt.Errorf("audit-type must be 'inline-audit' or 'inline-checkpoint', got %q", auditType)
	}

	if max <= 0 {
		max = 3
	}

	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	statePath := iterationStatePath(runDir, auditType)
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create state dir: %w", err)
	}

	state := iterationState{
		Type:             auditType,
		CurrentIteration: 1,
		MaxIterations:    max,
		History: []iterationEntry{
			{
				Iteration: 1,
				StartedAt: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write state file: %w", err)
	}

	relPath, _ := filepath.Rel(cwd, statePath)

	return &auditIterationStartResult{
		OK:        true,
		Iteration: 1,
		Max:       max,
		Remaining: max - 1,
		Allowed:   true,
		StatePath: relPath,
	}, nil
}

// AuditIterationStart initializes iteration tracking for an audit.
func AuditIterationStart(cwd, runDir, auditType string, max int, raw bool) {
	result, err := auditIterationStartCore(cwd, runDir, auditType, max)
	if err != nil {
		Fail(err.Error(), raw)
	}
	Output(result, result.StatePath, raw)
}

// --- Check ---

type auditIterationCheckResult struct {
	OK        bool   `json:"ok"`
	Allowed   bool   `json:"allowed"`
	Iteration int    `json:"iteration"`
	Max       int    `json:"max"`
	Remaining int    `json:"remaining"`
	Message   string `json:"message"`
}

func auditIterationCheckCore(cwd, runDir, auditType string) (*auditIterationCheckResult, error) {
	if runDir == "" {
		return nil, fmt.Errorf("audit-iteration check requires <run-dir>")
	}
	if auditType == "" {
		return nil, fmt.Errorf("audit-iteration check requires <audit-type>")
	}
	if auditType != "inline-audit" && auditType != "inline-checkpoint" {
		return nil, fmt.Errorf("audit-type must be 'inline-audit' or 'inline-checkpoint', got %q", auditType)
	}

	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	statePath := iterationStatePath(runDir, auditType)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("no iteration state found, run audit-iteration start first")
	}

	var state iterationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("invalid iteration state: %w", err)
	}

	allowed := state.CurrentIteration < state.MaxIterations
	remaining := state.MaxIterations - state.CurrentIteration
	if remaining < 0 {
		remaining = 0
	}

	message := "OK - YOU CAN TRY AGAIN"
	if !allowed {
		message = "BLOCKED - NO MORE TRIES, LET THE MAIN AGENT KNOW"
	}

	return &auditIterationCheckResult{
		OK:        true,
		Allowed:   allowed,
		Iteration: state.CurrentIteration,
		Max:       state.MaxIterations,
		Remaining: remaining,
		Message:   message,
	}, nil
}

// AuditIterationCheck checks whether another iteration is allowed.
func AuditIterationCheck(cwd, runDir, auditType string, raw bool) {
	result, err := auditIterationCheckCore(cwd, runDir, auditType)
	if err != nil {
		Fail(err.Error(), raw)
	}
	Output(result, result.Message, raw)
}

// --- Advance ---

type auditIterationAdvanceResult struct {
	OK          bool `json:"ok"`
	AdvancedTo  int  `json:"advanced_to"`
	Max         int  `json:"max"`
	Remaining   int  `json:"remaining"`
	NextAllowed bool `json:"next_allowed"`
}

func auditIterationAdvanceCore(cwd, runDir, auditType, result string) (*auditIterationAdvanceResult, error) {
	if runDir == "" {
		return nil, fmt.Errorf("audit-iteration advance requires <run-dir>")
	}
	if auditType == "" {
		return nil, fmt.Errorf("audit-iteration advance requires <audit-type>")
	}
	if auditType != "inline-audit" && auditType != "inline-checkpoint" {
		return nil, fmt.Errorf("audit-type must be 'inline-audit' or 'inline-checkpoint', got %q", auditType)
	}
	if result == "" {
		return nil, fmt.Errorf("audit-iteration advance requires <result>")
	}

	if !filepath.IsAbs(runDir) {
		runDir = filepath.Join(cwd, runDir)
	}

	statePath := iterationStatePath(runDir, auditType)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("no iteration state found, run audit-iteration start first")
	}

	var state iterationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("invalid iteration state: %w", err)
	}

	if state.CurrentIteration >= state.MaxIterations {
		return nil, fmt.Errorf("cannot advance past max iterations (%d)", state.MaxIterations)
	}

	// Record the result of the current (finishing) iteration in its history entry.
	for i := range state.History {
		if state.History[i].Iteration == state.CurrentIteration && state.History[i].Result == "" {
			state.History[i].Result = result
			break
		}
	}

	state.CurrentIteration++
	state.History = append(state.History, iterationEntry{
		Iteration: state.CurrentIteration,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	})

	newData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}
	if err := os.WriteFile(statePath, newData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write state file: %w", err)
	}

	remaining := state.MaxIterations - state.CurrentIteration
	if remaining < 0 {
		remaining = 0
	}
	nextAllowed := state.CurrentIteration < state.MaxIterations

	return &auditIterationAdvanceResult{
		OK:          true,
		AdvancedTo:  state.CurrentIteration,
		Max:         state.MaxIterations,
		Remaining:   remaining,
		NextAllowed: nextAllowed,
	}, nil
}

// AuditIterationAdvance increments the iteration counter and records the result.
func AuditIterationAdvance(cwd, runDir, auditType, result string, raw bool) {
	res, err := auditIterationAdvanceCore(cwd, runDir, auditType, result)
	if err != nil {
		Fail(err.Error(), raw)
	}
	Output(res, fmt.Sprintf("%d", res.AdvancedTo), raw)
}
