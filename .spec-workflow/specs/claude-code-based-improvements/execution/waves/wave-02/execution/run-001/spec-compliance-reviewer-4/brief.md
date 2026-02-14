# Brief: spec-compliance-reviewer-4

## Inputs
- Implementation: cli/internal/validate/status.go
- Tests: cli/internal/validate/status_test.go
- Task Definition: .spec-workflow/specs/claude-code-based-improvements/tasks.md (Task 4)
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
**Review Task 4: Enhanced status.json validation with graduated enforcement**

Verify exact adherence to requirements:
- REQ-003: Enhanced status.json validation
- REQ-007: Validation infrastructure

### Review Criteria
1. ValidateStatus(data, strict) validates all 5 fields with graduated enforcement
2. Default mode: 2 required + 3 optional (warn)
3. Strict mode: all 5 required
4. StatusValidationResult with field-level errors
5. All test cases covered per tasks.md

### Verification Command
go test ./cli/internal/validate/ -run TestValidateStatus -v

## Output Contract
Write your output to:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/spec-compliance-reviewer-4/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/spec-compliance-reviewer-4/status.json
