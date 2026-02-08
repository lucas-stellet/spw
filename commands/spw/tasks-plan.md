---
name: spw:tasks-plan
description: Subagent-driven tasks.md generation for waves, parallelism, and per-task TDD
argument-hint: "<spec-name> [--mode initial|next-wave] [--max-wave-size <N>] [--allow-no-test-exception true|false]"
---

<objective>
Subagent-driven tasks.md generation for waves, parallelism, and per-task TDD.
</objective>

<execution_context>
@.claude/workflows/spw/tasks-plan.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/spw/tasks-plan.md` end-to-end.
Preserve existing guardrails, gates, and output contracts.
</process>
