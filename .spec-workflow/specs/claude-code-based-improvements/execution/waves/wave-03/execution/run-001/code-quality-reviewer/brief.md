# Brief: code-quality-reviewer

## Inputs
- cli/internal/validate/audit.go
- cli/internal/validate/audit_test.go
- .spec-workflow/specs/claude-code-based-improvements/design.md

## Task
Review Task 6 implementation (audit confidence gate logic) for code quality:

1. Maintainability: Code is clean, well-structured, follows Go best practices
2. Safety: No security vulnerabilities, proper error handling
3. Regression risk: No breaking changes to existing validate package exports
4. Tests: Adequate test coverage, edge cases covered

Run go vet ./cli/internal/validate/ and go build ./cli/... to verify.

Report PASS if code quality is acceptable, BLOCKED with specific issues if not.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/code-quality-reviewer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/code-quality-reviewer/status.json

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
