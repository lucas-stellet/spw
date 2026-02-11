# MCP Approval Reconciliation

Approval source of truth is MCP.

When `spec-status` is incomplete or ambiguous:
1. Resolve `approvalId` from `spec-status` fields.
2. If missing, inspect `.spec-workflow/approvals/<spec-name>/approval_*.json`.
3. If `approvalId` exists, call MCP `approvals status`.
4. Never infer approval from phase labels or `overallStatus` alone.

## MCP Unavailability

When MCP tools (`spec-status`, `request-approval`, `get-approval-status`)
are unavailable, fail, or return errors:

1. Do NOT silently skip approval.
2. Do NOT assume approval from absence of status.
3. Log warning to `<run-dir>/_handoff.md` under `## Warnings`:
   ```
   WARNING: MCP approval unavailable. [spec-status call failed | MCP tools not loaded].
   Approval was not requested. Manual approval required before proceeding.
   ```
4. Check local `.spec-workflow/approvals/<spec-name>/approval_*.json` as last resort.
5. If no approval record exists, treat as `WAITING_FOR_APPROVAL` (never auto-approve).
6. Report `WAITING_FOR_APPROVAL` in completion guidance with instructions to approve
   via MCP dashboard or rerun after MCP is available.
