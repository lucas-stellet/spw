---
name: oraculo:exec
description: Subagent-driven task execution in batches with mandatory checkpoints
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<objective>
Subagent-driven task execution in batches with mandatory checkpoints.
</objective>

<execution_context>
@.claude/workflows/oraculo/exec.md
@.claude/workflows/oraculo/overlays/active/exec.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/exec.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/exec.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
