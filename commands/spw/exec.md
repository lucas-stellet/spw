---
name: spw:exec
description: Subagent-driven task execution in batches with mandatory checkpoints
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<objective>
Execute tasks in batches, with mandatory pauses for checkpoint quality gates.
</objective>

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
1. Read `tasks.md` and select pending tasks by wave.
2. Execute up to `batch-size` tasks per batch (prefer safe parallelism).
3. For each task:
   - mark `[-]`
   - dispatch `task-implementer`
   - dispatch `spec-compliance-reviewer`
   - dispatch `code-quality-reviewer`
   - if all gates pass: log implementation details and mark `[x]`
   - if any gate fails: mark BLOCKED and stop current batch
4. At end of batch, run `spw:checkpoint <spec-name>`.
5. Continue only if checkpoint passes.
</workflow>

<strict_mode>
With `--strict true` (default):
- block continuation when checkpoint fails.
- block continuation when a task has no requirement traceability.
</strict_mode>
