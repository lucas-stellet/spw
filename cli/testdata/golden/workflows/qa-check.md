---
name: spw:qa-check
description: Validate QA test plan selectors, traceability, and data feasibility against actual code
argument-hint: "<spec-name>"
---

<dispatch_pattern>
category: audit
subcategory: code
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

Canonical runtime config path is `.spec-workflow/spw-config.toml`.

Transitional compatibility:
- If `.spec-workflow/spw-config.toml` is missing, fallback to `.spw/spw-config.toml`.

When shell logic is required, prefer:
- `spw tools config-get <section.key> --default <value> [--raw]`

This keeps workflow behavior stable during migration and avoids hardcoded path drift.

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
Validate QA test plan against actual code. Confirm selectors/endpoints exist, verify requirement traceability, and check data feasibility. Produces a verified selector map that `spw:qa-exec` consumes without re-reading source files.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`
- Implementation source files (selector-verifier only)

output:
- `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md`

comms:
- `.spec-workflow/specs/<spec-name>/qa/_comms/qa-check/<run-id>/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — auditors + aggregator
     ============================================================ -->

<subagents>
- `qa-traceability-auditor` (model: complex_reasoning)
  - Performs bidirectional mapping: scenario ↔ requirement.
  - Flags untested requirements and orphaned scenarios.
- `qa-selector-verifier` (model: implementation)
  - Searches source files for each selector, route, and endpoint in the test plan.
  - Confirms existence and outputs a verified map: test-id → selector → file:line.
  - This is the ONLY subagent that reads implementation source files.
- `qa-data-feasibility-checker` (model: implementation)
  - Validates test data assumptions: seeds, fixtures, accounts, environment variables.
  - Flags missing or unreachable test prerequisites.
- `qa-check-aggregator` (model: complex_reasoning)
  - Consumes all auditor/verifier reports.
  - Produces PASS/BLOCKED decision with severity-ranked findings.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the audit dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: verify QA-TEST-PLAN.md exists .................. -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>` and stop BLOCKED if missing.
2. Verify `qa/QA-TEST-PLAN.md` exists in SPEC_DIR; stop BLOCKED if missing → recommend `spw:qa <spec-name>`.
3. Inspect existing qa-check run dirs and apply resume decision gate.
4. Read context files:
   - `qa/QA-TEST-PLAN.md`
   - `requirements.md`
   - `design.md`
</pre_pipeline>

<!-- post_pipeline: generate QA-CHECK.md .......................... -->
<post_pipeline>
1. Generate `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md` containing:
   - PASS or BLOCKED status
   - Verified selector map (test-id → selector/endpoint → file:line)
   - Traceability percentage (tested requirements / total requirements)
   - Data feasibility assessment
   - Severity-ranked findings
2. Write `<run-dir>/_handoff.md` linking all auditor outputs.
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

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] QA-TEST-PLAN.md was read and validated against source code.
- [ ] Every selector/endpoint in the plan was verified against implementation files.
- [ ] Bidirectional traceability (scenario ↔ requirement) was assessed.
- [ ] Test data/fixture feasibility was checked.
- [ ] PASS/BLOCKED decision is justified by aggregated findings.
- [ ] Verified selector map is included in QA-CHECK.md.
- [ ] File-first handoff exists under `qa/_comms/qa-check/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision was respected.
- [ ] Orchestrator never read report.md from any auditor (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On PASS:
- Confirm output path: `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md`.
- Recommend next command: `spw:qa-exec <spec-name>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- If selectors are missing or invalid, recommend updating the test plan and rerunning `spw:qa-check <spec-name>`.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
</completion_guidance>
