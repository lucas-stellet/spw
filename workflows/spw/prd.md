---
name: spw:prd
description: Zero-to-PRD discovery flow with subagents to generate requirements.md
argument-hint: "<spec-name> [--source <url-or-file.md>]"
---

<objective>
Create or update `.spec-workflow/specs/<spec-name>/requirements.md` in PRD format using a subagent-first process.

This command combines:
- GSD strengths: v1/v2/out-of-scope scoping, REQ-IDs, testable criteria, traceability.
- superpowers strengths: one-question-at-a-time discovery, recommendation + trade-off framing, incremental validation.
</objective>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
</shared_policies>

<when_to_use>
- Use when the spec does NOT have approved requirements yet (zero-to-PRD).
- Use when requirements need to be revisited with new product sources.
</when_to_use>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- web_research -> default `haiku`
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<post_mortem_memory>
Resolve from `.spec-workflow/spw-config.toml` `[post_mortem_memory]`:
- `enabled` (default `true`)
- `max_entries_for_design` (default `5`)
- `prefer_same_spec` (default `true`)

If enabled and index exists:
1. Read `.spec-workflow/post-mortems/INDEX.md`.
2. Select up to `max_entries_for_design` relevant entries:
   - same `<spec-name>` first when `prefer_same_spec=true`
   - then by tag/topic similarity and recency
3. Load selected reports and extract reusable guardrails for PRD quality.

If index/report files are missing, continue with warning (non-blocking).
</post_mortem_memory>

<subagents>
- `source-reader-web` (model: web_research)
  - Reads URLs and extracts only requirement-relevant signals.
- `source-reader-mcp` (model: implementation)
  - Reads MCP-backed sources (GitHub/Linear/ClickUp) and normalizes output.
- `feedback-analyzer` (model: complex_reasoning)
  - Converts approval comments into concrete requirement deltas and open questions.
- `codebase-impact-scanner` (model: implementation)
  - Checks feasibility/impact against existing code patterns and boundaries.
- `revision-planner` (model: complex_reasoning)
  - Produces an explicit revision plan before any document edits.
- `requirements-structurer` (model: complex_reasoning)
  - Produces v1/v2/out-of-scope, REQ-IDs, acceptance criteria draft.
- `prd-editor` (model: implementation)
  - Writes final PRD into template format.
- `prd-critic` (model: complex_reasoning)
  - Performs strict quality gate before approval request.
</subagents>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder:
- `.spec-workflow/specs/<spec-name>/_agent-comms/prd/<run-id>/`

For each subagent, use:
- `<run-dir>/<subagent>/brief.md` (written by orchestrator before dispatch)
- `<run-dir>/<subagent>/report.md` (written by subagent after execution)
- `<run-dir>/<subagent>/status.json` (written by subagent)

Status schema (minimum):
- `status`: `pass|blocked`
- `summary`: short result
- `inputs`: key files/URLs used
- `outputs`: generated artifacts
- `open_questions`: unresolved items

For revision loops, also create:
- `.spec-workflow/specs/<spec-name>/_agent-comms/prd-revision/<run-id>/`

After each phase, write:
- `<run-dir>/_handoff.md` (orchestrator synthesis of subagent outputs)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<resume_policy>
Before creating a new run, inspect existing PRD run folders:
- `.spec-workflow/specs/<spec-name>/_agent-comms/prd/<run-id>/`
- during revision protocol: `.spec-workflow/specs/<spec-name>/_agent-comms/prd-revision/<run-id>/`

A run is `unfinished` when any of these is true:
- `_handoff.md` is missing
- any subagent directory is missing `brief.md`, `report.md`, or `status.json`
- any subagent `status.json` reports `status=blocked`

Resume decision gate (mandatory):
1. Find latest unfinished run (if multiple, sort by mtime descending and use the newest).
2. If found, ask user once (AskUserQuestion) with options:
   - `continue-unfinished` (Recommended): continue with that run directory.
   - `delete-and-restart`: delete that unfinished run directory and start a new run.
3. Never choose automatically. Do not infer user intent.
4. If explicit user decision is unavailable, stop with `WAITING_FOR_USER_DECISION`.
5. Do not create a new run-id before this decision.

If user chooses `continue-unfinished`:
- Reuse completed subagent outputs (`report.md` + `status.json` with `status=pass`).
- Redispatch only missing/blocked subagents.
- Always rerun `prd-critic` before final approval request.

