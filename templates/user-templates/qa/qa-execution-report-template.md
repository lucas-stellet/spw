---
spw:
  schema: 1
  spec: "<spec-name>"
  doc: "qa-execution-report"
  status: "draft"
  source: "spw:qa-exec"
  updated_at: "YYYY-MM-DD"
  inputs:
    - ".spec-workflow/specs/<spec-name>/_generated/QA-TEST-PLAN.md"
  requirements: []
  risk: "medium"
  open_questions: []
---

# QA Execution Report

## Run Metadata
- Date/time:
- Environment:
- Spec:
- Build/commit:
- Executor:
- Tool strategy used: playwright | bruno | hybrid

## Planned vs Executed
| Metric | Planned | Executed | Notes |
|--------|---------|----------|-------|
| Total scenarios | | | |
| P0 scenarios | | | |
| P1 scenarios | | | |
| API scenarios | | | |
| Browser scenarios | | | |

## Result Summary
| Status | Count |
|--------|-------|
| Passed | |
| Failed | |
| Blocked | |
| Not Run | |

## Scenario Results
| Test ID | Priority | Tool | Status | Evidence | Defect ID |
|---------|----------|------|--------|----------|-----------|
| T-001 | P0 | Playwright MCP | PASS | | |
| T-002 | P1 | Bruno CLI | FAIL | | BUG-001 |

## Playwright MCP Evidence
- MCP server: playwright
- Screenshots captured:
- Console messages collected:
- Generated test scripts (if available):
- Extra notes:

## Bruno CLI Evidence
- Run command:
- Report paths:
  - junit:
  - json:
  - html:
- Environment/sandbox notes:

## Defects and Blockers
- Critical defects:
- High defects:
- Blocked scenarios and reason:

## Residual Risks
- Risk still open:
- Impact if released:
- Proposed mitigation:

## Recommendation
- Decision: GO | NO-GO | GO WITH CONDITIONS
- Conditions (if any):
- Retest scope:
