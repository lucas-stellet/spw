---
name: oraculo:qa-exec
description: Execute validated QA test plan using verified selectors from QA-CHECK.md
argument-hint: "<spec-name> [--scope smoke|regression|full] [--rerun-failed true|false]"
---

<dispatch_pattern>
category: wave-execution
subcategory: validation
phase: qa
comms_path: qa/_comms/qa-exec/waves/wave-{wave}
policy: (inlined below)

# Wave Execution Dispatch Pattern

Iterative dispatch over a set of work items, split into waves. Each wave dispatches
subagents for a bounded group of items, completes, then the next wave starts.
A synthesizer at the end consolidates everything into the command's final artifact.

## Thin-Dispatch Rules

These rules are mandatory for all wave execution commands. They override any command-specific behavior.

### 1. Status-Only Reads

After dispatching any subagent, read ONLY `<subagent>/status.json`.
- If `status=pass`: proceed to next subagent or wave step.
- If `status=blocked`: read `<subagent>/report.md` to decide action (log + skip, or stop BLOCKED).
- Never read `report.md` when status is pass.

### 2. Path-Based Briefs

When dispatching subsequent subagents or the synthesizer:
- Write **filesystem paths** to previous report files in the brief.
- Never copy or summarize report content into the brief.

### 3. Synthesizer Reads From Filesystem

The final synthesizer receives a brief listing ALL wave summary paths and report paths.
It reads them directly from disk — the orchestrator does not relay content.

### 4. Run Structure

```
<phase>/waves/wave-NN/
  <stage>/run-NNN/
    <subagent-1>/brief.md, report.md, status.json
    <subagent-2>/brief.md, report.md, status.json
    _handoff.md
  _wave-summary.json
  _latest.json
```

### 5. Resume Policy

On `continue-unfinished`:
- Scout inspects `_wave-summary.json` per wave.
- Skip completed waves entirely.
- Resume from first incomplete wave.
- Always rerun synthesizer.

## Wave Lifecycle

```
orchestrator:
  dispatch state-scout → resume state (compact)
  resolve waves from work items + wave size config
  for each wave-NN:
    for each subagent in wave:
      dispatch subagent → read status.json only
      on blocked: read report.md, decide action
    write wave-NN/_wave-summary.json (from status.json data)
  dispatch synthesizer (brief includes paths to all wave summaries + reports)
  synthesizer reads from fs → final artifact
```

## Extension Points

Wave execution commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir, before scout dispatch. Use for precondition checks, skill loading.
- **`inter_wave`**: Between waves, after wave summary is written. Use for quality gates (checkpoint), user authorization, re-authentication.
- **`per_task`**: Within a wave, around each task/scenario dispatch. Use for git hygiene, commit policy, per-item gates.
- **`post_pipeline`**: After all waves complete and synthesizer runs, before final _handoff.md. Use for artifact generation, drift reporting, next-step guidance.

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
Execute the validated QA test plan. All selectors/routes come from QA-CHECK.md — never reads implementation source files. Produces QA-EXECUTION-REPORT.md and QA-DEFECT-REPORT.md with GO/NO-GO decision.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`
- `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md` (verified selectors)
- `.spec-workflow/specs/<spec-name>/qa/QA-EXECUTION-REPORT.md` (if resuming)
- **Never reads implementation source files**

output:
- `.spec-workflow/specs/<spec-name>/qa/QA-EXECUTION-REPORT.md`
- `.spec-workflow/specs/<spec-name>/qa/QA-DEFECT-REPORT.md`
- `.spec-workflow/specs/<spec-name>/qa/qa-artifacts/wave-NN/`

comms:
- `.spec-workflow/specs/<spec-name>/qa/_comms/qa-exec/<run-id>/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `qa-state-scout` (model: implementation)
  - Produces compact resume state: completed/failed/next scenarios.
  - Max 12 bullets + JSON summary.
  - Runs first to enable fast resume.
- `qa-test-runner` (model: implementation)
  - Executes scenarios via Playwright MCP (headless) or Bruno CLI.
  - **Must NOT read source files** — uses only verified selector map from QA-CHECK.md.
  - Reports PASS/FAIL/BLOCKED per scenario with evidence paths.
- `qa-evidence-collector` (model: implementation)
  - Gathers traces, screenshots, junit/json/html reports.
  - Maps each artifact to its test ID.
  - Organizes evidence under `qa/qa-artifacts/`.
