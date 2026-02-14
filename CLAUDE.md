# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Oraculo

Oraculo is a command/template kit for `spec-workflow-mcp` that provides stricter agent execution patterns (planning gates, waves, checkpoints) with subagent-first orchestration and model routing (haiku for web scouting, opus for complex reasoning, sonnet for implementation).

## Canonical Sources (read order)

1. `README.md` — installation, usage, workflow reference
2. `AGENTS.md` — operational rules for agents and contributors (Portuguese)
3. `config/oraculo.toml` — runtime defaults

## Validation & Testing

There are no unit tests or a test framework. Validation is done via a checklist of syntax checks and smoke runs:

```bash
# Validate all shell scripts parse correctly
bash -n bin/oraculo
bash -n scripts/bootstrap.sh
bash -n scripts/install-oraculo-bin.sh
bash -n scripts/validate-thin-orchestrator.sh
bash -n copy-ready/install.sh

# Validate thin-orchestrator contract (wrapper sizes, workflow refs, mirror sync)
scripts/validate-thin-orchestrator.sh

# Build and smoke-test Go CLI
go build -o /tmp/oraculo ./cli/cmd/oraculo && PATH="/tmp:$PATH"

# Hooks (each reads JSON from stdin)
echo '{"workspace":{"current_dir":"'"$(pwd)"'"}}' | oraculo hook statusline
echo '{"prompt":"/oraculo:plan"}' | oraculo hook guard-prompt
echo '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":"README.md"}}' | oraculo hook guard-paths
echo '{}' | oraculo hook guard-stop
echo '{}' | oraculo hook session-start

# User-facing commands
oraculo finalizar --help
oraculo view --help
oraculo search --help
oraculo summary --help
oraculo init --help

# Inspection commands
oraculo tasks state --help
oraculo wave state --help
oraculo spec list --help
```

## Architecture

### Thin-Orchestrator Pattern

Commands and workflows are separated into two layers:

- **`commands/oraculo/*.md`** — Thin wrappers (max 60 lines) that define frontmatter metadata and point to a workflow via `<execution_context>` referencing `@.claude/workflows/oraculo/<command>.md`. These are what Claude Code slash commands (`/oraculo:exec`, `/oraculo:prd`, etc.) invoke.
- **`workflows/oraculo/*.md`** — Full orchestration logic: subagent definitions, policies, gates, state machines. Shared policy fragments live in `workflows/oraculo/shared/` (config resolution, file handoff, resume policy, skills policy, approval reconciliation).

Agent Teams uses base + overlay via symlinks: each command references `workflows/oraculo/overlays/active/<command>.md`, which is a symlink pointing to `../noop.md` (teams off) or `../teams/<command>.md` (teams on). The installer switches symlinks; no separate command directory needed.

### Mirror System

Source files in this repo must stay in sync with their `copy-ready/` counterparts:

| Source | Mirror |
|--------|--------|
| `commands/oraculo/` | `copy-ready/.claude/commands/oraculo/` |
| `workflows/oraculo/` | `copy-ready/.claude/workflows/oraculo/` |
| `workflows/oraculo/overlays/noop.md` | `copy-ready/.claude/workflows/oraculo/overlays/noop.md` |
| `workflows/oraculo/overlays/active/*.md` | `copy-ready/.claude/workflows/oraculo/overlays/active/*.md` (symlinks) |
| `templates/user-templates/` | `copy-ready/.spec-workflow/user-templates/` |
| `config/oraculo.toml` | `copy-ready/.spec-workflow/oraculo.toml` |
| `templates/claude-md-snippet.md` | `copy-ready/.claude.md.snippet` |
| `templates/agents-md-snippet.md` | `copy-ready/.agents.md.snippet` |
| `workflows/oraculo/shared/dispatch-implementation.md` | `copy-ready/.claude/workflows/oraculo/shared/dispatch-implementation.md` |

`scripts/validate-thin-orchestrator.sh` enforces mirror integrity via `diff -rq`. Always update both sides in the same patch.

