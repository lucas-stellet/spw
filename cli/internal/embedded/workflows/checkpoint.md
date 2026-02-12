---
name: spw:checkpoint
description: Subagent-driven quality gate between execution batches/waves
argument-hint: "<spec-name> [--scope batch|wave|phase]"
---

<dispatch_pattern>
category: audit
subcategory: code
phase: execution
comms_path: execution/waves/wave-{wave}/checkpoint
policy: @.claude/workflows/spw/shared/dispatch-audit.md
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
Validate that the executed batch truly meets spec intent, code quality, and integration safety before moving forward.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/tasks.md`
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`
- Implementation source files (evidence-collector)
- Git status (when clean-worktree gate enabled)

output:
- `.spec-workflow/specs/<spec-name>/execution/CHECKPOINT-REPORT.md`

comms:
- `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/checkpoint/<run-id>/`

Wave container:
- `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/`
- `_wave-summary.md`
- `_latest.json`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — auditors + aggregator
     ============================================================ -->

<subagents>
- `evidence-collector` (model: implementation)
  - Collects task state, test/lint/typecheck outputs, implementation logs, and git status.
- `traceability-judge` (model: complex_reasoning)
  - Verifies requirements/design/tasks alignment for delivered changes.
- `release-gate-decider` (model: complex_reasoning)
  - Produces final PASS/BLOCKED decision and corrective actions.
</subagents>

<subagent_artifact_map>
| Subagent | Artifact | Dispatch | Model |
|----------|----------|----------|-------|
| evidence-collector | (report.md only) | task | implementation |
| traceability-judge | (report.md only) | task | complex_reasoning |
| release-gate-decider | CHECKPOINT-REPORT.md | task | complex_reasoning |
</subagent_artifact_map>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the audit dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: wave-awareness, skills, resume ................. -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Apply skills policy: run implementation skills preflight and write `SKILLS-CHECKPOINT.md`.
3. Resolve current wave ID (`wave-<NN>`) and create canonical wave directory:
   - `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/`
4. Inspect existing checkpoint run dirs for the current wave and apply resume decision gate.
</pre_pipeline>

<!-- post_pipeline: generate CHECKPOINT-REPORT.md + wave updates .. -->
<post_pipeline>
1. Generate `.spec-workflow/specs/<spec-name>/execution/CHECKPOINT-REPORT.md` with:
   - status: PASS | BLOCKED
   - critical issues
   - corrective actions
   - recommended next step
   - implementation log coverage by task ID
2. Write `<run-dir>/_handoff.md` linking all subagent outputs.
3. Update wave-level pointers/summaries:
   - `<wave-dir>/_latest.json`
   - `<wave-dir>/_wave-summary.md`
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES
     ============================================================ -->

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
   - `.spec-workflow/specs/<spec-name>/execution/SKILLS-CHECKPOINT.md`
2. Avoid loading full skill content in main context (subagent-first).
3. If `[execution].tdd_default=true`, treat `test-driven-development` as required for this phase (effective required set).
4. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
5. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<git_gate>
Resolve from `.spec-workflow/spw-config.toml` `[execution].require_clean_worktree_for_wave_pass` (default `true`).

If enabled:
- include `git status --porcelain` evidence in the report
- return BLOCKED when uncommitted tracked changes exist
- recommend exact commit commands before rerunning checkpoint
</git_gate>

<implementation_log_gate>
Checkpoint must enforce implementation logs for completed tasks in the evaluated scope.

Rules:
- For every task marked `[x]` in the current batch/wave, there must be a corresponding implementation log entry.
- `evidence-collector` must output a mapping:
  - completed task IDs
  - implementation log IDs/paths
  - missing log entries (if any)
- If one or more completed tasks are missing implementation logs, return BLOCKED.
</implementation_log_gate>

<orchestrator_boundary>
The checkpoint orchestrator is a read-only observer. It MUST NOT:
- Create/modify/delete implementation logs, source files, or spec artifacts to resolve a BLOCKED auditor.
- Commit code or stage files on behalf of a blocked auditor.
- Mark tasks as complete/in-progress in tasks.md.
- Write files outside checkpoint comms (`execution/waves/wave-<NN>/checkpoint/<run-id>/`) and `CHECKPOINT-REPORT.md`.

On BLOCKED auditor: record in _handoff.md, propagate BLOCKED, report corrective actions, stop.
</orchestrator_boundary>

<gate_rule>
If status is BLOCKED, do not proceed to the next batch/wave.
</gate_rule>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/spw/overlays/active/checkpoint.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] File-based handoff exists under `execution/waves/wave-<NN>/checkpoint/<run-id>/`.
- [ ] `CHECKPOINT-REPORT.md` decision is traceable to subagent reports.
- [ ] Wave-level summary/pointers are updated (`_wave-summary.md`, `_latest.json`).
- [ ] Every completed task in scope has a corresponding implementation log entry.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
- [ ] Orchestrator never read report.md from any auditor (thin-dispatch).
- [ ] Orchestrator did not create/modify/delete any artifact outside checkpoint comms to resolve a BLOCKED auditor (anti-self-heal).
- [ ] No brief asserts codebase facts (file-handoff rule 2b).
</acceptance_criteria>

<completion_guidance>
On PASS:
- Show concise go/no-go summary and recommend next command: `spw:exec <spec-name>` (next batch/wave).

On BLOCKED:
- Show critical issues first, with exact corrective actions.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Recommend remediation command(s) and rerun: `spw:checkpoint <spec-name>`.
</completion_guidance>
