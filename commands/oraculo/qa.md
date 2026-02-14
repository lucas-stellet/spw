---
name: oraculo:qa
description: Build a QA validation plan with concrete selectors for a spec using Playwright MCP, Bruno CLI, or hybrid strategy
argument-hint: "<spec-name> [--focus <what-to-validate>] [--tool auto|playwright|bruno|hybrid]"
---

<objective>
Build a QA validation plan for a spec using Playwright MCP, Bruno CLI, or hybrid strategy.
</objective>

<execution_context>
@.claude/workflows/oraculo/qa.md
@.claude/workflows/oraculo/overlays/active/qa.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/oraculo/qa.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/oraculo/overlays/active/qa.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
