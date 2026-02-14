---
name: oraculo:qa-check
description: Validate QA test plan selectors, traceability, and data feasibility against actual code
argument-hint: "<spec-name>"
---

<objective>
Validate QA test plan selectors, traceability, and data feasibility against actual code.
</objective>

<execution_context>
@.claude/workflows/oraculo/qa-check.md
@.claude/workflows/oraculo/overlays/active/qa-check.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/qa-check.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/qa-check.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
