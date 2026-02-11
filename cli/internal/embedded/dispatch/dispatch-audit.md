# Audit Dispatch Pattern

Multiple independent auditors examine the same artifact(s) from different angles.
An aggregator synthesizes their findings into a PASS/BLOCKED decision.

## Thin-Dispatch Rules

These rules are mandatory for all audit commands. They override any command-specific behavior.

### 1. Status-Only Reads

After dispatching any auditor, read ONLY `<auditor>/status.json`.
- If `status=pass`: proceed to next auditor or aggregator.
- If `status=blocked`: read `<auditor>/report.md` to decide action (log + continue, or stop BLOCKED).
- Never read `report.md` when status is pass.

### 2. Path-Based Briefs

When dispatching the aggregator:
- Write **filesystem paths** to all auditor `report.md` files in `aggregator/brief.md`.
- Never copy or summarize auditor report content into the brief.

### 3. Aggregator Reads From Filesystem

The aggregator receives a brief listing ALL auditor report paths.
It reads them directly from disk — the orchestrator does not relay content.

### 4. Run Structure

```
<phase>/_comms/<command>/run-NNN/
  <auditor-1>/brief.md, report.md, status.json
  <auditor-2>/brief.md, report.md, status.json
  <auditor-3>/brief.md, report.md, status.json
  <aggregator>/brief.md, report.md, status.json
  _handoff.md
```

### 5. Resume Policy

On `continue-unfinished`:
- Skip auditors where `status.json` exists with `status=pass`.
- Redispatch missing or blocked auditors.
- Always rerun aggregator.

## Dispatch Modes

Auditors may be dispatched in **parallel** (when fully independent) or **sequentially** (when one auditor informs another). The command workflow specifies the mode.

## Extension Points

Audit commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir, before first auditor dispatch. Use for precondition checks (e.g., verify target artifact exists).
- **`pre_dispatch(<auditor>)`**: Before writing a specific auditor's brief. Use for conditional skip logic.
- **`post_dispatch(<auditor>)`**: After reading an auditor's status.json. Use for early-exit decisions.
- **`post_pipeline`**: After aggregator completes, before writing _handoff.md. Use for artifact generation, next-step guidance.
