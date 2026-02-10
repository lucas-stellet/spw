---
name: spw:qa-check
description: Validate QA test plan selectors, traceability, and data feasibility against actual code
argument-hint: "<spec-name>"
---

<objective>
Validate QA test plan against actual code. Confirm selectors/endpoints exist, verify requirement traceability, and check data feasibility. Produces a verified selector map that `spw:qa-exec` consumes without re-reading source files.
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
Reads:
- `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`
- Implementation source files (selector-verifier only)

Writes:
- `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md`
- `.spec-workflow/specs/<spec-name>/_agent-comms/qa-check/<run-id>/`
</artifact_boundary>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder:
- `.spec-workflow/specs/<spec-name>/_agent-comms/qa-check/<run-id>/`

For each subagent, use:
- `<run-dir>/<subagent>/brief.md` (written by orchestrator before dispatch)
- `<run-dir>/<subagent>/report.md` (written by subagent after execution)
- `<run-dir>/<subagent>/status.json` (written by subagent)

Status schema (minimum):
- `status`: `pass|blocked`
- `summary`: short result
- `inputs`: key files used
- `outputs`: generated artifacts
- `open_questions`: unresolved items
- `skills_used`: skills actually used by the subagent
- `skills_missing`: required skills not available for the subagent (if any)

After validation, write:
- `<run-dir>/_handoff.md` (orchestrator summary of check results)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<resume_policy>
Before creating a new run, inspect existing qa-check run folders:
- `.spec-workflow/specs/<spec-name>/_agent-comms/qa-check/<run-id>/`

A run is `unfinished` when any of these is true:
- `_handoff.md` is missing
- any subagent directory is missing `brief.md`, `report.md`, or `status.json`
- any subagent `status.json` reports `status=blocked`

Resume decision gate (mandatory):
1. Find latest unfinished run (if multiple, sort by mtime descending and use the newest).
2. If found, ask user once (AskUserQuestion) with options:
   - `continue-unfinished` (Recommended): continue with that run directory.
   - `delete-and-restart`: delete that unfinished run directory and start a new run.
3. Never choose automatically. Do not infer user intent.
4. If explicit user decision is unavailable, stop with `WAITING_FOR_USER_DECISION`.
5. Do not create a new run-id before this decision.

If user chooses `continue-unfinished`:
- Reuse completed auditor outputs (`report.md` + `status.json` with `status=pass`).
- Redispatch only missing/blocked auditors.
- Always rerun `qa-check-aggregator` before final PASS/BLOCKED output.

If user chooses `delete-and-restart`:
- Delete the selected unfinished run dir.
- Continue workflow with a fresh run-id.
- Record deleted path in final output.
</resume_policy>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `use_for_phases`

When `enabled=true` and `qa-check` is included in `use_for_phases`:
- create a team and set `teammate_mode`
- map subagent roles to teammates (do not exceed `max_teammates`)
- each teammate must still write `brief.md`, `report.md`, `status.json` in the run dir
</agent_teams_policy>

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

<workflow>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>` and stop BLOCKED if missing.
2. Verify `qa/QA-TEST-PLAN.md` exists in SPEC_DIR; stop BLOCKED if missing → recommend `spw:qa <spec-name>`.
3. Inspect existing qa-check run dirs and apply `<resume_policy>` decision gate.
4. Read context files:
   - `.spec-workflow/specs/<spec-name>/qa/QA-TEST-PLAN.md`
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
5. If Agent Teams are enabled for this phase, create a team before dispatching subagents.
6. Write briefs and dispatch 3 auditors in parallel:
   - `qa-traceability-auditor`
   - `qa-selector-verifier`
   - `qa-data-feasibility-checker`
   - If resuming, redispatch only missing/blocked auditors.
7. Require `report.md` + `status.json` from each auditor; stop BLOCKED if missing.
8. Dispatch `qa-check-aggregator` with file handoff to produce PASS/BLOCKED decision.
   - If resuming, always rerun `qa-check-aggregator`.
9. Generate `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md` containing:
   - PASS or BLOCKED status
   - Verified selector map (test-id → selector/endpoint → file:line)
   - Traceability percentage (tested requirements / total requirements)
   - Data feasibility assessment
   - Severity-ranked findings
10. Write `<run-dir>/_handoff.md` linking all auditor outputs and resume decision taken.
</workflow>

<acceptance_criteria>
- [ ] QA-TEST-PLAN.md was read and validated against source code.
- [ ] Every selector/endpoint in the plan was verified against implementation files.
- [ ] Bidirectional traceability (scenario ↔ requirement) was assessed.
- [ ] Test data/fixture feasibility was checked.
- [ ] PASS/BLOCKED decision is justified by aggregated findings.
- [ ] Verified selector map is included in QA-CHECK.md.
- [ ] File-first handoff exists under `.spec-workflow/specs/<spec-name>/_agent-comms/qa-check/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision was respected.
</acceptance_criteria>

<completion_guidance>
On PASS:
- Confirm output path: `.spec-workflow/specs/<spec-name>/qa/QA-CHECK.md`.
- Recommend next command: `spw:qa-exec <spec-name>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- If selectors are missing or invalid, recommend updating the test plan and rerunning `spw:qa-check <spec-name>`.
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
</completion_guidance>
