---
name: spw:prd
description: Zero-to-PRD discovery flow with subagents to generate requirements.md
argument-hint: "<spec-name> [--source <url-or-file.md>]"
---

<objective>
Zero-to-PRD discovery flow with subagents to generate requirements.md.
</objective>

<execution_context>
@.claude/workflows/spw/prd.md
@.claude/workflows/spw/overlays/active/prd.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/spw/prd.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/spw/overlays/active/prd.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
