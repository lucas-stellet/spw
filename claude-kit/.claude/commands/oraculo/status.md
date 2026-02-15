---
name: oraculo:status
description: Summarize current spec stage, blockers, and exact next commands
argument-hint: "[<spec-name>] [--all false|true]"
---

<objective>
Summarize current spec stage, blockers, and exact next commands.
</objective>

<execution_context>
@.claude/workflows/oraculo/status.md
@.claude/workflows/oraculo/overlays/active/status.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/status.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/status.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
