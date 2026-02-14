package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditIterationStart(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	result, err := auditIterationStartCore(tmp, runDir, "inline-audit", 3)
	if err != nil {
		t.Fatal(err)
	}

	if !result.OK {
		t.Error("expected ok=true")
	}
	if result.Iteration != 1 {
		t.Errorf("iteration = %d, want 1", result.Iteration)
	}
	if result.Max != 3 {
		t.Errorf("max = %d, want 3", result.Max)
	}
	if result.Remaining != 2 {
		t.Errorf("remaining = %d, want 2", result.Remaining)
	}
	if !result.Allowed {
		t.Error("expected allowed=true")
	}

	// Verify state file was created with correct content
	statePath := iterationStatePath(runDir, "inline-audit")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state file not created: %v", err)
	}

	var state iterationState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("invalid state JSON: %v", err)
	}
	if state.Type != "inline-audit" {
		t.Errorf("state.Type = %q, want %q", state.Type, "inline-audit")
	}
	if state.CurrentIteration != 1 {
		t.Errorf("state.CurrentIteration = %d, want 1", state.CurrentIteration)
	}
	if state.MaxIterations != 3 {
		t.Errorf("state.MaxIterations = %d, want 3", state.MaxIterations)
	}
	if len(state.History) != 1 {
		t.Errorf("state.History length = %d, want 1", len(state.History))
	}
}

func TestAuditIterationStartDefaultMax(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	result, err := auditIterationStartCore(tmp, runDir, "inline-audit", 0)
	if err != nil {
		t.Fatal(err)
	}

	if result.Max != 3 {
		t.Errorf("max = %d, want 3 (default)", result.Max)
	}
}

func TestAuditIterationCheckAllowed(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	// Start with max=3 (iteration 1 of 3)
	_, err := auditIterationStartCore(tmp, runDir, "inline-audit", 3)
	if err != nil {
		t.Fatal(err)
	}

	result, err := auditIterationCheckCore(tmp, runDir, "inline-audit")
	if err != nil {
		t.Fatal(err)
	}

	if !result.Allowed {
		t.Error("expected allowed=true at iteration 1 of 3")
	}
	if result.Iteration != 1 {
		t.Errorf("iteration = %d, want 1", result.Iteration)
	}
	if result.Remaining != 2 {
		t.Errorf("remaining = %d, want 2", result.Remaining)
	}
	if result.Message != "OK - YOU CAN TRY AGAIN" {
		t.Errorf("message = %q, want %q", result.Message, "OK - YOU CAN TRY AGAIN")
	}
}

