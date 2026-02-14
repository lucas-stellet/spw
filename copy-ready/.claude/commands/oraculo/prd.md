---
name: oraculo:prd
description: Zero-to-PRD discovery flow with subagents to generate requirements.md
argument-hint: "<spec-name> [--source <url-or-file.md>]"
---

<objective>
Zero-to-PRD discovery flow with subagents to generate requirements.md.
</objective>

<execution_context>
@.claude/workflows/oraculo/prd.md
@.claude/workflows/oraculo/overlays/active/prd.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/prd.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/prd.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
