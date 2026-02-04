---
name: spw:plan
description: Technical planning from existing requirements, orchestrated by subagents
argument-hint: "<spec-name> [--max-wave-size 3]"
---

<objective>
Run the full technical planning flow for a spec with existing `requirements.md`, using subagents to keep main-context size stable and process deterministic.
</objective>

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
- Always do this sequence:
  1) call `spec-status`
  2) if `documents.requirements.approved == true`: proceed immediately
  3) if not approved:
     - call `request-approval` (idempotent)
     - call `get-approval-status` once
     - stop with `WAITING_FOR_APPROVAL` (or BLOCKED on rejected/changes-requested)
- If status is `pending`, do not poll in a loop; instruct user to approve in Spec Workflow UI and rerun the command.
</approval_protocol>

<pipeline>
0. Dispatch `requirements-approval-gate`:
   - run the approval protocol above
1. Dispatch `planning-stage-orchestrator` for:
   - `spw:design-research <spec-name>`
   - `spw:design-draft <spec-name>`
   - `spw:tasks-plan <spec-name> --max-wave-size <N>`
   - `spw:tasks-check <spec-name>`
2. If `tasks-check` is BLOCKED, revise and repeat stage 1 as needed.
</pipeline>

<rules>
- Mandatory gate: requirements without MCP approval blocks `spw:plan`.
- Do not infer requirements in this command.
- Do not start execution before design/tasks are approved.
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
- Provide corrective action and rerun command (`spw:plan <spec-name>` or specific stage command).
</completion_guidance>
