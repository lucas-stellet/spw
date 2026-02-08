---
name: spw:plan
description: Technical planning from existing requirements, orchestrated by subagents
argument-hint: "<spec-name> [--max-wave-size <N>]"
---

<objective>
Technical planning from existing requirements, orchestrated by subagents.
</objective>

<execution_context>
@.claude/workflows/spw/plan.md
@.claude/workflows/spw/overlays/teams/plan.md
</execution_context>

<process>
Follow the base workflow from `@.claude/workflows/spw/plan.md` end-to-end.
Then apply the teams overlay from `@.claude/workflows/spw/overlays/teams/plan.md` as additional policy.
Preserve existing guardrails, gates, and output contracts.
</process>
