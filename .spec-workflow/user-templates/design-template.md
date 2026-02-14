---
oraculo:
  schema: 1
  spec: "<spec-name>"
  doc: "design"
  status: "draft"
  source: "oraculo:design-draft"
  updated_at: "YYYY-MM-DD"
  inputs:
    - ".spec-workflow/specs/<spec-name>/requirements.md"
    - ".spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md"
  requirements: []
  decisions: []
  risk: "low"
  open_questions: []
---

# Design Document

## Overview
- Technical objective of the solution:
- Relationship to the product problem:

## Requirement Mapping (Required)
| Requirement | Technical Decision | Components/Files | Test Strategy |
|-------------|--------------------|------------------|---------------|
| REQ-001 | | | |
| REQ-002 | | | |

## Considered Options and Trade-offs
### Option A (recommended)
- Pros:
- Cons:
- When not to use:

### Option B
- Pros:
- Cons:

## Alignment with Steering Docs
### tech.md
- Applied standards:

### structure.md
- Applied organization conventions:

## Code Reuse
- Modules/files to reuse:
- Planned extensions:
- Code to avoid / critical legacy:

## Architecture
### Boundaries
- Contexts/modules:
- Allowed dependencies between modules:

### Main flow
```mermaid
graph TD
  Client[Client] --> API[Entry Layer]
  API --> Domain[Domain]
  Domain --> Infra[Infra/Adapters]
```

## Contracts and Data Models
- Relevant entities and fields:
- API/event contracts:
- Migrations/schema (if applicable):

## Error Strategy
- Expected errors and handling:
- Unexpected errors and fallback:
- Observability (logs/metrics/traces):

## Test Strategy
### Unit
-

### Integration
-

### E2E/UAT
-

## Risks and Mitigations
- Risk:
- Mitigation:
