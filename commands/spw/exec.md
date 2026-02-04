---
name: spw:exec
description: Execute tasks.md in batches with mandatory checkpoints
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<objective>
Execute tasks in batches, with mandatory pauses for checkpoint quality gates.
</objective>

<workflow>
1. Read `tasks.md` and select pending tasks by wave.
2. Execute up to `batch-size` tasks per batch (prefer safe parallelism).
3. For each task:
   - mark `[-]`
   - execute with a subagent
   - validate spec compliance + code quality
   - log implementation details
   - mark `[x]`
4. At end of batch, run `spw:checkpoint <spec-name>`.
5. Continue only if checkpoint passes.
</workflow>

<strict_mode>
With `--strict true` (default):
- block continuation when checkpoint fails.
- block continuation when a task has no requirement traceability.
</strict_mode>
