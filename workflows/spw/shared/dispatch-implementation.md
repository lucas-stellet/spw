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
   Never assert codebase facts in ## Task — instruct the subagent to verify instead.

c) Dispatch Task tool with:
   - model: from dispatch-setup output
   - prompt: "Read <brief_path> and follow its instructions"

d) Read status:
```
spw tools dispatch-read-status <name> --run-dir <RUN_DIR>
```
**CRITICAL:** Always use `dispatch-read-status` to read subagent status. Never infer status from TaskOutput text, task-notification content, or any other source. `dispatch-read-status` is the single source of truth — it reads and validates `status.json` from the filesystem.

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

### 5. Inline Audit Dispatch

After a producer subagent generates a primary artifact (tasks.md, QA plan,
implementation code), run an inline audit to validate it before proceeding.

Follow the pattern defined in `@.claude/workflows/spw/shared/dispatch-inline-audit.md`.

**Quick reference:**

a) Initialize iteration tracking:
```
spw tools audit-iteration start --run-dir <RUN_DIR> --type <inline-audit|inline-checkpoint>
```

b) Create nested audit directory:
```
spw tools dispatch-init-audit --run-dir <RUN_DIR> --type <type> [--iteration N]
```

c) Setup and dispatch auditors inside the audit directory (use dispatch-setup
   with the audit_dir as run-dir).

d) Read aggregator/decider status via dispatch-read-status.

e) On BLOCKED, check retry budget and re-dispatch writer if allowed:
```
spw tools audit-iteration check --run-dir <RUN_DIR> --type <type>
spw tools audit-iteration advance --run-dir <RUN_DIR> --type <type> --result blocked
```

f) On exhausted retries, recommend the standalone check command.

**Per-task self-check and spot-check** (exec only):

After each task-implementer completes, use verify-task for both
self-check (inside subagent) and spot-check (orchestrator):

```
spw tools impl-log register --spec <name> --task-id <N> --wave <NN> ...
spw tools verify-task --spec <name> --task-id <N> --check-commit
```

See `@.claude/workflows/spw/shared/file-handoff.md` section 6 for the
full self-check policy.
