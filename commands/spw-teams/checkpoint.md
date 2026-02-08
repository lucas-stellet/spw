---
name: spw:checkpoint
description: Subagent-driven quality gate between execution batches/waves
argument-hint: "<spec-name> [--scope batch|wave|phase]"
---

<objective>
Subagent-driven quality gate between execution batches/waves.
</objective>

<execution_context>
@.claude/workflows/spw/checkpoint.md
@.claude/workflows/spw/overlays/teams/checkpoint.md
</execution_context>

<process>
Follow the base workflow from `@.claude/workflows/spw/checkpoint.md` end-to-end.
Then apply the teams overlay from `@.claude/workflows/spw/overlays/teams/checkpoint.md` as additional policy.
Preserve existing guardrails, gates, and output contracts.
</process>
