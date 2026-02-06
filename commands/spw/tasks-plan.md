---
name: spw:tasks-plan
description: Subagent-driven tasks.md generation for waves, parallelism, and per-task TDD
argument-hint: "<spec-name> [--mode initial|next-wave] [--max-wave-size <N>] [--allow-no-test-exception true|false]"
---

<objective>
Generate `.spec-workflow/specs/<spec-name>/tasks.md` for predictable parallel execution.
</objective>

<planning_defaults>
Resolve planning defaults from `.spec-workflow/spw-config.toml` `[planning]`:
- `tasks_generation_strategy` (`rolling-wave|all-at-once`, default `rolling-wave`)
- `max_wave_size` (default `3`)

Precedence:
1. Explicit CLI args (`--mode`, `--max-wave-size`)
2. Config values from `[planning]`
3. Built-in defaults (`rolling-wave`, `3`)
</planning_defaults>

<mode_policy>
`--mode` controls how tasks are generated:
- `initial`:
  - create only Wave 1 executable tasks
  - keep later ideas only as deferred notes/backlog (not executable waves)
- `next-wave`:
  - append/update only the next executable wave based on implemented reality
  - do not regenerate or rewrite completed waves

If `--mode` is omitted:
- resolve `[planning].tasks_generation_strategy` (default `rolling-wave`)
- if strategy is `rolling-wave`:
  - if `tasks.md` does not exist -> behave as `initial`
  - if `tasks.md` exists -> behave as `next-wave`
- if strategy is `all-at-once`:
  - generate all executable waves in one planning pass
  - do not require prior completed/checkpointed wave
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

<resume_policy>
Before creating a new run, inspect existing tasks-plan run folders:
- `.spec-workflow/specs/<spec-name>/agent-comms/tasks-plan/<run-id>/`

A run is `unfinished` when any of these is true:
- `_handoff.md` is missing
- any subagent directory is missing `brief.md`, `report.md`, or `status.json`
- any subagent `status.json` reports `status=blocked`

Resume decision gate (mandatory):
1. Find latest unfinished run (if multiple, sort by mtime descending and use the newest).
2. If found, ask user once (AskUserQuestion) with options:
   - `continue-unfinished` (Recommended): continue with that run directory.
   - `delete-and-restart`: delete that unfinished run directory and start a new run.
3. Never choose automatically. Do not infer user intent.
4. If explicit user decision is unavailable, stop with `WAITING_FOR_USER_DECISION`.
5. Do not create a new run-id before this decision.

If user chooses `continue-unfinished`:
- Reuse completed subagent outputs (`report.md` + `status.json` with `status=pass`) from decomposition/checker roles.
- Redispatch only missing/blocked subagents.
- Always rerun `tasks-writer` before finalizing `tasks.md`.

If user chooses `delete-and-restart`:
- Delete the selected unfinished run dir.
- Continue workflow with a fresh run-id.
- Record deleted path in final output.
</resume_policy>

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
- In `all-at-once` strategy (with omitted `--mode`), emit all executable waves in one pass.
- Always enforce effective `max_wave_size` (CLI override or `[planning].max_wave_size`).
</rules>

<dashboard_markdown_profile>
`tasks.md` must remain compatible with `spec-workflow-mcp` dashboard parser/validator:
- Use checkbox markers only on real task lines:
  - `- [ ] <id>. <description>`
  - `- [-] <id>. <description>`
  - `- [x] <id>. <description>`
- Use `-` as task list marker (never `*` for task rows).
- Every task line must start with numeric ID and IDs must be globally unique in the file.
- Never use nested checkboxes inside metadata sections (DoD, notes, test details).
- Keep metadata lines as regular bullets (`- ...`), never checkbox bullets.
- Use parseable metadata lines:
  - `- Files: path/to/file.ext, test/path/to/file_test.ext` (single-line CSV)
  - `- _Requirements: ..._` (underscore-delimited)
  - `- _Leverage: ..._` (when reuse targets exist)
  - `- _Prompt: ..._` (must close with `_`)
