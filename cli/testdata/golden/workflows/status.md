---
name: spw:status
description: Summarize current spec stage, blockers, and exact next commands
argument-hint: "[<spec-name>] [--all false|true]"
---

<objective>
Show where the workflow stopped and what to run next, with explicit approval/execution blockers.
</objective>

<shared_policies>
# Config Resolution

Canonical runtime config path is `.spec-workflow/spw-config.toml`.

Transitional compatibility:
- If `.spec-workflow/spw-config.toml` is missing, fallback to `.spw/spw-config.toml`.

When shell logic is required, prefer:
- `spw tools config-get <section.key> --default <value> [--raw]`

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

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<subagents>
- `state-inspector` (model: implementation)
  - Inspects artifacts, task progress, and wave state.
- `approval-auditor` (model: implementation)
  - Reads approval status through MCP only.
- `next-step-planner` (model: complex_reasoning)
  - Produces ordered, minimal next actions.
</subagents>

<scope_resolution>
1. If `<spec-name>` is provided, use it.
2. Otherwise, inspect `.spec-workflow/specs/*`:
   - if only one spec exists, use it.
   - if multiple specs exist and `--all=true`, summarize all.
   - if multiple specs exist and `--all=false`, ask user to choose one via AskUserQuestion.
</scope_resolution>

<approval_reconciliation>
Approval state must be resolved with MCP-first reconciliation (never from stale summaries):

For each document (`requirements`, `design`, `tasks`):
1. Call `spec-status` and read (case-insensitive):
   - `documents.<doc>.approved`
   - `documents.<doc>.status`
   - `approvals.<doc>.status`
   - optional IDs:
     - `documents.<doc>.approvalId`
     - `approvals.<doc>.approvalId`
     - `approvals.<doc>.id`
2. If status is missing/unknown OR conflicts with artifact reality, run fallback:
   - resolve approval ID in this order:
     - ID returned by `spec-status`
     - latest `.spec-workflow/approvals/<spec-name>/approval_*.json` with matching `filePath`
   - if ID exists: call MCP `approvals status` and use that result as source of truth
   - if no ID exists: treat as not requested
3. Never infer approval state from:
   - `overallStatus` or phase labels alone
   - `.spec-workflow/specs/<spec-name>/STATUS-SUMMARY.md`
</approval_reconciliation>

<workflow>
1. Resolve target spec(s) from `.spec-workflow/specs/`.
2. For each spec, dispatch `state-inspector` to collect:
   - artifact presence: `requirements.md`, `design/DESIGN-RESEARCH.md`, `design.md`, `tasks.md`
   - tasks progress counts: `[ ]`, `[-]`, `[x]`
   - active wave/blocked/manual markers when present
   - wave comms state from:
     - `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/_latest.json`
     - `.spec-workflow/specs/<spec-name>/execution/waves/wave-<NN>/_wave-summary.md`
   - legacy comms paths present (warn user to re-run spec with current SPW):
     - `.spec-workflow/specs/<spec-name>/_agent-comms/` (pre-2.0 layout; current layout uses `<phase>/_comms/`)
3. Dispatch `approval-auditor`:
   - call `spec-status`
   - resolve document approval state via `<approval_reconciliation>`
4. Dispatch `next-step-planner` to classify stage:
   - product discovery
   - design research
   - design approval
   - tasks planning/approval
   - execution (wave/batch)
   - done
5. Produce:
   - console summary (current stage, blockers, next command)
   - `.spec-workflow/specs/<spec-name>/STATUS-SUMMARY.md` for each inspected spec (output-only snapshot; never reuse as approval input)
</workflow>

<output_contract>
Always include:
- Current stage (single label)
- Why this stage was inferred (artifacts + approval status)
- Blocking reason (if any)
- Exact next command(s), in order

When execution is in progress, also include:
- task counters (`todo/in-progress/done`)
- whether human authorization is required between waves
- current wave folder path and latest run pointers (`execution/checkpoint/post-check`)
</output_contract>

<completion_guidance>
If status is clear:
- Recommend exactly one immediate next command.

If blocked:
- Show exact unblock action and rerun command (`spw:status <spec-name>`).

If multiple specs:
- Recommend explicit spec target in next command (`<spec-name>`).
</completion_guidance>
