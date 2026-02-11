---
name: spw:qa
description: Build a QA validation plan for a spec using Playwright MCP, Bruno CLI, or hybrid strategy
argument-hint: "<spec-name> [--focus <what-to-validate>] [--tool auto|playwright|bruno|hybrid]"
---

<dispatch_pattern>
category: pipeline
subcategory: synthesis
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

Canonical runtime config path is `.spec-workflow/spw-config.toml`.

Transitional compatibility:
- If `.spec-workflow/spw-config.toml` is missing, fallback to `.spw/spw-config.toml`.

When shell logic is required, prefer:
- `spw tools config-get <section.key> --default <value> [--raw]`

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
Create a risk-based QA validation plan for the target spec and select the best validation toolchain:
- Playwright MCP for browser flows
- Bruno CLI for API behavior/contracts
- Hybrid when both are required
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`
- `.spec-workflow/specs/<spec-name>/tasks.md` (if present)
- `.spec-workflow/specs/<spec-name>/execution/CHECKPOINT-REPORT.md` (if present)
- Router/route config files (e.g. `router.ex`, `routes.ts`) for URL paths
- Template/component files for `data-testid` and CSS selectors

output:
- `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`

comms:
- `.spec-workflow/specs/<spec-name>/qa/_comms/qa/run-NNN/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `qa-scope-analyst` (model: complex_reasoning)
  Maps user intent + spec risks to a test strategy.
  Inputs: requirements.md, design.md, tasks.md, CHECKPOINT-REPORT.md (paths in brief).

- `browser-test-designer` (model: implementation)
  Produces Playwright MCP scenarios with concrete CSS/data-testid selectors.
  Inputs: qa-scope-analyst/report.md path, router/template source files.
  Conditional: dispatched only when tool=playwright or tool=hybrid.

- `api-test-designer` (model: implementation)
  Produces Bruno CLI scenarios with endpoints, methods, schemas.
  Inputs: qa-scope-analyst/report.md path, router/controller source files.
  Conditional: dispatched only when tool=bruno or tool=hybrid.

- `qa-plan-synthesizer` (model: complex_reasoning)
  Generates final QA-TEST-PLAN.md with Coverage Matrix including Selector/Endpoint column.
  Inputs: all previous report paths. Reads from filesystem.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the pipeline dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: user intent + tool selection ................. -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>` and stop BLOCKED if missing.
2. Apply skills policy: if `[skills].enabled=true`, load `qa-validation-planning` first.
3. Apply user intent gate (§ user_intent_gate).
4. Apply tool selection (§ tool_selection_policy).
5. Read router/template files to extract concrete selectors for designer briefs.
   This is the ONE planning phase where implementation files should be read.
6. Inspect existing qa run dirs and apply resume decision gate.
</pre_pipeline>

<!-- pre_dispatch: conditional designer selection ............... -->
<pre_dispatch subagent="browser-test-designer">
Skip this subagent if tool selection resolved to `bruno`.
</pre_dispatch>

<pre_dispatch subagent="api-test-designer">
Skip this subagent if tool selection resolved to `playwright`.
</pre_dispatch>

<!-- post_pipeline: artifact generation + guidance .............. -->
<post_pipeline>
1. Verify QA-TEST-PLAN.md includes Selector/Endpoint column in Coverage Matrix.
2. Verify all scenarios have concrete identifiers (per § concrete_selector_policy).
   Mark scenarios missing selectors as INCOMPLETE — does not block plan generation.
3. Write `<run-dir>/_handoff.md` linking evidence, tool rationale, unresolved risks.
4. Recommend next: `spw:qa-check <spec-name>` then `/clear`.
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

<user_intent_gate>
When `--focus` is missing:
1. AskUserQuestion with options:
   - `Browser journey validation (Recommended)`
   - `API contract/behavior validation`
   - `Release regression (UI + API)`
2. Ask follow-up clarifiers:
   - target environment/base URL
   - critical user flows or endpoints in scope
   - hard blockers or known flaky areas
Do not assume validation scope without this step.
</user_intent_gate>

<tool_selection_policy>
If `--tool` is explicit, honor it.
If `--tool=auto` (or omitted), choose by risk/scope:
- `playwright`: multi-page UI behavior, rendering, navigation, browser-only defects.
- `bruno`: API status codes, schema/contracts, auth, idempotency, error payloads.
- `hybrid`: user journeys AND API side effects must be validated together.
Write rationale in the plan.
</tool_selection_policy>

<concrete_selector_policy>
Every test scenario must contain concrete identifiers, not abstract descriptions.

Browser scenarios: CSS or `data-testid` selectors, URL routes, expected DOM state.
API scenarios: endpoint paths + HTTP methods, request schemas, expected responses.

Synthesizer must verify all scenarios have concrete identifiers.
INCOMPLETE scenarios are listed with warning; they do not block plan generation.
</concrete_selector_policy>

<playwright_runtime_policy>
Playwright MCP is a pre-configured MCP server.
- Must be registered: `claude mcp add playwright -- npx @playwright/mcp@latest --headless --isolated`
- If unavailable, stop BLOCKED with setup instructions.
- Use browser tools from the `playwright` MCP server — never invoke npx/node directly.
- Discover available tools at runtime — do not assume specific tool names.
</playwright_runtime_policy>

<skills_policy>
If `[skills].enabled=true`, load `qa-validation-planning` first.

Skill gate:
- If skill exists: use it.
- If skill is missing:
  - `[skills.design].enforce_required=true` -> BLOCKED
  - otherwise -> warn and continue.
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
- [ ] User validation target was explicitly captured.
- [ ] Tool selection (`playwright|bruno|hybrid`) is justified by risk/scope.
- [ ] Plan includes test levels, priority, data/env strategy, and pass/fail gates.
- [ ] Every test scenario contains concrete selectors/endpoints.
- [ ] Coverage Matrix includes `Selector/Endpoint` column.
- [ ] File-first handoff exists under `qa/_comms/qa/run-NNN/`.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
- [ ] All browser interactions used Playwright MCP server tools.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`.
- Recommend next command: `spw:qa-check <spec-name>` to validate selectors and traceability before execution.
- Recommend running `/clear` before validation.

On BLOCKED:
- Show missing input/decision.
- If waiting on resume choice, ask for `continue-unfinished` or `delete-and-restart`.
</completion_guidance>
