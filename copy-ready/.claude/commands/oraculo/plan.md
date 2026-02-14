---
name: oraculo:plan
description: Technical planning from existing requirements, orchestrated by subagents
argument-hint: "<spec-name> [--max-wave-size <N>]"
---

<objective>
Technical planning from existing requirements, orchestrated by subagents.
</objective>

<execution_context>
@.claude/workflows/oraculo/plan.md
@.claude/workflows/oraculo/overlays/active/plan.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/plan.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/plan.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