- `_Prompt` must include: `Role: ... | Task: ... | Restrictions: ... | Success: ...`.
</dashboard_markdown_profile>

<workflow>
1. Run design skills preflight (availability + load mode) and write `SKILLS-TASKS-PLAN.md`.
2. Inspect existing tasks-plan run dirs and apply `<resume_policy>` decision gate.
3. Determine active run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create:
     `.spec-workflow/specs/<spec-name>/agent-comms/tasks-plan/<run-id>/`
4. Resolve effective planning behavior (per mode_policy + planning_defaults):
   - resolve effective `max_wave_size`
   - resolve effective generation mode:
     - explicit `--mode initial` -> `initial`
     - explicit `--mode next-wave` -> `next-wave`
     - omitted `--mode` + `rolling-wave` -> derive `initial` or `next-wave` from `tasks.md` presence
     - omitted `--mode` + `all-at-once` -> `all-at-once`
   - validate preconditions:
     - `initial` requires no prior completed executable wave
     - `next-wave` requires existing `tasks.md` plus at least one completed or checkpointed wave
     - `all-at-once` does not require prior completed/checkpointed wave
5. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/specs/<spec-name>/tasks.md` (required for `next-wave`; optional for `all-at-once` reconciliation)
   - `.spec-workflow/specs/<spec-name>/CHECKPOINT-REPORT.md` (if present, for reconciliation)
   - `.spec-workflow/user-templates/tasks-template.md` (preferred)
   - fallback: `.spec-workflow/templates/tasks-template.md`
6. Write briefs (including mode + required skills per role) and dispatch `task-decomposer`.
7. Write briefs (including mode + required skills per role) and dispatch `dependency-graph-builder`.
8. Write briefs (including mode + required skills per role) and dispatch `parallel-conflict-checker`.
9. Write briefs (including mode + required skills per role) and dispatch `test-policy-enforcer`.
   - if resuming, redispatch only missing/blocked subagents
10. Require subagent `report.md` + `status.json` (with skill fields); BLOCKED if missing.
11. Dispatch `tasks-writer` with file handoff and save `.spec-workflow/specs/<spec-name>/tasks.md`:
    - `initial`: only Wave 1 executable tasks
    - `next-wave`: only one newly appended executable wave
    - `all-at-once`: full executable plan (all waves) respecting `max_wave_size`
    - if resuming, always rerun `tasks-writer` before final save
12. Write `<run-dir>/_handoff.md` with mode decisions, DAG rationale, conflict/test policy outcomes, and resume decision taken (`continue-unfinished` or `delete-and-restart`).
13. Handle approval via MCP only:
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
- [ ] Task IDs are numeric and globally unique (no duplicate IDs).
- [ ] Every executable task has a parseable single-line `Files:` metadata entry.
- [ ] Prompt fields follow `Role|Task|Restrictions|Success` and close with underscore.
- [ ] All tasks have tests or a documented exception.
- [ ] Waves respect the configured limit.
- [ ] Conflict checker returns no critical same-wave file collisions.
- [ ] If effective mode is `initial`, only Wave 1 executable tasks were produced.
- [ ] If effective mode is `next-wave`, exactly one new executable wave was produced.
- [ ] If effective mode is `all-at-once`, a full executable multi-wave plan was produced.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/agent-comms/tasks-plan/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/tasks.md`.
- Confirm effective planning mode (`initial`, `next-wave`, or `all-at-once`).
- Confirm effective `max_wave_size` source (CLI override or `[planning].max_wave_size`).
- Confirm approval request status for tasks.
- Recommend next command: `spw:tasks-check <spec-name>`.

If blocked:
- Show mode/precondition/decomposition/dependency/conflict/test-policy failures.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide rerun command:
  - default from config: `spw:tasks-plan <spec-name>`
  - explicit rolling override: `spw:tasks-plan <spec-name> --mode initial --max-wave-size <N>`
  - explicit rolling append: `spw:tasks-plan <spec-name> --mode next-wave --max-wave-size <N>`.
</completion_guidance>
