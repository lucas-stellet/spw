# Dispatch Patterns

SPW commands follow a **thin-dispatch** model: the orchestrator (main agent) never accumulates subagent output in its own context. It dispatches subagents, reads only `status.json` after each, and passes filesystem paths — not content — between stages. This document defines the three dispatch categories, their subcategories, and the rules that apply to all of them.

## Core Rules (all categories)

1. **Status-only reads.** After dispatching a subagent, the orchestrator reads only `<subagent>/status.json`. It never reads `report.md` in the normal flow.
2. **Report reads on failure only.** When `status.json` reports `status=blocked`, the orchestrator reads `report.md` to decide the next action (log and continue, or stop BLOCKED).
3. **Paths, not content.** When subagent-B depends on the output of subagent-A, the orchestrator writes the *path* to `subagent-A/report.md` in `subagent-B/brief.md`. It never relays report content.
4. **Synthesizers read from filesystem.** The final subagent in any command (synthesizer, aggregator, writer) receives a brief listing all relevant report paths and reads them directly from disk.
5. **File-first handoff contract unchanged.** Every subagent writes `brief.md` (by orchestrator), `report.md`, and `status.json`. Every run writes `_handoff.md`.
6. **CLI-enforced dispatch.** Use `spw tools dispatch-init`, `dispatch-setup`, `dispatch-read-status`, and `dispatch-handoff` to create directories and validate structure. Never create run dirs or subagent dirs manually.

---

## Category 1: Pipeline

Sequential chain of subagents where each produces output that feeds the next. A synthesizer at the end consolidates everything into the command's final artifact.

**Directory structure:** `<phase>/_comms/<command>/run-NNN/<subagent>/`

Phase mapping: `prd` → `prd/_comms/run-NNN/`, `design-research`/`design-draft` → `design/_comms/<command>/run-NNN/`, `tasks-plan` → `planning/_comms/tasks-plan/run-NNN/`, `qa` → `qa/_comms/qa/run-NNN/`, `post-mortem` → `post-mortem/_comms/run-NNN/`.

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
| `prd` | scope-analyst → research scouts → requirements-synthesizer | PRD.md, requirements.md |
| `design-research` | research dispatches | DESIGN-RESEARCH.md |

**Distinguishing traits:**
- May dispatch multiple scouts for different sources (web URLs, code, prototypes).
- May include user interaction gates mid-pipeline (e.g., `prd` revision loop).
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
- Mandatory checkpoint (`spw:checkpoint`) between waves.
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
| `status` | Read-only | No subagents. Reads artifacts and produces STATUS-SUMMARY.md. |

---

## Command → Category Map

| Command | Category | Subcategory |
|---------|----------|-------------|
| `prd` | Pipeline | Research |
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
| `status` | Utility | — |

---

## Shared Policy Reference

The thin-dispatch rules are codified in three category-specific policies referenced by all command workflows via `<shared_policies>`:
- `workflows/spw/shared/dispatch-pipeline.md` — pipeline sequencing (sequential chain → synthesizer)
- `workflows/spw/shared/dispatch-audit.md` — audit parallelism (auditors → aggregator)
- `workflows/spw/shared/dispatch-wave.md` — wave iteration (scout → waves → synthesizer)

Step-by-step CLI implementation procedure: `workflows/spw/shared/dispatch-implementation.md`.
The `spw tools dispatch-init` command enforces these category mappings deterministically.