If user chooses `delete-and-restart`:
- Delete the selected unfinished run dir.
- Continue workflow with a fresh run-id.
- Record deleted path in final output.
</resume_policy>

<revision_protocol>
Trigger this protocol when either:
- MCP approval status for requirements is `changes-requested` or `rejected`, or
- user asks to analyze/adjust reviewed requirements.

Protocol (mandatory):
1. Inspect existing `prd-revision` run dirs and apply `<resume_policy>` decision gate.
2. Determine active revision run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create:
     `.spec-workflow/specs/<spec-name>/_agent-comms/prd-revision/<run-id>/`
3. Read approval feedback from MCP and existing `requirements.md`.
4. Dispatch `feedback-analyzer` (with file handoff) to classify:
   - accepted changes
   - ambiguous/conflicting feedback
   - out-of-scope suggestions
   - if resuming, redispatch only when output is missing/blocked
5. Dispatch `codebase-impact-scanner` (if enabled in config `[reviews]`) with file handoff.
   - if resuming, redispatch only when output is missing/blocked
6. Dispatch `revision-planner` with file handoff to create:
   - `.spec-workflow/specs/<spec-name>/_generated/PRD-REVISION-PLAN.md`
   - `.spec-workflow/specs/<spec-name>/_generated/PRD-REVISION-QUESTIONS.md` (if needed)
   - if resuming, redispatch only when output is missing/blocked
7. Ask targeted clarification questions before editing if ambiguity/conflict exists.
8. Only after clarification, dispatch `prd-editor` with file handoff to apply approved deltas.
9. Save revision summary:
   - `.spec-workflow/specs/<spec-name>/_generated/PRD-REVISION-NOTES.md`
10. Write revision handoff:
   - `.spec-workflow/specs/<spec-name>/_agent-comms/prd-revision/<run-id>/_handoff.md`
   - include resume decision taken (`continue-unfinished` or `delete-and-restart`)

Never directly edit requirements immediately after reading review comments.
</revision_protocol>

<source_handling>
If `--source` is provided and looks like a URL (`http://` or `https://`) or markdown (`.md`), run a source-reading gate:

1. Ask with AskUserQuestion:
   - header: "Source"
   - question: "I detected an external source. Do you want to use a specific MCP to read it?"
   - options:
     - "Yes, choose MCP (Recommended)" — explicit connector selection
     - "Auto" — try compatible MCP first, fallback to direct read
     - "No" — read without MCP

2. If user selects "Yes, choose MCP", ask:
   - header: "MCP"
   - question: "Which MCP should be used for this source?"
   - options:
     - "GitHub"
     - "Linear"
     - "ClickUp"
     - "Web/Browser"
     - "Local markdown file"

3. If selected MCP is unavailable, clearly report and ask fallback:
   - "Read without MCP"
   - "Choose another MCP"
</source_handling>

<ui_approval_markdown_profile>
`requirements.md` must stay render-safe and review-friendly in Spec Workflow UI:
- Use plain Markdown (avoid raw HTML blocks unless strictly necessary).
- Use ATX headings (`#`, `##`, `###`) with consistent hierarchy.
- Keep tables valid with explicit header separator rows.
- Keep fenced code blocks balanced and language-tagged when applicable.
- Keep emphasis/underscore delimiters balanced (no dangling `_` or `**`).
- Avoid task-style checkboxes in requirements content (`- [ ]`, `- [-]`, `- [x]`).
- Keep requirement IDs canonical and unique (`REQ-001`, `REQ-002`, ...).
</ui_approval_markdown_profile>

<approval_reconciliation>
Resolve requirements approval with MCP-first reconciliation:
- Primary source:
  - `documents.requirements.approved`
  - `documents.requirements.status`
  - `approvals.requirements.status`
  - optional IDs:
    - `documents.requirements.approvalId`
    - `approvals.requirements.approvalId`
    - `approvals.requirements.id`
- If status is missing/unknown or inconsistent, fallback:
  1. Resolve approval ID from `spec-status` fields above.
  2. If still missing, read latest `.spec-workflow/approvals/<spec-name>/approval_*.json`
     where `filePath` is `.spec-workflow/specs/<spec-name>/requirements.md`.
  3. If approval ID exists, call MCP `approvals status` and use it as source of truth.
  4. If approval ID does not exist, treat as not requested.
