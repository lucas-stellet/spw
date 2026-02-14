# Dispatch Patterns

Oraculo commands follow a **thin-dispatch** model: the orchestrator (main agent) never accumulates subagent output in its own context. It dispatches subagents, reads only `status.json` after each, and passes filesystem paths — not content — between stages. This document defines the three dispatch categories, their subcategories, and the rules that apply to all of them.

## Core Rules (all categories)

1. **Status-only reads.** After dispatching a subagent, the orchestrator reads only `<subagent>/status.json`. It never reads `report.md` in the normal flow.
2. **Report reads on failure only.** When `status.json` reports `status=blocked`, the orchestrator reads `report.md` to decide the next action (log and continue, or stop BLOCKED).
3. **Paths, not content.** When subagent-B depends on the output of subagent-A, the orchestrator writes the *path* to `subagent-A/report.md` in `subagent-B/brief.md`. It never relays report content.
4. **No codebase assertions.** Briefs instruct subagents to verify codebase facts rather than asserting them. Orchestrator findings go to `_orchestrator-context/`.
5. **Synthesizers read from filesystem.** The final subagent in any command (synthesizer, aggregator, writer) receives a brief listing all relevant report paths and reads them directly from disk.
6. **File-first handoff contract unchanged.** Every subagent writes `brief.md` (by orchestrator), `report.md`, and `status.json`. Every run writes `_handoff.md`.
7. **CLI-enforced dispatch.** Use `oraculo tools dispatch-init`, `dispatch-setup`, `dispatch-read-status`, and `dispatch-handoff` to create directories and validate structure. Never create run dirs or subagent dirs manually.

---

## Category 1: Pipeline

Sequential chain of subagents where each produces output that feeds the next. A synthesizer at the end consolidates everything into the command's final artifact.

**Directory structure:** `<phase>/_comms/<command>/run-NNN/<subagent>/`

Phase mapping: `discover` → `discover/_comms/run-NNN/`, `design-research`/`design-draft` → `design/_comms/<command>/run-NNN/`, `tasks-plan` → `planning/_comms/tasks-plan/run-NNN/`, `qa` → `qa/_comms/qa/run-NNN/`, `post-mortem` → `post-mortem/_comms/run-NNN/`.

**Dispatch pattern:**
```
orchestrator:
  dispatch subagent-A → read status.json only
  dispatch subagent-B (brief includes path to A/report.md) → read status.json only
  dispatch synthesizer (brief includes paths to all reports) → synthesizer reads from fs
```

**Resume:** redispatch missing/blocked subagents; always rerun synthesizer.

### Subcategory 1a: Research Pipeline

Gathers information, potentially from external sources (URLs, web search, Playwright MCP for SPAs). May have conditional branches (revision loops, multiple scouts for different URLs).

| Command | Subagent chain | Final artifact |
|---------|---------------|----------------|
| `discover` | scope-analyst → research scouts → requirements-synthesizer | PRD.md, requirements.md |
| `design-research` | research dispatches | DESIGN-RESEARCH.md |

**Distinguishing traits:**
- May dispatch multiple scouts for different sources (web URLs, code, prototypes).
- May include user interaction gates mid-pipeline (e.g., `discover` revision loop).
- External source reads (WebFetch, Playwright MCP) happen inside subagents, not in orchestrator.
- **MCP inline exception:** When a subagent needs session-scoped MCP tools (Linear, Playwright), the orchestrator runs dispatch-setup as normal but executes the work inline — still writing report.md and status.json to the subagent directory.

### Subcategory 1b: Synthesis Pipeline

Takes existing artifacts (spec documents, code, execution history) and produces a consolidated document. Strictly linear — no external source fetching, no conditional branches.

| Command | Subagent chain | Final artifact |
|---------|---------------|----------------|
| `design-draft` | design analysis → design-synthesizer | design.md |
| `tasks-plan` | task-decomposer → dependency-builder → tasks-writer | tasks.md |
| `qa` | scope-analyst → test-designer(s) → plan-synthesizer | QA-TEST-PLAN.md |
| `post-mortem` | analyzers → lessons-synthesizer | post-mortem report |

**Distinguishing traits:**
- All inputs are local artifacts (spec files, code, execution logs).
- `qa` has a tool selection branch (playwright/bruno/hybrid) that determines which designer subagent(s) run, but the pipeline is still linear within each branch.
- `post-mortem` may update memory index as a side effect.

---

## Category 2: Audit

Multiple independent auditors examine the same artifact(s) from different angles. An aggregator synthesizes their findings into a PASS/BLOCKED decision.

