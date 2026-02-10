# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is SPW

SPW is a command/template kit for `spec-workflow-mcp` that provides stricter agent execution patterns (planning gates, waves, checkpoints) with subagent-first orchestration and model routing (haiku for web scouting, opus for complex reasoning, sonnet for implementation).

## Canonical Sources (read order)

1. `README.md` — installation, usage, workflow reference
2. `AGENTS.md` — operational rules for agents and contributors (Portuguese)
3. `config/spw-config.toml` — runtime defaults

## Validation & Testing

There are no unit tests or a test framework. Validation is done via a checklist of syntax checks and smoke runs:

```bash
# Validate all shell scripts parse correctly
bash -n bin/spw
bash -n scripts/bootstrap.sh
bash -n scripts/install-spw-bin.sh
bash -n scripts/validate-thin-orchestrator.sh
bash -n hooks/session-start-sync-tasks-template.sh
bash -n copy-ready/install.sh

# Validate thin-orchestrator contract (wrapper sizes, workflow refs, mirror sync)
scripts/validate-thin-orchestrator.sh

# Smoke-test Node.js hooks (each reads JSON from stdin)
node hooks/spw-statusline.js <<< '{"workspace":{"current_dir":"'"$(pwd)"'"}}'
node hooks/spw-guard-user-prompt.js <<< '{"prompt":"/spw:plan"}'
node hooks/spw-guard-paths.js <<< '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":"README.md"}}'
node hooks/spw-guard-stop.js <<< '{}'
```

## Architecture

### Thin-Orchestrator Pattern

Commands and workflows are separated into two layers:

- **`commands/spw/*.md`** — Thin wrappers (max 60 lines) that define frontmatter metadata and point to a workflow via `<execution_context>` referencing `@.claude/workflows/spw/<command>.md`. These are what Claude Code slash commands (`/spw:exec`, `/spw:prd`, etc.) invoke.
- **`workflows/spw/*.md`** — Full orchestration logic: subagent definitions, policies, gates, state machines. Shared policy fragments live in `workflows/spw/shared/` (config resolution, file handoff, resume policy, skills policy, approval reconciliation).

Agent Teams uses base + overlay via symlinks: each command references `workflows/spw/overlays/active/<command>.md`, which is a symlink pointing to `../noop.md` (teams off) or `../teams/<command>.md` (teams on). The installer switches symlinks; no separate command directory needed.

### Mirror System

Source files in this repo must stay in sync with their `copy-ready/` counterparts:

| Source | Mirror |
|--------|--------|
| `commands/spw/` | `copy-ready/.claude/commands/spw/` |
| `workflows/spw/` | `copy-ready/.claude/workflows/spw/` |
| `workflows/spw/overlays/noop.md` | `copy-ready/.claude/workflows/spw/overlays/noop.md` |
| `workflows/spw/overlays/active/*.md` | `copy-ready/.claude/workflows/spw/overlays/active/*.md` (symlinks) |
| `templates/user-templates/` | `copy-ready/.spec-workflow/user-templates/` |
| `config/spw-config.toml` | `copy-ready/.spec-workflow/spw-config.toml` |
| `hooks/*.js\|*.sh` | `copy-ready/.claude/hooks/` |

`scripts/validate-thin-orchestrator.sh` enforces mirror integrity via `diff -rq`. Always update both sides in the same patch.

### Hooks (Node.js + Bash)

All hooks read JSON from stdin and use `hooks/spw-hook-lib.js` as shared library for TOML config reading, workspace detection, and violation reporting.

- **`spw-statusline.js`** — StatusLine hook: detects active spec from git diff/cache
- **`spw-guard-user-prompt.js`** — UserPromptSubmit: validates spec arg presence in SPW commands
- **`spw-guard-paths.js`** — PreToolUse (Write/Edit): prevents writes outside spec-workflow paths
- **`spw-guard-stop.js`** — Stop: checks file-first handoff completeness in recent runs
- **`session-start-sync-tasks-template.sh`** — SessionStart: syncs active tasks template variant based on TDD config

Hook enforcement mode is configured in `config/spw-config.toml` under `[hooks]`: `warn` (diagnostics only) or `block` (deny violating actions).

### CLI (`bin/spw`)

The `spw` CLI is a bash wrapper that caches the kit from GitHub and delegates to `copy-ready/install.sh`. Key commands: `spw install`, `spw update`, `spw doctor`, `spw status`, `spw skills`. Environment variables: `SPW_REPO`, `SPW_REF`, `SPW_HOME`, `SPW_KIT_DIR`, `SPW_AUTO_UPDATE`.

### Runtime Config

Canonical path: `.spec-workflow/spw-config.toml` (legacy fallback: `.spw/spw-config.toml`). This TOML controls model routing, execution gates (TDD, wave approval, commit-per-task tri-state, clean worktree), planning strategy (rolling-wave vs all-at-once), per-stage skill enforcement, hook behavior, and Agent Teams (with `exclude_phases` deny-list).

