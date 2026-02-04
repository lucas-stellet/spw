---
name: spw:checkpoint
description: Subagent-driven quality gate between execution batches/waves
argument-hint: "<spec-name> [--scope batch|wave|phase]"
---

<objective>
Validate that the executed batch truly meets spec intent, code quality, and integration safety before moving forward.
</objective>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.implementation].required`
- `[skills.implementation].optional`
- `[skills.implementation].enforce_required` (boolean)

Backward compatibility:
- if `[skills.implementation].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Skill loading gate (mandatory when `skills.enabled=true`):
1. Explicitly invoke every required implementation skill before checkpoint analysis.
2. Record loaded/missing skills in:
   - `.spec-workflow/specs/<spec-name>/SKILLS-CHECKPOINT.md`
3. If any required skill is missing/not invoked:
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
1. Run implementation skill loading gate and write `SKILLS-CHECKPOINT.md`.
2. Dispatch `evidence-collector`.
3. Dispatch `traceability-judge` using collected evidence.
4. Dispatch `release-gate-decider`.
5. Generate `.spec-workflow/specs/<spec-name>/CHECKPOINT-REPORT.md` with:
   - status: PASS | BLOCKED
   - critical issues
   - corrective actions
   - recommended next step
</workflow>

<gate_rule>
If status is BLOCKED, do not proceed to the next batch/wave.
</gate_rule>

<completion_guidance>
On PASS:
- Show concise go/no-go summary and recommend next command: `spw:exec <spec-name>` (next batch/wave).

On BLOCKED:
- Show critical issues first, with exact corrective actions.
- Recommend remediation command(s) and rerun: `spw:checkpoint <spec-name>`.
</completion_guidance>
