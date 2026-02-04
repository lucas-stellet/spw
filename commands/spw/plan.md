---
name: spw:plan
description: Technical planning from existing requirements, with MCP approval gate
argument-hint: "<spec-name> [--max-wave-size 3]"
---

<objective>
Run the full technical planning flow for a spec that already has an existing `requirements.md`.
Before continuing, this command validates and, if needed, requests approval via MCP.
This command does NOT create a PRD and does NOT run initial product discovery.
</objective>

<when_to_use>
- Use when the spec already has product context and an existing `requirements.md`.
- Expected input: `.spec-workflow/specs/<spec-name>/requirements.md`.
</when_to_use>

<preconditions>
- `requirements.md` exists for `<spec-name>`.
- If it does not exist, stop with BLOCKED and instruct: `run /spw:prd <spec-name>`.
- Do not assume approval from file existence; validate approval via MCP.
</preconditions>

<pipeline>
0. Validate existence of `.spec-workflow/specs/<spec-name>/requirements.md`.
0.1 Validate status via MCP `spec-status`:
    - check `documents.requirements.approved`
0.2 If not approved:
    - request approval via MCP `request-approval` for `docType: "requirements"`
    - ask user to review in dashboard/UI
    - poll with `get-approval-status`
    - continue only when status = `approved`
    - if status = `rejected` or `changes-requested`, stop with BLOCKED
0.3 If already approved:
    - continue pipeline
1. `spw:design-research <spec-name>`
2. `spw:design-draft <spec-name>`
3. `spw:tasks-plan <spec-name> --max-wave-size <N>`
4. `spw:tasks-check <spec-name>`
</pipeline>

<rules>
- If `tasks-check` returns BLOCKED, fix `tasks.md` and run the check again.
- Do not start code execution without approved design and tasks.
- Do not "infer requirements" in this command; requirements come from PRD/requirements with MCP approval.
- Mandatory gate: requirements without MCP approval blocks `spw:plan`.
</rules>
