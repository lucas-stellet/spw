# Brief: spec-compliance-reviewer

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/design.md
- .spec-workflow/specs/claude-code-based-improvements/requirements.md
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer/report.md
- cli/internal/validate/schema.go
- cli/internal/validate/schema_test.go

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
<!-- Describe what this subagent must do -->

### Review Task 1: Create validate package foundation

Verify implementation against:
- **Requirements:** REQ-001, REQ-007
- **Definition of Done from tasks.md:**
  - FieldRule, Violation, ValidationResult, ValidationStats types exported
  - Helper functions for field type validation work correctly
  - Package compiles with no errors
- **Test Plan:** FieldRule validation helpers — type checking for string, string_array, enum. Enum match/mismatch. Required vs optional field logic.

**Compliance Check:**
1. Read the implementation files (schema.go, schema_test.go)
2. Verify all exported types match the design contract
3. Verify helper functions work correctly
4. Run `go build ./cli/...` to confirm compilation
5. Run tests to verify test coverage

**Restrictions:**
- No dependencies on tools, hook, or cli packages
- Pure validation types only

Report pass/blocked with specific evidence of compliance or gaps.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/spec-compliance-reviewer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/spec-compliance-reviewer/status.json

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
