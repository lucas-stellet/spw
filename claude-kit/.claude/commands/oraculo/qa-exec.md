---
name: oraculo:qa-exec
description: Execute validated QA test plan using verified selectors from QA-CHECK.md
argument-hint: "<spec-name> [--scope smoke|regression|full] [--rerun-failed true|false]"
---

<objective>
Execute validated QA test plan using verified selectors from QA-CHECK.md.
</objective>

<execution_context>
@.claude/workflows/oraculo/qa-exec.md
@.claude/workflows/oraculo/overlays/active/qa-exec.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/qa-exec.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/qa-exec.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
