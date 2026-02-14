# Brief: parallel-conflict-checker

## Context
- **Spec:** claude-code-based-improvements
- **Run:** run-003
- **Mode:** next-wave
- **Planning for:** Wave 3 (tasks 6, 8, 9)

## Inputs
- Task decomposition: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/task-decomposer/report.md`
- Dependency analysis: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/dependency-graph-builder/report.md`

## Instructions
Detect same-wave file/state conflicts for Wave 3:

### Wave 3 Tasks
- Task 6: cli/internal/validate/audit.go, audit_test.go
- Task 8: cli/internal/cli/validate_cmd.go, root.go
- Task 9: cli/internal/tools/dispatch_status.go

## Output
Write to: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/parallel-conflict-checker/report.md`

Check for:
- Same file modified by multiple tasks
- State conflicts (e.g., same config key)
- Resource conflicts

Then write status.json:
```json
{"status": "pass", "summary": "Wave 3 tasks verified safe for parallel execution - no conflicts"}
```
