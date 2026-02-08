---
name: spw:qa
description: Build a QA validation plan for a spec using Playwright MCP, Bruno CLI, or hybrid strategy
argument-hint: "<spec-name> [--focus <what-to-validate>] [--tool auto|playwright|bruno|hybrid]"
---

<objective>
Create a risk-based QA validation plan for the target spec and select the best validation toolchain:
- Playwright MCP for browser flows
- Bruno CLI for API behavior/contracts
- Hybrid when both are required
</objective>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
</shared_policies>

<path_conventions>
- Canonical spec root: `.spec-workflow/specs/<spec-name>/`
- Never use `.specs/`
- QA outputs must stay inside the active spec directory
</path_conventions>

<artifact_boundary>
Write outputs under:
- `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`
- `.spec-workflow/specs/<spec-name>/qa/QA-EXECUTION-REPORT.md`
- `.spec-workflow/specs/<spec-name>/qa/QA-DEFECT-REPORT.md`

Communication/handoff:
- `.spec-workflow/specs/<spec-name>/agent-comms/qa/<run-id>/`
</artifact_boundary>

<file_handoff_protocol>
Subagent communication must be file-first.

For each subagent, use:
- `<run-dir>/<subagent>/brief.md`
- `<run-dir>/<subagent>/report.md`
- `<run-dir>/<subagent>/status.json`

After synthesis, write:
- `<run-dir>/_handoff.md`

If a required handoff file is missing, stop BLOCKED.
</file_handoff_protocol>

<resume_policy>
Before creating a new run, inspect existing QA run folders:
- `.spec-workflow/specs/<spec-name>/agent-comms/qa/<run-id>/`

A run is `unfinished` when any of these is true:
- `_handoff.md` is missing
- any subagent folder is missing `brief.md`, `report.md`, or `status.json`
- any `status.json` reports `status=blocked`

Resume decision gate (mandatory):
1. Find latest unfinished run.
2. Ask user once (AskUserQuestion):
   - `continue-unfinished` (Recommended)
   - `delete-and-restart`
3. Never choose automatically.
4. If no explicit decision, stop with `WAITING_FOR_USER_DECISION`.
</resume_policy>

<skills_policy>
If `[skills].enabled=true`, load `qa-validation-planning` first.

Skill gate:
- If skill exists: use it.
- If skill is missing:
  - `[skills.design].enforce_required=true` -> BLOCKED
  - otherwise -> warn and continue.
</skills_policy>

<agent_teams_policy>
Resolve Agent Teams config from `.spw/spw-config.toml` `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `use_for_phases`

When `enabled=true` and `qa` is included in `use_for_phases`:
- create a team and set `teammate_mode`
- map QA roles to teammates (scope analyst, browser/API designers, synthesizer)
- assign only active designer roles based on selected tool (`playwright|bruno|hybrid`)
- each teammate must still write `brief.md`, `report.md`, and `status.json`
</agent_teams_policy>

<user_intent_gate>
This command must ask what the user wants to validate before drafting the plan.

When `--focus` is missing:
1. AskUserQuestion with options:
   - `Browser journey validation (Recommended)`
   - `API contract/behavior validation`
   - `Release regression (UI + API)`
2. Ask follow-up clarifiers in plain language:
   - target environment/base URL
   - critical user flows or endpoints in scope
   - hard blockers or known flaky areas

Do not assume validation scope without this step.
</user_intent_gate>

<tool_selection_policy>
If `--tool` is explicit, honor it.

If `--tool=auto` (or omitted), choose by risk/scope:
- choose `playwright` when confidence depends on multi-page UI behavior, rendering states, navigation, or browser-only defects.
- choose `bruno` when confidence depends on API status codes, schema/contracts, auth behavior, idempotency, or error payloads.
- choose `hybrid` when user journeys and API side effects must be validated together.

Always write the rationale in the plan.
</tool_selection_policy>

<playwright_runtime_policy>
For any browser validation in this command, Playwright MCP must run in headless mode.

Rules:
- always include `--headless` in Playwright MCP runtime args
- never switch to headed mode in `spw:qa`
- if a user requests headed browser, keep `--headless` and document the restriction in the plan notes
</playwright_runtime_policy>

<subagents>
- `qa-scope-analyst` (model: complex_reasoning)
  - Maps user intent + spec risks to a test strategy.
- `browser-test-designer` (model: implementation)
  - Produces Playwright MCP scenarios, evidence strategy, and execution order.
- `api-test-designer` (model: implementation)
  - Produces Bruno collection execution strategy, env matrix, and assertions.
- `qa-plan-synthesizer` (model: complex_reasoning)
  - Generates final QA artifacts and go/no-go guidance.
</subagents>

<workflow>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>` and stop BLOCKED if missing.
2. Apply `<resume_policy>` and determine active run dir.
3. Apply `<user_intent_gate>` and capture explicit validation target.
4. Read only required context:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/specs/<spec-name>/tasks.md` (if present)
   - `.spec-workflow/specs/<spec-name>/CHECKPOINT-REPORT.md` (if present)
5. If Agent Teams are enabled for this phase, create a team before dispatching subagents.
6. Dispatch `qa-scope-analyst`.
7. Apply `<tool_selection_policy>`.
8. If Agent Teams are enabled for this phase, assign active roles to teammates based on selected tool.
9. Dispatch designers based on selected tool:
   - `playwright` -> `browser-test-designer`
   - `bruno` -> `api-test-designer`
   - `hybrid` -> both in parallel
10. Enforce `<playwright_runtime_policy>` for all Playwright MCP scenarios.
11. Dispatch `qa-plan-synthesizer` with previous outputs.
12. Generate artifacts under `.spec-workflow/specs/<spec-name>/qa/`:
   - `QA-TEST-PLAN.md`
   - `QA-EXECUTION-REPORT.md` (template prefilled for this run)
   - `QA-DEFECT-REPORT.md` (template prefilled for this run)
13. Write `<run-dir>/_handoff.md` linking evidence, selected tool rationale, and unresolved risks.
</workflow>

<acceptance_criteria>
- [ ] User validation target was explicitly captured.
- [ ] Tool selection (`playwright|bruno|hybrid`) is justified by risk/scope.
- [ ] Plan includes test levels, priority, data/env strategy, and pass/fail gates.
- [ ] Execution and defect report templates were generated for the same spec.
- [ ] File-first handoff exists under `.spec-workflow/specs/<spec-name>/agent-comms/qa/<run-id>/`.
- [ ] If Agent Teams are enabled for `qa`, teammate assignment was applied for active roles.
- [ ] All Playwright MCP runtime commands include `--headless`.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output paths in `.spec-workflow/specs/<spec-name>/qa/`.
- Recommend immediate next step:
  - if plan only: execute planned tests and fill `QA-EXECUTION-REPORT.md`.
  - if defects found: log in `QA-DEFECT-REPORT.md` and rerun `spw:qa <spec-name>` for regression scope.

On BLOCKED:
- Show missing input/decision.
- If waiting on resume choice, ask for `continue-unfinished` or `delete-and-restart`.
</completion_guidance>
