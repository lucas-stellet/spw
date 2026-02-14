---
name: oraculo:design-research
description: Subagent-driven technical research to prepare design.md
argument-hint: "<spec-name> [--focus <topic>] [--web-depth low|medium|high]"
---

<dispatch_pattern>
category: pipeline
subcategory: research
phase: design
comms_path: design/_comms/design-research
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
Generate architecture and implementation research inputs for the spec design.
Output: `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md`.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/steering/*.md` (if present)
- post-mortem memory entries (if enabled)

output:
- `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md`

comms:
- `.spec-workflow/specs/<spec-name>/design/_comms/design-research/run-NNN/`

Forbidden output locations for generated research:
- `docs/*`
- project root
- `.spec-workflow/steering/*`
- `.spec-workflow/user-templates/*`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `codebase-pattern-scanner` (model: implementation)
  - Finds reusable patterns, boundaries, integration points.
  - Investigates cross-boundary integrations: verify whether framework features (e.g., LiveView navigation, event handlers) work inside third-party-managed DOM (e.g., Leaflet popups, phx-update="ignore" zones). Factual questions must be resolved here, not deferred to design-draft.
- `web-pattern-scout-*` (model: web_research, parallel)
  - Performs external web/library/pattern scans.
- `risk-analyst` (model: complex_reasoning)
  - Identifies architecture/operational risks and mitigations.
- `research-synthesizer` (model: complex_reasoning)
  - Produces final consolidated recommendation set.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the pipeline dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: SPA/prototype URL detection, skills, resume .... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Apply skills policy: run design skills preflight and write `SKILLS-DESIGN-RESEARCH.md`.
3. Verify preconditions: `requirements.md` exists and is approved (MCP `spec-status`).
4. Load post-mortem memory inputs via `<post_mortem_memory>`.
5. Inspect existing design-research run dirs and apply resume decision gate.
</pre_pipeline>

<!-- pre_dispatch: Playwright MCP fallback for SPA scouts ......... -->
<pre_dispatch subagent="web-pattern-scout-*">
Apply `<prototype_url_policy>`: if scout target is an SPA or known prototype domain, use Playwright MCP. If unavailable, warn and continue with WebFetch.
</pre_dispatch>

<!-- post_pipeline: artifact generation + guidance ................. -->
<post_pipeline>
1. Write `<run-dir>/_handoff.md` with recommendation summary, unresolved risks, and subagent report references.
2. Ensure final sections include:
   - primary recommendations
   - alternatives and trade-offs
   - references/patterns to adopt
   - technical risks and mitigations
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

<preconditions>
- The spec has `requirements.md` and it is approved.
- Also read steering docs when present (`product.md`, `tech.md`, `structure.md`).
- Approval check must come from MCP `spec-status`; never ask approval in chat.
</preconditions>

<model_policy>
Resolve models from `.spec-workflow/oraculo.toml` `[models]`:
- web_research -> default `haiku`
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`

Multimodal override: when a `web-pattern-scout-*` must analyze images (screenshots, prototype visuals), use `implementation` instead of `web_research`. Record the override reason in `status.json` `model_override_reason` field.
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
3. Load selected post-mortems and derive design constraints:
   - failure patterns to avoid
   - missing decisions to enforce
   - review/test checks to include in recommendations

If index/report files are missing, continue with warning (non-blocking).
</post_mortem_memory>

<prototype_url_policy>
When a web scout fetches a URL that returns an SPA shell (minimal HTML with only JS bundle references, no meaningful text content), or the URL matches a known prototype/deploy-preview domain (`*.lovable.app`, `*.vercel.app`, `*.netlify.app`, `*.framer.app`, `*.webflow.io`, `*.stackblitz.com`):

1. Use Playwright MCP to navigate the URL, take screenshots, and extract visible content.
   - Playwright MCP is a pre-configured MCP server; discover its tools at runtime. Never invoke `npx` or Node scripts directly for browser automation.
2. If Playwright MCP tools are not available in the current session:
   - Warn the user: "Playwright MCP is not configured — prototype content may be incomplete."
   - Continue with whatever `WebFetch` returned.
3. Include extracted prototype content in the scout's `report.md`.
</prototype_url_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/oraculo.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/design/SKILLS-DESIGN-RESEARCH.md`
2. Avoid loading full skill content in main context (subagent-first).
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] Every relevant functional requirement has at least one technical recommendation.
- [ ] Existing-code reuse section is included.
- [ ] Risks and recommended decisions section is included.
- [ ] Web-only findings came from web_research model (or implementation with multimodal override documented in status.json).
- [ ] File-based handoff exists under `design/_comms/design-research/run-NNN/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md`.
- Recommend next command: `oraculo:design-draft <spec-name>`.

If blocked:
- List missing inputs (requirements approval, steering docs, source failures).
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide rerun command: `oraculo:design-research <spec-name>`.
</completion_guidance>
