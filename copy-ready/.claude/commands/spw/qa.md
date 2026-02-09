---
name: spw:qa
description: Build a QA validation plan for a spec using Playwright MCP, Bruno CLI, or hybrid strategy
argument-hint: "<spec-name> [--focus <what-to-validate>] [--tool auto|playwright|bruno|hybrid]"
---

<objective>
Build a QA validation plan for a spec using Playwright MCP, Bruno CLI, or hybrid strategy.
</objective>

<execution_context>
@.claude/workflows/spw/qa.md
@.claude/workflows/spw/overlays/active/qa.md
</execution_context>

<process>
Follow the workflow from `@.claude/workflows/spw/qa.md` end-to-end.
Apply any overlay policy from `@.claude/workflows/spw/overlays/active/qa.md`.
Preserve existing guardrails, gates, and output contracts.
</process>
