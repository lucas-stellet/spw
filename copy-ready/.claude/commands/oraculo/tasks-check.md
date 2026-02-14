---
name: oraculo:tasks-check
description: Subagent-driven tasks.md validation (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<objective>
Subagent-driven tasks.md validation (traceability, dependencies, tests).
</objective>

<execution_context>
@.claude/workflows/oraculo/tasks-check.md
@.claude/workflows/oraculo/overlays/active/tasks-check.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/tasks-check.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/tasks-check.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
