# Oraculo 

![Version](https://img.shields.io/badge/version-1.16.0-blue)
![License: MIT](https://img.shields.io/badge/license-MIT-green)

> *"Know thyself."* — Inscription at the Temple of Apollo at Delphi

## Index

- [What is the Oraculo?](#what-is-the-Oraculo)
- [Why Oraculo?](#why-Oraculo)
- [Quick Start](#quick-start)
- [Where to start](#where-to-start)
- [Installation](#installation)
- [Local Storage](#local-storage)
- [Input Commands](#input-commands)
- [Orchestrator Architecture] [Lean](#lean-orchestrator-architecture)
- [Dashboard Compatibility](#dashboard-compatibility-spec-workflow-mcp)
- [Mermaid for Architectural Design](#mermaid-for-architectural-design)
- [QA Validation (3 Phases)](#qa-validation-3-phases)
- [Glossary](#glossary)

## What is the Oraculo?

First of all, it is the Portuguese word for Oracle.

In the temples of Ancient Greece, the Oraculo of Delphi was consulted before any important undertaking—wars, city foundings, political decisions. No one acted without first seeking the wisdom of the Pythia, who channeled Apollo's knowledge into structured and interpretable prophecies.

**Oraculo** brings this same principle to software development with Claude Code. Just as the ancients did not go into battle without consulting Delphi, Oraculo prevents an AI agent from attacking an entire feature at once. Instead, the work is broken down into ritualized phases—each with its quality gates and approval points—like the stages of an oracular consultation:

| Phase | Analogy | What happens |
|------|----------|----------------|
| **Prophecy** (PRD) | The question to the Oraculo | Requirements are extracted and structured |
| **Interpretation** (Design) | The Pythia translates the vision | Research, architecture, and technical decisions |
| **Tablets** (Planning) | The stone tablets with the answer | Executable tasks are generated in waves |
| **Execution** (Exec) | The generals implement the prophecy | Implementation with automatic checkpoints |
| **Judgment** (QA) | The tribunal validates compliance | Tests planned, verified, and executed | Each phase dispatches **specialized agents** with model routing: Haiku performs light reconnaissance (like scouts), Opus conducts complex reasoning (like temple sages), and Sonnet executes implementation (like artisans). Agents communicate via artifacts in the file system—not chat—making each handoff reproducible and auditable.

You orchestrate everything with slash commands in Claude Code (e.g., `/oraculo:discover`, `/oraculo:exec`) while `spec-workflow-mcp` serves as the truth source for artifacts and approvals.

## Why Oraculo?

The Oraculo of Delphi wasn't simply a diviner. It was a **system**:

- **Ritualistic structure** — Questions had a format. Answers followed protocol. Nothing was improvised. Oraculo imposes the same discipline: each phase has its format, its artifacts, its gates.

- **Specialized Intermediaries** — The Pythia prophesied, the priests interpreted, the scribes recorded. The Oraculo replicates this with sub-agents: each has its role, its model, its scope.

- **Accumulated Wisdom** — The temple kept records of past consultations. The Oraculo's post-mortems index lessons learned that inform future decisions.

- **Never Act Without Consulting** — The greatest sin in Ancient Greece was *hubris*: acting arrogantly, without seeking guidance. The Oraculo ensures that no agent implements code without first passing through the planning gates.

## Quick Start

After [installation](#installation), run the following within a Claude Code session:

1. `/oraculo:discover my-feature` — Generates the requirements document from the description
2. `/oraculo:plan my-feature` — Creates the design and decomposes it into executable tasks
3. `/oraculo:exec my-feature` — Deploys in waves with automatic checkpoints
4. `/oraculo:qa my-feature` — Builds and executes the QA validation plan

Each command handles sub-agent dispatch, file handoff, and quality gates. Between steps, artifacts are located in `.spec-workflow/specs/my-feature/` and approvals flow through `spec-workflow-mcp`.

## Where to start

- This file is the primary source for usage and operation.

- Operational rules for agents and contributors are in `AGENTS.md`.

- `docs/ORACULO-WORKFLOW.md`, `hooks/README.md`, and `copy-ready/README.md` are lightweight pointers to this README.

## Installation

### 1. Install the CLI

The bootstrap script downloads the compiled Go binary from the latest release on GitHub and installs it in `~/.local/bin/oraculo`. Requires `curl` and `tar`.

``bash
curl -fsSL https://raw.githubusercontent.com/lucas-stellet/oraculo/main/scripts/bootstrap.sh | bash
```

**From a local clone (build from source):**

```bash
cd cli && go build -o ~/.local/bin/oraculo ./cmd/oraculo/
```

Execute `oraculo` without arguments to see the available commands. Use `oraculo update` to auto-update.

### 2. Install in the project

In the project root:

```bash
oraculo install
```

Copies commands, workflows, hooks, config, and skills to the project. For manual installation: `cp -R /path/to/oraculo/copy-ready/. .`

### 3. Post-installation checklist

Required:

1. Merge `.claude/settings.json.example` into your `.claude/settings.json` (if necessary).

2. Review `.spec-workflow/oraculo.toml`, especially `[planning].tasks_generation_strategy` and `[planning].max_wave_size`.

3. Start a new Claude Code session for the SessionStart hook to synchronize the task template.

Optional:
- Enable skill enforcement per phase: `skills.design.enforce_required` and `skills.implementation.enforce_required` in `oraculo.toml`.

- Enable Oraculo statusline (see `.claude/settings.json.example`).

- Enable enforcement hooks with `hooks.enforcement_mode = "warn"` or `"block"` in `oraculo.toml`.

- For browser-based QA validation and URL exploration in planning, add Playwright MCP:

```
claude mcp add playwright --npx @playwright/mcp@latest --headless --isolated
```

Default skills are copied to `.claude/skills/` during installation. The `test-driven-development` skill is in the default catalog; `qa-validation-planning` is available for QA phases. In implementation phases (`oraculo:exec`, `oraculo:checkpoint`), TDD is only required when `[execution].tdd_default = true`.

TDD template variants: `user-templates/variants/` contains `tasks-template.tdd-on.md` and `tasks-template.tdd-off.md`. The `[templates].tasks_template_mode` key controls the selection (`auto`, `on`, `off`). The SessionStart hook synchronizes the active variant at the start of each session.

> **Legacy Path:** Oraculo also checks `.spw/spw-config.toml` as a fallback if `.spec-workflow/oraculo.toml` is not found.

### Global Installation

For those working on multiple projects, Oraculo supports a two-tier installation that avoids duplicating ~72 files per project:

| Mode | Command | What it installs | Where |

|------|---------|---------------|------|
| **Global** | `oraculo install --global` | Commands, workflows, hooks, skills | `~/.claude/` |

| **Project Initialization** | `oraculo init` | Config, templates, snippets, .gitattributes | `.spec-workflow/`, `CLAUDE.md`, `AGENTS.md` |

| **Complete (default)** | `oraculo install` | Everything (unchanged behavior) | `.claude/` + `.spec-workflow/` |

**Setup:**

```bash
# Once: install globally
oraculo install --global

# Per project: initialize specific config
cd my-project
oraculo init
```

**How it works:** Claude Code resolves paths `@.claude/` with local priority, global fallback. If the project has a local install (`oraculo install`), it takes precedence over the global one.

**Limitations:**

- Global workflows use default config (without project guidelines). Projects that need custom guidelines should use `oraculo install`.

- Agent Teams overlays are globally noop. Projects using Agent Teams need `oraculo install` for local activation.

### CLI Commands

| Command | Description |

|---------|-----------|

| `oraculo install` | Installs Oraculo in the current project (complete local installation) |

| `oraculo install --global` | Installs commands, workflows, hooks, and skills in `~/.claude/` |

| `oraculo init` | Initializes project-specific config, templates, and snippets |

| `oraculo update` | Auto-updates the binary via GitHub Releases |

| `oraculo doctor` | Checks the health of the installation (version, config, hooks, commands, workflows, skills) |

| `oraculo status` | Quick summary of the kit and skills |

| `oraculo skills` | Status of installed/available/missing skills |

| `oraculo skills install` | Installs general skills |

<details>
<summary>Quick config reference (all sections)</summary>

| Section | Key(s) | Description |

|-------|----------|-----------|

| `[statusline]` | `cache_ttl_seconds`, `base_branches`, `sticky_spec`, `show_token_cost` | StatusLine hook behavior |
| `[templates]` | `sync_tasks_template_on_session_start`, `tasks_template_mode` | Task template variant selection |
| `[safety]` | `backup_before_overwrite` | Backup before overwriting spec files |
| `[verification]` | `inline_audit_max_iterations` | Max inline audit retry attempts |
| `[qa]` | `max_scenarios_per_wave` | QA wave sizing |
| `[hooks]` | `verbose`, `recent_run_window_minutes`, `guard_prompt_require_spec`, `guard_paths`, `guard_wave_layout`, `guard_stop_handoff` | Toggles by hook guard |
| `[execution]` | `require_clean_worktree_for_wave_pass`, `manual_tasks_require_human_handoff`, `tdd_default` | Execution Gates |
| `[planning]` | `tasks_generation_strategy`, `max_wave_size` | Wave Planning Strategy |
| `[post_mortem_memory]` | `enabled`, `max_entries_for_design` | Indexing post-mortem lessons |
| `[agent_teams]` | `enabled`, `exclude_phases`, `require_delegate_mode` | Agent Teams Toggle | See `.spec-workflow/oraculo.toml` for complete documentation of each key.

</details>

| `oraculo finalizar <spec>` | Marks spec as complete, generates a summary with frontmatter YAML |

| `oraculo view <spec> [type]` | Views artifacts in the terminal or VS Code |

| `oraculo search <query>` | Searches full-text (FTS5) in indexed specs |

| `oraculo summary <spec>` | Generates progress summary on demand |

#### Workflow tools (used by sub-agents)

| Command | Description |

|---------|-----------|

| `oraculo tools verify-task <spec> --task-id N [--check-commit]` | Checks for the existence of task artifacts |

| `oraculo tools impl-log register <spec> --task-id N --wave NN --title T --files F --changes C` | Creates deployment log for completed task |
| `oraculo tools impl-log check <spec> --task-ids 1,2,3` | Checks if deployment logs exist |
| `oraculo tools task-mark <spec> --task-id N --status done` | Updates task checkbox in tasks.md |
| `oraculo tools wave-status <spec>` | Full wave status resolution |
| `oraculo tools wave-update <spec> --wave NN --status pass --tasks 3,4,7` | Writes wave summary and status JSON |
| `oraculo tools dispatch-init-audit --run-dir R --type T` | Creates audit directory within a run |
| `oraculo tools audit-iteration start --run-dir R --type T [--max N]` | Initializes audit iteration tracking |

| `oraculo tools audit-iteration check --run-dir R --type T` | Checks if another retry is allowed |

| `oraculo tools audit-iteration advance --run-dir R --type T --result R` | Advances iteration counter |

### Agent Teams (optional)

Agent Teams is disabled by default. To enable it, set `[agent_teams].enabled = true` in `oraculo.toml`. The installer reads this setting and toggles symlinks in `.claude/workflows/oraculo/overlays/active/`.

Additional configuration (automatic by the installer):
- `env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1"` in `.claude/settings.json`
- `teammateMode = "in-process"` (change manually to `"tmux"` if desired)
- Overlay symlinks: `cd .claude/workflows/oraculo/overlays/active && ln -sf ../teams/<cmd>.md <cmd>.md`

When enabled, Oraculo creates a team for any phase not listed in `[agent_teams].exclude_phases`. `oraculo:exec` requires delegate mode when `[agent_teams].require_delegate_mode = true`.

## Local Storage

Oraculo stores structured data in SQLite databases (pure Go driver, no CGO, WAL mode):

- **`spec.db`** — Database per spec in `.spec-workflow/specs/<spec-name>/spec.db`. The dispatch system automatically retrieves artifacts from sub-agents (briefs, reports, status) in the DB during handoff (dual-write). Three files managed by the MCP remain on disk as a source of truth: `requirements.md`, `design.md`, `tasks.md`.

- **`.oraculo-index.db`** — Global index in `.spec-workflow/.oraculo-index.db` with FTS5 full-text search. Feeds `oraculo search <query>`.

`oraculo finalizar <spec>` marks a spec as complete and generates a frontmatter YAML summary block, making the spec searchable in the global index.

## Input Commands

The phases follow the Oraculo's journey:

- `oraculo:discover` → zero-to-PRD requirements flow *(the question to the Oraculo)*
- `oraculo:plan` → design/task meta-orchestrator: chains research → draft → plan → checkout *(the interpretation)*
- `oraculo:tasks-plan` → task generation (`rolling-wave` or `all-at-once`) *(the tablets with the answer)*
- `oraculo:exec` → batch execution with checkpoints *(the generals implement)*
- `oraculo:checkpoint` → quality gate (PASS/BLOCKED) *(the temple validates)*
- `oraculo:status` → summary of where the workflow left off + next commands *(consultation with the temple)*
- `oraculo:post-mortem` → Analyzes post-spec commits and logs lessons *(the chronicles)*
- `oraculo:qa` → Builds QA validation plan *(the tribunal)*
- `oraculo:qa-check` → Validates selectors and traceability *(verification of witnesses)*
- `oraculo:qa-exec` → Executes validated plan *(the verdict)*

## Lean Orchestrator Architecture

Oraculo uses lean orchestrators with a dispatch pattern system:
- command wrappers in `.claude/commands/oraculo/*.md`
- detailed workflows in `.claude/workflows/oraculo/*.md`
- shared dispatch policies in `.claude/workflows/oraculo/shared/dispatch-{pipeline,audit,wave}.md`
- cross-cutting policies in `.claude/workflows/oraculo/shared/*.md`

### Dispatch Categories

Each workflow declares a `<dispatch_pattern>` section as the **single source of truth** for dispatch metadata (`category`, `phase`, `comms_path`, `artifacts`). The CLI parses this section of the workflows embedded in the initialization.

| Category | Policy | Commands |

|-----------|----------|----------|

| **Pipeline** | `dispatch-pipeline.md` | `discover`, `design-research`, `design-draft`, `tasks-plan`, `qa`, `post-mortem` |

| **Audit** | `dispatch-audit.md` | `tasks-check`, `qa-check`, `checkpoint` |

| **Wave Execution** | `dispatch-wave.md` | `exec`, `qa-exec` |

Checkpoint guardrails (audit commands):
- Orchestrators are read-only observers — they NEVER create/modify/delete artifacts outside of commands to resolve a BLOCKED auditor (anti-self-heal).

- If ANY auditor returns `blocked`, the final verdict MUST be BLOCKED (handoff consistency).

- Briefs never state facts about the code — they instruct auditors to verify.

- `oraculo:exec` should stop and instruct the user to run `oraculo:checkpoint` in a separate session (session isolation).

The 5 core rules of thin-dispatch:
1. Orchestrator only reads `status.json` after dispatch (never `report.md` in case of pass).

2. Briefs contain filesystem paths to previous reports (never content).

3. Synthesizers/aggregators read directly from disk.

4. Run structure follows category layout.

5. Resume skips complete sub-agents, always re-executes the final stage.

Specific logic per command is injected via `<extensions>` at named points (`pre_pipeline`, `pre_dispatch`, `post_dispatch`, `post_pipeline`, `inter_wave`, `per_task`).

Agent Teams

Agent Teams uses base + overlay via symlinks:
- Workflow base: `.claude/workflows/oraculo/<command>.md`
- Active overlay: `.claude/workflows/oraculo/overlays/active/<command>.md` (symlink)
- Teams off: symlink → `../noop.md`
- Teams on: symlink → `../teams/<command>.md`

Wrappers remain intentionally lean and delegate 100% of the logic to workflows.

Execution context guardrail (`oraculo:exec`):
- Before wide reads, dispatch `execution-state-scout` (deployment model, default `sonnet`).

- Scout returns only compact state: checkpoint status, task `[-]` in progress, next executable tasks, and required action (`resume|wait-user-authorization|manual-handoff|done|blocked`).

- Orchestrator then reads only task-scoped files (avoids full `requirements.md`/`design.md` files, except for blockers).

Planning defaults in `.spec-workflow/oraculo.toml`:

```toml
[planning]
tasks_generation_strategy = "rolling-wave" # or "all-at-once"
max_wave_size = 3

```

- `rolling-wave`: each planning cycle creates an executable wave.

- Typical loop: `boards` → `exec` → `checkpoint` → `boards` (next wave)...
- `all-at-once`: a planning step creates all waves.

- Explicit CLI arguments override config (`--mode`, `--max-wave-size`).

Post-mortem memory in `.spec-workflow/oraculo.toml`:

```toml
[post_mortem_memory]
enabled = true
max_entries_for_design = 5

```

- `oraculo:pos-mortem` writes reports to `.spec-workflow/post-mortems/<spec-name>/`.

- Shared index: `.spec-workflow/post-mortems/INDEX.md` (used by design/planning when enabled).

- Design/planning phases carry indexed lessons prioritized by recency/tags.

Handling unfinished runs for long commands:
- Before creating a new run-id, inspect the phase's runs folder.

- If an unfinished run exists, ask for an explicit user decision:

- `continue-unfinished`

- `delete-and-restart`

- Never chooses automatically.

- If a decision is unavailable, stop with `WAITING_FOR_USER_DECISION`.

Reconciliation of approvals for commands with MCP gates:

- First read the approval status from the `spec-status` fields.

- If the status is missing/unknown/inconsistent, resolve the approval ID and confirm via MCP `approvals status`.

- `STATUS-SUMMARY.md` is output-only, never a source of truth.

File-first communication between subagents is stored in `_comms/` directories organized by phase:
- discover: `.spec-workflow/specs/<spec-name>/discover/_comms/run-NNN/`
- design: `.spec-workflow/specs/<spec-name>/design/_comms/{design-research,design-draft}/run-NNN/`
- planning: `.spec-workflow/specs/<spec-name>/planning/_comms/{tasks-plan,tasks-check}/run-NNN/`
- execution: `.spec-workflow/specs/<spec-name>/execution/waves/wave-NN/{execution,checkpoint}/run-NNN/`
- qa: `.spec-workflow/specs/<spec-name>/qa/_comms/{qa,qa-check}/run-NNN/`
- qa-exec: `.spec-workflow/specs/<spec-name>/qa/_comms/qa-exec/waves/wave-NN/run-NNN/`

- post-mortem: `.spec-workflow/specs/<spec-name>/post-mortem/_comms/run-NNN/`

Format of `<run-id>`: `run-NNN` (sequential with zero-padding, e.g., `run-001`).

YAML frontmatter (optional metadata) is included in the spec templates under the key `Oraculo` for classifying documents by sub-agents.

## Dashboard Compatibility (`spec-workflow-mcp`)

To keep `tasks.md` compatible with Dashboard rendering + parsing + approval validation:

- Checkbox markers only on actual task lines:

- `- [ ] <id>. <description>`
- `- [-] <id>. <description>`

- `- [x] <id>. <description>`
- Use `-` as a marker (never `*`).

- Never use nested checkboxes in metadata blocks.

- Numeric IDs at the beginning (`1`, `1.1`, `2.3`, ...), unique in the entire file.

- Metadata as regular bullets (`- ...`), never checkboxes.

- `Files` parsable in a single line:

- `- Files: path/to/file.ext, test/path/to/file_test.ext`
- Metadata fields with underscores:

- `_Requirements: ..._`

- `_Leverage: ..._`

- `_Prompt: ..._`
- `_Prompt` structured as:

- `Role: ... | Task: ... | Restrictions: ... | Success: ...

## Mermaid for Architecture Design

Oraculo includes the `mermaid-architecture` skill for design phases:

- skill file: `skills/mermaid-architecture/SKILL.md`
- default config: listed in `[skills.design].optional`

Examples of covered architecture:

- module/layer boundaries (`flowchart`)
- container/system view (`flowchart`)
- request flow with success/error paths (`sequenceDiagram`)
- event-driven pipeline (`flowchart`)
- workflow lifecycle (`stateDiagram-v2`)

In `oraculo:design-draft`, `design.md` must include at least one valid Mermaid diagram in the `## Architecture` section.

## QA Validation (3 Phases)

QA follows a plan → check → execute chain, like a Greek court:

```
Oraculo:judgment (plan) → Oraculo:judgment-check (validate) → Oraculo:judgment-exec (execute)

```

### `Oraculo:judgment` (planning)
- Asks the user what to validate when the focus is not explicit
- Selects `Playwright MCP`, `Bruno CLI`, or `hybrid` by risk/scope
- Produces `QA-TEST-PLAN.md` with concrete selectors/endpoints per scenario
- Uses Playwright MCP's pre-configured browser automation tools

### `Oraculo:judgment-check` (validation)
- Validates the test plan against real code (the ONLY phase that reads (Implementation files)
- Checks for the existence of selectors/endpoints via `qa-selector-verifier`
- Checks data traceability and feasibility
- Produces `QA-CHECK.md` with a verified map (test-id → selector → file:line)
- PASS/BLOCKED decision triggers `oraculo:julgamento-exec`

### `oraculo:julgamento-exec` (execution)
- Executes validated plan using only verified selectors from `QA-CHECK.md`
- **Never reads implementation source files** — selector drift is logged as a defect
- Supports `--scope smoke|regression|full` and `--rerun-failed true|false`
- Produces `QA-EXECUTION-REPORT.md` and `QA-DEFECT-REPORT.md` with GO/NO-GO decision

Enforcement of Hooks:
- `warn` → diagnostics only
- `block` → negates violating actions
- Details: `AGENTS.md` + `.spec-workflow/oraculo.toml`

## Glossary

- **Agent Teams**: Optional mode where Oraculo instantiates multiple Claude Code agents to work in parallel in a phase. Controlled by `[agent_teams].enabled` in `oraculo.toml`.

- **Checkpoint**: Quality gate executed after each wave via `oraculo:checkpoint`. Produces a PASS/BLOCKED report that determines if the next wave can proceed.

- **Dispatch Pattern**: Orchestration strategy for a command. One of three categories: Pipeline (sequential stages with synthesizer), Audit (parallel reviewers with aggregator), or Wave Execution (iterative cycles with checkpoints). Declared via `<dispatch_pattern>` in each workflow.

- **File-First Communication**: Sub-agents communicate exclusively via artifacts in the filesystem (`brief.md`, `report.md`, `status.json`) — never via chat. Stored in `_comms/` directories.

- **Overlay**: Symlink-based mechanism that alternates behavior between solo mode (symlink to `noop.md`) and Agent Teams mode (symlink to `teams/<cmd>.md`).

- **Rolling Wave**: Strategy where tasks are generated one wave at a time, allowing future waves to incorporate lessons from the previous execution. Config: `[planning].tasks_generation_strategy = "rolling-wave"`.

- **Scout**: Lightweight sub-agent dispatched before a wave to collect execution state without reading complete specs. Returns compact resume state to the orchestrator.

- **Synthesizer**: Final sub-agent in a pipeline that reads all previous reports from disk and produces the consolidated artifact.

- **Thin Dispatch**: Core architectural principle: orchestrators read only `status.json` after each sub-agent, pass paths between stages, and delegate detailed logic to workflows.

- **Wave**: Batch of tasks executed together in `oraculo:exec`. Each wave is followed by a checkpoint. Size controlled by `[planning].max_wave_size`.

- **spec.db**: SQLite database per spec that stores artifacts collected from sub-agents. Created automatically via dual-write during dispatch handoff. - **Harvest**: A pattern where dispatch-handoff collects files from sub-agents (`brief.md`, `report.md`, `status.json`) to `spec.db` after each sub-agent completes.

---

<p align="center">
<em>"Μηδὲν ἄγαν"</em> — Nothing in excess.<br>

<small>Second inscription on the Temple of Apollo at Delphi.</small>
</p>