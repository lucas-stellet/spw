# File-First Handoff Contract

Required files for each dispatched subagent:
- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

If any required handoff file is missing, return `BLOCKED`.

**status.json schema:**
```json
{"status": "pass"|"blocked", "summary": "one-line description"}
```

**CRITICAL — Run-id format**: MUST be `run-NNN` (zero-padded 3-digit sequential).
Examples: `run-001`, `run-002`, `run-003`.
NEVER use dates, timestamps, or any other format (e.g. `run-20260209-1` is WRONG).
To create a new run, scan existing sibling directories, extract the highest NNN, and increment by 1.

## Thin-Dispatch Integration

This contract defines the **file structure**. The category-level dispatch policies define **how the orchestrator interacts** with these files:

- `dispatch-pipeline.md` — sequential chain, status-only reads, path-based briefs
- `dispatch-audit.md` — parallel auditors, aggregator reads from filesystem
- `dispatch-wave.md` — wave iteration, wave summaries, scout-based resume

The 5 core thin-dispatch rules apply on top of this contract:
1. Orchestrator reads only `status.json` after dispatch (never `report.md` on pass).
2. Briefs contain filesystem paths to prior reports (never content).
   **Corollary to rule 2**: Orchestrator-generated context (prototype observations, user clarifications, MCP extraction notes) must also be persisted to `<run-dir>/_orchestrator-context/` files and referenced by path — never embedded inline in briefs.
   **Rule 2b**: Briefs must not assert codebase facts. Instruct subagents to verify
   (e.g., "check if test framework exists") instead of stating (e.g., "no test framework").
3. Synthesizers/aggregators read from disk directly.
4. Run structure follows category layout.
5. Resume skips completed subagents, always reruns final stage.
