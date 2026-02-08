---
name: spw:exec
description: Subagent-driven task execution in batches with mandatory checkpoints
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<objective>
Subagent-driven task execution in batches with mandatory checkpoints.
</objective>

<execution_context>
@.claude/workflows/spw/exec.md
@.claude/workflows/spw/overlays/teams/exec.md
</execution_context>

<process>
Follow the base workflow from `@.claude/workflows/spw/exec.md` end-to-end.
Then apply the teams overlay from `@.claude/workflows/spw/overlays/teams/exec.md` as additional policy.
Preserve existing guardrails, gates, and output contracts.
</process>
