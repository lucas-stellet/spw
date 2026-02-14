# Parallel Conflict Checker Brief

## Objective
Detect same-wave file/lock conflicts that would prevent safe parallel execution.

## Inputs
- Dependency graph: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/dependency-graph-builder/report.md
- Task decomposition: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/task-decomposer/report.md

## Wave 1 Tasks to Check
- Task 1: Creates cli/internal/validate/schema.go
- Task 2: Creates cli/internal/validate/prompts.go, cli/internal/validate/prompts_test.go (depends on Task 1)
- Task 5: Modifies cli/internal/config/config.go, cli/internal/config/config_test.go

## Output
Write `report.md` listing any file conflicts between parallel tasks within the same wave, and whether they are critical (blocking) or acceptable.