### Go CLI (`cli/`)

The Go CLI (`cli/cmd/oraculo/`) provides hooks, inspection commands, user-facing commands, and workflow tools. Full CLI reference: `.claude/docs/oraculo-cli-reference.md`.

#### Hooks

All hooks are implemented in Go at `cli/internal/hook/` and invoked via `oraculo hook <event>`. Each reads JSON from stdin and follows the same exit-code contract: 0 = ok, 2 = block.

- **`oraculo hook statusline`** — StatusLine: displays context in the editor status bar. Uses 3 detection strategies (git diff, cache, sticky spec) to find the active spec. Shows token/cost estimates when `show_token_cost = true`. Cache TTL controlled by `statusline.cache_ttl_seconds`.
- **`oraculo hook guard-prompt`** — UserPromptSubmit: validates spec arg presence in Oraculo commands. Controlled by `hooks.guard_prompt_require_spec`.
- **`oraculo hook guard-paths`** — PreToolUse (Write/Edit): prevents writes outside spec-workflow paths. Also enforces wave-NN directory format and blocks legacy `_agent-comms/` paths when `hooks.guard_wave_layout = true`. Controlled by `hooks.guard_paths`.
- **`oraculo hook guard-stop`** — Stop: checks file-first handoff completeness in recent runs. Scans runs within `hooks.recent_run_window_minutes`. Controlled by `hooks.guard_stop_handoff`.
- **`oraculo hook session-start`** — SessionStart: syncs active tasks template variant based on TDD config. Also auto-re-renders workflows when config changes are detected.

Hook enforcement mode is configured in `config/oraculo.toml` under `[hooks]`: `warn` (diagnostics only) or `block` (deny violating actions).

#### Local Storage

Oraculo stores structured data in SQLite databases (pure Go driver, no CGO, WAL mode):

- **`spec.db`** — Per-spec database at `.spec-workflow/specs/<spec-name>/spec.db`. The dispatch-handoff dual-writes subagent artifacts (briefs, reports, status) into the DB when the store is available. Three MCP-managed files remain on disk as source of truth: `requirements.md`, `design.md`, `tasks.md`.
- **`.oraculo-index.db`** — Global index at `.spec-workflow/.oraculo-index.db` with FTS5 full-text search across all specs. Updated by `oraculo finalizar` and queried by `oraculo search`.

#### User-Facing Commands

| Command | Description |
|---------|-------------|
| `oraculo finalizar <spec>` | Mark spec as completed, harvest artifacts, generate summary with YAML frontmatter, index in global FTS5 |
| `oraculo view <spec> [type]` | View spec artifacts (`overview`, `report`, `brief`, `checkpoint`, `implementation-log`, `wave-summary`, `completion-summary`) |
| `oraculo search <query>` | FTS5 full-text search across indexed specs |
| `oraculo summary <spec>` | Generate on-demand progress summary |

#### Inspection Commands

- **`oraculo tasks`** — Task state resolution: `state`, `next`, `mark`, `count`, `files`, `validate`, `complexity`
- **`oraculo wave`** — Wave inspection: `state`, `summary`, `checkpoint`, `resume`
- **`oraculo spec`** — Spec lifecycle: `artifacts`, `stage`, `prereqs`, `approval`, `list`

### CLI Wrapper (`bin/oraculo`)

The `oraculo` CLI is a bash wrapper that caches the kit from GitHub and delegates to `copy-ready/install.sh`. Key commands: `oraculo install`, `oraculo install --global`, `oraculo init`, `oraculo update`, `oraculo doctor`, `oraculo status`, `oraculo skills`. Environment variables: `ORACULO_REPO`, `ORACULO_REF`, `ORACULO_HOME`, `ORACULO_KIT_DIR`, `ORACULO_AUTO_UPDATE`, `INSTALL_DIR`.

#### Two-Tier Installation

