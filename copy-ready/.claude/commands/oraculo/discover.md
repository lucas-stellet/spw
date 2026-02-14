---
name: oraculo:discover
description: Discovery flow with subagents to generate requirements.md
argument-hint: "<spec-name> [--source <url-or-file.md>]"
---

<objective>
Discovery flow with subagents to generate requirements.md.
</objective>

<execution_context>
@.claude/workflows/oraculo/discover.md
@.claude/workflows/oraculo/overlays/active/discover.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/discover.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/discover.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
