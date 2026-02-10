# SPW

SPW is a command/template kit that combines:
- `spec-workflow-mcp` as the source of truth for artifacts and approvals
- stricter agent execution patterns (planning gates, waves, checkpoints)
- subagent-first orchestration with model routing:
  - web scouting -> `haiku`
  - complex reasoning -> `opus`
  - implementation/drafting -> `sonnet`

## Where to start

- This file is the main source of truth for usage and operations.
- Agent/contributor operational rules are in `AGENTS.md`.
- Keep `docs/SPW-WORKFLOW.md`, `hooks/README.md`, and `copy-ready/README.md` as lightweight pointers to this README.

## Install `spw` CLI

One-liner bootstrap via GitHub CLI (latest `main`):

```bash
gh api 'repos/lucas-stellet/spw/contents/scripts/bootstrap.sh?ref=main' -H 'Accept: application/vnd.github.raw' | bash
```

Public raw fallback (latest `main`):

```bash
curl -fsSL https://raw.githubusercontent.com/lucas-stellet/spw/main/scripts/bootstrap.sh | bash
```

From this repository:

```bash
bash ./scripts/install-spw-bin.sh
```

From anywhere with GitHub CLI:

```bash
tmp_dir="$(mktemp -d)"
gh repo clone lucas-stellet/spw "${tmp_dir}/spw"
bash "${tmp_dir}/spw/scripts/install-spw-bin.sh"
rm -rf "${tmp_dir}"
```

The installed `spw` wrapper caches the kit from GitHub and runs `copy-ready/install.sh`.
Default CLI behavior:
- `spw` prints help output
- `spw install` performs installation in the current project

Useful commands:
- `spw update` (self-update the `spw` wrapper first, then clear cache, fetch fresh repo/ref, and print `ref@commit` + update timestamp)
- `spw doctor` (show current repo/ref/cache configuration, including `ref@commit` and last update timestamp)

## Quick install in another project

Option 1 (recommended, from target project root):

```bash
spw install
```

Optional:

```bash
spw status
spw skills
spw update
spw doctor
```

`spw status` prints a quick kit/skills summary.  
`spw skills` installs default SPW skills only (the default catalog no longer includes `requesting-code-review`).

Option 2 (manual copy):

```bash
cp -R /path/to/spw/copy-ready/. .
```

After install:
1. Merge `.claude/settings.json.example` into your `.claude/settings.json` (if needed).
2. Review `.spec-workflow/spw-config.toml` (fallback legado: `.spw/spw-config.toml`) especially `[planning].tasks_generation_strategy` and `[planning].max_wave_size`.
3. Set per-stage skill enforcement as needed:
   - `skills.design.enforce_required = true|false`
   - `skills.implementation.enforce_required = true|false`
4. Start a new session so SessionStart hook can sync the active tasks template.
5. (Optional) Enable SPW statusline from `.claude/settings.json.example`.
6. Default SPW skills are copied into `.claude/skills/` when local sources are found (best effort).
   - `test-driven-development` belongs to the common/default catalog.
   - `qa-validation-planning` is available for QA planning (`spw:qa`) with Playwright MCP/Bruno CLI guidance.
   - In implementation phases (`spw:exec`, `spw:checkpoint`), this skill is treated as required only when `[execution].tdd_default=true`.
7. (Optional) enable SPW enforcement hooks with `hooks.enforcement_mode=warn|block`.

8. (Optional) For QA browser validation (`spw:qa`, `spw:qa-exec`) and prototype/SPA URL exploration in planning stages (`spw:prd`, `spw:design-research`), configure Playwright MCP:
   ```
   claude mcp add playwright -- npx @playwright/mcp@latest --headless --isolated
   ```

Optional: Agent Teams (disabled by default)
- Enable via installer: `spw install --enable-teams`
- The installer switches symlinks in `.claude/workflows/spw/overlays/active/` from `../noop.md` to `../teams/<cmd>.md`.
- To disable teams, run `spw install` without `--enable-teams` (symlinks reset to `../noop.md`).
- Or manually:
  - set `[agent_teams].enabled = true` in `.spec-workflow/spw-config.toml` (fallback legado: `.spw/spw-config.toml`)
  - add `env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1"` in `.claude/settings.json`
  - set `teammateMode = "in-process"` (change to `"tmux"` manually if desired)
  - switch symlinks: `cd .claude/workflows/spw/overlays/active && ln -sf ../teams/<cmd>.md <cmd>.md`
- When enabled and the phase is NOT listed in `[agent_teams].exclude_phases`, SPW creates a team.
- `spw:exec` enforces delegate mode when `[agent_teams].require_delegate_mode = true`.
- Team overlays are available for all subagent-first entrypoints:
  `spw:prd`, `spw:plan`, `spw:design-research`, `spw:design-draft`,
  `spw:tasks-plan`, `spw:tasks-check`, `spw:exec`, `spw:checkpoint`,
  `spw:post-mortem`, `spw:qa`, `spw:qa-check`, `spw:qa-exec`, `spw:status`.
