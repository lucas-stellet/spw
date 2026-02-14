# Dependency Graph Builder Brief

## Objective
Build a dependency DAG from the decomposed tasks and organize them into waves respecting max_wave_size=3.

## Inputs
- Task decomposition: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/task-decomposer/report.md
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md

## Context
- Effective mode: `initial` (only Wave 1 executable tasks; later waves as deferred notes)
- max_wave_size: 3
- 11 tasks identified by decomposer

## Constraints
- Wave 1 must contain at most 3 tasks (max_wave_size=3)
- Wave 1 tasks must have no inter-dependencies (safe parallelism)
- Wave 1 tasks must be foundation tasks that unblock later work
- Later waves are deferred (not executable) in initial mode
- Respect code boundaries: validate package depends on config but not on tools/hook/cli

## Output
Write `report.md` with:
1. Dependency DAG (task ID -> depends on)
2. Wave assignment (which tasks go in which wave)
3. Parallelism notes (which tasks within a wave can run in parallel)
4. Rationale for Wave 1 selection
