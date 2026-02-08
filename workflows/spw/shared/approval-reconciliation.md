# MCP Approval Reconciliation

Approval source of truth is MCP.

When `spec-status` is incomplete or ambiguous:
1. Resolve `approvalId` from `spec-status` fields.
2. If missing, inspect `.spec-workflow/approvals/<spec-name>/approval_*.json`.
3. If `approvalId` exists, call MCP `approvals status`.
4. Never infer approval from phase labels or `overallStatus` alone.
