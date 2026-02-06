---
name: mermaid-architecture
description: Use when drafting architecture/design docs, creating system diagrams, mapping request flows, event-driven pipelines, state machines, or deployment topology. Produces valid Mermaid blocks for spec-workflow dashboards.
---

# Mermaid Architecture

Use this skill during design phases to transform architecture decisions into clear Mermaid diagrams that render reliably in `spec-workflow-mcp` dashboards.

## Output Contract

- Always use fenced code blocks: ```` ```mermaid ... ``` ````
- Prefer one focused diagram per section.
- Keep node IDs short and stable; keep labels descriptive.
- Use explicit direction (`flowchart TD` or `flowchart LR`).
- Avoid HTML/script/custom injections in labels.
- Keep labels consistent with terms used in `design.md`.

## Choose The Right Diagram

- Module/layer boundaries: `flowchart`
- Temporal request interaction: `sequenceDiagram`
- Async/event pipelines: `flowchart` with queue/topic nodes
- Workflow lifecycle: `stateDiagram-v2`
- Environment/deployment topology: `flowchart` grouped by environment

## Common Architecture Examples

### 1) Layered Architecture
```mermaid
flowchart LR
  Client[Client/UI] --> API[HTTP API]
  API --> Service[Domain Service]
  Service --> Repo[Repository]
  Repo --> DB[(PostgreSQL)]
  Service --> Events[[Domain Events]]
```

### 2) C4-Style Container View
```mermaid
flowchart TB
  User[End User]
  Web[Web App]
  BFF[BFF/API]
  Worker[Background Worker]
  DB[(Main DB)]
  Cache[(Redis)]
  Broker[[Message Broker]]

  User --> Web
  Web --> BFF
  BFF --> DB
  BFF --> Cache
  BFF --> Broker
  Broker --> Worker
  Worker --> DB
```

### 3) Request Sequence With Auth + Error Path
```mermaid
sequenceDiagram
  actor U as User
  participant FE as Frontend
  participant API as API
  participant Auth as Auth Service
  participant DB as Database

  U->>FE: Submit form
  FE->>API: POST /orders
  API->>Auth: Validate token
  Auth-->>API: Token valid
  API->>DB: Insert order
  DB-->>API: order_id
  API-->>FE: 201 Created

  alt validation fails
    API-->>FE: 422 Validation Error
  end
```

### 4) Event-Driven Processing Pipeline
```mermaid
flowchart LR
  API[API] --> TopicOrders[[orders.created]]
  TopicOrders --> Billing[Billing Worker]
  TopicOrders --> Email[Email Worker]
  Billing --> TopicPaid[[orders.paid]]
  TopicPaid --> Notify[Notification Worker]
```

### 5) State Machine For Approval/Delivery
```mermaid
stateDiagram-v2
  [*] --> Draft
  Draft --> PendingApproval: submit
  PendingApproval --> Approved: approve
  PendingApproval --> Rejected: reject
  Rejected --> Draft: revise
  Approved --> Implemented: deliver
  Implemented --> [*]
```

## How To Apply In `design.md`

1. In `## Architecture`, include one main-flow diagram.
2. In `## Error Strategy` or `## Contracts`, include one focused sequence/state diagram for critical risk.
3. Make sure diagram edges/states map to requirement groups.
4. Update Mermaid when architectural decisions change.

## Quality Checklist

- Diagram renders in dashboard preview.
- Names in diagram match document terminology.
- Dependencies are real (no speculative edges).
- Critical error/retry behavior is explicit.
- No orphan component or state remains undocumented.
