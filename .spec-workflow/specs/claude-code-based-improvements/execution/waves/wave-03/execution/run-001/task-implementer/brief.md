# Brief: task-implementer

## Inputs
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/design.md
- .spec-workflow/specs/claude-code-based-improvements/requirements.md
- .spec-workflow/specs/claude-code-based-improvements/planning/SKILLS-EXEC.md

## Task
Implement **Task 6: Implement audit confidence gate logic**.

From tasks.md:
- Files: cli/internal/validate/audit.go, cli/internal/validate/audit_test.go
- Requirements: REQ-004, REQ-007
- Test Plan:
  - Unit: Confidence exactly at threshold (0.8) stays blocked. Below threshold (0.79) downgraded to warning. Above threshold (0.81) stays blocked. validated=false always downgraded. Missing confidence field treated as 0 (downgraded). Custom threshold from config.
- Verification Command: go test ./cli/internal/validate/ -run TestAuditGate -v
- Definition of Done:
  - ApplyAuditGate checks confidence and validated fields
  - Findings below threshold downgraded to warnings
  - AuditGateResult includes original and effective status
  - Boundary tests pass at exactly audit_min_confidence
  - All test cases pass

Skill required: effective-go

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/task-implementer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/task-implementer/status.json

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
