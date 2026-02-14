# Spec Directory Structure

Each spec organizes its artifacts by **workflow phase**. Generated outputs and agent communications live together inside each phase directory — no separate `_generated/` or `_agent-comms/` top-level dumps.

## Design Principles

1. **Phase ownership.** Each phase directory owns its artifacts and its agent comms. To understand how `QA-CHECK.md` was produced, look inside `qa/` — both the output and the `_comms/` runs that generated it are there.
2. **Dashboard files stay at root.** `requirements.md`, `design.md`, and `tasks.md` remain at the spec root because the MCP dashboard reads them from there.
3. **Consistent comms structure.** Every phase uses `_comms/<command>/run-NNN/` (Pipeline, Audit) or `_comms/<command>/waves/wave-NN/run-NNN/` (Wave Execution). The file-first handoff contract (`brief.md`, `report.md`, `status.json`, `_handoff.md`) is unchanged.
4. **Waves are universal.** Both `execution/` and `qa/` use `waves/wave-NN/` for iterative work. The wave format is consistent across dispatch categories.

## Directory Layout

```
.spec-workflow/specs/<spec-name>/
│
├── requirements.md                        ← dashboard (MCP approval)
├── design.md                              ← dashboard (MCP approval)
├── tasks.md                               ← dashboard (MCP approval)
├── STATUS-SUMMARY.md                      ← output-only (not source of truth)
│
├── discover/                              ← phase: requirements
│   ├── PRD.md
│   ├── PRD-SOURCE-NOTES.md
│   ├── PRD-STRUCTURE.md
│   ├── PRD-REVISION-PLAN.md
│   ├── PRD-REVISION-QUESTIONS.md
│   ├── PRD-REVISION-NOTES.md
│   └── _comms/
│       └── run-NNN/
│           ├── <subagent>/brief.md, report.md, status.json
│           └── _handoff.md
│
├── design/                                ← phase: design
│   ├── DESIGN-RESEARCH.md
│   ├── SKILLS-DESIGN.md
│   └── _comms/
│       ├── design-research/run-NNN/
│       └── design-draft/run-NNN/
│
├── planning/                              ← phase: planning
│   ├── TASKS-CHECK.md
│   ├── SKILLS-EXEC.md
│   └── _comms/
│       ├── tasks-plan/run-NNN/
│       │   └── _inline-audit/             ← inline audit (from tasks-plan)
│       │       ├── _iteration-state.json
│       │       └── iteration-N/
│       │           ├── <auditor>/brief.md, report.md, status.json
│       │           └── _handoff.md
│       └── tasks-check/run-NNN/
│
├── execution/                             ← phase: implementation (waves)
│   ├── CHECKPOINT-REPORT.md
│   ├── _implementation-logs/
│   └── waves/
│       └── wave-NN/
│           ├── execution/run-NNN/
│           │   ├── <subagent>/brief.md, report.md, status.json
│           │   └── _handoff.md
│           ├── checkpoint/run-NNN/
│           │   ├── <subagent>/brief.md, report.md, status.json
│           │   └── _handoff.md
│           ├── _inline-checkpoint/        ← inline checkpoint (from exec)
│           │   ├── _iteration-state.json
│           │   ├── <auditor>/brief.md, report.md, status.json
│           │   └── _handoff.md
│           ├── _wave-summary.json
│           └── _latest.json
│
├── qa/                                    ← phase: validation
│   ├── QA-TEST-PLAN.md
│   ├── QA-CHECK.md
│   ├── QA-EXECUTION-REPORT.md
│   ├── QA-DEFECT-REPORT.md
│   ├── qa-artifacts/
│   │   └── wave-NN/                       ← evidence per wave
│   └── _comms/
│       ├── qa/run-NNN/                    ← pipeline: plan creation
│       │   └── _inline-audit/             ← inline audit (from qa)
│       │       ├── _iteration-state.json
│       │       └── iteration-N/
│       │           ├── <auditor>/brief.md, report.md, status.json
│       │           └── _handoff.md
│       ├── qa-check/run-NNN/              ← audit: selector verification
│       └── qa-exec/
│           └── waves/
│               └── wave-NN/
│                   ├── run-NNN/
│                   │   ├── <subagent>/brief.md, report.md, status.json
│                   │   └── _handoff.md
│                   ├── _wave-summary.json
│                   └── _latest.json
│
└── post-mortem/                           ← phase: retrospective
    ├── report.md
    └── _comms/
        └── run-NNN/
```

