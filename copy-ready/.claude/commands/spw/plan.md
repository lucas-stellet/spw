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
  - Validates approval state through MCP and handles approval loop.
- `planning-stage-orchestrator` (model: complex_reasoning)
  - Runs stage-by-stage orchestration for design/tasks generation.
</subagents>

<preconditions>
- `requirements.md` exists for `<spec-name>`.
- If it does not exist, stop with BLOCKED and instruct: `run /spw:prd <spec-name>`.
- Do not assume approval from file existence; validate approval via MCP.
</preconditions>

<pipeline>
0. Dispatch `requirements-approval-gate`:
   - check MCP `spec-status` -> `documents.requirements.approved`
   - if not approved: request via `request-approval`, poll `get-approval-status`
   - if status = `rejected` or `changes-requested`: stop BLOCKED
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
- Provide corrective action and rerun command (`spw:plan <spec-name>` or specific stage command).
</completion_guidance>