- All phases are eligible by default (`exclude_phases = []`).

## Command entry points

- `spw:prd` -> zero-to-PRD requirements flow
- `spw:plan` -> design/tasks planning from existing requirements (with MCP approval gate)
- `spw:tasks-plan` -> config-driven task generation (`rolling-wave` or `all-at-once`)
- `spw:exec` -> batch execution with checkpoints
- `spw:checkpoint` -> quality gate report (PASS/BLOCKED)
- `spw:status` -> summarize where workflow stopped + next commands
- `spw:post-mortem` -> analyze post-spec commits and write reusable lessons
- `spw:qa` -> asks validation target and builds a QA test plan with concrete selectors (Playwright MCP/Bruno CLI/hybrid strategy)
- `spw:qa-check` -> validates test plan selectors, traceability, and data feasibility against actual code
- `spw:qa-exec` -> executes validated test plan using verified selectors (never reads source files)

## Thin-Orchestrator Architecture

SPW uses thin orchestrators with a dispatch pattern system:
- command wrappers live in `.claude/commands/spw/*.md`
- detailed orchestration workflows live in `.claude/workflows/spw/*.md`
- shared dispatch policies live in `.claude/workflows/spw/shared/dispatch-{pipeline,audit,wave}.md`
- shared cross-cutting policies live in `.claude/workflows/spw/shared/*.md`

### Dispatch Categories

Every workflow declares a `<dispatch_pattern>` referencing one of three shared policies:

| Category | Policy | Commands |
|----------|--------|----------|
| **Pipeline** | `dispatch-pipeline.md` | `prd`, `design-research`, `design-draft`, `tasks-plan`, `qa`, `post-mortem` |
| **Audit** | `dispatch-audit.md` | `tasks-check`, `qa-check`, `checkpoint` |
| **Wave Execution** | `dispatch-wave.md` | `exec`, `qa-exec` |

All categories enforce the 5 core thin-dispatch rules:
1. Orchestrator reads only `status.json` after dispatch (never `report.md` on pass).
2. Briefs contain filesystem paths to prior reports (never content).
3. Synthesizers/aggregators read from disk directly.
4. Run structure follows category layout.
5. Resume skips completed subagents, always reruns final stage.

Command-specific logic is injected via `<extensions>` at named points (`pre_pipeline`, `pre_dispatch`, `post_dispatch`, `post_pipeline`, `inter_wave`, `per_task`).

### Agent Teams

Agent Teams uses base + overlay via symlinks:
- base workflow: `.claude/workflows/spw/<command>.md`
- active overlay: `.claude/workflows/spw/overlays/active/<command>.md` (symlink)
- teams off: symlink -> `../noop.md` (empty placeholder)
- teams on: symlink -> `../teams/<command>.md`

Wrappers stay intentionally thin and delegate 100% of detailed logic to workflows.

Execution context guardrail (`spw:exec`):
- Before broad reads, dispatch `execution-state-scout` (implementation model, default `sonnet`).
- Scout returns only compact resume state: checkpoint status, task `[-]` in progress, next executable tasks, and required action (`resume|wait-user-authorization|manual-handoff|done|blocked`).
- Orchestrator then reads only task-scoped files for the selected IDs (avoid full `requirements.md`/`design.md` unless needed for blockers).

Planning defaults are configured in `.spec-workflow/spw-config.toml` (fallback legado: `.spw/spw-config.toml`):

```toml
[planning]
tasks_generation_strategy = "rolling-wave" # or "all-at-once"
max_wave_size = 3
```

- `rolling-wave`: each planning cycle creates one executable wave.
  - Typical loop: `tasks-plan` -> `exec` -> `checkpoint` -> `tasks-plan` (next wave)...
- `all-at-once`: one planning pass creates all executable waves.
- Explicit CLI args still override config (`--mode`, `--max-wave-size`).

Post-mortem memory defaults are configured in `.spec-workflow/spw-config.toml` (fallback legado: `.spw/spw-config.toml`):

```toml
[post_mortem_memory]
enabled = true
max_entries_for_design = 5
```

- `spw:post-mortem` writes reports to `.spec-workflow/post-mortems/<spec-name>/`.
- Shared index: `.spec-workflow/post-mortems/INDEX.md` (used by design/planning commands when enabled).
- Design/planning phases (`spw:prd`, `spw:design-research`, `spw:design-draft`, `spw:tasks-plan`, `spw:tasks-check`) load indexed lessons with recency/tag prioritization.

