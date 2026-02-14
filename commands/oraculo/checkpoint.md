---
name: oraculo:checkpoint
description: Subagent-driven quality gate between execution batches/waves
argument-hint: "<spec-name> [--scope batch|wave|phase]"
---

<objective>
Subagent-driven quality gate between execution batches/waves.
</objective>

<execution_context>
@.claude/workflows/oraculo/checkpoint.md
@.claude/workflows/oraculo/overlays/active/checkpoint.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/checkpoint.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/checkpoint.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
