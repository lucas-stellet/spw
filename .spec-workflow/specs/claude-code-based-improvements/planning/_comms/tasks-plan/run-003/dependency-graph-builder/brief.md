# Brief: dependency-graph-builder

## Context
- **Spec:** claude-code-based-improvements
- **Run:** run-003
- **Mode:** next-wave
- **Planning for:** Wave 3 (tasks 6, 8, 9)

## Inputs
- Task decomposition: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/task-decomposer/report.md`
- Current tasks: `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- Design: `.spec-workflow/specs/claude-code-based-improvements/design.md`

## Instructions
Build and validate the dependency DAG for Wave 3 tasks:

### Wave 3 Tasks (from task-decomposer)
- Task 6: Audit confidence gate logic (depends on tasks 1, 5 - complete)
- Task 8: CLI command wiring (depends on tasks 2, 3 - complete)
- Task 9: Dispatch-read-status integration (depends on task 6)

### Constraints
- max_wave_size: 3
- No cycles allowed
- Must respect existing task IDs

## Output
Validate the DAG and write to: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/dependency-graph-builder/report.md`

Include:
- DAG validation (no cycles, all dependencies satisfiable)
- Wave grouping recommendation
- Any dependency issues

Write status.json:
```json
{"status": "pass", "summary": "Wave 3 DAG validated with N tasks, no cycles"}
```
