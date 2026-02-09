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
@.claude/workflows/spw/overlays/active/status.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/spw/status.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/spw/overlays/active/status.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
