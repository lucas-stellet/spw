# Oráculo

![Version](https://img.shields.io/badge/version-1.16.0-blue)
![License: MIT](https://img.shields.io/badge/license-MIT-green)

## Table of Contents

- [What is Oraculo?](#what-is-oraculo)
- [Quick Start](#quick-start)
- [Where to start](#where-to-start)
- [Installation](#installation)
- [Command entry points](#command-entry-points)
- [Thin-Orchestrator Architecture](#thin-orchestrator-architecture)
- [Dashboard Markdown Compatibility](#dashboard-markdown-compatibility-spec-workflow-mcp)
- [Mermaid for Architecture Design](#mermaid-for-architecture-design)
- [QA Validation (3-Phase)](#qa-validation-3-phase)
- [Glossary](#glossary)

## What is Oraculo?

Oraculo is a toolkit that adds structured workflow orchestration to Claude Code projects. Instead of letting an agent tackle an entire feature in one shot, Oráculo breaks work into phases -- requirements, design, planning, implementation, and QA -- each with its own quality gates and approval checkpoints.

Every phase dispatches specialized subagents with model routing: haiku handles lightweight web scouting, opus drives complex reasoning, and sonnet does the implementation drafting. Agents communicate through filesystem artifacts (not chat), so handoffs are reproducible and auditable. You drive the whole process through slash commands in Claude Code (e.g., `/oraculo:discover`, `/oraculo:exec`) while `spec-workflow-mcp` serves as the source of truth for artifacts and approvals.

## Quick Start

After [installing](#installation), run these commands inside a Claude Code session:

1. `/oraculo:discover my-feature` -- Generate a requirements document from your feature description
2. `/oraculo:plan my-feature` -- Create a design document and break it into executable tasks
3. `/oraculo:exec my-feature` -- Implement tasks in waves with automatic checkpoints
4. `/oraculo:qa my-feature` -- Build and run a QA validation plan

Each command handles subagent dispatch, file handoff, and quality gates automatically. Between steps, artifacts are stored under `.spec-workflow/specs/my-feature/` and approvals flow through `spec-workflow-mcp`.

## Where to start

- This file is the main source of truth for usage and operations.
- Agent/contributor operational rules are in `AGENTS.md`.
- Keep `docs/ORACULO-WORKFLOW.md` and `claude-kit/README.md` as lightweight pointers to this README.

## Installation

### 1. Install the CLI

The bootstrap script downloads the compiled Go binary from the latest GitHub Release and installs it to `~/.local/bin/oraculo`. Requires `curl` and `tar`.

```bash
curl -fsSL https://raw.githubusercontent.com/lucas-stellet/oraculo/main/scripts/bootstrap.sh | bash
```

**From a local clone (build from source):**

```bash
cd cli && go build -o ~/.local/bin/oraculo ./cmd/oraculo/
```

Run `oraculo` with no arguments to see available commands. Use `oraculo update` to self-update from the latest release.

### 2. Install in your project

From your project root:

```bash
oraculo install
```

This copies commands, workflows, hooks, config, and skills into your project. For a manual install, you can also run `cp -R /path/to/oraculo/claude-kit/. .` instead.

### 3. Post-install checklist

Required:
1. Merge `.claude/settings.json.example` into your `.claude/settings.json` (if needed).
2. Review `.spec-workflow/oraculo.toml`, especially `[planning].tasks_generation_strategy` and `[planning].max_wave_size`.
3. Start a new Claude Code session so the SessionStart hook can sync the active tasks template.

Optional:
- Set per-stage skill enforcement: `skills.design.enforce_required` and `skills.implementation.enforce_required` in `oraculo.toml`.
- Enable Oraculo statusline (see `.claude/settings.json.example`).
- Enable enforcement hooks with `hooks.enforcement_mode = "warn"` or `"block"` in `oraculo.toml`.
- For QA browser validation and URL exploration in planning stages, add Playwright MCP:
  ```
  claude mcp add playwright -- npx @playwright/mcp@latest --headless --isolated
  ```

Default Oraculo skills are copied into `.claude/skills/` during install (best effort). The `test-driven-development` skill is in the default catalog; `qa-validation-planning` is available for QA phases. In implementation phases (`oraculo:exec`, `oraculo:checkpoint`), TDD is treated as required only when `[execution].tdd_default = true`.

> **Legacy path:** Oraculo also checks `.spw/spw-config.toml` as a fallback if `.spec-workflow/oraculo.toml` is not found.

### CLI commands

| Command | Description |
|---------|-------------|
| `oraculo install` | Install Oráculo in the current project |
| `oraculo update` | Self-update the CLI, clear cache, fetch latest kit |
| `oraculo doctor` | Show current repo/ref/cache configuration |
| `oraculo status` | Print a quick kit and skills summary |
| `oraculo skills` | Show installed/available/missing skills status |
| `oraculo skills install` | Install general skills |

#### Workflow tools (used by subagents and workflows)

| Command | Description |
|---------|-------------|
| `oraculo tools verify-task <spec> --task-id N [--check-commit]` | Verify task artifacts exist (impl log, optional commit check) |
| `oraculo tools impl-log register <spec> --task-id N --wave NN --title T --files F --changes C` | Create implementation log for a completed task |
| `oraculo tools impl-log check <spec> --task-ids 1,2,3` | Check if implementation logs exist for given task IDs |
| `oraculo tools task-mark <spec> --task-id N --status done` | Update task checkbox status in tasks.md (`in-progress`, `done`, `blocked`) |
| `oraculo tools wave-status <spec>` | Comprehensive wave state resolution (current wave, resume action, progress) |
| `oraculo tools wave-update <spec> --wave NN --status pass --tasks 3,4,7` | Write wave summary and latest JSON files |
| `oraculo tools dispatch-init-audit --run-dir R --type T` | Create nested audit directory inside a run (`_inline-audit/` or `_inline-checkpoint/`) |
| `oraculo tools audit-iteration start --run-dir R --type T [--max N]` | Initialize inline audit iteration tracking |
| `oraculo tools audit-iteration check --run-dir R --type T` | Check if another audit retry is allowed |
| `oraculo tools audit-iteration advance --run-dir R --type T --result R` | Increment audit iteration counter and record result |

### Agent Teams (optional)

Agent Teams is disabled by default. To enable it, set `[agent_teams].enabled = true` in `oraculo.toml`. The installer (`oraculo install`) reads this setting and switches symlinks in `.claude/workflows/oraculo/overlays/active/` from `../noop.md` to `../teams/<cmd>.md` accordingly.

Additional setup (done automatically by the installer):
- `env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1"` in `.claude/settings.json`
- `teammateMode = "in-process"` (change to `"tmux"` manually if desired)
- Overlay symlinks: `cd .claude/workflows/oraculo/overlays/active && ln -sf ../teams/<cmd>.md <cmd>.md`

When enabled, Oraculo creates a team for any phase not listed in `[agent_teams].exclude_phases` (all phases are eligible by default). `oraculo:exec` enforces delegate mode when `[agent_teams].require_delegate_mode = true`. Team overlays are available for all subagent-first entrypoints: `oraculo:discover`, `oraculo:plan`, `oraculo:design-research`, `oraculo:design-draft`, `oraculo:tasks-plan`, `oraculo:tasks-check`, `oraculo:exec`, `oraculo:checkpoint`, `oraculo:post-mortem`, `oraculo:qa`, `oraculo:qa-check`, `oraculo:qa-exec`, `oraculo:status`.

## Command entry points

- `oraculo:discover` -> zero-to-PRD requirements flow
- `oraculo:plan` -> design/tasks planning from existing requirements (with MCP approval gate)
- `oraculo:tasks-plan` -> config-driven task generation (`rolling-wave` or `all-at-once`)
- `oraculo:exec` -> batch execution with checkpoints
- `oraculo:checkpoint` -> quality gate report (PASS/BLOCKED)
- `oraculo:status` -> summarize where workflow stopped + next commands
- `oraculo:post-mortem` -> analyze post-spec commits and write reusable lessons
- `oraculo:qa` -> asks validation target and builds a QA test plan with concrete selectors (Playwright MCP/Bruno CLI/hybrid strategy)
- `oraculo:qa-check` -> validates test plan selectors, traceability, and data feasibility against actual code
- `oraculo:qa-exec` -> executes validated test plan using verified selectors (never reads source files)

## Thin-Orchestrator Architecture

Oráculo uses thin orchestrators with a dispatch pattern system:
- command wrappers live in `.claude/commands/oraculo/*.md`
- detailed orchestration workflows live in `.claude/workflows/oraculo/*.md`
- shared dispatch policies live in `.claude/workflows/oraculo/shared/dispatch-{pipeline,audit,wave}.md`
- shared cross-cutting policies live in `.claude/workflows/oraculo/shared/*.md`

### Dispatch Categories

Every workflow declares a `<dispatch_pattern>` section that serves as the **single source of truth** for dispatch metadata (`category`, `phase`, `comms_path`, `artifacts`). The CLI parses this section from embedded workflow files at startup — adding a new command only requires creating the workflow `.md` file with a valid `<dispatch_pattern>`.

| Category | Policy | Commands |
|----------|--------|----------|
| **Pipeline** | `dispatch-pipeline.md` | `discover`, `design-research`, `design-draft`, `tasks-plan`, `qa`, `post-mortem` |
| **Audit** | `dispatch-audit.md` | `tasks-check`, `qa-check`, `checkpoint` |
| **Wave Execution** | `dispatch-wave.md` | `exec`, `qa-exec` |

Checkpoint guardrails (audit commands):
- Orchestrators are read-only observers — they MUST NOT create/modify/delete artifacts outside comms to resolve a BLOCKED auditor (anti-self-heal).
- If ANY auditor returns `blocked`, the final verdict MUST be BLOCKED (handoff consistency).
- Briefs must never assert codebase facts — instruct auditors to verify instead.
- `oraculo:exec` must stop and instruct the user to run `oraculo:checkpoint` in a separate session (session isolation).

All categories enforce the 5 core thin-dispatch rules:
1. Orchestrator reads only `status.json` after dispatch (never `report.md` on pass).
2. Briefs contain filesystem paths to prior reports (never content).
3. Synthesizers/aggregators read from disk directly.
4. Run structure follows category layout.
5. Resume skips completed subagents, always reruns final stage.

Command-specific logic is injected via `<extensions>` at named points (`pre_pipeline`, `pre_dispatch`, `post_dispatch`, `post_pipeline`, `inter_wave`, `per_task`).

### Agent Teams

Agent Teams uses base + overlay via symlinks:
- base workflow: `.claude/workflows/oraculo/<command>.md`
- active overlay: `.claude/workflows/oraculo/overlays/active/<command>.md` (symlink)
- teams off: symlink -> `../noop.md` (empty placeholder)
- teams on: symlink -> `../teams/<command>.md`

Wrappers stay intentionally thin and delegate 100% of detailed logic to workflows.

Execution context guardrail (`oraculo:exec`):
- Before broad reads, dispatch `execution-state-scout` (implementation model, default `sonnet`).
- Scout returns only compact resume state: checkpoint status, task `[-]` in progress, next executable tasks, and required action (`resume|wait-user-authorization|manual-handoff|done|blocked`).
- Orchestrator then reads only task-scoped files for the selected IDs (avoid full `requirements.md`/`design.md` unless needed for blockers).

Planning defaults are configured in `.spec-workflow/oraculo.toml` (legacy fallback: `.spw/spw-config.toml`):

```toml
[planning]
tasks_generation_strategy = "rolling-wave" # or "all-at-once"
max_wave_size = 3
```

- `rolling-wave`: each planning cycle creates one executable wave.
  - Typical loop: `tasks-plan` -> `exec` -> `checkpoint` -> `tasks-plan` (next wave)...
- `all-at-once`: one planning pass creates all executable waves.
- Explicit CLI args still override config (`--mode`, `--max-wave-size`).

Post-mortem memory defaults are configured in `.spec-workflow/oraculo.toml` (legacy fallback: `.spw/spw-config.toml`):

```toml
[post_mortem_memory]
enabled = true
max_entries_for_design = 5
```

- `oraculo:post-mortem` writes reports to `.spec-workflow/post-mortems/<spec-name>/`.
- Shared index: `.spec-workflow/post-mortems/INDEX.md` (used by design/planning commands when enabled).
- Design/planning phases (`oraculo:discover`, `oraculo:design-research`, `oraculo:design-draft`, `oraculo:tasks-plan`, `oraculo:tasks-check`) load indexed lessons with recency/tag prioritization.

Unfinished-run handling for long subagent commands (`oraculo:discover`, `oraculo:design-research`, `oraculo:tasks-plan`, `oraculo:tasks-check`, `oraculo:checkpoint`, `oraculo:post-mortem`, `oraculo:qa`, `oraculo:qa-check`, `oraculo:qa-exec`):
- Before creating a new run-id, inspect the phase run folder (for `checkpoint`, inspect current wave folder first).
- If latest unfinished run exists, ask explicit user decision:
  - `continue-unfinished`
  - `delete-and-restart`
- Never choose automatically.
- If explicit decision is unavailable, stop with `WAITING_FOR_USER_DECISION`.
- On `continue-unfinished`, reuse completed `status=pass` outputs, redispatch missing/blocked subagents, and rerun the phase final decision/synthesis subagent before final artifact output.

Approval reconciliation for MCP-gated commands (`oraculo:discover`, `oraculo:status`, `oraculo:plan`, `oraculo:design-draft`, `oraculo:tasks-plan`):
- First read approval state from `spec-status` document fields.
- If status is missing/unknown/inconsistent, resolve approval ID (from `spec-status` or approval records under `.spec-workflow/approvals/<spec-name>/`) and confirm via MCP `approvals status`.
- `STATUS-SUMMARY.md` is output-only and must not be used as approval source of truth.

File-first subagent communication is stored under phase-based `_comms/` directories:
- discover: `.spec-workflow/specs/<spec-name>/discover/_comms/run-NNN/`
- design: `.spec-workflow/specs/<spec-name>/design/_comms/{design-research,design-draft}/run-NNN/`
- planning: `.spec-workflow/specs/<spec-name>/planning/_comms/{tasks-plan,tasks-check}/run-NNN/`
- execution: `.spec-workflow/specs/<spec-name>/execution/waves/wave-NN/{execution,checkpoint}/run-NNN/`
- qa: `.spec-workflow/specs/<spec-name>/qa/_comms/{qa,qa-check}/run-NNN/`
- qa-exec: `.spec-workflow/specs/<spec-name>/qa/_comms/qa-exec/waves/wave-NN/run-NNN/`
- post-mortem: `.spec-workflow/specs/<spec-name>/post-mortem/_comms/run-NNN/`

`<run-id>` format: `run-NNN` (zero-padded sequential, e.g. `run-001`).

YAML frontmatter (optional metadata) is included in spec templates under the
`oraculo` key to help subagents classify documents. It does not replace MCP
approvals or status.
- `schema`, `spec`, `doc`, `status`, `source`, `updated_at`
- `inputs`, `requirements`, `decisions`, `task_ids`, `test_required`
- `risk`, `open_questions`

## Dashboard Markdown Compatibility (`spec-workflow-mcp`)

To keep `tasks.md` fully compatible with Dashboard rendering + parsing + approval validation:

- Use checkbox markers only on real task lines:
  - `- [ ] <id>. <description>`
  - `- [-] <id>. <description>`
  - `- [x] <id>. <description>`
- Use `-` as task list marker for task rows (never `*`).
- Never use nested checkboxes in metadata blocks (for example DoD).
- Always start task lines with numeric IDs (`1`, `1.1`, `2.3`, ...), and keep IDs unique in the whole file.
- Keep metadata rows as regular bullets (`- ...`), never checkbox bullets.
- Keep `Files` parseable in a single line:
  - `- Files: path/to/file.ext, test/path/to/file_test.ext`
- Prefer underscore-delimited metadata fields when applicable:
  - `_Requirements: ..._`
  - `_Leverage: ..._`
  - `_Prompt: ..._` (closing underscore required)
- Keep `_Prompt` structured as:
  - `Role: ... | Task: ... | Restrictions: ... | Success: ...`

Oraculo task templates and `oraculo:tasks-plan` are aligned with this compatibility profile.

## Mermaid for Architecture Design

Oraculo now includes the `mermaid-architecture` skill for design phases, with common
diagram patterns and syntax guidance:
- skill file: `skills/mermaid-architecture/SKILL.md`
- default config: listed in `[skills.design].optional`

Common architecture examples covered by the skill:
- layered/module boundaries (`flowchart`)
- container/system view (`flowchart`)
- request flow with success/error path (`sequenceDiagram`)
- event-driven pipeline (`flowchart`)
- workflow lifecycle (`stateDiagram-v2`)

In `oraculo:design-draft`, `design.md` should include at least one valid Mermaid
diagram in the `## Architecture` section, using fenced lowercase `mermaid`
code blocks.

Skills use subagent-first loading by default to reduce main-context growth.

## QA Validation (3-Phase)

QA follows a plan → check → execute chain:

```
oraculo:qa (plan) → oraculo:qa-check (validate) → oraculo:qa-exec (execute)
```

### `oraculo:qa` (planning)
- Asks user what should be validated when focus is not explicitly provided
- Selects `Playwright MCP`, `Bruno CLI`, or `hybrid` by risk/scope
- Produces `QA-TEST-PLAN.md` with concrete selectors/endpoints per scenario
- Uses browser automation tools from pre-configured Playwright MCP server
- Stores file-first communications under `.spec-workflow/specs/<spec-name>/qa/_comms/qa/<run-id>/`

### `oraculo:qa-check` (validation)
- Validates test plan against actual code (the ONE phase that reads implementation files)
- Verifies selectors/endpoints exist via `qa-selector-verifier`
- Checks requirement traceability and data feasibility
- Produces `QA-CHECK.md` with verified selector map (test-id → selector → file:line)
- PASS/BLOCKED decision gates `oraculo:qa-exec`
- Stores file-first communications under `.spec-workflow/specs/<spec-name>/qa/_comms/qa-check/<run-id>/`

### `oraculo:qa-exec` (execution)
- Executes validated test plan using only verified selectors from `QA-CHECK.md`
- **Never reads implementation source files** — selector drift is logged as defect
- Supports `--scope smoke|regression|full` and `--rerun-failed true|false`
- Produces `QA-EXECUTION-REPORT.md` and `QA-DEFECT-REPORT.md` with GO/NO-GO decision
- Stores file-first communications under `.spec-workflow/specs/<spec-name>/qa/_comms/qa-exec/waves/wave-NN/<run-id>/`

Hook enforcement:
- `warn` -> diagnostics only
- `block` -> deny violating actions
- details: `AGENTS.md` + `.spec-workflow/oraculo.toml` comments (legacy fallback: `.spw/spw-config.toml`)

## Glossary

- **Agent Teams**: Optional mode where Oraculo spawns multiple Claude Code agents to work in parallel on a phase. Controlled by `[agent_teams].enabled` in `oraculo.toml`.

- **Checkpoint**: Quality gate run after each execution wave via `oraculo:checkpoint`. Produces a PASS/BLOCKED report that determines whether the next wave can proceed.

- **Dispatch Pattern**: The orchestration strategy a command uses. One of three categories: Pipeline (sequential stages leading to a synthesizer), Audit (parallel reviewers feeding an aggregator), or Wave Execution (iterative implementation cycles with checkpoints). Declared via `<dispatch_pattern>` in each workflow.

- **File-First Communication**: Subagents communicate exclusively via filesystem artifacts (`brief.md`, `report.md`, `status.json`) rather than chat messages. Artifacts are stored in `_comms/` directories under each phase.

- **Overlay**: Symlink-based mechanism that switches command behavior between solo mode (symlink to `noop.md`) and Agent Teams mode (symlink to `teams/<cmd>.md`). Located in `.claude/workflows/oraculo/overlays/active/`.

- **Rolling Wave**: Planning strategy where tasks are generated one wave at a time, allowing later waves to incorporate lessons from earlier execution. Set via `[planning].tasks_generation_strategy = "rolling-wave"` in `oraculo.toml`. Alternative: `all-at-once`.

- **Scout**: Lightweight subagent dispatched before a wave to gather execution state (checkpoint status, in-progress tasks, next actions) without reading full spec files. Returns compact resume state used by the orchestrator to scope its reads.

- **Synthesizer**: Final subagent in a Pipeline dispatch that reads all prior subagent reports from disk and produces the consolidated output artifact. Follows thin-dispatch rules (receives filesystem paths, not inline content).

- **Thin Dispatch**: Core architectural principle: orchestrators read only `status.json` after each subagent (never full reports), pass filesystem paths between stages, and delegate all detailed logic to workflows. Enforced by the 5 core thin-dispatch rules (see Dispatch Categories).

- **Wave**: A batch of tasks executed together in `oraculo:exec`. Each wave is followed by a checkpoint. Wave size is controlled by `[planning].max_wave_size` in `oraculo.toml`.