- Never infer approval from `overallStatus`/phase labels alone.
</approval_reconciliation>

<workflow>
1. Inspect existing `prd` run dirs and apply `<resume_policy>` decision gate.
2. Determine active run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create:
     `.spec-workflow/specs/<spec-name>/_agent-comms/prd/<run-id>/`
3. Read existing context:
   - `.spec-workflow/specs/<spec-name>/requirements.md` (if present)
   - `.spec-workflow/specs/<spec-name>/design.md` (if present)
   - `.spec-workflow/steering/*.md` (if present)
   - post-mortem memory inputs via `<post_mortem_memory>`
4. If `--source` is present, write briefs and dispatch source-reader subagents:
   - web-only fetches -> `source-reader-web`
   - MCP-backed reads -> `source-reader-mcp`
   - save normalized notes to `.spec-workflow/specs/<spec-name>/_generated/PRD-SOURCE-NOTES.md`
   - if resuming, redispatch only when output is missing/blocked
5. Require source-reader `report.md` + `status.json`; BLOCKED if missing.
6. Run one-question-at-a-time discovery with user.
7. Dispatch `requirements-structurer` with file handoff to produce a structured draft:
   - `.spec-workflow/specs/<spec-name>/_generated/PRD-STRUCTURE.md`
   - if resuming, redispatch only when output is missing/blocked
8. Dispatch `prd-editor` with file handoff to fill template using:
   - `.spec-workflow/user-templates/prd-template.md` (preferred)
   - fallback: `.spec-workflow/templates/prd-template.md`
   - enforce `<ui_approval_markdown_profile>`
   - if resuming, redispatch only when output is missing/blocked
9. Dispatch `prd-critic` with file handoff and enforce gate:
   - if BLOCKED, revise and re-run critic
   - if PASS, continue
   - if resuming, always rerun `prd-critic` before final approval flow
10. Save artifacts:
   - canonical: `.spec-workflow/specs/<spec-name>/requirements.md`
   - product mirror: `.spec-workflow/specs/<spec-name>/_generated/PRD.md`
11. Write `<run-dir>/_handoff.md` referencing source/structure/editor/critic outputs and resume decision taken (`continue-unfinished` or `delete-and-restart`).
12. Handle approval via MCP only:
   - call `spec-status`
   - resolve requirements status via `<approval_reconciliation>`
   - if approved, continue without re-requesting
   - if `needs-revision`/`changes-requested`/`rejected`, run revision_protocol first (subagent-driven), then continue through critic + approval flow
   - if pending, stop with `WAITING_FOR_APPROVAL` and instruct UI approval + rerun
   - only if approval was never requested (missing/empty/unknown status):
     - call `request-approval` then `get-approval-status` once
     - if pending, stop with `WAITING_FOR_APPROVAL`
     - if needs revision, run revision_protocol
   - never ask for approval in chat
</workflow>

<acceptance_criteria>
- [ ] Subagent outputs exist and are traceable (`_generated/PRD-SOURCE-NOTES.md`, `_generated/PRD-STRUCTURE.md`).
- [ ] Final document is PRD format and remains compatible with spec-workflow requirements flow.
- [ ] Every functional requirement has REQ-ID, priority, and verifiable acceptance criteria.
- [ ] REQ-IDs are unique and follow canonical format (`REQ-001`, `REQ-002`, ...).
- [ ] Explicit separation exists for v1, v2, and out-of-scope.
- [ ] If `--source` was provided, MCP usage was explicitly asked.
- [ ] On revision cycles, subagent analysis + codebase impact scan happened before edits.
- [ ] Clarification questions were asked when feedback was ambiguous/conflicting.
- [ ] PRD is approved before moving to design/tasks.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/_agent-comms/prd/<run-id>/` (and revision run dir when applicable).
- [ ] If unfinished run exists (`prd` or `prd-revision`), explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm PRD approval status and show artifact paths.
- Recommend next command: `spw:plan <spec-name>`.
- Recommend running `/clear` before `spw:plan` to keep context clean.

If blocked:
- Show the blocking reason (approval pending/rejected, missing source context, quality gate failure).
- If blocked by revision ambiguity, show pending clarification questions and do not edit artifacts until answered.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide exact fix action and the command to rerun: `spw:prd <spec-name>`.
</completion_guidance>
