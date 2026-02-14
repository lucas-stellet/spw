# Brief: spec-compliance-reviewer

## Inputs
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/task-implementer/report.md
- cli/internal/validate/audit.go
- cli/internal/validate/audit_test.go

## Task
Verify Task 6 implementation (audit confidence gate logic) against tasks.md Definition of Done:

1. ApplyAuditGate checks confidence and validated fields - verify function exists and has correct logic
2. Findings below threshold downgraded to warnings - verify threshold comparison logic
3. AuditGateResult includes original and effective status - verify struct has both fields
4. Boundary tests pass at exactly audit_min_confidence (0.8) - run tests
5. All test cases pass - run go test ./cli/internal/validate/ -run TestAuditGate -v

Report PASS if all criteria met, BLOCKED with specific issues if not.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/spec-compliance-reviewer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/spec-compliance-reviewer/status.json

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
