# Brief: task-implementer

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/design.md
- .spec-workflow/specs/claude-code-based-improvements/requirements.md

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
<!-- Describe what this subagent must do -->

### Task 1: Create validate package foundation with shared schema types

**Files to create:**
- `cli/internal/validate/schema.go`

**Requirements:** REQ-001, REQ-007

**Definition of Done:**
- FieldRule, Violation, ValidationResult, ValidationStats types exported
- Helper functions for field type validation work correctly
- Package compiles with no errors

**Verification Command:**
```
go build ./cli/...
```

**Restrictions:**
- No dependencies on tools, hook, or cli packages
- Pure validation types only

**Test Plan (required):**
- Unit: FieldRule validation helpers — type checking for string, string_array, enum. Enum match/mismatch. Required vs optional field logic.

Execute the task per the Definition of Done. Write tests to verify the helper functions work correctly. Ensure the package compiles before marking complete.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer/status.json

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["skill-name"],
  "skills_missing": [],
  "model_override_reason": null
}
```
