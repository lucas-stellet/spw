---
name: qa-validation-planning
description: Plan and structure QA validation for a spec using risk-based testing, selecting Playwright MCP for browser journeys, Bruno CLI for API validations, or a hybrid strategy. Use when users ask to validate features, design test plans, define QA scope, produce execution/defect reports, or choose between browser and API test tooling.
---

# QA Validation Planning

Use this skill to design QA coverage before execution.

## Core Workflow

1. Capture validation intent from the user.
2. Classify risk and test level.
3. Choose tool strategy (`playwright`, `bruno`, or `hybrid`).
4. Draft a test plan with explicit entry/exit criteria.
5. Prepare execution and defect reporting artifacts.

## Required Inputs

- Spec name and target scope
- Environment/base URL/API host
- Critical journeys or endpoints
- Release risk notes (known flaky areas, high-impact paths)

## Tool Selection Matrix

Use `Playwright MCP` when validation depends on browser behavior:
- multi-page user journey
- rendering/state transitions
- navigation + dialogs + upload/download
- browser-observed regressions

Use `Bruno CLI` when validation depends on API behavior:
- status codes and payload contracts
- auth/authorization outcomes
- idempotency and error responses
- data-driven API matrix (env, tags, folders)

Use `Hybrid` when both UI journey and API side-effects are critical.

## QA Principles To Enforce

Apply the ISTQB-aligned principles during planning and reporting:
- Testing shows presence of defects, not their absence.
- Exhaustive testing is impossible; prioritize by risk.
- Start testing activities early.
- Defects cluster; focus around high-defect areas.
- Beware pesticide paradox; refresh test data/cases.
- Testing is context-dependent.
- Absence-of-errors fallacy: correctness must meet user/business needs.

## Planning Rules

- Prioritize by business impact x likelihood.
- Keep scenarios traceable to requirements.
- Define test data and cleanup strategy.
- Separate must-pass smoke checks from deeper regression checks.
- Define evidence expected per scenario (logs, screenshots, traces, responses).
- Playwright MCP must be a pre-configured MCP server; agent uses the server's browser automation tools — never invokes npx or node scripts directly.

## Reporting Rules

Always produce three artifacts:
- Test plan: scope, risks, strategy, pass/fail criteria.
- Execution report: what ran, what failed, what is blocked, evidence links.
- Defect report: reproducible bug records with severity and impact.

If needed, load these references:
- Playwright MCP details: `references/playwright-mcp.md`
- Bruno CLI details: `references/bruno-cli.md`
- QA principles and report fields: `references/qa-principles-and-reports.md`
