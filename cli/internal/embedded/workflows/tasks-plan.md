---
name: spw:tasks-plan
description: Subagent-driven tasks.md generation for waves, parallelism, and per-task TDD
argument-hint: "<spec-name> [--mode initial|next-wave] [--max-wave-size <N>] [--allow-no-test-exception true|false]"
---

<dispatch_pattern>
category: pipeline
subcategory: synthesis
phase: planning
comms_path: planning/_comms/tasks-plan
policy: @.claude/workflows/spw/shared/dispatch-pipeline.md
</dispatch_pattern>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
- @.claude/workflows/spw/shared/dispatch-implementation.md
- @.claude/workflows/spw/shared/dispatch-inline-audit.md
</shared_policies>

<objective>
Generate `.spec-workflow/specs/<spec-name>/tasks.md` for predictable parallel execution.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`
- `.spec-workflow/specs/<spec-name>/tasks.md` (required for `next-wave`; optional for reconciliation)
- `.spec-workflow/specs/<spec-name>/execution/CHECKPOINT-REPORT.md` (if present)
- post-mortem memory entries (if enabled)

output:
- `.spec-workflow/specs/<spec-name>/tasks.md`

comms:
- `.spec-workflow/specs/<spec-name>/planning/_comms/tasks-plan/run-NNN/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

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

<subagent_artifact_map>
| Subagent | Artifact | Dispatch | Model |
|----------|----------|----------|-------|
| task-decomposer | (report.md only) | task | complex_reasoning |
| dependency-graph-builder | (report.md only) | task | complex_reasoning |
| parallel-conflict-checker | (report.md only) | task | implementation |
| test-policy-enforcer | (report.md only) | task | complex_reasoning |
| tasks-writer | tasks.md | task | implementation |
</subagent_artifact_map>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the pipeline dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: mode selection, skills, resume .................. -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Apply skills policy: run design skills preflight and write `SKILLS-TASKS-PLAN.md`.
3. Resolve effective planning behavior (per `<mode_policy>` + `<planning_defaults>`):
   - resolve effective `max_wave_size`
   - resolve effective generation mode (`initial`, `next-wave`, or `all-at-once`)
   - validate preconditions per mode
4. Load post-mortem memory inputs via `<post_mortem_memory>`.
5. Read templates:
   - `.spec-workflow/user-templates/tasks-template.md` (preferred)
   - fallback: `.spec-workflow/templates/tasks-template.md`
6. Use config values from dispatch-init output:
   - `execution.tdd_default` is the source of truth for TDD policy.
   - If template contains `tdd_default: managed-by-config`, resolve to
     the dispatch-init value when writing the final artifact.
   - dispatch-setup already injects Config Context into each brief —
     do not override or restate these values in ## Task sections.
7. Inspect existing tasks-plan run dirs and apply resume decision gate.
</pre_pipeline>

<!-- post_pipeline: inline audit + dashboard compatibility + approval  -->
<post_pipeline>

### Inline Audit

After tasks-writer produces tasks.md, validate it before proceeding to approval.
Follow @.claude/workflows/spw/shared/dispatch-inline-audit.md.

1. `spw tools audit-iteration start --run-dir <RUN_DIR> --type inline-audit`
2. `spw tools dispatch-init-audit --run-dir <RUN_DIR> --type inline-audit --iteration 1`
3. Inside the audit dir, dispatch auditors via dispatch-setup:
   - traceability-auditor (complex_reasoning model)
   - dag-validator (implementation model)
   - test-policy-auditor (implementation model)
4. Dispatch decision-aggregator to synthesize results
5. `spw tools dispatch-read-status decision-aggregator --run-dir <audit_dir>`
6. If PASS → proceed to dashboard verification and approval below
7. If BLOCKED:
   a. `spw tools audit-iteration check --run-dir <RUN_DIR> --type inline-audit`
   b. If allowed: `spw tools audit-iteration advance --run-dir <RUN_DIR> --type inline-audit --result blocked`, re-dispatch tasks-writer with aggregator report path as feedback, then repeat from step 2 with next iteration
   c. If exhausted: STOP, recommend `spw:tasks-check <spec-name>`

### Dashboard Verification and Approval

1. Verify `tasks.md` satisfies `<dashboard_markdown_profile>`.
2. Write `<run-dir>/_handoff.md` with mode decisions, DAG rationale, conflict/test policy outcomes.
3. Handle approval via MCP only:
   - If MCP tools are unavailable or fail: log WARNING to `_handoff.md` per `<approval_reconciliation>` § MCP Unavailability, stop `WAITING_FOR_APPROVAL`.
   - call `spec-status`, resolve via `<approval_reconciliation>`
   - if approved: continue
   - if `needs-revision`/`changes-requested`/`rejected`: stop BLOCKED
   - if pending: stop with `WAITING_FOR_APPROVAL`
   - only if never requested: call `request-approval` then `get-approval-status` once
   - never ask for approval in chat
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

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
3. Load selected reports and inject lessons into decomposition/wave/test policy decisions.

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
   - `.spec-workflow/specs/<spec-name>/planning/SKILLS-TASKS-PLAN.md`
2. Avoid loading full skill content in main context (subagent-first).
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

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

<approval_reconciliation>
Resolve tasks approval with MCP-first reconciliation:
- Primary source:
  - `documents.tasks.approved`
  - `documents.tasks.status`
  - `approvals.tasks.status`
  - optional IDs:
    - `documents.tasks.approvalId`
    - `approvals.tasks.approvalId`
    - `approvals.tasks.id`
- If status is missing/unknown or inconsistent, fallback:
  1. Resolve approval ID from `spec-status` fields above.
  2. If still missing, read latest `.spec-workflow/approvals/<spec-name>/approval_*.json`
     where `filePath` is `.spec-workflow/specs/<spec-name>/tasks.md`.
  3. If approval ID exists, call MCP `approvals status` and use it as source of truth.
  4. If approval ID does not exist, treat as not requested.
- Never infer approval from `overallStatus`/phase labels alone.
</approval_reconciliation>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/spw/overlays/active/tasks-plan.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

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
- [ ] File-based handoff exists under `planning/_comms/tasks-plan/run-NNN/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/tasks.md`.
- Confirm effective planning mode (`initial`, `next-wave`, or `all-at-once`).
- Confirm effective `max_wave_size` source (CLI override or `[planning].max_wave_size`).
- Confirm approval request status for tasks.
- Confirm inline audit result (PASS or iteration count).
- Recommend next command: `spw:tasks-check <spec-name>` (for re-validation or CI gate).

If blocked:
- Show mode/precondition/decomposition/dependency/conflict/test-policy failures.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide rerun command:
  - default from config: `spw:tasks-plan <spec-name>`
  - explicit rolling override: `spw:tasks-plan <spec-name> --mode initial --max-wave-size <N>`
  - explicit rolling append: `spw:tasks-plan <spec-name> --mode next-wave --max-wave-size <N>`.
</completion_guidance>
