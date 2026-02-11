# Dispatch Implementation

Procedural companion to dispatch-pipeline/audit/wave.
Those files define RULES. This file defines HOW via CLI.

## Dispatch Lifecycle

### 1. Init run (once per command invocation)

```
spw tools dispatch-init <command> <spec-name> [--wave NN]
```

Returns: run_dir, run_id, phase, category, subcategory, dispatch_policy, models.

The `dispatch_policy` field tells you which shared policy governs this run:
- `dispatch-pipeline` → sequential chain, synthesizer reads all reports from fs
- `dispatch-audit` → parallel/sequential auditors, aggregator reads all from fs
- `dispatch-wave` → iterative waves with _wave-summary.json per wave

Follow the rules in the corresponding `shared/<dispatch_policy>.md`.

### 2a. Write orchestrator context files (if needed)

Before filling any brief, persist orchestrator-generated context to files:

```
mkdir -p <RUN_DIR>/_orchestrator-context/
```

Write context files as needed:
- Prototype observations from Playwright/WebFetch sessions
- User clarification decisions from AskUserQuestion interactions
- Inline MCP extraction summaries

These files become inputs in subsequent briefs (referenced by path).

### 2b. For each subagent

a) Setup:
```
spw tools dispatch-setup <name> --run-dir <RUN_DIR> --model-alias <alias>
```
Returns: subagent_dir, brief_path, report_path, status_path, resolved model.

b) Edit brief.md: fill ## Inputs (file PATHS only, never content) and ## Task.

c) Dispatch Task tool with:
   - model: from dispatch-setup output
   - prompt: "Read <brief_path> and follow its instructions"

d) Read status:
```
spw tools dispatch-read-status <name> --run-dir <RUN_DIR>
```
If pass → proceed to next subagent. Do NOT read report.md.
If blocked → read report.md for decision (this is the ONLY case you read it).
If status.json missing (subagent failed/killed) → apply Subagent Failure Policy (see dispatch-pipeline.md).

### 3. MCP Inline Exception

When a subagent needs session-scoped MCP tools (Linear, Playwright, etc.):
- Run dispatch-setup as normal (creates dir + brief)
- Execute work inline (orchestrator does the subagent's task directly)
- Write report.md and status.json to the subagent directory
- Continue with dispatch-read-status

### 4. Finalize

```
spw tools dispatch-handoff --run-dir <RUN_DIR> [--command <cmd>]
```

Generates _handoff.md from all subagent status.json files.
For wave commands: also validates _wave-summary.json presence.
