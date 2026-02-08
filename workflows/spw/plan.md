---
name: spw:plan
description: Technical planning from existing requirements, orchestrated by subagents
argument-hint: "<spec-name> [--max-wave-size <N>]"
---

<objective>
Run the full technical planning flow for a spec with existing `requirements.md`, using subagents to keep main-context size stable and process deterministic.
</objective>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
</shared_policies>

<when_to_use>
- Use when the spec already has product context and an existing `requirements.md`.
- Expected input: `.spec-workflow/specs/<spec-name>/requirements.md`.
</when_to_use>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- web_research -> default `haiku`
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<planning_defaults>
Resolve planning defaults from `.spec-workflow/spw-config.toml` `[planning]`:
- `tasks_generation_strategy` (`rolling-wave|all-at-once`, default `rolling-wave`)
- `max_wave_size` (default `3`)

Rule:
- Call `spw:tasks-plan` without forcing `--mode`/`--max-wave-size`, unless user explicitly asked for overrides.
</planning_defaults>

<subagents>
- `requirements-approval-gate` (model: complex_reasoning)
  - Validates approval state strictly through MCP only.
- `planning-stage-orchestrator` (model: complex_reasoning)
  - Runs stage-by-stage orchestration for design/tasks generation.
</subagents>

<preconditions>
- `requirements.md` exists for `<spec-name>`.
- If it does not exist, stop with BLOCKED and instruct: `run /spw:prd <spec-name>`.
- Do not assume approval from file existence; validate approval via MCP.
</preconditions>

<approval_protocol>
- Approval source of truth is MCP status only.
- Never ask for approval in chat (no AskUserQuestion/manual "approve now?" options).
- Resolve requirements approval from MCP using both boolean and status fields:
  - `documents.requirements.approved`
  - `documents.requirements.status`
  - `approvals.requirements.status`
- When `spec-status` returns missing/unknown/inconsistent requirements status, reconcile before deciding:
  1) resolve approval ID from `spec-status` fields (if present):
     - `documents.requirements.approvalId`
     - `approvals.requirements.approvalId`
     - `approvals.requirements.id`
  2) if still missing, read latest `.spec-workflow/approvals/<spec-name>/approval_*.json`
     where `filePath` is `.spec-workflow/specs/<spec-name>/requirements.md`
  3) if approval ID exists, call MCP `approvals status` and use that result as source of truth
  4) if no approval ID exists, treat as not requested
- Never infer approval from `overallStatus` or phase labels alone.
- Treat status values case-insensitively:
  - approved: `approved`
  - pending: `pending`
  - needs revision: `needs-revision`, `changes-requested`, `rejected`
  - not requested: missing/empty/unknown
- Always do this sequence:
  1) call `spec-status`
  2) if status is approved: proceed immediately (never re-request approval)
  3) if status is needs revision: stop BLOCKED and route to `spw:prd <spec-name>` revision protocol (never request approval first)
  4) if status is pending: stop with `WAITING_FOR_APPROVAL` and instruct UI approval + rerun
  5) only if status is not requested:
     - call `request-approval` (idempotent)
     - call `get-approval-status` once
     - if approved: proceed
     - if pending: stop with `WAITING_FOR_APPROVAL`
     - if needs revision: stop BLOCKED and route to `spw:prd <spec-name>`
- If status is `pending`, do not poll in a loop; instruct user to approve in Spec Workflow UI and rerun the command.
</approval_protocol>

<pipeline>
0. Dispatch `requirements-approval-gate`:
   - run the approval protocol above
1. Dispatch `planning-stage-orchestrator` for:
   - `spw:design-research <spec-name>`
   - `spw:design-draft <spec-name>`
   - `spw:tasks-plan <spec-name>`
   - `spw:tasks-check <spec-name>`
2. If `tasks-check` is BLOCKED, revise and repeat stage 1 as needed.
</pipeline>

<ui_approval_docs_policy>
Documents reviewed/approved in Spec Workflow UI must follow strict markdown profiles:
- `requirements.md`: render-safe markdown (valid headings/tables/fences, no task-style checkboxes).
- `design.md`: render-safe markdown plus at least one valid fenced lowercase Mermaid block in `## Architecture`.
- `tasks.md`: strict dashboard parser profile (`- [ ]/-[-]/-[x]`, numeric unique IDs, parseable metadata, prompt closure).

If any stage output violates its profile, stop BLOCKED and route user to rerun the specific stage command.
</ui_approval_docs_policy>

<artifact_boundary>
Planning artifacts must stay under:
- `.spec-workflow/specs/<spec-name>/`

Research/supporting files must stay under:
- `.spec-workflow/specs/<spec-name>/research/`

Do not generate planning/research artifacts under generic folders like `docs/`.
</artifact_boundary>

<rules>
- Mandatory gate: requirements without MCP approval blocks `spw:plan`.
- Do not infer requirements in this command.
- Do not start execution before design/tasks are approved.
- Enforce `<ui_approval_docs_policy>` for stage outputs before advancing.
</rules>

<completion_guidance>
On success:
- Summarize generated artifacts (`DESIGN-RESEARCH.md`, `design.md`, `tasks.md`, `TASKS-CHECK.md`).
- Confirm approval state for design/tasks.
- Recommend next command: `spw:exec <spec-name> --batch-size <N>`.
- Recommend running `/clear` before `spw:exec` for fresh execution context.

If blocked:
- Show exactly which stage blocked (approval gate, design, tasks-plan, tasks-check).
- If waiting on approval, explicitly state: "Approve in Spec Workflow UI, then rerun `/spw:plan <spec-name>`."
- If requirements are `changes-requested`/`rejected`, route to `spw:prd <spec-name>` revision protocol before planning.
- Provide corrective action and rerun command (`spw:plan <spec-name>` or specific stage command).
</completion_guidance>
