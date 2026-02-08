---
name: spw:tasks-check
description: Subagent-driven tasks.md validation (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<objective>
Subagent-driven tasks.md validation (traceability, dependencies, tests).
</objective>

<execution_context>
@.claude/workflows/spw/tasks-check.md
@.claude/workflows/spw/overlays/teams/tasks-check.md
</execution_context>

<process>
Follow the base workflow from `@.claude/workflows/spw/tasks-check.md` end-to-end.
Then apply the teams overlay from `@.claude/workflows/spw/overlays/teams/tasks-check.md` as additional policy.
Preserve existing guardrails, gates, and output contracts.
</process>
