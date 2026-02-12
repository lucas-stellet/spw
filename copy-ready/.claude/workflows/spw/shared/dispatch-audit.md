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

### 6. Auditor Failure Policy

When a dispatched auditor fails without writing `status.json`:
1. If `report.md` exists and is substantial: write `status.json` with pass and proceed.
2. If not: redispatch (same brief, same model). Maximum 1 retry.
3. Never resolve a BLOCKED auditor by creating, modifying, or deleting artifacts outside the auditor's comms directory (implementation logs, source files, spec files, dashboard files).
4. If an auditor returns `status=blocked`, the orchestrator MUST NOT take corrective action. Record the block and proceed to aggregator. Only a new run may resolve a previously blocked auditor.

### 7. Handoff Consistency

- If ANY auditor `status.json` reports `blocked`, the final verdict MUST be BLOCKED.
- The aggregator may override only in a NEW run where the blocked auditor is re-dispatched.
- `_handoff.md` MUST list every auditor's final status. If it shows any blocked but the artifact says PASS, the run is invalid.

### 8. No Codebase Assertions in Briefs

Briefs must never assert codebase facts. Instruct auditors to verify instead.
- BAD: "task 5 may not have a log since it's CSS/i18n"
- GOOD: "Verify whether task 5 has a corresponding implementation log"

## Dispatch Modes

Auditors may be dispatched in **parallel** (when fully independent) or **sequentially** (when one auditor informs another). The command workflow specifies the mode.

## Extension Points

Audit commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir, before first auditor dispatch. Use for precondition checks (e.g., verify target artifact exists).
- **`pre_dispatch(<auditor>)`**: Before writing a specific auditor's brief. Use for conditional skip logic.
- **`post_dispatch(<auditor>)`**: After reading an auditor's status.json. Use for early-exit decisions.
- **`post_pipeline`**: After aggregator completes, before writing _handoff.md. Use for artifact generation, next-step guidance.

## Implementation Procedure

For step-by-step CLI commands to execute this pattern, see `dispatch-implementation.md`.
The CLI handles directory creation, brief skeleton generation, status validation, and handoff generation.