## Phase → Command Mapping

| Phase | Commands | Dispatch category |
|-------|----------|-------------------|
| `discover/` | `oraculo:discover` | Pipeline / Research |
| `design/` | `oraculo:design-research`, `oraculo:design-draft` | Pipeline / Research + Synthesis |
| `planning/` | `oraculo:tasks-plan`, `oraculo:tasks-check` | Pipeline / Synthesis + Audit / Artifact |
| `execution/` | `oraculo:exec`, `oraculo:checkpoint` | Wave Execution / Implementation + Audit / Code |
| `qa/` | `oraculo:qa`, `oraculo:qa-check`, `oraculo:qa-exec` | Pipeline / Synthesis + Audit / Code + Wave Execution / Validation |
| `post-mortem/` | `oraculo:post-mortem` | Pipeline / Synthesis |

When a phase contains commands from different dispatch categories (e.g., `qa/` has pipeline, audit, and wave), each command uses its own `_comms/` subdirectory with the appropriate structure for its category.

## Migration From Current Structure

| Current path | New path |
|--------------|----------|
| `_generated/PRD.md` | `discover/PRD.md` |
| `_generated/PRD-SOURCE-NOTES.md` | `discover/PRD-SOURCE-NOTES.md` |
| `_generated/PRD-STRUCTURE.md` | `discover/PRD-STRUCTURE.md` |
| `_generated/PRD-REVISION-*.md` | `discover/PRD-REVISION-*.md` |
| `_generated/DESIGN-RESEARCH.md` | `design/DESIGN-RESEARCH.md` |
| `_generated/SKILLS-*.md` | `design/SKILLS-DESIGN.md` or `planning/SKILLS-EXEC.md` |
| `_generated/TASKS-CHECK.md` | `planning/TASKS-CHECK.md` |
| `_generated/CHECKPOINT-REPORT.md` | `execution/CHECKPOINT-REPORT.md` |
| `_generated/QA-TEST-PLAN.md` | `qa/QA-TEST-PLAN.md` |
| `_generated/QA-CHECK.md` | `qa/QA-CHECK.md` |
| `_generated/QA-EXECUTION-REPORT.md` | `qa/QA-EXECUTION-REPORT.md` |
| `_generated/QA-DEFECT-REPORT.md` | `qa/QA-DEFECT-REPORT.md` |
| `_generated/qa-artifacts/` | `qa/qa-artifacts/` |
| `_generated/STATUS-SUMMARY.md` | `STATUS-SUMMARY.md` (spec root) |
| `_agent-comms/discover/run-NNN/` | `discover/_comms/run-NNN/` |
| `_agent-comms/design-research/run-NNN/` | `design/_comms/design-research/run-NNN/` |
| `_agent-comms/tasks-plan/run-NNN/` | `planning/_comms/tasks-plan/run-NNN/` |
| `_agent-comms/tasks-check/run-NNN/` | `planning/_comms/tasks-check/run-NNN/` |
| `_agent-comms/waves/wave-NN/` | `execution/waves/wave-NN/` |
| `_agent-comms/post-mortem/run-NNN/` | `post-mortem/_comms/run-NNN/` |
| `_agent-comms/qa/run-NNN/` | `qa/_comms/qa/run-NNN/` |
| `_agent-comms/qa-check/run-NNN/` | `qa/_comms/qa-check/run-NNN/` |
| `_agent-comms/qa-exec/run-NNN/` | `qa/_comms/qa-exec/waves/wave-NN/run-NNN/` |
| `_generated/` | **Eliminated** — artifacts live in their phase |
| `_agent-comms/` | **Eliminated** — `_comms/` lives inside each phase |
| `_implementation-logs/` | `execution/_implementation-logs/` |

## Conventions

- **Run-id format:** `run-NNN` (zero-padded 3-digit sequential). Never dates.
- **Wave format:** `wave-NN` (zero-padded 2-digit sequential).
- **`_comms/` prefix:** Underscore prefix signals internal/agent-only content, same convention as the previous `_agent-comms/` and `_generated/`.
- **Dashboard files:** `requirements.md`, `design.md`, `tasks.md` are the only files at spec root that the MCP dashboard reads. They must not move.
- **Phase directories are created on demand.** If `oraculo:qa` has never run, the `qa/` directory does not exist.
