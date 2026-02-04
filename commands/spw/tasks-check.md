---
name: spw:tasks-check
description: Subagent-driven tasks.md validation (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<objective>
Validate whether `tasks.md` is ready for subagent execution.
</objective>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder:
- `.spec-workflow/specs/<spec-name>/agent-comms/tasks-check/<run-id>/`

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

After validation, write:
- `<run-dir>/_handoff.md` (orchestrator summary of audit results)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Backward compatibility:
- if `[skills.design].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Skill loading gate (mandatory when `skills.enabled=true`):
1. Explicitly invoke every required design skill before validation.
2. Record loaded/missing skills in:
   - `.spec-workflow/specs/<spec-name>/SKILLS-TASKS-CHECK.md`
3. If any required skill is missing/not invoked:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<subagents>
- `traceability-auditor` (model: complex_reasoning)
- `dag-validator` (model: implementation)
- `test-policy-auditor` (model: complex_reasoning)
- `decision-aggregator` (model: complex_reasoning)
</subagents>

<workflow>
1. Run design skill loading gate and write `SKILLS-TASKS-CHECK.md`.
2. Create communication run directory:
   - `.spec-workflow/specs/<spec-name>/agent-comms/tasks-check/<run-id>/`
3. Read `.spec-workflow/specs/<spec-name>/tasks.md` + requirements/design docs.
4. Write briefs and dispatch in parallel:
   - `traceability-auditor`
   - `dag-validator`
   - `test-policy-auditor`
5. Require auditor `report.md` + `status.json`; BLOCKED if missing.
6. Dispatch `decision-aggregator` with file handoff to produce PASS/BLOCKED decision.
7. Generate `.spec-workflow/specs/<spec-name>/TASKS-CHECK.md` containing:
   - PASS/BLOCKED
   - findings by severity
   - recommended fixes
8. Write `<run-dir>/_handoff.md` linking all auditor/aggregator outputs.
</workflow>

<acceptance_criteria>
- [ ] Every task references at least one requirement.
- [ ] Every requirement maps to at least one task.
- [ ] DAG has no cycles and wave order is valid.
- [ ] Test policy gate is satisfied.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/agent-comms/tasks-check/<run-id>/`.
</acceptance_criteria>

<completion_guidance>
On PASS:
- Confirm output path: `.spec-workflow/specs/<spec-name>/TASKS-CHECK.md`.
- Recommend next command: `spw:exec <spec-name> --batch-size <N>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- Recommend fix path: update `tasks.md`, then rerun `spw:tasks-check <spec-name>`.
</completion_guidance>
