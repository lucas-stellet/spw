---
name: spw:prd
description: Zero-to-PRD discovery flow with subagents to generate requirements.md
argument-hint: "<spec-name> [--source <url-or-file.md>]"
---

<dispatch_pattern>
category: pipeline
subcategory: research
policy: @.claude/workflows/spw/shared/dispatch-pipeline.md
</dispatch_pattern>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
</shared_policies>

<objective>
Create or update `.spec-workflow/specs/<spec-name>/requirements.md` in PRD format using a subagent-first process.

This command combines:
- GSD strengths: v1/v2/out-of-scope scoping, REQ-IDs, testable criteria, traceability.
- superpowers strengths: one-question-at-a-time discovery, recommendation + trade-off framing, incremental validation.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/requirements.md` (if present)
- `.spec-workflow/specs/<spec-name>/design.md` (if present)
- `.spec-workflow/steering/*.md` (if present)
- post-mortem memory entries (if enabled)

output:
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/prd/PRD.md` (product mirror)
- `.spec-workflow/specs/<spec-name>/prd/PRD-SOURCE-NOTES.md`
- `.spec-workflow/specs/<spec-name>/prd/PRD-STRUCTURE.md`

comms:
- `.spec-workflow/specs/<spec-name>/prd/_comms/run-NNN/`
- `.spec-workflow/specs/<spec-name>/prd/_comms/prd-revision/run-NNN/` (revision loop)
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `source-reader-web` (model: web_research)
  - Reads URLs and extracts only requirement-relevant signals.
- `source-reader-mcp` (model: implementation)
  - Reads MCP-backed sources (GitHub/Linear/ClickUp) and normalizes output.
- `feedback-analyzer` (model: complex_reasoning)
  - Converts approval comments into concrete requirement deltas and open questions.
- `codebase-impact-scanner` (model: implementation)
  - Checks feasibility/impact against existing code patterns and boundaries.
  - Scope: file paths, component names, architectural boundaries, current behavior.
  - NOT in scope: implementation recommendations, alternative approaches, effort estimates, code snippets for proposed changes. Those belong to later subagents.
- `revision-planner` (model: complex_reasoning)
  - Produces an explicit revision plan before any document edits.
- `requirements-structurer` (model: complex_reasoning)
  - Produces v1/v2/out-of-scope, REQ-IDs, acceptance criteria draft.
- `prd-editor` (model: implementation)
  - Writes final PRD into template format.
- `prd-critic` (model: complex_reasoning)
  - Performs strict quality gate before approval request.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the pipeline dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: preflight, source handling, user intent ........ -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Apply skills policy: run design skills preflight (availability).
3. Load post-mortem memory inputs via `<post_mortem_memory>`.
4. If `--source` is present, apply `<source_handling>` gate.
5. Inspect existing `prd` run dirs and apply resume decision gate.
6. Determine active run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create new run dir.
</pre_pipeline>

<!-- pre_dispatch: conditional scout branching ..................... -->
<pre_dispatch subagent="source-reader-web">
Dispatch only when `--source` looks like a URL. Apply `<prototype_url_policy>` for SPA detection.
</pre_dispatch>

<pre_dispatch subagent="source-reader-mcp">
Dispatch only when user selected an MCP-based source in `<source_handling>`.
</pre_dispatch>

<!-- post_dispatch: mid-pipeline user interaction .................. -->
<post_dispatch subagent="requirements-structurer">
Run one-question-at-a-time discovery with user before proceeding to prd-editor.

Procedure:
1. Read `status.json` only (thin-dispatch). Extract `summary` for clarification count.
2. For each [NEEDS_CLARIFICATION] or CLARIFY item flagged by the structurer,
   ask ONE AskUserQuestion call with that single question.
   - Include a recommendation and trade-off context in the question options.
   - Wait for user answer before asking the next question.
3. After all clarifications are resolved, write decisions to:
   `<run-dir>/_orchestrator-context/user-clarifications.md`
4. Reference that file in the prd-editor brief.

Exception: if the structurer flagged 4+ items AND they are independent (no dependencies between them), batch up to 4 in a single AskUserQuestion call. Document the batching reason in user-clarifications.md.
</post_dispatch>

<post_dispatch subagent="prd-critic">
If critic returns BLOCKED, revise with `prd-editor` and re-run critic.
If resuming, always rerun `prd-critic` before final approval flow.
</post_dispatch>

<!-- post_pipeline: artifact save + approval ....................... -->
<post_pipeline>
1. Save artifacts:
   - canonical: `.spec-workflow/specs/<spec-name>/requirements.md`
   - product mirror: `.spec-workflow/specs/<spec-name>/prd/PRD.md`
2. Write `<run-dir>/_handoff.md` referencing source/structure/editor/critic outputs.
3. Handle approval via MCP only:
   - call `spec-status`, resolve via `<approval_reconciliation>`
   - if approved: continue
   - if `needs-revision`/`changes-requested`/`rejected`: run `<revision_protocol>` (subagent-driven)
   - if pending: stop with `WAITING_FOR_APPROVAL`
   - only if never requested: call `request-approval` then `get-approval-status` once
   - never ask for approval in chat
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

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

If enabled and index exists:
1. Read `.spec-workflow/post-mortems/INDEX.md`.
2. Select up to `max_entries_for_design` relevant entries:
   - same `<spec-name>` first
   - then by tag/topic similarity and recency
3. Load selected reports and extract reusable guardrails for PRD quality.

If index/report files are missing, continue with warning (non-blocking).
</post_mortem_memory>

<source_handling>
**Shortcut:** If the user's command includes `use <mcp-name>` (e.g., `use linear-server`, `use github`), skip the AskUserQuestion flow and treat it as if the user selected that MCP directly. Proceed to the matching MCP dispatch.

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

<prototype_url_policy>
When fetching a URL provided in a source (issue, brief, user input):

1. If `WebFetch` returns an SPA shell (minimal HTML with only JS bundle references, no meaningful text content), or the URL matches a known prototype/deploy-preview domain (`*.lovable.app`, `*.vercel.app`, `*.netlify.app`, `*.framer.app`, `*.webflow.io`, `*.stackblitz.com`):
   - **Discover Playwright MCP tools first**: check if Playwright MCP tools are available in the current session before declaring them absent. Look for tool names containing "playwright" or browser automation capabilities.
   - If available: use Playwright MCP to navigate the URL, take screenshots, and extract visible content. Never invoke `npx` or Node scripts directly.
   - Write all prototype observations (screenshot descriptions, UI patterns, interaction findings) to:
     `<run-dir>/_orchestrator-context/prototype-observations.md`
   - Reference that file in subsequent briefs (source-reader, requirements-structurer).
2. If Playwright MCP tools are not available in the current session:
   - Warn the user: "Playwright MCP is not configured — prototype content may be incomplete. Run `claude mcp add playwright -- npx @playwright/mcp@latest --headless --isolated` to enable."
   - Continue with whatever `WebFetch` returned.
3. Include any screenshots or extracted content in the PRD source notes (`PRD-SOURCE-NOTES.md`).
</prototype_url_policy>

<revision_protocol>
Trigger this protocol when either:
- MCP approval status for requirements is `changes-requested` or `rejected`, or
- user asks to analyze/adjust reviewed requirements.

Protocol (mandatory):
1. Inspect existing `prd-revision` run dirs and apply resume decision gate.
2. Determine active revision run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create:
     `.spec-workflow/specs/<spec-name>/prd/_comms/prd-revision/run-NNN/`
3. Read approval feedback from MCP and existing `requirements.md`.
4. Dispatch `feedback-analyzer` (with file handoff) to classify:
   - accepted changes
   - ambiguous/conflicting feedback
   - out-of-scope suggestions
   - if resuming, redispatch only when output is missing/blocked
5. Dispatch `codebase-impact-scanner` with file handoff.
   - if resuming, redispatch only when output is missing/blocked
6. Dispatch `revision-planner` with file handoff to create:
   - `.spec-workflow/specs/<spec-name>/prd/PRD-REVISION-PLAN.md`
   - `.spec-workflow/specs/<spec-name>/prd/PRD-REVISION-QUESTIONS.md` (if needed)
   - if resuming, redispatch only when output is missing/blocked
7. Ask targeted clarification questions before editing if ambiguity/conflict exists.
8. Only after clarification, dispatch `prd-editor` with file handoff to apply approved deltas.
9. Save revision summary:
   - `.spec-workflow/specs/<spec-name>/prd/PRD-REVISION-NOTES.md`
10. Write revision handoff:
   - `.spec-workflow/specs/<spec-name>/prd/_comms/prd-revision/run-NNN/_handoff.md`
   - include resume decision taken (`continue-unfinished` or `delete-and-restart`)

Never directly edit requirements immediately after reading review comments.
</revision_protocol>

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

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/spw/overlays/active/prd.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] Subagent outputs exist and are traceable (`prd/PRD-SOURCE-NOTES.md`, `prd/PRD-STRUCTURE.md`).
- [ ] Final document is PRD format and remains compatible with spec-workflow requirements flow.
- [ ] Every functional requirement has REQ-ID, priority, and verifiable acceptance criteria.
- [ ] REQ-IDs are unique and follow canonical format (`REQ-001`, `REQ-002`, ...).
- [ ] Explicit separation exists for v1, v2, and out-of-scope.
- [ ] If `--source` was provided, MCP usage was explicitly asked.
- [ ] On revision cycles, subagent analysis + codebase impact scan happened before edits.
- [ ] Clarification questions were asked when feedback was ambiguous/conflicting.
- [ ] PRD is approved before moving to design/tasks.
- [ ] File-based handoff exists under `prd/_comms/run-NNN/` (and revision run dir when applicable).
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
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