func TestAuditIterationCheckBlocked(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	// Create state file at iteration 3 of 3
	statePath := iterationStatePath(runDir, "inline-audit")
	os.MkdirAll(filepath.Dir(statePath), 0755)
	state := iterationState{
		Type:             "inline-audit",
		CurrentIteration: 3,
		MaxIterations:    3,
		History:          []iterationEntry{},
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(statePath, data, 0644)

	result, err := auditIterationCheckCore(tmp, runDir, "inline-audit")
	if err != nil {
		t.Fatal(err)
	}

	if result.Allowed {
		t.Error("expected allowed=false at iteration 3 of 3")
	}
	if result.Remaining != 0 {
		t.Errorf("remaining = %d, want 0", result.Remaining)
	}
	if result.Message != "BLOCKED - NO MORE TRIES, LET THE MAIN AGENT KNOW" {
		t.Errorf("message = %q, want blocked message", result.Message)
	}
}

func TestAuditIterationCheckNoState(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	_, err := auditIterationCheckCore(tmp, runDir, "inline-audit")
	if err == nil {
		t.Error("expected error when no state file exists")
	}
}

func TestAuditIterationAdvance(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	// Start with max=3
	_, err := auditIterationStartCore(tmp, runDir, "inline-audit", 3)
	if err != nil {
		t.Fatal(err)
	}

	result, err := auditIterationAdvanceCore(tmp, runDir, "inline-audit", "blocked")
	if err != nil {
		t.Fatal(err)
	}

	if result.AdvancedTo != 2 {
		t.Errorf("advanced_to = %d, want 2", result.AdvancedTo)
	}
	if result.Max != 3 {
		t.Errorf("max = %d, want 3", result.Max)
	}
	if result.Remaining != 1 {
		t.Errorf("remaining = %d, want 1", result.Remaining)
	}
	if !result.NextAllowed {
		t.Error("expected next_allowed=true")
	}

	// Verify state file was updated
	statePath := iterationStatePath(runDir, "inline-audit")
	data, _ := os.ReadFile(statePath)
	var state iterationState
	json.Unmarshal(data, &state)
	if state.CurrentIteration != 2 {
		t.Errorf("state.CurrentIteration = %d, want 2", state.CurrentIteration)
	}
	if len(state.History) != 2 {
		t.Fatalf("state.History length = %d, want 2", len(state.History))
	}

	// Verify the result is recorded on the correct history entry (iteration 1, not 2)
	if state.History[0].Iteration != 1 {
		t.Errorf("history[0].Iteration = %d, want 1", state.History[0].Iteration)
	}
	if state.History[0].Result != "blocked" {
		t.Errorf("history[0].Result = %q, want %q (result belongs to finishing iteration)", state.History[0].Result, "blocked")
	}
	if state.History[1].Iteration != 2 {
		t.Errorf("history[1].Iteration = %d, want 2", state.History[1].Iteration)
	}
	if state.History[1].Result != "" {
		t.Errorf("history[1].Result = %q, want empty (new iteration hasn't finished)", state.History[1].Result)
	}
}

func TestAuditIterationFullLifecycle(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	// Start: iteration 1 of 3
	startResult, err := auditIterationStartCore(tmp, runDir, "inline-audit", 3)
	if err != nil {
		t.Fatal(err)
	}
	if startResult.Iteration != 1 || startResult.Max != 3 {
		t.Fatalf("start: iteration=%d max=%d", startResult.Iteration, startResult.Max)
	}

	// Check: should be allowed (1 < 3)
	checkResult, err := auditIterationCheckCore(tmp, runDir, "inline-audit")
	if err != nil {
		t.Fatal(err)
	}
	if !checkResult.Allowed {
		t.Fatal("check after start: expected allowed=true")
	}

	// Advance to 2
	advResult, err := auditIterationAdvanceCore(tmp, runDir, "inline-audit", "blocked")
	if err != nil {
		t.Fatal(err)
	}
	if advResult.AdvancedTo != 2 {
		t.Fatalf("advance 1: advanced_to=%d, want 2", advResult.AdvancedTo)
	}
	if !advResult.NextAllowed {
		t.Fatal("advance 1: expected next_allowed=true")
	}

	// Check: should still be allowed (2 < 3)
	checkResult, err = auditIterationCheckCore(tmp, runDir, "inline-audit")
	if err != nil {
		t.Fatal(err)
	}
	if !checkResult.Allowed {
		t.Fatal("check after advance to 2: expected allowed=true")
	}
	if checkResult.Remaining != 1 {
		t.Errorf("check after advance to 2: remaining=%d, want 1", checkResult.Remaining)
	}

	// Advance to 3
	advResult, err = auditIterationAdvanceCore(tmp, runDir, "inline-audit", "blocked")
	if err != nil {
		t.Fatal(err)
	}
	if advResult.AdvancedTo != 3 {
		t.Fatalf("advance 2: advanced_to=%d, want 3", advResult.AdvancedTo)
	}
	if advResult.NextAllowed {
		t.Fatal("advance 2: expected next_allowed=false")
	}
	if advResult.Remaining != 0 {
		t.Errorf("advance 2: remaining=%d, want 0", advResult.Remaining)
	}

	// Check: should be blocked (3 >= 3)
	checkResult, err = auditIterationCheckCore(tmp, runDir, "inline-audit")
	if err != nil {
		t.Fatal(err)
	}
	if checkResult.Allowed {
		t.Fatal("check after advance to 3: expected allowed=false")
	}
	if checkResult.Message != "BLOCKED - NO MORE TRIES, LET THE MAIN AGENT KNOW" {
		t.Errorf("check after advance to 3: message=%q", checkResult.Message)
	}
}

func TestAuditIterationAdvancePastMax(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	// Start with max=1
	_, err := auditIterationStartCore(tmp, runDir, "inline-audit", 1)
	if err != nil {
		t.Fatal(err)
	}

	// Attempting to advance past max should return an error
	_, err = auditIterationAdvanceCore(tmp, runDir, "inline-audit", "blocked")
	if err == nil {
		t.Error("expected error when advancing past max iterations")
	}
}

func TestAuditIterationHistoryPreserved(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	// Start with max=4
	_, err := auditIterationStartCore(tmp, runDir, "inline-audit", 4)
	if err != nil {
		t.Fatal(err)
	}

	// Advance 1->2 with "blocked"
	_, err = auditIterationAdvanceCore(tmp, runDir, "inline-audit", "blocked")
	if err != nil {
		t.Fatal(err)
	}

	// Advance 2->3 with "pass"
	_, err = auditIterationAdvanceCore(tmp, runDir, "inline-audit", "pass")
	if err != nil {
		t.Fatal(err)
	}

	// Read final state and verify all history entries
	statePath := iterationStatePath(runDir, "inline-audit")
	data, _ := os.ReadFile(statePath)
	var state iterationState
	json.Unmarshal(data, &state)

	if len(state.History) != 3 {
		t.Fatalf("history length = %d, want 3", len(state.History))
	}

	// Iteration 1 should have result "blocked"
	if state.History[0].Result != "blocked" {
		t.Errorf("history[0].Result = %q, want %q", state.History[0].Result, "blocked")
	}
	// Iteration 2 should have result "pass"
	if state.History[1].Result != "pass" {
		t.Errorf("history[1].Result = %q, want %q", state.History[1].Result, "pass")
	}
	// Iteration 3 (current) should have no result yet
	if state.History[2].Result != "" {
		t.Errorf("history[2].Result = %q, want empty", state.History[2].Result)
	}
}

func TestAuditIterationCheckpointType(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "run-001")
	os.MkdirAll(runDir, 0755)

	// Verify inline-checkpoint uses _inline-checkpoint dir
	result, err := auditIterationStartCore(tmp, runDir, "inline-checkpoint", 2)
	if err != nil {
		t.Fatal(err)
	}

	statePath := iterationStatePath(runDir, "inline-checkpoint")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Errorf("state file not created at %s", statePath)
	}

	if result.Max != 2 {
		t.Errorf("max = %d, want 2", result.Max)
	}
}