### SPW Command Entry Points

`spw:prd` (requirements) → `spw:plan` (design+tasks) → `spw:design-research` → `spw:design-draft` → `spw:tasks-plan` → `spw:tasks-check` → `spw:exec` (implementation) → `spw:checkpoint` (quality gate) → `spw:post-mortem` → `spw:qa` (validation planning) → `spw:qa-check` (plan validation) → `spw:qa-exec` (test execution) → `spw:status` (resume guidance)

### File-First Subagent Communication

Subagent handoffs use filesystem artifacts, not chat. Required files per subagent:
- `<subagent>/brief.md`, `<subagent>/report.md`, `<subagent>/status.json`
- `<run-dir>/_handoff.md`

Stored under `.spec-workflow/specs/<spec-name>/<phase>/_comms/` within each phase directory.

### Dispatch Categories

All commands follow a **thin-dispatch** model: the orchestrator reads only `status.json` after each subagent (never `report.md` unless blocked), and passes filesystem paths between stages — never inline content. Synthesizers/aggregators read all reports directly from disk. See `docs/DISPATCH-PATTERNS.md` for the full reference.

Commands are organized into three dispatch categories:

| Category | Pattern | Commands |
|----------|---------|----------|
| **Pipeline** | Sequential subagents → synthesizer | `prd`, `design-research`, `design-draft`, `tasks-plan`, `qa`, `post-mortem` |
| **Audit** | Parallel auditors → aggregator | `tasks-check`, `qa-check`, `checkpoint` |
| **Wave Execution** | Scout → iterative waves → synthesizer | `exec`, `qa-exec` |

Pipeline has two subcategories: **Research** (external sources, may branch — `prd`, `design-research`) and **Synthesis** (local artifacts, linear — the rest). Audit splits into **Artifact** (document-only — `tasks-check`) and **Code** (reads source — `qa-check`, `checkpoint`). Wave Execution splits into **Implementation** (code changes + checkpoints — `exec`) and **Validation** (no code changes — `qa-exec`).

### Spec Directory Structure

Artifacts are organized by **workflow phase**, not in flat dumps. Each phase directory owns its generated outputs and agent communications (`_comms/`). See `docs/SPEC-DIRECTORY-STRUCTURE.md` for the full reference and migration table.

```
.spec-workflow/specs/<spec-name>/
├── requirements.md                    ← dashboard (MCP approval)
├── design.md                          ← dashboard (MCP approval)
├── tasks.md                           ← dashboard (MCP approval)
├── STATUS-SUMMARY.md                  ← output-only (not source of truth)
│
├── prd/                               ← PRD.md, PRD-SOURCE-NOTES.md, ...
│   └── _comms/run-NNN/
├── design/                            ← DESIGN-RESEARCH.md, SKILLS-DESIGN.md
│   └── _comms/{design-research,design-draft}/run-NNN/
├── planning/                          ← TASKS-CHECK.md, SKILLS-EXEC.md
│   └── _comms/{tasks-plan,tasks-check}/run-NNN/
├── execution/                         ← CHECKPOINT-REPORT.md, _implementation-logs/
│   └── waves/wave-NN/{execution,checkpoint}/run-NNN/
├── qa/                                ← QA-TEST-PLAN.md, QA-CHECK.md, QA-*-REPORT.md
│   ├── qa-artifacts/wave-NN/
│   └── _comms/{qa,qa-check}/run-NNN/
│   └── _comms/qa-exec/waves/wave-NN/run-NNN/
└── post-mortem/                       ← report.md
    └── _comms/run-NNN/
```

### PR Review Optimization

Spec-workflow files are marked as `linguist-generated` via `.gitattributes` so GitHub collapses them by default in PR diffs. Reviewers see only feature code changes; spec artifacts are expandable on demand. The installer adds the rule automatically during `spw install`. See `docs/PR-REVIEW-OPTIMIZATION.md`.

```gitattributes
.spec-workflow/specs/** linguist-generated=true
```

## Key Constraints

- Approval is MCP-only (via `spec-workflow-mcp`), never manual chat approval. `STATUS-SUMMARY.md` is output-only, not a source of truth.
- `tasks.md` must follow dashboard compatibility: checkbox markers only on task lines with numeric IDs, `-` as list marker (never `*`), no nested checkboxes in metadata, `Files` in single line.
- Unfinished runs must prompt user decision (`continue-unfinished` or `delete-and-restart`), never auto-restart.
- `spw:exec` orchestrator never implements code directly — always dispatches subagents, even for single-task waves.
- When modifying behavior, defaults, or guardrails, update `README.md`, `AGENTS.md`, `docs/SPW-WORKFLOW.md`, `hooks/README.md`, and `copy-ready/README.md` in the same patch.