**Directory structure:** `<phase>/_comms/<command>/run-NNN/<auditor>/`

Phase mapping: `tasks-check` → `planning/_comms/tasks-check/run-NNN/`, `qa-check` → `qa/_comms/qa-check/run-NNN/`, `checkpoint` → `execution/waves/wave-NN/checkpoint/run-NNN/`.

**Dispatch pattern:**
```
orchestrator:
  dispatch auditor-1 ─┐
  dispatch auditor-2 ──┤ (parallel or sequential)
  dispatch auditor-3 ─┘
  read status.json from each
  dispatch aggregator (brief includes paths to all auditor reports)
  aggregator reads from fs → PASS/BLOCKED
```

**Resume:** redispatch only missing/blocked auditors; always rerun aggregator.

### Subcategory 2a: Artifact Audit

Validates documents against other documents. No source code reads.

| Command | Auditors | Aggregator | Output |
|---------|----------|-----------|--------|
| `tasks-check` | validation auditors | aggregator | TASKS-CHECK.md |

**Distinguishing traits:**
- All inputs are spec documents (requirements.md, design.md, tasks.md).
- Auditors are fully independent — safe to dispatch in parallel.

### Subcategory 2b: Code Audit

Validates artifacts against actual source code. At least one auditor reads implementation files.

| Command | Auditors | Aggregator | Output |
|---------|----------|-----------|--------|
| `qa-check` | traceability-auditor, selector-verifier, data-feasibility-checker | check-aggregator | QA-CHECK.md |
| `checkpoint` | evidence-collector, traceability-judge | release-gate-decider | CHECKPOINT-REPORT.md |

**Distinguishing traits:**
- One or more auditors read source files (heavier context per auditor).
- `checkpoint` is wave-aware: runs within `execution/waves/wave-NN/checkpoint/run-NNN/`.
- `qa-check`'s selector-verifier is the **only** QA subagent that reads implementation files.

---

## Category 3: Wave Execution

Iterative dispatch over a set of work items, split into waves. Each wave dispatches subagents for a bounded group of items, completes, then the next wave starts.

**Directory structure:**
```
execution/waves/wave-NN/execution/run-NNN/<subagent>/       (exec)
qa/_comms/qa-exec/waves/wave-NN/run-NNN/<subagent>/        (qa-exec)
```

**Dispatch pattern:**
```
orchestrator:
  dispatch state-scout → resume state (compact)
  resolve waves from work items + wave size config
  for each wave-NN:
    for each subagent in wave:
      dispatch subagent → read status.json only
      on blocked: read report.md, decide action
    write wave-NN/_wave-summary.json (from status.json data)
  dispatch synthesizer (brief includes paths to all wave summaries + reports)
  synthesizer reads from fs → final artifact
```

**Resume:** scout inspects `_wave-summary.json` per wave; skip completed waves; resume from first incomplete wave. Always rerun synthesizer.

### Subcategory 3a: Implementation Waves

Produce code changes. Require quality gates between waves and git hygiene.

| Command | Work items | Subagents per wave | Output |
|---------|-----------|-------------------|--------|
| `exec` | Tasks from tasks.md | task-implementer, compliance-reviewer, quality-reviewer | Code + commits |

**Distinguishing traits:**
- Side effects: code changes, git commits.
- Mandatory checkpoint (`oraculo:checkpoint`) between waves.
- User authorization gate between waves (configurable).
- Wave size from `[planning].max_wave_size`.

### Subcategory 3b: Validation Waves

Execute tests or checks without modifying code. Lighter gates between waves.

| Command | Work items | Subagents per wave | Output |
|---------|-----------|-------------------|--------|
| `qa-exec` | Scenarios from QA-TEST-PLAN.md | test-runner, evidence-collector | QA-EXECUTION-REPORT.md |

**Distinguishing traits:**
- No code side effects — only evidence artifacts (screenshots, traces, reports).
- No checkpoint between waves (synthesizer runs once at end).
- Wave size from `[qa].max_scenarios_per_wave`.
- Each wave re-authenticates (clean browser state with `--isolated`).

---

## Utility

| Command | Pattern | Notes |
|---------|---------|-------|
| `status` | Pipeline (3 subagents) | Dispatches `state-inspector`, `approval-auditor`, `next-step-planner` to gather state and produce STATUS-SUMMARY.md. |

---

## Command → Category Map

