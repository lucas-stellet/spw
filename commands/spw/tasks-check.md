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
@.claude/workflows/spw/overlays/active/tasks-check.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/spw/tasks-check.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/spw/overlays/active/tasks-check.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
