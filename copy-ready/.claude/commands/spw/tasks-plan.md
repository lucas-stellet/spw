---
name: spw:tasks-plan
description: Subagent-driven tasks.md generation for waves, parallelism, and per-task TDD
argument-hint: "<spec-name> [--max-wave-size 3] [--allow-no-test-exception true|false]"
---

<objective>
Generate `.spec-workflow/specs/<spec-name>/tasks.md` for predictable parallel execution.
</objective>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<subagents>
- `task-decomposer` (model: complex_reasoning)
  - Creates atomic tasks from requirements/design.
- `dependency-graph-builder` (model: complex_reasoning)
  - Builds DAG and wave grouping.
- `parallel-conflict-checker` (model: implementation)
  - Detects same-wave file/lock conflicts.
- `test-policy-enforcer` (model: complex_reasoning)
  - Enforces test-per-task and valid exceptions.
- `tasks-writer` (model: implementation)
  - Writes final markdown in template format.
</subagents>

<rules>
- Each task must be self-contained.
- Each task must include tests and a verification command.
- No-test exception is allowed only with explicit justification.
- Plan dependencies to maximize safe parallelism per wave.
</rules>

<workflow>
1. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/user-templates/tasks-template.md` (preferred)
   - fallback: `.spec-workflow/templates/tasks-template.md`
2. Dispatch `task-decomposer`.
3. Dispatch `dependency-graph-builder`.
4. Dispatch `parallel-conflict-checker`.
5. Dispatch `test-policy-enforcer`.
6. Dispatch `tasks-writer` and save `.spec-workflow/specs/<spec-name>/tasks.md`.
7. Request approval.
</workflow>

<acceptance_criteria>
- [ ] All tasks have `Requirements`.
- [ ] All tasks have tests or a documented exception.
- [ ] Waves respect the configured limit.
- [ ] Conflict checker returns no critical same-wave file collisions.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/tasks.md`.
- Confirm approval request status for tasks.
- Recommend next command: `spw:tasks-check <spec-name>`.

If blocked:
- Show decomposition/dependency/conflict/test-policy failures.
- Provide rerun command: `spw:tasks-plan <spec-name> --max-wave-size <N>`.
</completion_guidance>
