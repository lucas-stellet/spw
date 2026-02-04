---
name: spw:exec
description: Subagent-driven task execution in batches with mandatory checkpoints
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<objective>
Execute tasks in batches, with mandatory pauses for checkpoint quality gates.
</objective>

<path_conventions>
- Canonical spec root: `.spec-workflow/specs/<spec-name>/`
- Canonical files:
  - `.spec-workflow/specs/<spec-name>/tasks.md`
  - `.spec-workflow/specs/<spec-name>/requirements.md`
  - `.spec-workflow/specs/<spec-name>/design.md`
- Legacy `.specs/` paths are NOT valid for SPW commands.
- Never read from or glob `.specs/*`.
</path_conventions>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills].load_mode` (`subagent-first|principal-first`)
- `[skills.implementation].required`
- `[skills.implementation].optional`
- `[skills.implementation].enforce_required` (boolean)

Backward compatibility:
- if `[skills.implementation].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Load modes:
- `subagent-first` (default): orchestrator does availability preflight only and
  delegates skill loading/use to task subagents.
- `principal-first` (legacy): orchestrator loads required skills before dispatch.

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/SKILLS-EXEC.md`
2. If `load_mode=subagent-first`, avoid loading full skill content in main context.
3. Require task subagent outputs/logs to explicitly mention skills used/missing.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<subagents>
- `task-implementer` (model: implementation)
  - Implements each task and runs task-level verification.
- `spec-compliance-reviewer` (model: complex_reasoning for complex/critical tasks; otherwise implementation)
  - Verifies exact adherence to requirements/design/task scope.
- `code-quality-reviewer` (model: implementation)
  - Reviews maintainability, safety, and regression risk.
</subagents>

<execution_mode>
- Subagent dispatch is mandatory for every task, including sequential waves with only one task.
- The main agent is an orchestrator only (selection, dispatch, aggregation, status updates).
- Do not implement task code directly in the main orchestration context.
</execution_mode>

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

When enabled, if the next task is manual/human-gated (for example external sandbox verification, production UI checks, stakeholder validation):
- do not auto-mark `[ ] -> [-]`
- do not auto-execute
- stop with `WAITING_FOR_HUMAN_ACTION`
- provide checklist + exact command to resume after user confirms completion
</manual_task_policy>

<git_hygiene>
Resolve from `.spec-workflow/spw-config.toml` `[execution]`:
- `commit_per_task` (default `true`)
- `auto_commit_on_task_completion` (default `true`)
- `require_clean_worktree_for_wave_pass` (default `true`)

Rules:
- For each completed implementation task, create an atomic commit before moving forward.
- Commit must include task-scoped code changes plus task status artifacts (`tasks.md`, implementation logs).
- Commit message must follow Conventional Commits:
  - `<type>(<spec-name>): task <task-id> - <short-title>`
  - type guidance: `feat|fix|refactor|test|docs|chore`
- If auto-commit is disabled, stop with exact `git add`/`git commit` commands.
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

<workflow>
0. Resolve spec directory:
   - `SPEC_DIR=.spec-workflow/specs/<spec-name>`
   - if `SPEC_DIR` does not exist, list available specs from `.spec-workflow/specs/*` and stop BLOCKED.
1. Run implementation skills preflight (availability + load mode) and write `SKILLS-EXEC.md`.
2. Read files from canonical paths:
   - `.spec-workflow/specs/<spec-name>/tasks.md`
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - if `tasks.md` is missing, stop BLOCKED and instruct to run `spw:tasks-plan <spec-name>`.
3. Select pending tasks by wave.
4. Execute up to `batch-size` tasks per batch (prefer safe parallelism).
5. For each task:
   - mark `[-]`
   - dispatch `task-implementer` (mandatory, even when single-task batch)
   - dispatch `spec-compliance-reviewer`
   - dispatch `code-quality-reviewer`
   - if all gates pass:
     - log implementation details and mark `[x]`
     - enforce git_hygiene commit policy for this task
   - if any gate fails: mark BLOCKED and stop current batch
6. At end of batch, run `spw:checkpoint <spec-name>`.
7. If checkpoint BLOCKED, stop.
8. If checkpoint PASS:
   - if `require_clean_worktree_for_wave_pass=true` and worktree is dirty: stop BLOCKED
   - if no remaining waves: finish
   - if remaining waves and `require_user_approval_between_waves=true`: request explicit authorization, then continue only if approved
   - if remaining waves and `require_user_approval_between_waves=false`: continue by policy
</workflow>

<strict_mode>
With `--strict true` (default):
- block continuation when checkpoint fails.
- block continuation when required implementation skills were not invoked under enforce mode.
- block continuation when a task has no requirement traceability.
- block continuation when implementation was done without required subagent dispatch.
- block continuation on out-of-scope edits (changes unrelated to active task IDs).
- block progression to next wave without required user authorization.
- block progression without required per-task commit.
- block progression when clean-worktree gate is enabled and there are uncommitted changes.
</strict_mode>

<completion_guidance>
After each batch:
- Show executed task IDs, commit/log evidence, and checkpoint status.
- If checkpoint PASS and there are remaining waves:
  - if authorization is required, stop in `WAITING_FOR_USER_AUTHORIZATION` and ask whether to continue.
  - otherwise, suggest/continue next batch per policy.
- If checkpoint BLOCKED: stop and show exact corrective actions.

If waiting on manual task:
- stop in `WAITING_FOR_HUMAN_ACTION` with checklist.
- keep manual task unchecked unless user explicitly confirms it started/completed.

After full execution success:
- Recommend final validation review and optionally `/clear` before any new planning cycle.

If blocked by spec resolution:
- Show the exact missing path and list discovered specs from `.spec-workflow/specs/`.
- Recommend the corrected command using one discovered spec name.
</completion_guidance>
