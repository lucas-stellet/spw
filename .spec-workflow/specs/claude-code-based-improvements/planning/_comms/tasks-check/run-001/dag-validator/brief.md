# Brief: dag-validator

## Inputs
- `.spec-workflow/specs/claude-code-based-improvements/tasks.md`

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
Verify the dependency graph (DAG) is valid and wave assignments are consistent:

1. **No dependency cycles.** Build the dependency graph from `Depends On:` fields and verify it is acyclic.
2. **All dependency references valid.** Every task ID in `Depends On:` must correspond to an actual task ID (1-11).
3. **Wave order consistency.** A task in Wave N must not depend on a task in Wave M where M > N.
4. **Parallel correctness.** Tasks in `Can Run In Parallel With:` must be in the same wave and must not depend on each other.
5. **Deferred task placement.** Verify tasks marked Wave 2+ are in the "Deferred Tasks" section, not the executable section.
6. **Max wave size.** Verify no wave exceeds `max_wave_size: 3`.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/dag-validator/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/dag-validator/status.json

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
