# Brief: code-quality-reviewer

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- cli/internal/validate/schema.go
- cli/internal/validate/schema_test.go
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer/report.md

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
<!-- Describe what this subagent must do -->

### Code Quality Review - Task 1

Review code quality, maintainability, safety, and regression risk for:
- `cli/internal/validate/schema.go`
- `cli/internal/validate/schema_test.go`

**Quality Criteria:**
1. **Maintainability:** Code is clean, well-organized, follows Go conventions
2. **Safety:** No nil pointer dereferences, proper error handling, bounds checking
3. **Regression Risk:** Low - changes are isolated to validate package
4. **Test Coverage:** Adequate test coverage for validation logic

**Review Actions:**
1. Read the implementation files
2. Run static analysis if available (golint, staticcheck)
3. Check for common Go anti-patterns
4. Verify tests are comprehensive

Report pass/blocked with specific quality findings.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/code-quality-reviewer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/code-quality-reviewer/status.json

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
