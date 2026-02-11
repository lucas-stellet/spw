# Pipeline Dispatch Pattern

Sequential chain of subagents where each produces output that feeds the next.
A synthesizer at the end consolidates everything into the command's final artifact.

## Thin-Dispatch Rules

These rules are mandatory for all pipeline commands. They override any command-specific behavior.

### 1. Status-Only Reads

After dispatching any subagent, read ONLY `<subagent>/status.json`.
- If `status=pass`: proceed to next step.
- If `status=blocked`: read `<subagent>/report.md` to decide action (log + skip, or stop BLOCKED).
- Never read `report.md` when status is pass.

### 2. Path-Based Briefs

When subagent-B depends on output from subagent-A:
- Write the **filesystem path** to `subagent-A/report.md` in `subagent-B/brief.md`.
- Never copy or summarize report content into the brief.

Example brief content:
```
## Inputs
- Scope analysis: <run-dir>/qa-scope-analyst/report.md
- Requirements: .spec-workflow/specs/<spec-name>/requirements.md
```

### 3. Orchestrator Context Files

When the orchestrator generates context that is NOT a subagent report (e.g., MCP-extracted data, user clarification decisions, prototype observations), it MUST write that context to a file in the run directory BEFORE referencing it in any brief.

Convention:
- `<run-dir>/_orchestrator-context/<topic>.md`

Examples:
- `_orchestrator-context/prototype-observations.md` — screenshots, SPA content
- `_orchestrator-context/user-clarifications.md` — resolved CLARIFY items
- `_orchestrator-context/mcp-source-context.md` — inline MCP extraction notes

Briefs then reference these files by path, same as subagent reports:
```
## Inputs
- Prototype observations: <run-dir>/_orchestrator-context/prototype-observations.md
- User clarifications: <run-dir>/_orchestrator-context/user-clarifications.md
```

Never embed orchestrator-generated content directly in a brief's ## Task section.

### 4. Synthesizer Reads From Filesystem

The last subagent (synthesizer/writer) receives a brief listing ALL previous report paths.
It reads them directly from disk — the orchestrator does not relay content.

### 5. Run Structure

```
<phase>/_comms/<command>/run-NNN/
  <subagent-1>/brief.md, report.md, status.json
  <subagent-2>/brief.md, report.md, status.json
  <synthesizer>/brief.md, report.md, status.json
  _handoff.md
```

### 6. Resume Policy

On `continue-unfinished`:
- Skip subagents where `status.json` exists with `status=pass`.
- Redispatch missing or blocked subagents.
- Always rerun synthesizer.

### 7. Subagent Failure Policy

When a dispatched subagent fails (error, killed, timeout) without writing `status.json`:

1. Check if `report.md` exists and is non-empty:
   - If yes AND content is substantial: write `status.json` with `{"status": "pass", "summary": "..."}` and proceed.
   - If no or content is partial/empty: redispatch the subagent (same brief, same model).
2. Never complete a subagent's work inline. If the subagent's task requires writing output artifacts (beyond report.md/status.json), the orchestrator must redispatch — not write those artifacts itself.
3. Maximum 1 retry per subagent. If the retry also fails, stop with BLOCKED and report the failure.

### 8. Artifact Save

When the pipeline's final subagent (synthesizer/writer) writes the command's output artifact to its `report.md`, the orchestrator saves it to the canonical path using filesystem copy — never by reading content into its own context.

```
cp <run-dir>/<writer>/report.md <canonical-output-path>
```

If the command requires post-save validation (Mermaid syntax, dashboard markdown profile, MDX compilation), run validation tools/scripts on the saved file — do not Read the file into orchestrator context. If validation fails, re-dispatch the writer with fix instructions in a new brief iteration, or apply the Surgical Fix Policy below.

### 9. Surgical Fix Policy

When a critic/reviewer returns BLOCKED with a specific, mechanical fix (e.g., arithmetic correction, typo, missing escape character):

- **Threshold:** fix touches ≤ 3 lines in the writer's `report.md` AND requires no design judgment (pure factual/syntactic correction).
- **Allowed:** orchestrator applies the fix directly to the writer's `report.md`.
- **Required:** log every inline fix in `<run-dir>/_handoff.md` under a `## Inline Fixes` section with: line(s) changed, reason, original value → new value.
- **Re-run critic:** always re-dispatch the critic after an inline fix.

If the fix exceeds the threshold (> 3 lines or requires design judgment), re-dispatch the writer subagent with the critic's feedback in a new brief.

## Extension Points

Pipeline commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir and reading config, before first dispatch. Use for user intent gates, preflight checks, skill loading.
- **`pre_dispatch(<subagent>)`**: Before writing a specific subagent's brief. Use for conditional dispatch (e.g., selecting which designer to run based on a gate decision).
- **`post_dispatch(<subagent>)`**: After reading a subagent's status.json. Use for mid-pipeline decisions that affect subsequent dispatches.
- **`post_pipeline`**: After synthesizer completes, before writing _handoff.md. Use for artifact generation, approval reconciliation, completion guidance.

## Implementation Procedure

For step-by-step CLI commands to execute this pattern, see `dispatch-implementation.md`.
The CLI handles directory creation, brief skeleton generation, status validation, and handoff generation.
