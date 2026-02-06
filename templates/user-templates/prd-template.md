---
spw:
  schema: 1
  spec: "<spec-name>"
  doc: "prd"
  status: "draft"
  source: "spw:prd"
  updated_at: "YYYY-MM-DD"
  inputs: []
  open_questions: []
  risk: "low"
---

# PRD (Product Requirements Document)

## Markdown Quality Rules (UI Approval)
- Keep Markdown render-safe for Spec Workflow UI (balanced headings/tables/fences/emphasis).
- Prefer plain Markdown over raw HTML blocks.
- Avoid task-style checkboxes (`- [ ]`, `- [-]`, `- [x]`) in requirements content.
- Keep REQ-IDs canonical and unique (`REQ-001`, `REQ-002`, ...).

## 1. Executive Summary
- Initiative name:
- Primary problem:
- Expected outcome:
- Stakeholders:

## 2. Context and Motivation
- Current situation:
- User/product pain:
- Why now:

## 3. Goals and Non-Goals
### Goals (v1)
-

### Non-goals (now)
-

## 4. Personas and Jobs To Be Done
### Primary persona
- Who they are:
- Usage context:
- JTBD:

### Secondary persona (if applicable)
- Who they are:
- JTBD:

## 5. Scope
### In Scope (v1)
-

### v2 / Later
-

### Out of Scope
-

## 6. Functional Requirements

### REQ-001 - [Title]
- User story:
- Acceptance criteria:
  - GIVEN [context] WHEN [action] THEN [outcome]
  - IF [condition] THEN [outcome]
- Priority: Must | Should | Could
- Dependencies:

### REQ-002 - [Title]
- User story:
- Acceptance criteria:
- Priority:
- Dependencies:

## 7. Non-Functional Requirements
- Performance:
- Security:
- Reliability/Availability:
- Observability:
- Accessibility:
- Compliance/Privacy:

## 8. UX and Main Flows
- Main flow:
- Error states:
- Empty states:
- Content/messaging considerations:

## 9. Success Metrics
- Primary KPI:
- Secondary KPI:
- Failure signal:

## 10. Risks and Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| | | |

## 11. Open Questions
-

## 12. Traceability to Design/Tasks
| REQ-ID | Design Section | Expected Tasks |
|--------|----------------|----------------|
| REQ-001 | | |
| REQ-002 | | |
