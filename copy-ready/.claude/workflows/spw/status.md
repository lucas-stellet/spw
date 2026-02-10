---
name: spw:status
description: Summarize current spec stage, blockers, and exact next commands
argument-hint: "[<spec-name>] [--all false|true]"
---

<objective>
Show where the workflow stopped and what to run next, with explicit approval/execution blockers.
</objective>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
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
