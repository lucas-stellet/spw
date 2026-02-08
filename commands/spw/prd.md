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
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/spw/prd.md` end-to-end.
Preserve existing guardrails, gates, and output contracts.
</process>
