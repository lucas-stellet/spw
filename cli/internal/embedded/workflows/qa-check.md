---
name: spw:qa-check
description: Validate QA test plan selectors, traceability, and data feasibility against actual code
argument-hint: "<spec-name>"
---

<dispatch_pattern>
category: audit
subcategory: code
phase: qa
comms_path: qa/_comms/qa-check
policy: @.claude/workflows/spw/shared/dispatch-audit.md
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
Validate QA test plan against actual code. Confirm selectors/endpoints exist, verify requirement traceability, and check data feasibility. Produces a verified selector map that `spw:qa-exec` consumes without re-reading source files.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`
- Implementation source files (selector-verifier only)

output:
- `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md`

comms:
- `.spec-workflow/specs/<spec-name>/qa/_comms/qa-check/<run-id>/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — auditors + aggregator
     ============================================================ -->

<subagents>
- `qa-traceability-auditor` (model: complex_reasoning)
  - Performs bidirectional mapping: scenario ↔ requirement.
  - Flags untested requirements and orphaned scenarios.
- `qa-selector-verifier` (model: implementation)
  - Searches source files for each selector, route, and endpoint in the test plan.
  - Confirms existence and outputs a verified map: test-id → selector → file:line.
  - This is the ONLY subagent that reads implementation source files.
- `qa-data-feasibility-checker` (model: implementation)
  - Validates test data assumptions: seeds, fixtures, accounts, environment variables.
  - Flags missing or unreachable test prerequisites.
- `qa-check-aggregator` (model: complex_reasoning)
  - Consumes all auditor/verifier reports.
  - Produces PASS/BLOCKED decision with severity-ranked findings.
</subagents>

<subagent_artifact_map>
| Subagent | Artifact | Dispatch | Model |
|----------|----------|----------|-------|
| qa-traceability-auditor | (report.md only) | task | complex_reasoning |
| qa-selector-verifier | (report.md only) | task | implementation |
| qa-data-feasibility-checker | (report.md only) | task | implementation |
| qa-check-aggregator | QA-CHECK.md | task | complex_reasoning |
</subagent_artifact_map>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the audit dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: verify QA-TEST-PLAN.md exists .................. -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>` and stop BLOCKED if missing.
2. Verify `qa/QA-TEST-PLAN.md` exists in SPEC_DIR; stop BLOCKED if missing → recommend `spw:qa <spec-name>`.
3. Inspect existing qa-check run dirs and apply resume decision gate.
4. Read context files:
   - `qa/QA-TEST-PLAN.md`
   - `requirements.md`
   - `design.md`
</pre_pipeline>

<!-- post_pipeline: generate QA-CHECK.md .......................... -->
<post_pipeline>
1. Generate `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md` containing:
   - PASS or BLOCKED status
   - Verified selector map (test-id → selector/endpoint → file:line)
   - Traceability percentage (tested requirements / total requirements)
   - Data feasibility assessment
   - Severity-ranked findings
2. Write `<run-dir>/_handoff.md` linking all auditor outputs.
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES
     ============================================================ -->

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/spw/overlays/active/qa-check.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] QA-TEST-PLAN.md was read and validated against source code.
- [ ] Every selector/endpoint in the plan was verified against implementation files.
- [ ] Bidirectional traceability (scenario ↔ requirement) was assessed.
- [ ] Test data/fixture feasibility was checked.
- [ ] PASS/BLOCKED decision is justified by aggregated findings.
- [ ] Verified selector map is included in QA-CHECK.md.
- [ ] File-first handoff exists under `qa/_comms/qa-check/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision was respected.
- [ ] Orchestrator never read report.md from any auditor (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
Note: Inline audit now runs automatically during `spw:qa`. Use this command for re-validation after manual fixes or as a standalone CI gate.

On PASS:
- Confirm output path: `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md`.
- Recommend next command: `spw:qa-exec <spec-name>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- If selectors are missing or invalid, recommend updating the test plan and rerunning `spw:qa-check <spec-name>`.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
</completion_guidance>
