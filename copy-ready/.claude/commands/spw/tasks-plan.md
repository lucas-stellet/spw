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

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Backward compatibility:
- if `[skills.design].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Skill loading gate (mandatory when `skills.enabled=true`):
1. Explicitly invoke every required design skill before decomposition/wave planning.
2. Record loaded/missing skills in:
   - `.spec-workflow/specs/<spec-name>/SKILLS-TASKS-PLAN.md`
3. If any required skill is missing/not invoked:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

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
1. Run design skill loading gate and write `SKILLS-TASKS-PLAN.md`.
2. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/user-templates/tasks-template.md` (preferred)
   - fallback: `.spec-workflow/templates/tasks-template.md`
3. Dispatch `task-decomposer`.
4. Dispatch `dependency-graph-builder`.
5. Dispatch `parallel-conflict-checker`.
6. Dispatch `test-policy-enforcer`.
7. Dispatch `tasks-writer` and save `.spec-workflow/specs/<spec-name>/tasks.md`.
8. Handle approval via MCP only:
   - call `spec-status`
   - resolve tasks status from:
     - `documents.tasks.approved`
     - `documents.tasks.status`
     - `approvals.tasks.status`
   - if approved, continue without re-requesting
   - if `needs-revision`/`changes-requested`/`rejected`, stop BLOCKED
   - if pending, stop with `WAITING_FOR_APPROVAL` and instruct UI approval + rerun
   - only if approval was never requested (missing/empty/unknown status):
     - call `request-approval` then `get-approval-status` once
     - if pending, stop with `WAITING_FOR_APPROVAL`
     - if needs revision, stop BLOCKED
   - never ask for approval in chat
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
