---
name: oraculo:post-mortem
description: Analyze post-spec commits and generate reusable process learnings
argument-hint: "<spec-name> [--since-commit <sha>] [--until-ref <ref>] [--tags <tag1,tag2>] [--topic <short-subject>]"
---

<objective>
Analyze post-spec commits and generate reusable process learnings.
</objective>

<execution_context>
@.claude/workflows/oraculo/post-mortem.md
@.claude/workflows/oraculo/overlays/active/post-mortem.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/post-mortem.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/post-mortem.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
