# QA Principles And Reports

Primary references:
- ISTQB CTFL v4.0 syllabus (testing principles, test management/reporting)
- https://www.istqb.org/certifications/certified-tester-foundation-level

## Planning Principles

Use the 7 testing principles when shaping scope and priority.

Operationally:
- apply risk-based prioritization first
- cover critical paths before broad permutations
- include negative/error-path checks
- refresh stale tests periodically (pesticide paradox)

## Minimum Test Plan Content

- objective and in-scope/out-of-scope
- requirement traceability
- strategy by level: smoke, regression, exploratory, API, browser
- environment/data prerequisites
- entry/exit criteria
- blockers and mitigation

## Minimum Execution Report Content

- run metadata (date, env, commit/spec)
- what was executed vs planned
- passed/failed/blocked counts
- failed scenarios with evidence links
- known gaps and residual risk
- release recommendation (go/no-go + conditions)

## Minimum Defect Report Content

- unique defect id and title
- severity/priority
- environment and build/version
- reproducible steps
- expected vs actual result
- evidence (screenshots, traces, response logs)
- impact/risk and workaround
- owner/status and retest notes
