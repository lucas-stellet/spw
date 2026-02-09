---
spw:
  schema: 1
  spec: "<spec-name>"
  doc: "qa-test-plan"
  status: "draft"
  source: "spw:qa"
  updated_at: "YYYY-MM-DD"
  inputs:
    - ".spec-workflow/specs/<spec-name>/requirements.md"
    - ".spec-workflow/specs/<spec-name>/design.md"
    - ".spec-workflow/specs/<spec-name>/tasks.md"
  requirements: []
  risk: "medium"
  open_questions: []
---

# QA Test Plan

## Validation Request
- Requested by:
- What to validate:
- Release context:
- In scope:
- Out of scope:

## Tool Strategy
- Selection mode: auto | forced
- Selected tool: playwright | bruno | hybrid
- Rationale:

## Risk Prioritization
| Risk ID | Risk Description | Impact | Likelihood | Priority | Mitigation Test |
|--------|------------------|--------|------------|----------|-----------------|
| R-001 | | | | P0 | |
| R-002 | | | | P1 | |

## Coverage Matrix
| Test ID | Requirement | Level (Smoke/Regression) | Type (UI/API) | Tool | Selector/Endpoint | Priority | Preconditions | Expected Evidence | Owner |
|---------|-------------|--------------------------|---------------|------|--------------------|----------|---------------|-------------------|-------|
| T-001 | REQ-001 | Smoke | UI | Playwright MCP | `[data-testid="login-btn"]` | P0 | | trace + screenshot | |
| T-002 | REQ-002 | Regression | API | Bruno CLI | `GET /api/v1/users` | P1 | | junit/json/html report | |

## Browser Validation Plan (Playwright MCP)
- Runtime mode: headless (mandatory)
- Target journeys:
- Data/accounts:
- Main assertions:
- Evidence required:
  - snapshots/screenshots
  - trace/session artifacts
  - network or console notes (if relevant)
- Execution notes:

## API Validation Plan (Bruno CLI)
- Target collections/folders/tags:
- Environment matrix:
- Main assertions:
- Bruno execution profile:
  - fail-fast or full-run policy:
  - sandbox mode: safe | developer
- Report artifacts required:
  - junit
  - json
  - html

## Entry Criteria
- Build/environment available:
- Required test data available:
- Known blockers resolved:

## Exit Criteria
- P0 tests pass:
- P1 failure tolerance:
- Open critical defects allowed: yes/no
- Residual risk threshold:

## Commands
```bash
# Playwright MCP (required headless mode)
npx @playwright/mcp@latest --headless --isolated --save-trace --output-dir .spec-workflow/specs/<spec-name>/qa/artifacts/playwright

# Bruno CLI (example execution)
bru run --env <env-name> --reporter-junit --reporter-json --reporter-html
```

## Reporting Outputs
- Execution report path: `.spec-workflow/specs/<spec-name>/qa/QA-EXECUTION-REPORT.md`
- Defect report path: `.spec-workflow/specs/<spec-name>/qa/QA-DEFECT-REPORT.md`