| Mode | Command | What it installs | Where |
|------|---------|------------------|-------|
| **Global** | `oraculo install --global` | Commands, workflows, hooks, skills | `~/.claude/` |
| **Project Init** | `oraculo init` | Config, templates, snippets, .gitattributes | `.spec-workflow/`, `CLAUDE.md`, `AGENTS.md` |
| **Full (default)** | `oraculo install` | Everything (unchanged behavior) | `.claude/` + `.spec-workflow/` |

Coexistence: if a project has a local install, it takes precedence over the global (Claude Code native path resolution).

<!-- ORACULO-KIT-START — managed by oraculo install, do not edit manually -->
## Oraculo (Spec-Workflow)

This project uses Oraculo for structured AI-driven development workflows.

### Commands

`/oraculo:prd` → `/oraculo:plan` → `/oraculo:design-research` → `/oraculo:design-draft` → `/oraculo:tasks-plan` → `/oraculo:tasks-check` → `/oraculo:exec` → `/oraculo:checkpoint` → `/oraculo:qa` → `/oraculo:qa-check` → `/oraculo:qa-exec` → `/oraculo:post-mortem` → `/oraculo:status`

### Dispatch CLI (used within workflows)

All Oraculo workflows use these CLI commands for subagent dispatch. The CLI creates directories, boilerplate files, and enforces the file-first handoff contract:

- `oraculo tools dispatch-init <command> <spec-name> [--wave NN]` — creates run-NNN dir, returns category/dispatch_policy/models
- `oraculo tools dispatch-setup <subagent> --run-dir <dir> --model-alias <alias>` — creates subagent dir + brief.md skeleton
- `oraculo tools dispatch-read-status <subagent> --run-dir <dir>` — reads status.json (ONLY read report.md if status=blocked)
- `oraculo tools dispatch-handoff --run-dir <dir>` — generates _handoff.md from all status.json files
- `oraculo tools resolve-model <alias>` — maps config alias (web_research/complex_reasoning/implementation) to model

### File-First Handoff Contract

Every subagent MUST produce: `brief.md` (written by orchestrator), `report.md`, `status.json`.
Every run MUST produce: `_handoff.md`.
Status.json format: `{"status": "pass"|"blocked", "summary": "one-line description"}`

### Config

Runtime config: `.spec-workflow/oraculo.toml`
<!-- ORACULO-KIT-END -->

<!-- ORACULO-KIT-START — managed by oraculo install, do not edit manually -->
## Oraculo (Spec-Workflow)

This project uses Oraculo for structured AI-driven development workflows.

### Commands

`/oraculo:prd` → `/oraculo:plan` → `/oraculo:design-research` → `/oraculo:design-draft` → `/oraculo:tasks-plan` → `/oraculo:tasks-check` → `/oraculo:exec` → `/oraculo:checkpoint` → `/oraculo:qa` → `/oraculo:qa-check` → `/oraculo:qa-exec` → `/oraculo:post-mortem` → `/oraculo:status`

### Dispatch CLI (used within workflows)

All Oraculo workflows use these CLI commands for subagent dispatch. The CLI creates directories, boilerplate files, and enforces the file-first handoff contract:

- `oraculo tools dispatch-init <command> <spec-name> [--wave NN]` — creates run-NNN dir, returns category/dispatch_policy/models
- `oraculo tools dispatch-setup <subagent> --run-dir <dir> --model-alias <alias>` — creates subagent dir + brief.md skeleton
- `oraculo tools dispatch-read-status <subagent> --run-dir <dir>` — reads status.json (ONLY read report.md if status=blocked)
- `oraculo tools dispatch-handoff --run-dir <dir>` — generates _handoff.md from all status.json files
- `oraculo tools resolve-model <alias>` — maps config alias (web_research/complex_reasoning/implementation) to model

### File-First Handoff Contract

Every subagent MUST produce: `brief.md` (written by orchestrator), `report.md`, `status.json`.
Every run MUST produce: `_handoff.md`.
Status.json format: `{"status": "pass"|"blocked", "summary": "one-line description"}`

### Config

Runtime config: `.spec-workflow/oraculo.toml`
<!-- ORACULO-KIT-END -->

