# Oraculo CLI Reference

Complete reference for all `oraculo` CLI commands. The Go binary is built from `cli/cmd/oraculo/`.

## Root Commands

| Command | Description |
|---------|-------------|
| `oraculo install` | Install Oraculo kit into the current project |
| `oraculo update` | Self-update the CLI binary from GitHub Releases |
| `oraculo doctor` | Check Oráculo installation health (version, config, hooks, commands, workflows, skills) |
| `oraculo status` | Show Oraculo kit presence and spec summary |
| `oraculo skills` | Show skills installation status |
| `oraculo skills install [--elixir]` | Install general skills (or Elixir-specific with flag) |
| `oraculo version` | Print version information |
| `oraculo render [command] [--all]` | Render composed workflow with all shared policies inlined |
| `oraculo hook <event>` | Handle Claude Code hook events (see [Hooks](#hooks)) |

## User-Facing Commands

| Command | Description |
|---------|-------------|
| `oraculo finalizar <spec>` | Mark spec as completed, harvest artifacts into DB, generate summary with YAML frontmatter, index in global FTS5. Flags: `--export`, `--force` (skip post-mortem check), `--raw` |
| `oraculo view <spec> [type]` | View spec artifacts. Types: `overview` (default), `report`, `brief`, `checkpoint`, `implementation-log`, `wave-summary`, `completion-summary`. Flags: `--wave`, `--run`, `--task`, `--vscode`, `--raw` |
| `oraculo search <query>` | FTS5 full-text search across indexed specs. Requires prior `oraculo finalizar`. Flags: `--spec` (filter), `--limit` (default 5), `--raw` |
| `oraculo summary <spec>` | Generate on-demand progress summary (task status, wave progress, files changed). Flags: `--export`, `--vscode`, `--raw` |

## Task Inspection (`oraculo tasks`)

Parse `tasks.md` body-first and resolve task state deterministically. All `tasks`, `wave`, and `spec` subcommands support `--raw` for machine-readable output.

| Subcommand | Args | Description |
|------------|------|-------------|
| `state <spec>` | spec name | Show full task state for a spec |
| `next <spec>` | spec name | Resolve next executable wave (executable tasks, deferred with resolved deps) |
| `mark <spec> <task-id> <status>` | spec, task ID, status (`done`, `in_progress`, `pending`) | Update task checkbox. Flag: `--require-impl-log` |
| `count <spec>` | spec name | Count tasks by status |
| `files <spec> <task-id>` | spec, task ID | List files for a specific task |
| `validate <spec>` | spec name | Validate tasks.md against dashboard rules |
| `complexity <spec> [task-id]` | spec, optional task ID | Score task complexity for model routing (haiku/sonnet/opus) |

## Wave Inspection (`oraculo wave`)

Inspect wave execution state, checkpoint status, and summaries.

| Subcommand | Args | Description |
|------------|------|-------------|
| `state <spec>` | spec name | Show state of all waves |
| `summary <spec> <wave-num>` | spec, wave number | Generate summary for a specific wave |
| `checkpoint <spec> <wave-num>` | spec, wave number | Resolve checkpoint status (`_latest.json`-first resolution) |
| `resume <spec>` | spec name | Compute resume state for a spec |

## Spec Inspection (`oraculo spec`)

Inspect spec artifacts, lifecycle stage, prerequisites, and approvals.

| Subcommand | Args | Description |
|------------|------|-------------|
| `artifacts <spec>` | spec name | Check which artifacts exist for a spec |
| `stage <spec>` | spec name | Classify the lifecycle stage of a spec |
| `prereqs <spec> <command>` | spec, Oráculo command | Check prerequisites for an Oráculo command |
| `approval <spec> <doc-type>` | spec, doc type (`requirements`, `design`, `tasks`) | Check approval status |
| `list` | none | List all specs |

## Workflow Tools (`oraculo tools`)

Internal tools used by subagents and workflows. All support `--raw` flag except `merge-config` and `merge-settings`.

### Config & Resolution

| Command | Description |
|---------|-------------|
| `config-get <section.key> [--default V]` | Read a config value |
| `spec-resolve-dir <spec>` | Resolve spec directory path |
| `wave-resolve-current <spec>` | Resolve current wave number |
| `runs-latest-unfinished <phase-dir>` | Find latest unfinished run directory |
| `resolve-model <alias>` | Resolve model alias to configured model name |

### Dispatch (Pipeline/Audit/Wave)

| Command | Description |
|---------|-------------|
| `dispatch-init <command> <spec>` | Initialize a dispatch run directory. Flag: `--wave` |
| `dispatch-setup <subagent> --run-dir R` | Create subagent directory with `brief.md` skeleton. Flag: `--model-alias` (valid aliases: `web_research`, `complex_reasoning`, `implementation`) |
| `dispatch-read-status <subagent> --run-dir R` | Read and validate subagent `status.json` |
| `dispatch-handoff --run-dir R --command C` | Generate `_handoff.md` from subagent status files |
| `dispatch-init-audit --run-dir R --type T` | Create audit subdirectory (`inline-audit` or `inline-checkpoint`). Flag: `--iteration` |

### Task Management

| Command | Description |
|---------|-------------|
| `verify-task <spec> --task-id N [--check-commit]` | Verify task has implementation log and optionally a commit |
| `impl-log register <spec> --task-id N --wave NN --title T --files F --changes C [--tests T]` | Create implementation log for a completed task |
| `impl-log check <spec> --task-ids 1,2,3` | Check if implementation logs exist for given task IDs |
| `task-mark <spec> --task-id N --status S` | Update task checkbox (`done`, `in_progress`, `pending`) |

### Wave Management

| Command | Description |
|---------|-------------|
| `wave-update <spec> --wave NN --status S --tasks T` | Write wave summary and `_latest.json`. Flags: `--checkpoint-run`, `--execution-run` |
| `wave-status <spec>` | Resolve comprehensive wave state |

### Audit Iteration

| Command | Description |
|---------|-------------|
| `audit-iteration start --run-dir R --type T [--max N]` | Initialize iteration tracking |
| `audit-iteration check --run-dir R --type T` | Check if another iteration is allowed |
| `audit-iteration advance --run-dir R --type T --result R` | Increment counter and record result |

### Handoff & Validation

| Command | Description |
|---------|-------------|
| `handoff-validate <run-dir>` | Validate file-first handoff completeness |
| `skills-effective-set <design\|implementation>` | List effective skills for a stage |
| `approval-fallback-id <spec> <doc-type>` | Get fallback approval ID for a document |

### Configuration Merging

| Command | Description |
|---------|-------------|
| `merge-config <template> <user> <output>` | Merge template TOML with user TOML, preserving user values |
| `merge-settings` | Merge Oraculo hooks into `.claude/settings.json`, preserving non-Oraculo entries |

## Hooks

All hooks read JSON from stdin. Exit codes: `0` = ok, `2` = block. Enforcement mode configured in `oraculo.toml` under `[hooks]`.

| Event | Hook Type | Description |
|-------|-----------|-------------|
| `statusline` | StatusLine | Detects active spec from git diff/cache |
| `session-start` | SessionStart | Syncs active tasks template variant based on TDD config |
| `guard-prompt` | UserPromptSubmit | Validates spec arg presence in Oraculo commands |
| `guard-paths` | PreToolUse (Write/Edit) | Prevents writes outside spec-workflow paths |
| `guard-stop` | Stop | Checks file-first handoff completeness in recent runs |

### Hook behavior details

- **statusline**: Uses 3 detection strategies (git diff, cache file, sticky spec) to find the active spec. Displays token/cost estimates when `show_token_cost = true`. Cache TTL controlled by `statusline.cache_ttl_seconds`. Current task detection shows in-progress task ID in the status bar.
- **guard-prompt**: Blocks Oráculo slash commands that lack a spec argument. Controlled by `hooks.guard_prompt_require_spec`.
- **guard-paths**: Prevents writes outside spec-workflow paths. Also enforces wave-NN directory format and blocks legacy `_agent-comms/` paths when `hooks.guard_wave_layout = true`. Controlled by `hooks.guard_paths`.
- **guard-stop**: Scans recent runs within `hooks.recent_run_window_minutes` for missing handoff artifacts. Controlled by `hooks.guard_stop_handoff`.
- **session-start**: Syncs the active tasks template variant (`tdd-on`/`tdd-off`) based on `[templates].tasks_template_mode`. Also auto-re-renders workflows when config changes are detected.
