# Brief: dependency-graph-builder (next-wave mode)

## Inputs
- Task decomposer report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/task-decomposer/report.md
- Current tasks.md: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true
- effective mode: next-wave

## Task
Build the dependency DAG for Wave 2 (tasks 3, 4, 7) and verify wave grouping.

Wave 1 completed: tasks 1, 2, 5 (all marked [x]).
Wave 2 candidates: tasks 3, 4, 7 (confirmed by task-decomposer).

1. Verify all dependencies for tasks 3, 4, 7 are satisfied by completed Wave 1 tasks.
2. Confirm tasks 3, 4, 7 have no inter-dependencies and can run in parallel.
3. Verify the DAG for the remaining deferred tasks (6, 8, 9, 10, 11) still holds after Wave 2 completion.
4. Produce a DAG showing the dependency graph across all waves.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/dependency-graph-builder/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/dependency-graph-builder/status.json

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
