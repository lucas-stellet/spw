# Tasks Writer Brief

## Objective
Write the final `tasks.md` in dashboard-compatible format, following the template and incorporating all subagent outputs.

## Inputs
- Task decomposition: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/task-decomposer/report.md
- Dependency graph: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/dependency-graph-builder/report.md
- Conflict check: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/parallel-conflict-checker/report.md
- Test policy: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/test-policy-enforcer/report.md
- Template: .spec-workflow/user-templates/tasks-template.md
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md

## Context
- Effective mode: `initial` (only Wave 1 executable tasks)
- max_wave_size: 3
- TDD: off (tdd_default=false)
- Wave 1 tasks: Task 1 (schema), Task 2 (prompts validation), Task 5 (config extension)

## Dashboard Markdown Profile (CRITICAL)
- Checkbox markers ONLY on real task lines: `- [ ] <id>. <description>`
- Use `-` as list marker (never `*`)
- Numeric IDs, globally unique
- No nested checkboxes in metadata
- Metadata as regular bullets
- `Files:` single-line CSV
- `_Requirements: ..._` underscore-delimited
- `_Prompt: Role: ... | Task: ... | Restrictions: ... | Success: ..._`

## Output
Write `report.md` containing the complete tasks.md content ready for filesystem copy.
