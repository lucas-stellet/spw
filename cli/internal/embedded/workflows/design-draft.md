---
name: spw:design-draft
description: Subagent-driven design.md drafting from requirements + DESIGN-RESEARCH
argument-hint: "<spec-name>"
---

<dispatch_pattern>
category: pipeline
subcategory: synthesis
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
Generate `.spec-workflow/specs/<spec-name>/design.md` with strong traceability back to requirements.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md` (required)
- `.spec-workflow/specs/<spec-name>/research/*` (optional)
- post-mortem memory entries (if enabled)

output:
- `.spec-workflow/specs/<spec-name>/design.md`

comms:
- `.spec-workflow/specs/<spec-name>/design/_comms/design-draft/run-NNN/`

Do not consume generated research from generic locations (for example `docs/*`).
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `traceability-mapper` (model: complex_reasoning)
  - Maps REQ-IDs to technical decisions, files, and tests.
- `design-writer` (model: implementation)
  - Produces design draft from mapped decisions.
- `design-critic` (model: complex_reasoning)
  - Runs consistency and completeness gate.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the pipeline dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: approval reconciliation, skills, preconditions .. -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Apply skills policy: run design skills preflight and write `SKILLS-DESIGN-DRAFT.md`.
3. Verify preconditions:
   - `requirements.md` exists.
   - `design/DESIGN-RESEARCH.md` exists; stop BLOCKED if missing → instruct `spw:design-research <spec-name>`.
4. Load post-mortem memory inputs via `<post_mortem_memory>`.
5. Read templates:
   - `.spec-workflow/user-templates/design-template.md` (preferred)
   - fallback: `.spec-workflow/templates/design-template.md`
6. Inspect existing design-draft run dirs and apply resume decision gate.
</pre_pipeline>

<!-- post_dispatch: critic gate ................................... -->
<post_dispatch subagent="design-critic">
If critic returns BLOCKED:
- Revise with `design-writer`
- Re-run `design-critic`
</post_dispatch>

<!-- post_pipeline: artifact save + Mermaid validation + approval .. -->
<post_pipeline>
1. Validate Mermaid diagram(s) per `<diagram_policy>`.
2. Validate markdown per `<ui_approval_markdown_profile>`.
3. Save to `.spec-workflow/specs/<spec-name>/design.md`.
4. Write `<run-dir>/_handoff.md`.
5. Handle approval via MCP only:
   - call `spec-status`, resolve via `<approval_reconciliation>`
   - if approved: continue
   - if `needs-revision`/`changes-requested`/`rejected`: stop BLOCKED
   - if pending: stop with `WAITING_FOR_APPROVAL`
   - only if never requested: call `request-approval` then `get-approval-status` once
   - never ask for approval in chat
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

<preconditions>
- `.spec-workflow/specs/<spec-name>/requirements.md` exists.
- `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md` exists (mandatory intermediate artifact).
- If `DESIGN-RESEARCH.md` is missing, stop BLOCKED and instruct:
  - `spw:design-research <spec-name>`
</preconditions>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
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
3. Load selected reports and convert lessons into explicit design guardrails.

If index/report files are missing, continue with warning (non-blocking).
</post_mortem_memory>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/design/SKILLS-DESIGN-DRAFT.md`
2. Avoid loading full skill content in main context (subagent-first).
3. Require subagent outputs to explicitly mention skills used/missing.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<diagram_policy>
For `design.md` output:
- Include at least one valid Mermaid diagram in `## Architecture` main flow.
- Use fenced lowercase Mermaid language marker: `mermaid`.
- Prefer diagrams that represent real boundaries and data/control flow.
- If `mermaid-architecture` skill is available, use it for diagram type selection and syntax quality.
- Keep diagram terms consistent with requirement IDs and section vocabulary.
</diagram_policy>

<ui_approval_markdown_profile>
`design.md` must stay render-safe and review-friendly in Spec Workflow UI:
- Use plain Markdown (avoid raw HTML blocks unless strictly necessary).
- Use ATX headings (`#`, `##`, `###`) with consistent hierarchy.
- Keep tables valid with explicit header separator rows.
- Keep fenced code blocks balanced and language-tagged.
- Keep emphasis/underscore delimiters balanced (no dangling `_` or `**`).
- Keep architecture diagrams as fenced lowercase Mermaid blocks.
</ui_approval_markdown_profile>

<approval_reconciliation>
Resolve design approval with MCP-first reconciliation:
- Primary source:
  - `documents.design.approved`
  - `documents.design.status`
  - `approvals.design.status`
  - optional IDs:
    - `documents.design.approvalId`
    - `approvals.design.approvalId`
    - `approvals.design.id`
- If status is missing/unknown or inconsistent, fallback:
  1. Resolve approval ID from `spec-status` fields above.
  2. If still missing, read latest `.spec-workflow/approvals/<spec-name>/approval_*.json`
     where `filePath` is `.spec-workflow/specs/<spec-name>/design.md`.
  3. If approval ID exists, call MCP `approvals status` and use it as source of truth.
  4. If approval ID does not exist, treat as not requested.
- Never infer approval from `overallStatus`/phase labels alone.
</approval_reconciliation>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/spw/overlays/active/design-draft.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] Requirements traceability matrix exists.
- [ ] Technical decisions are justified.
- [ ] Test strategy is explicit.
- [ ] Architecture section contains at least one valid Mermaid diagram.
- [ ] Mermaid diagram uses fenced lowercase language marker `mermaid`.
- [ ] Document satisfies UI-safe markdown profile (headings/tables/fences/emphasis balanced).
- [ ] Critic gate returned PASS before approval request.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/design.md`.
- Confirm approval request status for design.
- Recommend next command: `spw:tasks-plan <spec-name>` (use config defaults, or override with `--mode` / `--max-wave-size` when needed).

If blocked:
- Show precondition/critic/review failures with required fixes.
- Provide rerun command: `spw:design-draft <spec-name>`.
</completion_guidance>
