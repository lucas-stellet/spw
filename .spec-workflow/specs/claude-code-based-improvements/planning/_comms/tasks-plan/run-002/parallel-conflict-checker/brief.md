# Brief: parallel-conflict-checker

## Inputs
- Dependency graph report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/dependency-graph-builder/report.md
- Current tasks.md: .spec-workflow/specs/claude-code-based-improvements/tasks.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
Check Wave 2 tasks (3, 4, 7) for same-wave file/lock conflicts.

Each task's file targets from tasks.md:
- Task 3: cli/internal/validate/mirror.go, cli/internal/validate/mirror_test.go
- Task 4: cli/internal/validate/status.go, cli/internal/validate/status_test.go
- Task 7: cli/internal/validate/iteration.go, cli/internal/validate/iteration_test.go

Verify:
1. No two tasks modify the same file
2. No two tasks import/modify shared state that could conflict
3. All tasks are in the same package (cli/internal/validate/) -- confirm no package-level init conflicts
4. No lock file or resource conflicts

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/parallel-conflict-checker/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/parallel-conflict-checker/status.json

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
