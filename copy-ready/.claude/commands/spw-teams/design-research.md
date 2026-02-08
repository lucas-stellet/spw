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
@.claude/workflows/spw/overlays/teams/design-research.md
</execution_context>

<process>
Follow the base workflow from `@.claude/workflows/spw/design-research.md` end-to-end.
Then apply the teams overlay from `@.claude/workflows/spw/overlays/teams/design-research.md` as additional policy.
Preserve existing guardrails, gates, and output contracts.
</process>
