# Brief: task-implementer-4

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
**Task 4: Implement enhanced status.json validation with graduated enforcement**

Files to create:
- cli/internal/validate/status.go
- cli/internal/validate/status_test.go

Requirements: REQ-003, REQ-007

### Test Cases Required
- All 5 fields present and valid
- Missing optional fields in default mode (warn)
- Missing optional fields in strict mode (error)
- Wrong types per field
- Invalid status enum
- Null vs missing distinction for model_override_reason
- Empty skills arrays valid

### Verification Command
go test ./cli/internal/validate/ -run TestValidateStatus -v

### Definition of Done
- ValidateStatus(data, strict) validates all 5 fields with graduated enforcement
- Default mode: 2 required + 3 optional (warn)
- Strict mode: all 5 required
- StatusValidationResult with field-level errors
- All test cases pass

## Output Contract
Write your output to:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/task-implementer-4/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/task-implementer-4/status.json

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
