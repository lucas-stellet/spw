---
name: oraculo:tasks-check
description: Subagent-driven tasks.md validation (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<dispatch_pattern>
category: audit
subcategory: artifact
phase: planning
comms_path: planning/_comms/tasks-check
policy: (inlined below)

# Audit Dispatch Pattern

Multiple independent auditors examine the same artifact(s) from different angles.
An aggregator synthesizes their findings into a PASS/BLOCKED decision.

## Thin-Dispatch Rules

These rules are mandatory for all audit commands. They override any command-specific behavior.

### 1. Status-Only Reads

After dispatching any auditor, read ONLY `<auditor>/status.json`.
- If `status=pass`: proceed to next auditor or aggregator.
- If `status=blocked`: read `<auditor>/report.md` to decide action (log + continue, or stop BLOCKED).
- Never read `report.md` when status is pass.

### 2. Path-Based Briefs

When dispatching the aggregator:
- Write **filesystem paths** to all auditor `report.md` files in `aggregator/brief.md`.
- Never copy or summarize auditor report content into the brief.

### 3. Aggregator Reads From Filesystem

The aggregator receives a brief listing ALL auditor report paths.
It reads them directly from disk — the orchestrator does not relay content.

### 4. Run Structure

```
<phase>/_comms/<command>/run-NNN/
  <auditor-1>/brief.md, report.md, status.json
  <auditor-2>/brief.md, report.md, status.json
  <auditor-3>/brief.md, report.md, status.json
  <aggregator>/brief.md, report.md, status.json
  _handoff.md
```

### 5. Resume Policy

On `continue-unfinished`:
- Skip auditors where `status.json` exists with `status=pass`.
- Redispatch missing or blocked auditors.
- Always rerun aggregator.

## Dispatch Modes

Auditors may be dispatched in **parallel** (when fully independent) or **sequentially** (when one auditor informs another). The command workflow specifies the mode.

## Extension Points

Audit commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir, before first auditor dispatch. Use for precondition checks (e.g., verify target artifact exists).
- **`pre_dispatch(<auditor>)`**: Before writing a specific auditor's brief. Use for conditional skip logic.
- **`post_dispatch(<auditor>)`**: After reading an auditor's status.json. Use for early-exit decisions.
- **`post_pipeline`**: After aggregator completes, before writing _handoff.md. Use for artifact generation, next-step guidance.

</dispatch_pattern>

<shared_policies>
# Config Resolution

Canonical runtime config path is `.spec-workflow/oraculo.toml`.

Transitional compatibility:
- If `.spec-workflow/oraculo.toml` is missing, fallback to `.oraculo/oraculo.toml`.

When shell logic is required, prefer:
- `oraculo tools config-get <section.key> --default <value> [--raw]`

This keeps workflow behavior stable and avoids hardcoded path drift.

# File-First Handoff Contract

Required files for each dispatched subagent:
- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

If any required handoff file is missing, return `BLOCKED`.

**CRITICAL — Run-id format**: MUST be `run-NNN` (zero-padded 3-digit sequential).
Examples: `run-001`, `run-002`, `run-003`.
NEVER use dates, timestamps, or any other format (e.g. `run-20260209-1` is WRONG).
To create a new run, scan existing sibling directories, extract the highest NNN, and increment by 1.

## Thin-Dispatch Integration

This contract defines the **file structure**. The category-level dispatch policies define **how the orchestrator interacts** with these files:

- `dispatch-pipeline.md` — sequential chain, status-only reads, path-based briefs
- `dispatch-audit.md` — parallel auditors, aggregator reads from filesystem
- `dispatch-wave.md` — wave iteration, wave summaries, scout-based resume

The 5 core thin-dispatch rules apply on top of this contract:
1. Orchestrator reads only `status.json` after dispatch (never `report.md` on pass).
2. Briefs contain filesystem paths to prior reports (never content).
3. Synthesizers/aggregators read from disk directly.
4. Run structure follows category layout.
5. Resume skips completed subagents, always reruns final stage.

# Resume Policy

For commands with run folders:
- Detect the latest unfinished run before creating a new run.
- Ask user explicitly: `continue-unfinished` or `delete-and-restart`.
- Never auto-restart without explicit user decision.

# Skills Policy Canonical Notes

- Skill loading is always subagent-first.
- Enforce per stage via `skills.<stage>.enforce_required` (default: `true`).

# MCP Approval Reconciliation

Approval source of truth is MCP.

When `spec-status` is incomplete or ambiguous:
1. Resolve `approvalId` from `spec-status` fields.
2. If missing, inspect `.spec-workflow/approvals/<spec-name>/approval_*.json`.
3. If `approvalId` exists, call MCP `approvals status`.
4. Never infer approval from phase labels or `overallStatus` alone.

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

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the audit dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: verify tasks.md exists, skills, resume ......... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Verify `tasks.md` exists in SPEC_DIR; stop BLOCKED if missing → recommend `oraculo:tasks-plan <spec-name>`.
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
Resolve models from `.spec-workflow/oraculo.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<post_mortem_memory>
Resolve from `.spec-workflow/oraculo.toml` `[post_mortem_memory]`:
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
Resolve skill policy from `.spec-workflow/oraculo.toml`:
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
- Recommend next command: `oraculo:exec <spec-name> --batch-size <N>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Recommend fix path: update `tasks.md`, then rerun `oraculo:tasks-check <spec-name>`.
</completion_guidance>
