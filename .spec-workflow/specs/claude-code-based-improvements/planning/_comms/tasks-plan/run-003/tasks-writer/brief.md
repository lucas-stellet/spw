# Brief: tasks-writer

## Context
- **Spec:** claude-code-based-improvements
- **Run:** run-003
- **Mode:** next-wave
- **Planning for:** Wave 3 (tasks 6, 8, 9)

## Inputs
- Task decomposition: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/task-decomposer/report.md`
- Dependency analysis: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/dependency-graph-builder/report.md`
- Conflict analysis: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/parallel-conflict-checker/report.md`
- Test policy: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/test-policy-enforcer/report.md`

## Current tasks.md Location
`.spec-workflow/specs/claude-code-based-improvements/tasks.md`

## Instructions
Write the final tasks.md by:
1. Reading the existing tasks.md
2. Promoting Wave 3 deferred tasks (6, 8, 9) to current wave
3. Following the dashboard-compatible markdown format

## Template
Use: `.spec-workflow/user-templates/tasks-template.md`

## Output
Write final tasks.md to: `.spec-workflow/specs/claude-code-based-improvements/tasks.md`

Write status.json:
```json
{"status": "pass", "summary": "Wave 3 tasks.md produced with tasks 6, 8, 9 promoted to current wave"}
```
