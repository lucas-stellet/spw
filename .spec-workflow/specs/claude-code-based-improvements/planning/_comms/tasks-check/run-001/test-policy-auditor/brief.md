# Brief: test-policy-auditor

## Inputs
- `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- `.spec-workflow/specs/claude-code-based-improvements/design.md`

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
Verify every task satisfies the test policy:

1. **Every task has a Test Plan.** When `require_test_per_task: true`, each task must have a `Test Plan:` section with at least one test type (Unit, Integration, E2E). If `allow_no_test_exception: true`, a task with `TDD: skip` and a `No-TDD Justification:` is acceptable.
2. **No-TDD exception justification.** If a task has `TDD: skip`, it must provide `No-TDD Justification:` with `Reason:` and `Alternative validation:`.
3. **Verification Command present.** Every task must have a `Verification Command:` field.
4. **Definition of Done present.** Every task must have a `Definition of Done:` with concrete criteria.
5. **Test plan alignment with design.** Verify test plans align with the test strategy in design.md (table-driven tests, golden files, boundary tests where specified).

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/test-policy-auditor/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/test-policy-auditor/status.json

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
