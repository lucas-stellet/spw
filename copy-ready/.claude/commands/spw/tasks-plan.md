---
name: spw:tasks-plan
description: Subagent-driven tasks.md generation for waves, parallelism, and per-task TDD
argument-hint: "<spec-name> [--mode initial|next-wave] [--max-wave-size 3] [--allow-no-test-exception true|false]"
---

<objective>
Generate `.spec-workflow/specs/<spec-name>/tasks.md` for predictable parallel execution.
</objective>

<mode_policy>
`--mode` controls how tasks are generated:
- `initial`:
  - create only Wave 1 executable tasks
  - keep later ideas only as deferred notes/backlog (not executable waves)
- `next-wave`:
  - append/update only the next executable wave based on implemented reality
  - do not regenerate or rewrite completed waves

If `--mode` is omitted:
- if `tasks.md` does not exist -> behave as `initial`
- if `tasks.md` exists -> behave as `next-wave`
</mode_policy>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder:
- `.spec-workflow/specs/<spec-name>/agent-comms/tasks-plan/<run-id>/`

For each subagent, use:
- `<run-dir>/<subagent>/brief.md` (written by orchestrator before dispatch)
- `<run-dir>/<subagent>/report.md` (written by subagent after execution)
- `<run-dir>/<subagent>/status.json` (written by subagent)

Status schema (minimum):
- `status`: `pass|blocked`
- `summary`: short result
- `inputs`: key files used
- `outputs`: generated artifacts
- `open_questions`: unresolved items
- `skills_used`: skills actually used by the subagent
- `skills_missing`: required skills not available for the subagent (if any)

After planning, write:
- `<run-dir>/_handoff.md` (orchestrator synthesis and final decisions)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills].load_mode` (`subagent-first|principal-first`)
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Backward compatibility:
- if `[skills.design].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Load modes:
- `subagent-first` (default): orchestrator does availability preflight only and
  delegates skill loading/use to subagents via briefs.
- `principal-first` (legacy): orchestrator loads required skills before dispatch.

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/SKILLS-TASKS-PLAN.md`
2. If `load_mode=subagent-first`, avoid loading full skill content in main context.
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
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
- In `initial` mode, do not emit Wave 2+ executable tasks.
- In `next-wave` mode, emit exactly one new executable wave.
</rules>

<workflow>
1. Run design skills preflight (availability + load mode) and write `SKILLS-TASKS-PLAN.md`.
2. Create communication run directory:
   - `.spec-workflow/specs/<spec-name>/agent-comms/tasks-plan/<run-id>/`
3. Resolve `mode` from args or defaults (per mode_policy) and validate preconditions:
   - `initial` requires no prior completed executable wave
   - `next-wave` requires existing `tasks.md` plus at least one completed or checkpointed wave
4. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/specs/<spec-name>/tasks.md` (required for `next-wave`)
   - `.spec-workflow/specs/<spec-name>/CHECKPOINT-REPORT.md` (if present, for reconciliation)
   - `.spec-workflow/user-templates/tasks-template.md` (preferred)
   - fallback: `.spec-workflow/templates/tasks-template.md`
5. Write briefs (including mode + required skills per role) and dispatch `task-decomposer`.
6. Write briefs (including mode + required skills per role) and dispatch `dependency-graph-builder`.
7. Write briefs (including mode + required skills per role) and dispatch `parallel-conflict-checker`.
8. Write briefs (including mode + required skills per role) and dispatch `test-policy-enforcer`.
9. Require subagent `report.md` + `status.json` (with skill fields); BLOCKED if missing.
10. Dispatch `tasks-writer` with file handoff and save `.spec-workflow/specs/<spec-name>/tasks.md`:
    - `initial`: only Wave 1 executable tasks
    - `next-wave`: only one newly appended executable wave
11. Write `<run-dir>/_handoff.md` with mode decisions, DAG rationale, and conflict/test policy outcomes.
12. Handle approval via MCP only:
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
- [ ] `initial` mode produced only Wave 1 executable tasks.
- [ ] `next-wave` mode produced exactly one new executable wave.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/agent-comms/tasks-plan/<run-id>/`.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/tasks.md`.
- Confirm mode used (`initial` or `next-wave`).
- Confirm approval request status for tasks.
- Recommend next command: `spw:tasks-check <spec-name>`.

If blocked:
- Show mode/precondition/decomposition/dependency/conflict/test-policy failures.
- Provide rerun command:
  - `spw:tasks-plan <spec-name> --mode initial --max-wave-size <N>`
  - or `spw:tasks-plan <spec-name> --mode next-wave --max-wave-size <N>`.
</completion_guidance>
