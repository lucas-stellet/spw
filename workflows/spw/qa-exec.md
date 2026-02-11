---
name: spw:qa-exec
description: Execute validated QA test plan using verified selectors from QA-CHECK.md
argument-hint: "<spec-name> [--scope smoke|regression|full] [--rerun-failed true|false]"
---

<dispatch_pattern>
category: wave-execution
subcategory: validation
phase: qa
comms_path: qa/_comms/qa-exec/waves/wave-{wave}
policy: @.claude/workflows/spw/shared/dispatch-wave.md
</dispatch_pattern>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
- @.claude/workflows/spw/shared/dispatch-implementation.md
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

<subagent_artifact_map>
| Subagent | Artifact | Dispatch | Model |
|----------|----------|----------|-------|
| qa-state-scout | (report.md only) | task | implementation |
| qa-test-runner | (report.md only) | inline-mcp | implementation |
| qa-evidence-collector | (report.md only) | task | implementation |
| qa-exec-synthesizer | QA-EXECUTION-REPORT.md, QA-DEFECT-REPORT.md | task | complex_reasoning |
</subagent_artifact_map>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the wave execution dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: verify prerequisites, scout .................... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>` and stop BLOCKED if missing.
2. Verify prerequisites exist in SPEC_DIR:
   - `qa/QA-TEST-PLAN.md` must exist; stop BLOCKED if missing → recommend `spw:qa <spec-name>`.
   - `qa/QA-CHECK.md` must exist and contain `PASS` status; stop BLOCKED if missing or BLOCKED → recommend `spw:qa-check <spec-name>`.
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
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<no_source_read_policy>
This is the core constraint of `spw:qa-exec`.

All selectors, routes, endpoints, and CSS identifiers must come from `QA-CHECK.md` (the verified selector map produced by `spw:qa-check`). No subagent in this command is allowed to read implementation source files (`.ex`, `.ts`, `.tsx`, `.py`, `.html`, `.heex`, etc.).

If a selector does not work at runtime:
- Log it as a "selector drift" defect in the execution report
- Do NOT search source files to find the correct selector
- Recommend rerunning `spw:qa-check` after the execution completes

This policy ensures `spw:qa-exec` remains fast and focused on execution, not re-discovery.
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
@.claude/workflows/spw/overlays/active/qa-exec.md
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
- [ ] Selector drift defects (if any) recommend `spw:qa-check` rerun.
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
- If failures are fixable: recommend `spw:qa-exec <spec-name> --rerun-failed true` after fixes.
- If failures need plan revision: recommend `spw:qa <spec-name>` → `spw:qa-check` → `spw:qa-exec`.

On selector drift:
- Log drift defects in QA-DEFECT-REPORT.md.
- Recommend `spw:qa-check <spec-name>` to re-verify selectors, then `spw:qa-exec <spec-name>`.
</completion_guidance>
