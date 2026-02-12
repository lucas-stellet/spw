---
name: spw:exec
description: Subagent-driven task execution in batches with mandatory checkpoints
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<dispatch_pattern>
category: wave-execution
subcategory: implementation
phase: execution
comms_path: execution/waves/wave-{wave}/execution
artifacts: execution/_implementation-logs
policy: @.claude/workflows/spw/shared/dispatch-wave.md
</dispatch_pattern>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
- @.claude/workflows/spw/shared/dispatch-implementation.md
</shared_policies>

<objective>
Execute tasks in batches, with mandatory pauses for checkpoint quality gates.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/tasks.md`
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`

output:
- Code changes + git commits
- `.spec-workflow/specs/<spec-name>/planning/SKILLS-EXEC.md`

comms:
- `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/execution/<run-id>/`
- `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/checkpoint/<run-id>/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `execution-state-scout` (model: implementation)
  - Reads workflow state and returns a compact resume decision for the orchestrator.
- `task-implementer` (model: implementation)
  - Implements each task and runs task-level verification.
- `spec-compliance-reviewer` (model: complex_reasoning for complex/critical tasks; otherwise implementation)
  - Verifies exact adherence to requirements/design/task scope.
- `code-quality-reviewer` (model: implementation)
  - Reviews maintainability, safety, and regression risk.
</subagents>

<subagent_artifact_map>
| Subagent | Artifact | Dispatch | Model |
|----------|----------|----------|-------|
| execution-state-scout | (report.md only) | task | implementation |
| task-implementer | code changes + commits | task | implementation |
| spec-compliance-reviewer | (report.md only) | task | complex_reasoning |
| code-quality-reviewer | (report.md only) | task | implementation |
</subagent_artifact_map>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the wave execution dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: scout, skills, preconditions ................... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
   - if missing, list available specs and stop BLOCKED.
2. Apply skills policy: run implementation skills preflight and write `SKILLS-EXEC.md`.
3. Dispatch `execution-state-scout` and require compact handoff contract.
   - if `tasks.md` is missing, stop BLOCKED → instruct `spw:tasks-plan <spec-name>`.
4. Resolve resume state from scout handoff.
5. Resolve current wave ID and ensure canonical wave comms folder exists.
6. Read only task-scoped context required for selected task IDs.
</pre_pipeline>

<!-- inter_wave: checkpoint + user authorization .................. -->
<inter_wave>
1. At end of batch, **stop execution** and instruct the user:
   "Wave complete. Run `/spw:checkpoint <spec-name>` in a new session (`/clear` first or new Claude Code session)."
   Do NOT invoke checkpoint via Skill, inline, or any other method within the current exec session.
2. If checkpoint BLOCKED, stop.
3. If checkpoint PASS:
   - if `require_clean_worktree_for_wave_pass=true` and worktree is dirty: stop BLOCKED
   - if no remaining waves: finish
   - if remaining waves and `require_user_approval_between_waves=true`: request explicit authorization (§ wave_authorization)
   - if remaining waves and `require_user_approval_between_waves=false`: continue
</inter_wave>

<!-- per_task: git hygiene, commit policy ......................... -->
<per_task>
1. Mark task `[-]`.
2. Dispatch `task-implementer` (mandatory, even when single-task batch).
3. Dispatch `spec-compliance-reviewer` (use `<complexity_routing>` for model selection).
4. Dispatch `code-quality-reviewer`.
5. If all gates pass:
   - log implementation details and mark `[x]`
   - enforce `<git_hygiene>` commit policy for this task
6. If any gate fails: mark BLOCKED and stop current batch.
</per_task>

<!-- post_pipeline: completion guidance ........................... -->
<post_pipeline>
1. After full execution success:
   - If `tasks_generation_strategy=rolling-wave` and no further waves, recommend:
     `spw:tasks-plan <spec-name>` then `spw:tasks-check <spec-name>`.
   - If `tasks_generation_strategy=all-at-once`, recommend final validation only.
2. Write `<run-dir>/_handoff.md` with execution summary.
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

<execution_mode>
- Subagent dispatch is mandatory for every task, including sequential waves with only one task.
- The main agent is an orchestrator only (selection, dispatch, aggregation, status updates).
- Do not implement task code directly in the main orchestration context.
</execution_mode>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.implementation].required`
- `[skills.implementation].optional`
- `[skills.implementation].enforce_required` (boolean)
- `[execution].tdd_default` (boolean)

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/planning/SKILLS-EXEC.md`
2. Avoid loading full skill content in main context (subagent-first).
3. If `[execution].tdd_default=true`, treat `test-driven-development` as required for this phase (effective required set).
4. Require task subagent outputs/logs to explicitly mention skills used/missing.
5. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<state_recon_policy>
Before broad reads in the main context, dispatch `execution-state-scout` (implementation model; default Sonnet) to inspect execution state.

Scout reads:
- `.spec-workflow/specs/<spec-name>/tasks.md`
- `.spec-workflow/specs/<spec-name>/execution/waves/**/_latest.json`
- `.spec-workflow/specs/<spec-name>/execution/waves/**/_wave-summary.md`
- latest checkpoint status artifacts available in wave comms
- `git status --porcelain` only when clean-worktree gate is enabled

Scout handoff contract (compact):
- `checkpoint_status`: `PASS|BLOCKED|MISSING`
- `current_wave`: `wave-<NN>|none`
- `in_progress_tasks`: task IDs currently `[-]`
- `next_executable_tasks`: ordered task IDs ready to run now
- `resume_action`: `resume-in-progress|start-next-task|wait-user-authorization|manual-handoff|done|blocked`
- `reason` and `evidence_paths` (max 5)

Output budget:
- max 12 bullets plus one machine-readable JSON block.
- no large excerpts from `tasks.md`, `requirements.md`, or `design.md`.

Main-context rule:
- use scout summary as the source for "where to resume".
- only read task-scoped files needed for selected task IDs after scout handoff.
</state_recon_policy>

<wave_comms_layout>
Execution/checkpoint communications must be grouped by wave:
- base: `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/`

Within each wave folder (`<run-id>` MUST be `run-NNN` format — e.g. `run-001`, never dates):
- `execution/<run-id>/` (task execution communications/evidence)
- `checkpoint/<run-id>/` (checkpoint communications/evidence)
- `post-check/<run-id>/` (optional post-check validations)
- `_wave-summary.md` (wave-level synthesis)
- `_latest.json` (pointers to latest execution/checkpoint/post-check run IDs)

Rules:
- Do not create top-level timestamped wave folders outside the canonical wave path.
- Keep all wave evidence under its canonical `wave-<NN>/` folder.
</wave_comms_layout>

<wave_authorization>
Resolve from `.spec-workflow/spw-config.toml` `[execution].require_user_approval_between_waves` (default `true`).

If `true`:
- After each checkpoint PASS, if there is at least one remaining wave, ask explicit user authorization.
- Never auto-start the next wave.
- Use AskUserQuestion with options:
  - "Continue to next wave (Recommended)"
  - "Pause here"
  - "Review checkpoint details first"
- Proceed only on explicit continue.
</wave_authorization>

<manual_task_policy>
Resolve from `.spec-workflow/spw-config.toml` `[execution].manual_tasks_require_human_handoff` (default `true`).

When enabled, if the next task is manual/human-gated:
- do not auto-mark `[ ] -> [-]`
- do not auto-execute
- stop with `WAITING_FOR_HUMAN_ACTION`
- provide checklist + exact command to resume after user confirms completion
</manual_task_policy>

<tasks_planning_strategy>
Resolve from `.spec-workflow/spw-config.toml` `[planning].tasks_generation_strategy` (default `rolling-wave`).

Post-execution behavior:
- `rolling-wave`: after current waves finish, recommend planning the next executable wave.
- `all-at-once`: do not force another planning cycle unless requirements/design changed.
</tasks_planning_strategy>

<git_hygiene>
Resolve from `.spec-workflow/spw-config.toml` `[execution]`:
- `commit_per_task` (`"auto"|"manual"|"none"`, default `"auto"`)
- `require_clean_worktree_for_wave_pass` (default `true`)

Rules:
- If `commit_per_task="auto"` or `"manual"`: for each completed implementation task, create an atomic commit before moving forward.
- Commit must include task-scoped code changes plus task status artifacts (`tasks.md`).
- Implementation logs should be recorded during execution, but missing logs are enforced only at `spw:checkpoint`.
- Commit message must follow Conventional Commits:
  - `<type>(<spec-name>): task <task-id> - <short-title>`
  - type guidance: `feat|fix|refactor|test|docs|chore`
- If `commit_per_task="manual"`, stop with exact `git add`/`git commit` commands.
- If `commit_per_task="none"`, skip per-task commit enforcement.
- If clean-worktree gate is enabled, checkpoint PASS cannot advance while `git status --porcelain` is non-empty.
</git_hygiene>

<scope_control>
- Execute only the currently selected task IDs for the active batch/wave.
- Do not "pre-fix" TODOs or unrelated failures from other tasks.
- If unrelated issues block the current task, stop with BLOCKED and report:
  - blocking file/test
  - why it is outside current task scope
  - suggested next task/order adjustment
</scope_control>

<complexity_routing>
Treat a task as complex/critical when it includes one or more:
- touches auth/security/payments/data migrations/concurrency boundaries
- modifies > 3 core files
- affects cross-context architecture or high-risk integrations

For complex/critical tasks, run spec-compliance review on `complex_reasoning` model.
</complexity_routing>

<strict_mode>
With `--strict true` (default):
- block continuation when checkpoint fails.
- block continuation when required implementation skills were not invoked under enforce mode.
- block continuation when a task has no requirement traceability.
- block continuation when implementation was done without required subagent dispatch.
- block continuation when `execution-state-scout` handoff is missing or incomplete.
- block continuation on out-of-scope edits (changes unrelated to active task IDs).
- block progression to next wave without required user authorization.
- block progression without required per-task commit.
- block progression when clean-worktree gate is enabled and there are uncommitted changes.
</strict_mode>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/spw/overlays/active/exec.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] Every task was dispatched to a subagent (no direct implementation in orchestrator).
- [ ] Per-task commit policy was enforced according to config.
- [ ] Checkpoint was run after each batch/wave.
- [ ] Wave authorization was respected between waves.
- [ ] File-based handoff exists under `execution/waves/wave-<NN>/execution/<run-id>/`.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
After each batch:
- Show state-recon decision (`resume_action`, selected task IDs, checkpoint status).
- Show executed task IDs, commit/log evidence, and checkpoint status.
- If checkpoint PASS and there are remaining waves:
  - if authorization is required, stop in `WAITING_FOR_USER_AUTHORIZATION` and ask whether to continue.
  - otherwise, suggest/continue next batch per policy.
- If checkpoint BLOCKED: stop and show exact corrective actions.

If waiting on manual task:
- stop in `WAITING_FOR_HUMAN_ACTION` with checklist.
- keep manual task unchecked unless user explicitly confirms it started/completed.

After full execution success:
- If no further executable waves are planned and `tasks_generation_strategy=rolling-wave`, recommend:
  - `spw:tasks-plan <spec-name>`
  - then `spw:tasks-check <spec-name>`
- If no further executable waves are planned and `tasks_generation_strategy=all-at-once`, skip forced re-planning and recommend final validation only.
- Recommend final validation review and optionally `/clear` before any new planning cycle.

If blocked by spec resolution:
- Show the exact missing path and list discovered specs from `.spec-workflow/specs/`.
- Recommend the corrected command using one discovered spec name.
</completion_guidance>
