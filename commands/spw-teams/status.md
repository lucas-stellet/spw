---
name: spw:status
description: Summarize current spec stage, blockers, and exact next commands
argument-hint: "[<spec-name>] [--all false|true]"
---

<objective>
Summarize current spec stage, blockers, and exact next commands.
</objective>

<execution_context>
@.claude/workflows/spw/status.md
@.claude/workflows/spw/overlays/teams/status.md
</execution_context>

<process>
Follow the base workflow from `@.claude/workflows/spw/status.md` end-to-end.
Then apply the teams overlay from `@.claude/workflows/spw/overlays/teams/status.md` as additional policy.
Preserve existing guardrails, gates, and output contracts.
</process>
