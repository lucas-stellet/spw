---
name: spw:tasks-check
description: Subagent-driven tasks.md validation (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<dispatch_pattern>
category: audit
subcategory: artifact
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
Validate whether `tasks.md` is ready for subagent execution.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/tasks.md`
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`

output:
- `.spec-workflow/specs/<spec-name>/planning/TASKS-CHECK.md`

comms:
- `.spec-workflow/specs/<spec-name>/planning/_comms/tasks-check/<run-id>/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — auditors + aggregator
     ============================================================ -->

<subagents>
- `traceability-auditor` (model: complex_reasoning)
- `dag-validator` (model: implementation)
- `test-policy-auditor` (model: complex_reasoning)
- `decision-aggregator` (model: complex_reasoning)
</subagents>

<subagent_artifact_map>
| Subagent | Artifact | Dispatch | Model |
|----------|----------|----------|-------|
| traceability-auditor | (report.md only) | task | complex_reasoning |
| dag-validator | (report.md only) | task | implementation |
| test-policy-auditor | (report.md only) | task | complex_reasoning |
| decision-aggregator | TASKS-CHECK.md | task | complex_reasoning |
</subagent_artifact_map>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the audit dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: verify tasks.md exists, skills, resume ......... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Verify `tasks.md` exists in SPEC_DIR; stop BLOCKED if missing → recommend `spw:tasks-plan <spec-name>`.
3. Apply skills policy: run design skills preflight and write `SKILLS-TASKS-CHECK.md`.
4. Load post-mortem memory inputs via `<post_mortem_memory>`.
5. Inspect existing tasks-check run dirs and apply resume decision gate.
</pre_pipeline>

<!-- post_pipeline: generate TASKS-CHECK.md ....................... -->
<post_pipeline>
1. Generate `.spec-workflow/specs/<spec-name>/planning/TASKS-CHECK.md` containing:
   - PASS/BLOCKED
   - findings by severity
   - recommended fixes
2. Write `<run-dir>/_handoff.md` linking all auditor/aggregator outputs.
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
   - `.spec-workflow/specs/<spec-name>/planning/SKILLS-TASKS-CHECK.md`
2. Avoid loading full skill content in main context (subagent-first).
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/spw/overlays/active/tasks-check.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] Every task references at least one requirement.
- [ ] Every requirement maps to at least one task.
- [ ] DAG has no cycles and wave order is valid.
- [ ] Test policy gate is satisfied.
- [ ] File-based handoff exists under `planning/_comms/tasks-check/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
- [ ] Orchestrator never read report.md from any auditor (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On PASS:
- Confirm output path: `.spec-workflow/specs/<spec-name>/planning/TASKS-CHECK.md`.
- Recommend next command: `spw:exec <spec-name> --batch-size <N>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Recommend fix path: update `tasks.md`, then rerun `spw:tasks-check <spec-name>`.
</completion_guidance>
