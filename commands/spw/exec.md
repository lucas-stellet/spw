---
name: spw:exec
description: Subagent-driven task execution in batches with mandatory checkpoints
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<objective>
Execute tasks in batches, with mandatory pauses for checkpoint quality gates.
</objective>

<path_conventions>
- Canonical spec root: `.spec-workflow/specs/<spec-name>/`
- Canonical files:
  - `.spec-workflow/specs/<spec-name>/tasks.md`
  - `.spec-workflow/specs/<spec-name>/requirements.md`
  - `.spec-workflow/specs/<spec-name>/design.md`
- Legacy `.specs/` paths are NOT valid for SPW commands.
- Never read from or glob `.specs/*`.
</path_conventions>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<subagents>
- `task-implementer` (model: implementation)
  - Implements each task and runs task-level verification.
- `spec-compliance-reviewer` (model: complex_reasoning for complex/critical tasks; otherwise implementation)
  - Verifies exact adherence to requirements/design/task scope.
- `code-quality-reviewer` (model: implementation)
  - Reviews maintainability, safety, and regression risk.
</subagents>

<complexity_routing>
Treat a task as complex/critical when it includes one or more:
- touches auth/security/payments/data migrations/concurrency boundaries
- modifies > 3 core files
- affects cross-context architecture or high-risk integrations

For complex/critical tasks, run spec-compliance review on `complex_reasoning` model.
</complexity_routing>

<workflow>
0. Resolve spec directory:
   - `SPEC_DIR=.spec-workflow/specs/<spec-name>`
   - if `SPEC_DIR` does not exist, list available specs from `.spec-workflow/specs/*` and stop BLOCKED.
1. Read files from canonical paths:
   - `.spec-workflow/specs/<spec-name>/tasks.md`
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - if `tasks.md` is missing, stop BLOCKED and instruct to run `spw:tasks-plan <spec-name>`.
2. Select pending tasks by wave.
3. Execute up to `batch-size` tasks per batch (prefer safe parallelism).
4. For each task:
   - mark `[-]`
   - dispatch `task-implementer`
   - dispatch `spec-compliance-reviewer`
   - dispatch `code-quality-reviewer`
   - if all gates pass: log implementation details and mark `[x]`
   - if any gate fails: mark BLOCKED and stop current batch
5. At end of batch, run `spw:checkpoint <spec-name>`.
6. Continue only if checkpoint passes.
</workflow>

<strict_mode>
With `--strict true` (default):
- block continuation when checkpoint fails.
- block continuation when a task has no requirement traceability.
</strict_mode>

<completion_guidance>
After each batch:
- Show executed task IDs, commit/log evidence, and checkpoint status.
- If checkpoint PASS: suggest continuing with next batch via `spw:exec <spec-name>`.
- If checkpoint BLOCKED: stop and show exact corrective actions.

After full execution success:
- Recommend final validation review and optionally `/clear` before any new planning cycle.

If blocked by spec resolution:
- Show the exact missing path and list discovered specs from `.spec-workflow/specs/`.
- Recommend the corrected command using one discovered spec name.
</completion_guidance>
