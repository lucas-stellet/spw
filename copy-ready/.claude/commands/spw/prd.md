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

<revision_protocol>
Trigger this protocol when either:
- MCP approval status for requirements is `changes-requested` or `rejected`, or
- user asks to analyze/adjust reviewed requirements.

Protocol (mandatory):
1. Read approval feedback from MCP and existing `requirements.md`.
2. Dispatch `feedback-analyzer` to classify:
   - accepted changes
   - ambiguous/conflicting feedback
   - out-of-scope suggestions
3. Dispatch `codebase-impact-scanner` (if enabled in config `[reviews]`).
4. Dispatch `revision-planner` to create:
   - `.spec-workflow/specs/<spec-name>/PRD-REVISION-PLAN.md`
   - `.spec-workflow/specs/<spec-name>/PRD-REVISION-QUESTIONS.md` (if needed)
5. Ask targeted clarification questions before editing if ambiguity/conflict exists.
6. Only after clarification, dispatch `prd-editor` to apply approved deltas.
7. Save revision summary:
   - `.spec-workflow/specs/<spec-name>/PRD-REVISION-NOTES.md`

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

<workflow>
1. Read existing context:
   - `.spec-workflow/specs/<spec-name>/requirements.md` (if present)
   - `.spec-workflow/specs/<spec-name>/design.md` (if present)
   - `.spec-workflow/steering/*.md` (if present)
2. If `--source` is present, dispatch source-reader subagents:
   - web-only fetches -> `source-reader-web`
   - MCP-backed reads -> `source-reader-mcp`
   - save normalized notes to `.spec-workflow/specs/<spec-name>/PRD-SOURCE-NOTES.md`
3. Run one-question-at-a-time discovery with user.
4. Dispatch `requirements-structurer` to produce a structured draft:
   - `.spec-workflow/specs/<spec-name>/PRD-STRUCTURE.md`
5. Dispatch `prd-editor` to fill template using:
   - `.spec-workflow/user-templates/prd-template.md` (preferred)
   - fallback: `.spec-workflow/templates/prd-template.md`
6. Dispatch `prd-critic` and enforce gate:
   - if BLOCKED, revise and re-run critic
   - if PASS, continue
7. Save artifacts:
   - canonical: `.spec-workflow/specs/<spec-name>/requirements.md`
   - product mirror: `.spec-workflow/specs/<spec-name>/PRD.md`
8. Handle approval via MCP only:
   - call `spec-status`
   - resolve status from:
     - `documents.requirements.approved`
     - `documents.requirements.status`
     - `approvals.requirements.status`
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
- [ ] Subagent outputs exist and are traceable (`PRD-SOURCE-NOTES.md`, `PRD-STRUCTURE.md`).
- [ ] Final document is PRD format and remains compatible with spec-workflow requirements flow.
- [ ] Every functional requirement has REQ-ID, priority, and verifiable acceptance criteria.
- [ ] Explicit separation exists for v1, v2, and out-of-scope.
- [ ] If `--source` was provided, MCP usage was explicitly asked.
- [ ] On revision cycles, subagent analysis + codebase impact scan happened before edits.
- [ ] Clarification questions were asked when feedback was ambiguous/conflicting.
- [ ] PRD is approved before moving to design/tasks.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm PRD approval status and show artifact paths.
- Recommend next command: `spw:plan <spec-name>`.
- Recommend running `/clear` before `spw:plan` to keep context clean.

If blocked:
- Show the blocking reason (approval pending/rejected, missing source context, quality gate failure).
- If blocked by revision ambiguity, show pending clarification questions and do not edit artifacts until answered.
- Provide exact fix action and the command to rerun: `spw:prd <spec-name>`.
</completion_guidance>
