# Brief: spec-compliance-reviewer-7

## Inputs
- Implementation: cli/internal/validate/iteration.go
- Tests: cli/internal/validate/iteration_test.go
- Task Definition: .spec-workflow/specs/claude-code-based-improvements/tasks.md (Task 7)
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
**Review Task 7: Iteration limit logic with state persistence**

Verify exact adherence to requirements:
- REQ-005: Iteration limits
- REQ-007: Validation infrastructure

### Review Criteria
1. CheckIterationLimit reads/creates _iteration_state.json
2. Counters increment correctly
3. Limit exceeded returns WAITING_FOR_HUMAN_DECISION
4. State file persisted in run directory
5. All test cases covered per tasks.md

### Verification Command
go test ./cli/internal/validate/ -run TestIterationLimit -v

## Output Contract
Write your output to:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/spec-compliance-reviewer-7/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/spec-compliance-reviewer-7/status.json
