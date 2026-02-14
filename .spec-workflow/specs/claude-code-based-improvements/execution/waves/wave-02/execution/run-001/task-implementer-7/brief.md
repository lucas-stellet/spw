# Brief: task-implementer-7

## Inputs
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md
- Tasks: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Existing Schema: cli/internal/validate/schema.go

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
**Task 7: Implement iteration limit logic with state persistence**

Files to create:
- cli/internal/validate/iteration.go
- cli/internal/validate/iteration_test.go

Requirements: REQ-005, REQ-007

### Test Cases Required
- State file creation on first call
- Counter increment on subsequent calls
- Threshold trigger (count > max)
- State persistence (read back after write)
- Config override for limits
- WAITING_FOR_HUMAN_DECISION returned when exceeded

### Verification Command
go test ./cli/internal/validate/ -run TestIterationLimit -v

### Definition of Done
- CheckIterationLimit reads/creates _iteration_state.json
- Counters increment correctly
- Limit exceeded returns WAITING_FOR_HUMAN_DECISION
- State file persisted in run directory
- All test cases pass

## Output Contract
Write your output to:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/task-implementer-7/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/task-implementer-7/status.json

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["effective-go"],
  "skills_missing": [],
  "model_override_reason": null
}
```
