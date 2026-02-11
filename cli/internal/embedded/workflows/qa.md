---
name: spw:qa
description: Build a QA validation plan for a spec using Playwright MCP, Bruno CLI, or hybrid strategy
argument-hint: "<spec-name> [--focus <what-to-validate>] [--tool auto|playwright|bruno|hybrid]"
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
@.claude/workflows/spw/overlays/active/qa.md
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
