---
name: spw:design-research
description: Subagent-driven technical research to prepare design.md
argument-hint: "<spec-name> [--focus <topic>] [--web-depth low|medium|high]"
---

<objective>
Subagent-driven technical research to prepare design.md.
</objective>

<execution_context>
@.claude/workflows/spw/design-research.md
@.claude/workflows/spw/overlays/active/design-research.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/spw/design-research.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/spw/overlays/active/design-research.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