| Command | Category | Subcategory |
|---------|----------|-------------|
| `discover` | Pipeline | Research |
| `design-research` | Pipeline | Research |
| `design-draft` | Pipeline | Synthesis |
| `tasks-plan` | Pipeline | Synthesis |
| `qa` | Pipeline | Synthesis |
| `post-mortem` | Pipeline | Synthesis |
| `tasks-check` | Audit | Artifact |
| `qa-check` | Audit | Code |
| `checkpoint` | Audit | Code |
| `exec` | Wave Execution | Implementation |
| `qa-exec` | Wave Execution | Validation |
| `status` | Utility (Pipeline) | 3 subagents |

---

## Inline Audit

Producer commands (`tasks-plan`, `qa`, `exec`) can run audit subagents inline before proceeding to approval, avoiding a separate check command round-trip. The inline audit runs inside the producer's own `run-NNN/` directory using nested audit subdirectories.

**Types:**

| Type | Used by | Scope |
|------|---------|-------|
| `inline-audit` | `tasks-plan`, `qa` | Validate producer artifact (tasks.md, QA plan) |
| `inline-checkpoint` | `exec` | Validate wave completion (evidence, traceability, gate) |

**Dispatch pattern:**
```
producer (tasks-plan / qa / exec):
  ... produce primary artifact ...
  audit-iteration start → initialize state (max from [verification].inline_audit_max_iterations)
  dispatch-init-audit   → create _inline-audit/iteration-1/ or _inline-checkpoint/
  dispatch auditors     → same subagents as standalone check command
  read aggregator status.json
  IF PASS → done
  IF BLOCKED:
    audit-iteration check → allowed?
      YES → audit-iteration advance, re-dispatch writer with feedback, repeat
      NO  → STOP, recommend standalone check command
```

**CLI commands:**
- `oraculo tools audit-iteration start --run-dir R --type T [--max N]` — initialize `_iteration-state.json`
- `oraculo tools dispatch-init-audit --run-dir R --type T [--iteration N]` — create nested audit directory
- `oraculo tools audit-iteration check --run-dir R --type T` — check if another retry is allowed
- `oraculo tools audit-iteration advance --run-dir R --type T --result R` — increment iteration counter

**Anti-self-heal rule:** The orchestrator must not fix artifacts directly on BLOCKED. It re-dispatches the writer subagent with the aggregator report path, preserving separation between production and validation.

**Fallback:** When inline audit exhausts iterations, the orchestrator recommends the standalone check command (`oraculo:tasks-check`, `oraculo:qa-check`, or `oraculo:checkpoint`).

Full reference: `workflows/oraculo/shared/dispatch-inline-audit.md`.

---

## Shared Policy Reference

The thin-dispatch rules are codified in four category-specific policies referenced by all command workflows via `<shared_policies>`:
- `workflows/oraculo/shared/dispatch-pipeline.md` — pipeline sequencing (sequential chain → synthesizer)
- `workflows/oraculo/shared/dispatch-audit.md` — audit parallelism (auditors → aggregator)
- `workflows/oraculo/shared/dispatch-wave.md` — wave iteration (scout → waves → synthesizer)
- `workflows/oraculo/shared/dispatch-inline-audit.md` — inline audit retry loop (producer → auditors → retry)

Step-by-step CLI implementation procedure: `workflows/oraculo/shared/dispatch-implementation.md`.
The `oraculo tools dispatch-init` command enforces these category mappings deterministically.

---

## Frontmatter-Driven Dispatch Registry

Dispatch metadata (`phase`, `category`, `subcategory`, `comms_path`, `artifacts`) is declared directly in each workflow's `<dispatch_pattern>` section. The CLI parses it at startup from embedded workflow files — **no hardcoded Go map required**.

Adding a new dispatch-capable command only requires creating the workflow `.md` file with a valid `<dispatch_pattern>` section. The registry (`cli/internal/registry/`) loads automatically.

### `<dispatch_pattern>` keys

| Key | Required | Description |
|-----|----------|-------------|
| `category` | yes | Dispatch category: `pipeline`, `audit`, `wave-execution` |
| `subcategory` | yes | Subcategory: `research`, `synthesis`, `artifact`, `code`, `implementation`, `validation` |
| `phase` | yes | Spec directory phase: `discover`, `design`, `planning`, `execution`, `qa`, `post-mortem` |
| `comms_path` | yes | Template path for comms dir. Use `{wave}` placeholder for wave-aware commands |
| `artifacts` | no | Comma-separated list of artifact dirs to create under spec dir (e.g. `execution/_implementation-logs`) |
| `policy` | yes | `@`-reference to the shared dispatch policy |

Wave-awareness is **derived**: if `comms_path` contains `{wave}`, the command is wave-aware (requires `--wave` argument at dispatch time).
