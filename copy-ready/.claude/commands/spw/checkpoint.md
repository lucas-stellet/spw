---
name: spw:checkpoint
description: Subagent-driven quality gate between execution batches/waves
argument-hint: "<spec-name> [--scope batch|wave|phase]"
---

<objective>
Validate that the executed batch truly meets spec intent, code quality, and integration safety before moving forward.
</objective>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder:
- `.spec-workflow/specs/<spec-name>/agent-comms/checkpoint/<run-id>/`

For each subagent, use:
- `<run-dir>/<subagent>/brief.md` (written by orchestrator before dispatch)
- `<run-dir>/<subagent>/report.md` (written by subagent after execution)
- `<run-dir>/<subagent>/status.json` (written by subagent)

Status schema (minimum):
- `status`: `pass|blocked`
- `summary`: short result
- `inputs`: evidence sources used
- `outputs`: generated artifacts
- `open_questions`: unresolved items
- `skills_used`: skills actually used by the subagent
- `skills_missing`: required skills not available for the subagent (if any)

After checkpoint, write:
- `<run-dir>/_handoff.md` (orchestrator final go/no-go reasoning)

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
- `[skills].load_mode` (`subagent-first|principal-first`)
- `[skills.implementation].required`
- `[skills.implementation].optional`
- `[skills.implementation].enforce_required` (boolean)

Backward compatibility:
- if `[skills.implementation].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Load modes:
- `subagent-first` (default): orchestrator does availability preflight only and
  delegates skill loading/use to subagents via briefs.
- `principal-first` (legacy): orchestrator loads required skills before dispatch.

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/SKILLS-CHECKPOINT.md`
2. If `load_mode=subagent-first`, avoid loading full skill content in main context.
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<subagents>
- `evidence-collector` (model: implementation)
  - Collects task state, test/lint/typecheck outputs, implementation logs, and git status.
- `traceability-judge` (model: complex_reasoning)
  - Verifies requirements/design/tasks alignment for delivered changes.
- `release-gate-decider` (model: complex_reasoning)
  - Produces final PASS/BLOCKED decision and corrective actions.
</subagents>

<git_gate>
Resolve from `.spec-workflow/spw-config.toml` `[execution].require_clean_worktree_for_wave_pass` (default `true`).

If enabled:
- include `git status --porcelain` evidence in the report
- return BLOCKED when uncommitted tracked changes exist
- recommend exact commit commands before rerunning checkpoint
</git_gate>

<workflow>
1. Run implementation skills preflight (availability + load mode) and write `SKILLS-CHECKPOINT.md`.
2. Create communication run directory:
   - `.spec-workflow/specs/<spec-name>/agent-comms/checkpoint/<run-id>/`
3. Write brief (including required skills for role) and dispatch `evidence-collector`.
4. Require `evidence-collector` output files (`report.md`, `status.json` with skill fields); BLOCKED if missing.
5. Write brief (including required skills for role) and dispatch `traceability-judge` using collected evidence files.
6. Require `traceability-judge` output files (`report.md`, `status.json` with skill fields); BLOCKED if missing.
7. Write brief (including required skills for role) and dispatch `release-gate-decider` using prior reports.
8. Generate `.spec-workflow/specs/<spec-name>/CHECKPOINT-REPORT.md` with:
   - status: PASS | BLOCKED
   - critical issues
   - corrective actions
   - recommended next step
9. Write `<run-dir>/_handoff.md` linking all subagent outputs and final decision.
</workflow>

<gate_rule>
If status is BLOCKED, do not proceed to the next batch/wave.
</gate_rule>

<acceptance_criteria>
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/agent-comms/checkpoint/<run-id>/`.
- [ ] `CHECKPOINT-REPORT.md` decision is traceable to subagent reports.
</acceptance_criteria>

<completion_guidance>
On PASS:
- Show concise go/no-go summary and recommend next command: `spw:exec <spec-name>` (next batch/wave).

On BLOCKED:
- Show critical issues first, with exact corrective actions.
- Recommend remediation command(s) and rerun: `spw:checkpoint <spec-name>`.
</completion_guidance>
