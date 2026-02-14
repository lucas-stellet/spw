---
name: oraculo:design-draft
description: Subagent-driven design.md drafting from requirements + DESIGN-RESEARCH
argument-hint: "<spec-name>"
---

<dispatch_pattern>
category: pipeline
subcategory: synthesis
phase: design
comms_path: design/_comms/design-draft
policy: (inlined below)

# Pipeline Dispatch Pattern

Sequential chain of subagents where each produces output that feeds the next.
A synthesizer at the end consolidates everything into the command's final artifact.

## Thin-Dispatch Rules

These rules are mandatory for all pipeline commands. They override any command-specific behavior.

### 1. Status-Only Reads

After dispatching any subagent, read ONLY `<subagent>/status.json`.
- If `status=pass`: proceed to next step.
- If `status=blocked`: read `<subagent>/report.md` to decide action (log + skip, or stop BLOCKED).
- Never read `report.md` when status is pass.

### 2. Path-Based Briefs

When subagent-B depends on output from subagent-A:
- Write the **filesystem path** to `subagent-A/report.md` in `subagent-B/brief.md`.
- Never copy or summarize report content into the brief.

Example brief content:
```
## Inputs
- Scope analysis: <run-dir>/qa-scope-analyst/report.md
- Requirements: .spec-workflow/specs/<spec-name>/requirements.md
```

### 3. Synthesizer Reads From Filesystem

The last subagent (synthesizer/writer) receives a brief listing ALL previous report paths.
It reads them directly from disk — the orchestrator does not relay content.

### 4. Run Structure

```
<phase>/_comms/<command>/run-NNN/
  <subagent-1>/brief.md, report.md, status.json
  <subagent-2>/brief.md, report.md, status.json
  <synthesizer>/brief.md, report.md, status.json
  _handoff.md
```

### 5. Resume Policy

On `continue-unfinished`:
- Skip subagents where `status.json` exists with `status=pass`.
- Redispatch missing or blocked subagents.
- Always rerun synthesizer.

### 6. Artifact Save

When the pipeline's final subagent (synthesizer/writer) writes the command's output artifact to its `report.md`, the orchestrator saves it to the canonical path using filesystem copy — never by reading content into its own context.

```
cp <run-dir>/<writer>/report.md <canonical-output-path>
```

If the command requires post-save validation (Mermaid syntax, dashboard markdown profile, MDX compilation), run validation tools/scripts on the saved file — do not Read the file into orchestrator context. If validation fails, re-dispatch the writer with fix instructions in a new brief iteration, or apply the Surgical Fix Policy below.

### 7. Surgical Fix Policy

When a critic/reviewer returns BLOCKED with a specific, mechanical fix (e.g., arithmetic correction, typo, missing escape character):

- **Threshold:** fix touches ≤ 3 lines in the writer's `report.md` AND requires no design judgment (pure factual/syntactic correction).
- **Allowed:** orchestrator applies the fix directly to the writer's `report.md`.
- **Required:** log every inline fix in `<run-dir>/_handoff.md` under a `## Inline Fixes` section with: line(s) changed, reason, original value → new value.
- **Re-run critic:** always re-dispatch the critic after an inline fix.

If the fix exceeds the threshold (> 3 lines or requires design judgment), re-dispatch the writer subagent with the critic's feedback in a new brief.

## Extension Points

Pipeline commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir and reading config, before first dispatch. Use for user intent gates, preflight checks, skill loading.
- **`pre_dispatch(<subagent>)`**: Before writing a specific subagent's brief. Use for conditional dispatch (e.g., selecting which designer to run based on a gate decision).
- **`post_dispatch(<subagent>)`**: After reading a subagent's status.json. Use for mid-pipeline decisions that affect subsequent dispatches.
- **`post_pipeline`**: After synthesizer completes, before writing _handoff.md. Use for artifact generation, approval reconciliation, completion guidance.

</dispatch_pattern>

<shared_policies>
# Config Resolution

Canonical runtime config path is `.spec-workflow/oraculo.toml`.

Transitional compatibility:
- If `.spec-workflow/oraculo.toml` is missing, fallback to `.oraculo/oraculo.toml`.

When shell logic is required, prefer:
- `oraculo tools config-get <section.key> --default <value> [--raw]`

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
   - `design/DESIGN-RESEARCH.md` exists; stop BLOCKED if missing → instruct `oraculo:design-research <spec-name>`.
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
1. Validate Mermaid diagram(s) per `<diagram_policy>` (run validation on writer's report.md file, do not Read into context).
2. Validate markdown per `<ui_approval_markdown_profile>` (grep-based check for unescaped angle brackets outside code fences).
3. If validation fails: apply Surgical Fix Policy (§ dispatch-pipeline.md) or re-dispatch `design-writer`.
4. Save via filesystem copy: `cp <run-dir>/design-writer/report.md <SPEC_DIR>/design.md`.
5. Write `<run-dir>/_handoff.md` via `dispatch-handoff`.
6. Handle approval via MCP only:
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
Verify file existence only (use Glob, never Read content into orchestrator context):
- `.spec-workflow/specs/<spec-name>/requirements.md` must exist.
- `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md` must exist (mandatory intermediate artifact).
- If `DESIGN-RESEARCH.md` is missing, stop BLOCKED and instruct:
  - `oraculo:design-research <spec-name>`
</preconditions>

<model_policy>
Resolve models from `.spec-workflow/oraculo.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<post_mortem_memory>
Resolve from `.spec-workflow/oraculo.toml` `[post_mortem_memory]`:
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
Resolve skill policy from `.spec-workflow/oraculo.toml`:
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
- Escape angle brackets that are NOT inside fenced code blocks: `<ComponentName>` → wrap in inline code backticks or escape as `\<ComponentName\>`. The Spec Workflow UI compiles markdown as MDX — unescaped `<...>` outside code fences causes compilation errors.
- Before saving, verify no unescaped `<` exists outside fenced code blocks (grep-based check is sufficient).
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
- [ ] No unescaped angle brackets (`<...>`) exist outside fenced code blocks (MDX safety).
- [ ] Critic gate returned PASS before approval request.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/design.md`.
- Confirm approval request status for design.
- Recommend next command: `oraculo:tasks-plan <spec-name>` (use config defaults, or override with `--mode` / `--max-wave-size` when needed).

If blocked:
- Show precondition/critic/review failures with required fixes.
- Provide rerun command: `oraculo:design-draft <spec-name>`.
</completion_guidance>
