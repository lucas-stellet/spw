---
name: spw:tasks-check
description: Subagent-driven tasks.md validation (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<objective>
Validate whether `tasks.md` is ready for subagent execution.
</objective>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
</shared_policies>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder (`<run-id>` MUST be `run-NNN` format — e.g. `run-001`, never dates):
- `.spec-workflow/specs/<spec-name>/_agent-comms/tasks-check/<run-id>/`

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

After validation, write:
- `<run-dir>/_handoff.md` (orchestrator summary of audit results)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<resume_policy>
Before creating a new run, inspect existing tasks-check run folders:
- `.spec-workflow/specs/<spec-name>/_agent-comms/tasks-check/<run-id>/`

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
- Reuse completed auditor outputs (`report.md` + `status.json` with `status=pass`).
- Redispatch only missing/blocked auditors.
- Always rerun `decision-aggregator` before final PASS/BLOCKED output.

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

<post_mortem_memory>
Resolve from `.spec-workflow/spw-config.toml` `[post_mortem_memory]`:
- `enabled` (default `true`)
- `max_entries_for_design` (default `5`)

If enabled and index exists:
1. Read `.spec-workflow/post-mortems/INDEX.md`.
2. Select up to `max_entries_for_design` relevant entries:
   - same `<spec-name>` first
   - then by tag/topic similarity and recency
3. Load selected reports and expand audit checks for previously missed issues.

If index/report files are missing, continue with warning (non-blocking).
</post_mortem_memory>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/_generated/SKILLS-TASKS-CHECK.md`
2. Avoid loading full skill content in main context (subagent-first).
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<subagents>
- `traceability-auditor` (model: complex_reasoning)
- `dag-validator` (model: implementation)
- `test-policy-auditor` (model: complex_reasoning)
- `decision-aggregator` (model: complex_reasoning)
</subagents>

<workflow>
1. Run design skills preflight (availability) and write `SKILLS-TASKS-CHECK.md`.
2. Inspect existing tasks-check run dirs and apply `<resume_policy>` decision gate.
3. Determine active run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create:
     `.spec-workflow/specs/<spec-name>/_agent-comms/tasks-check/<run-id>/`
4. Read `.spec-workflow/specs/<spec-name>/tasks.md` + requirements/design docs + post-mortem memory inputs via `<post_mortem_memory>`.
5. Write briefs (including required skills per role) and dispatch in parallel:
   - `traceability-auditor`
   - `dag-validator`
   - `test-policy-auditor`
   - if resuming, redispatch only missing/blocked auditors
6. Require auditor `report.md` + `status.json` (with skill fields); BLOCKED if missing.
7. Dispatch `decision-aggregator` with file handoff to produce PASS/BLOCKED decision.
   - if resuming, always rerun `decision-aggregator`
8. Generate `.spec-workflow/specs/<spec-name>/_generated/TASKS-CHECK.md` containing:
   - PASS/BLOCKED
   - findings by severity
   - recommended fixes
9. Write `<run-dir>/_handoff.md` linking all auditor/aggregator outputs and resume decision taken (`continue-unfinished` or `delete-and-restart`).
</workflow>

<acceptance_criteria>
- [ ] Every task references at least one requirement.
- [ ] Every requirement maps to at least one task.
- [ ] DAG has no cycles and wave order is valid.
- [ ] Test policy gate is satisfied.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/_agent-comms/tasks-check/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
</acceptance_criteria>

<completion_guidance>
On PASS:
- Confirm output path: `.spec-workflow/specs/<spec-name>/_generated/TASKS-CHECK.md`.
- Recommend next command: `spw:exec <spec-name> --batch-size <N>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Recommend fix path: update `tasks.md`, then rerun `spw:tasks-check <spec-name>`.
</completion_guidance>