- `qa-exec-synthesizer` (model: complex_reasoning)
  - Consumes all test runner and evidence collector outputs.
  - Fills QA-EXECUTION-REPORT.md and QA-DEFECT-REPORT.md.
  - Produces GO/NO-GO decision with pass/fail counts and risk assessment.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the wave execution dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: verify prerequisites, scout .................... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>` and stop BLOCKED if missing.
2. Verify prerequisites exist in SPEC_DIR:
   - `qa/QA-TEST-PLAN.md` must exist; stop BLOCKED if missing → recommend `oraculo:qa <spec-name>`.
   - `qa/QA-CHECK.md` must exist and contain `PASS` status; stop BLOCKED if missing or BLOCKED → recommend `oraculo:qa-check <spec-name>`.
3. Dispatch `qa-state-scout` for compact resume state.
4. Inspect existing qa-exec run dirs and apply resume decision gate.
5. Apply skills policy: if `[skills].enabled=true`, load `qa-validation-planning`.
</pre_pipeline>

<!-- inter_wave: re-auth (--isolated) ............................. -->
<inter_wave>
Between scenario waves:
- Re-authenticate browser state (clean `--isolated` session) for Playwright scenarios.
- No checkpoint between waves (lighter than implementation waves).
</inter_wave>

<!-- per_task: scenario execution .................................. -->
<per_task>
For each scenario batch (grouped by tool type — Playwright MCP / Bruno CLI):
1. Write brief and dispatch `qa-test-runner` with:
   - scenario details from QA-TEST-PLAN.md
   - verified selectors from QA-CHECK.md
   - execution scope from `--scope` filter
2. Enforce `<playwright_runtime_policy>` for all Playwright MCP scenarios.
3. Write brief and dispatch `qa-evidence-collector` to gather artifacts.
4. Record PASS/FAIL/BLOCKED per scenario from test runner output.
</per_task>

<!-- post_pipeline: synthesizer + drift reporting .................. -->
<post_pipeline>
1. Dispatch `qa-exec-synthesizer` with all test runner and evidence collector outputs.
   - If resuming, always rerun `qa-exec-synthesizer`.
2. Generate artifacts:
   - `QA-EXECUTION-REPORT.md` (results, evidence paths, GO/NO-GO)
   - `QA-DEFECT-REPORT.md` (defects found, including selector drift)
3. Write `<run-dir>/_handoff.md` linking all subagent outputs, evidence paths, and execution summary.
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

<model_policy>
Resolve models from `.spec-workflow/oraculo.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<no_source_read_policy>
This is the core constraint of `oraculo:qa-exec`.

All selectors, routes, endpoints, and CSS identifiers must come from `QA-CHECK.md` (the verified selector map produced by `oraculo:qa-check`). No subagent in this command is allowed to read implementation source files (`.ex`, `.ts`, `.tsx`, `.py`, `.html`, `.heex`, etc.).

If a selector does not work at runtime:
- Log it as a "selector drift" defect in the execution report
- Do NOT search source files to find the correct selector
- Recommend rerunning `oraculo:qa-check` after the execution completes

This policy ensures `oraculo:qa-exec` remains fast and focused on execution, not re-discovery.
</no_source_read_policy>

<execution_scope_policy>
Scope filtering:
- `--scope smoke` → execute only scenarios with Level=Smoke
- `--scope regression` → execute only scenarios with Level=Regression
- `--scope full` (default) → execute all scenarios

Rerun filtering:
- `--rerun-failed false` (default) → execute all scenarios in scope
- `--rerun-failed true` → execute only FAIL/BLOCKED scenarios from previous QA-EXECUTION-REPORT.md
</execution_scope_policy>

<playwright_runtime_policy>
Playwright MCP is a pre-configured MCP server that exposes browser automation tools.

Prerequisites:
- Server must be registered before the session: `claude mcp add playwright -- npx @playwright/mcp@latest --headless --isolated`
- If Playwright MCP tools are not available in the current session, stop BLOCKED with setup instructions above.

Rules:
- Use the browser automation tools provided by the `playwright` MCP server — never invoke npx or node scripts
- Discover available tools from the `playwright` server at runtime — do not assume specific tool names
- Selectors come from QA-CHECK.md verified map; use them in MCP tool calls
- Collect evidence: take screenshots after assertions, capture console messages for logs
- If a selector fails via MCP tool, log as "selector drift" defect — do NOT search source files
</playwright_runtime_policy>

<state_recon_policy>
Before execution begins, dispatch `qa-state-scout` to produce a compact resume state.

Scout output must be max 12 bullet points + a JSON summary containing:
- completed scenario IDs and their PASS/FAIL status
- failed scenario IDs with failure reason
- next scenario batch to execute
- overall progress percentage

This enables fast resume without re-reading full execution artifacts.
</state_recon_policy>

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
- [ ] QA-CHECK.md was verified as PASS before execution started.
- [ ] No implementation source files were read during execution.
- [ ] All scenarios in scope were executed or marked BLOCKED with reason.
- [ ] All browser interactions used tools from the Playwright MCP server with selectors from QA-CHECK.md.
- [ ] Evidence artifacts (traces, screenshots, reports) are mapped to test IDs.
- [ ] GO/NO-GO decision is justified by pass/fail counts and risk assessment.
- [ ] QA-EXECUTION-REPORT.md and QA-DEFECT-REPORT.md were generated.
- [ ] File-first handoff exists under `qa/_comms/qa-exec/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision was respected.
- [ ] Selector drift defects (if any) recommend `oraculo:qa-check` rerun.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On GO:
- Show pass/fail/blocked counts and overall result.
- Confirm output paths:
  - `.spec-workflow/specs/<spec-name>/qa/QA-EXECUTION-REPORT.md`
  - `.spec-workflow/specs/<spec-name>/qa/QA-DEFECT-REPORT.md`

On NO-GO:
- Show failed scenarios with defect IDs.
- If failures are fixable: recommend `oraculo:qa-exec <spec-name> --rerun-failed true` after fixes.
- If failures need plan revision: recommend `oraculo:qa <spec-name>` → `oraculo:qa-check` → `oraculo:qa-exec`.

On selector drift:
- Log drift defects in QA-DEFECT-REPORT.md.
- Recommend `oraculo:qa-check <spec-name>` to re-verify selectors, then `oraculo:qa-exec <spec-name>`.
</completion_guidance>
