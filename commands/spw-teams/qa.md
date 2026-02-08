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
@.claude/workflows/spw/overlays/teams/qa.md
</execution_context>

<process>
Follow the base workflow from `@.claude/workflows/spw/qa.md` end-to-end.
Then apply the teams overlay from `@.claude/workflows/spw/overlays/teams/qa.md` as additional policy.
Preserve existing guardrails, gates, and output contracts.
</process>