Unfinished-run handling for long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`, `spw:post-mortem`, `spw:qa`, `spw:qa-check`, `spw:qa-exec`):
- Before creating a new run-id, inspect the phase run folder (for `checkpoint`, inspect current wave folder first).
- If latest unfinished run exists, ask explicit user decision:
  - `continue-unfinished`
  - `delete-and-restart`
- Never choose automatically.
- If explicit decision is unavailable, stop with `WAITING_FOR_USER_DECISION`.
- On `continue-unfinished`, reuse completed `status=pass` outputs, redispatch missing/blocked subagents, and rerun the phase final decision/synthesis subagent before final artifact output.

Approval reconciliation for MCP-gated commands (`spw:prd`, `spw:status`, `spw:plan`, `spw:design-draft`, `spw:tasks-plan`):
- First read approval state from `spec-status` document fields.
- If status is missing/unknown/inconsistent, resolve approval ID (from `spec-status` or approval records under `.spec-workflow/approvals/<spec-name>/`) and confirm via MCP `approvals status`.
- `STATUS-SUMMARY.md` is output-only and must not be used as approval source of truth.

File-first subagent communication is stored under phase-based `_comms/` directories:
- prd: `.spec-workflow/specs/<spec-name>/prd/_comms/run-NNN/`
- design: `.spec-workflow/specs/<spec-name>/design/_comms/{design-research,design-draft}/run-NNN/`
- planning: `.spec-workflow/specs/<spec-name>/planning/_comms/{tasks-plan,tasks-check}/run-NNN/`
- execution: `.spec-workflow/specs/<spec-name>/execution/waves/wave-NN/{execution,checkpoint}/run-NNN/`
- qa: `.spec-workflow/specs/<spec-name>/qa/_comms/{qa,qa-check}/run-NNN/`
- qa-exec: `.spec-workflow/specs/<spec-name>/qa/_comms/qa-exec/waves/wave-NN/run-NNN/`
- post-mortem: `.spec-workflow/specs/<spec-name>/post-mortem/_comms/run-NNN/`

`<run-id>` format: `run-NNN` (zero-padded sequential, e.g. `run-001`).

YAML frontmatter (optional metadata) is included in spec templates under the
`spw` key to help subagents classify documents. It does not replace MCP
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

SPW task templates and `spw:tasks-plan` are aligned with this compatibility profile.

## Mermaid for Architecture Design

SPW now includes the `mermaid-architecture` skill for design phases, with common
diagram patterns and syntax guidance:
- skill file: `skills/mermaid-architecture/SKILL.md`
- default config: listed in `[skills.design].optional`

Common architecture examples covered by the skill:
- layered/module boundaries (`flowchart`)
- container/system view (`flowchart`)
- request flow with success/error path (`sequenceDiagram`)
- event-driven pipeline (`flowchart`)
- workflow lifecycle (`stateDiagram-v2`)

In `spw:design-draft`, `design.md` should include at least one valid Mermaid
diagram in the `## Architecture` section, using fenced lowercase `mermaid`
code blocks.

Skills use subagent-first loading by default to reduce main-context growth.

## QA Validation (3-Phase)

QA follows a plan → check → execute chain:

```
spw:qa (plan) → spw:qa-check (validate) → spw:qa-exec (execute)
```

### `spw:qa` (planning)
- Asks user what should be validated when focus is not explicitly provided
- Selects `Playwright MCP`, `Bruno CLI`, or `hybrid` by risk/scope
- Produces `QA-TEST-PLAN.md` with concrete selectors/endpoints per scenario
- Uses browser automation tools from pre-configured Playwright MCP server
- Stores file-first communications under `.spec-workflow/specs/<spec-name>/_agent-comms/qa/<run-id>/`

### `spw:qa-check` (validation)
- Validates test plan against actual code (the ONE phase that reads implementation files)
- Verifies selectors/endpoints exist via `qa-selector-verifier`
- Checks requirement traceability and data feasibility
- Produces `QA-CHECK.md` with verified selector map (test-id → selector → file:line)
- PASS/BLOCKED decision gates `spw:qa-exec`
- Stores file-first communications under `.spec-workflow/specs/<spec-name>/_agent-comms/qa-check/<run-id>/`

### `spw:qa-exec` (execution)
- Executes validated test plan using only verified selectors from `QA-CHECK.md`
- **Never reads implementation source files** — selector drift is logged as defect
- Supports `--scope smoke|regression|full` and `--rerun-failed true|false`
- Produces `QA-EXECUTION-REPORT.md` and `QA-DEFECT-REPORT.md` with GO/NO-GO decision
- Stores file-first communications under `.spec-workflow/specs/<spec-name>/_agent-comms/qa-exec/<run-id>/`

Hook enforcement:
- `warn` -> diagnostics only
- `block` -> deny violating actions
- details: `AGENTS.md` + `.spec-workflow/spw-config.toml` comments (fallback legado: `.spw/spw-config.toml`)
