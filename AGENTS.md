# AGENTS.md

## Project purpose

SPW is a command/template kit for `spec-workflow-mcp`, with subagent-first execution, explicit approval gates, and per-wave checkpoints.

## Canonical sources (read order)

1. `README.md` (primary source for installation/usage/workflow)
2. `AGENTS.md` (operational rules for agents and contributors)
3. `config/spw-config.toml` (operational defaults)

Note:
- `docs/SPW-WORKFLOW.md`, `hooks/README.md`, and `copy-ready/README.md` should remain lean and point to `README.md`.

## Mirror file map

- `commands/spw/*.md` <-> `copy-ready/.claude/commands/spw/*.md`
- `workflows/spw/*.md` <-> `copy-ready/.claude/workflows/spw/*.md`
- `workflows/spw/overlays/teams/*.md` <-> `copy-ready/.claude/workflows/spw/overlays/teams/*.md`
- `workflows/spw/overlays/noop.md` <-> `copy-ready/.claude/workflows/spw/overlays/noop.md`
- `workflows/spw/overlays/active/*.md` <-> `copy-ready/.claude/workflows/spw/overlays/active/*.md` (symlinks)
- `templates/user-templates/**` <-> `copy-ready/.spec-workflow/user-templates/**`
- `config/spw-config.toml` <-> `copy-ready/.spec-workflow/spw-config.toml`
- `hooks/*.js|*.sh` <-> `copy-ready/.claude/hooks/*`
- `hooks/claude-hooks.snippet.json` aligned with `copy-ready/.claude/settings.json.example`

## Mandatory operational rules

1. Respect canonical SPW paths: use `.spec-workflow/specs/<spec-name>/` (never `.specs/`).
2. Canonical runtime config: `.spec-workflow/spw-config.toml` (with legacy fallback to `.spw/spw-config.toml`).
3. Keep artifact locality: research/planning artifacts stay inside the active spec; supporting material goes in `.spec-workflow/specs/<spec-name>/research/`.
4. Approval is MCP-only: check status via MCP; never substitute with manual chat approval.
5. Preserve command contracts (`spw:prd`, `spw:plan`, `spw:tasks-plan`, `spw:exec`, `spw:checkpoint`, `spw:status`, `spw:post-mortem`, `spw:qa`, `spw:qa-check`, `spw:qa-exec`) and update docs if behavior changes.
6. Thin-orchestrator pattern is mandatory: `commands/` are thin wrappers (max 60 lines) and detailed logic lives in `workflows/`.
7. In `spw:tasks-plan`, maintain semantics + precedence:
   - `--mode initial`: generates only the initial executable wave
   - `--mode next-wave`: adds only the next executable wave
   - without `--mode`, use `[planning].tasks_generation_strategy`:
     - `rolling-wave`: generates one executable wave per cycle
     - `all-at-once`: generates all executable waves in a single run
   - `--max-wave-size` overrides `[planning].max_wave_size`; without the argument, use config
8. In `spw:exec`, execution is via per-task subagents (including sequential waves of 1 task); the orchestrator never implements code directly.
8b. In `spw:checkpoint` (and all audit commands), the orchestrator MUST NOT create, modify, or delete artifacts outside the comms directory to resolve a BLOCKED auditor. If BLOCKED, propagate to final verdict and stop.
9. If `execution.require_user_approval_between_waves=true`, do not advance to the next wave without explicit user authorization.
10. If `execution.commit_per_task="auto"` or `"manual"`, require atomic commit per task; if `"manual"`, stop with explicit git commands; if `"none"`, skip per-task commit enforcement. Respect the clean worktree gate when enabled.
11. `spw update` must first update the binary itself (`spw`) and then clear the local kit cache before updating, to avoid stale templates/commands.
12. In long-running commands with subagents (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`, `spw:post-mortem`, `spw:qa`, `spw:qa-check`, `spw:qa-exec`), if an incomplete run exists, AskUserQuestion is mandatory (`continue-unfinished` or `delete-and-restart`); the agent must not choose to restart on its own.
13. Dashboard compatibility (`spec-workflow-mcp`) in `tasks.md` is mandatory:
   - checkboxes only on task lines (`- [ ]`, `- [-]`, `- [x]` with numeric ID)
   - task IDs must be unique within the file (no duplicates)
   - task rows use `-` (never `*` for task lines)
   - never use nested checkboxes in DoD/metadata
   - metadata must be plain bullets (`- ...`), never checkboxes
   - `Files` must be parseable on a single line (`- Files: a, b`)
   - use underscore-delimited metadata: `_Requirements: ..._`, `_Leverage: ..._` (when applicable), `_Prompt: ..._` (closing with `_`)
   - `_Prompt` must include `Role|Task|Restrictions|Success`
14. In `design.md`, include at least one valid Mermaid diagram in `## Architecture` (main flow), preferring the `mermaid-architecture` skill for standardization.
   - use a fenced code block with lowercase `mermaid` language marker
15. CLI UX: `spw` must show help by default; installation is explicit via `spw install`.
16. In approval gates (`spw:prd`, `spw:status`, `spw:plan`, `spw:design-draft`, `spw:tasks-plan`), when `spec-status` returns incomplete/ambiguous, reconcile via MCP `approvals status` (resolving `approvalId` from `spec-status` and, if needed, from `.spec-workflow/approvals/<spec-name>/`); never decide based on `overallStatus`/phases alone and never use `STATUS-SUMMARY.md` as source of truth.
17. In `spw:post-mortem`, save reports to `.spec-workflow/post-mortems/<spec-name>/` with YAML front matter (`spec`, `topic`, `tags`, `range_from`, `range_to`) and update `.spec-workflow/post-mortems/INDEX.md`.
18. When `[post_mortem_memory].enabled=true`, design/planning commands (`spw:prd`, `spw:design-research`, `spw:design-draft`, `spw:tasks-plan`, `spw:tasks-check`) must consult the post-mortems index and apply at most `[post_mortem_memory].max_entries_for_design` relevant entries.
19. Default skill catalog: do not include `requesting-code-review`; keep alignment between `copy-ready/install.sh`, `config/spw-config.toml`, and `copy-ready/.spec-workflow/spw-config.toml`.
20. `test-driven-development` belongs to the common catalog; in `spw:exec`/`spw:checkpoint`, it only becomes mandatory when `[execution].tdd_default=true`.
21. In `spw:exec` (normal and teams), before broad reading the orchestrator must dispatch `execution-state-scout` (implementation/sonnet model by default) to consolidate checkpoint, in-progress `[-]` task, next executable task(s), and resume action; the main agent must consume only the compact summary and then read context per task.
22. In `spw:qa`, when the focus is not provided, explicitly ask the user for the validation target and choose `playwright|bruno|hybrid` with risk/scope justification. The plan must include concrete selectors/endpoints per scenario (CSS, `data-testid`, routes, HTTP methods).
23. In Playwright validations within `spw:qa`/`spw:qa-exec`, use pre-configured Playwright MCP server tools; never invoke npx or Node scripts directly for browser automation.
24. In `spw:prd` and `spw:design-research`, when a URL returns an SPA shell (minimal HTML with only JS bundle refs) or belongs to a prototype domain (`*.lovable.app`, `*.vercel.app`, etc.), use Playwright MCP to navigate and extract visible content; if unavailable, warn the user and continue with the WebFetch result.
25. Agent Teams coverage for subagent-first commands uses symlinks in `workflows/spw/overlays/active/` (pointing to `../noop.md` when disabled or `../teams/<cmd>.md` when enabled); by default all phases are eligible (`[agent_teams].exclude_phases = []`); phases can be excluded by adding them to `exclude_phases`.
26. In `spw:qa-check`, validate plan selectors/endpoints against actual source code (the only QA command that reads implementation files); produce a verified map in `QA-CHECK.md`.
27. In `spw:qa-exec`, never read implementation source files; use only verified selectors from `QA-CHECK.md`. If a selector fails at runtime, log it as a "selector drift" defect and recommend `spw:qa-check`.

## Thin-dispatch pattern (mandatory)

All SPW commands follow the thin-dispatch model (`docs/DISPATCH-PATTERNS.md`):

28. The orchestrator reads ONLY `status.json` after each subagent dispatch. It never reads `report.md` in the normal flow -- only when `status=blocked` (to decide on action).
29. Between subagents, the orchestrator passes file paths in `brief.md`, never inline content. Subagent-B receives the path to subagent-A's `report.md`, not its content.
30. Synthesizers and aggregators read all previous reports directly from the filesystem via paths received in the brief.
31. Commands are categorized into three dispatch patterns:
    - **Pipeline** (sequence -> synthesizer): `prd`, `design-research`, `design-draft`, `tasks-plan`, `qa`, `post-mortem`
    - **Audit** (parallel auditors -> aggregator): `tasks-check`, `qa-check`, `checkpoint`
    - **Wave Execution** (scout -> iterative waves -> synthesizer): `exec`, `qa-exec`
32. In Wave Execution, iterative work (tasks, scenarios) is divided into waves (`wave-NN`). Each wave dispatches subagents sequentially, writes `_wave-summary.json`, and the orchestrator only accumulates status -- never full results.

## Phase-based directory structure (mandatory)

Artifacts are organized by **workflow phase**, not in flat dumps. Each phase owns its outputs and its agent comms (`_comms/`). Full reference: `docs/SPEC-DIRECTORY-STRUCTURE.md`.

33. Generated artifacts belong to the phase directory that produced them (e.g., `qa/QA-CHECK.md`, not `_generated/QA-CHECK.md`). The top-level `_generated/` and `_agent-comms/` directories no longer exist.
34. Agent comms go in `<phase>/_comms/<command>/run-NNN/` (Pipeline and Audit) or `<phase>/_comms/<command>/waves/wave-NN/run-NNN/` (Wave Execution).
35. Phases: `prd/`, `design/`, `planning/`, `execution/`, `qa/`, `post-mortem/`. When a phase contains commands from different categories (e.g., `qa/` has pipeline, audit, and wave), each command uses the appropriate `_comms/` subdirectory.
36. Dashboard files (`requirements.md`, `design.md`, `tasks.md`) remain at the spec root -- the MCP dashboard reads from here.
37. Phase directories are created on demand. If `spw:qa` has never run, `qa/` does not exist.

## PR review optimization (mandatory)

Spec-workflow files are marked as `linguist-generated` so GitHub collapses them by default in PR diffs. Full reference: `docs/PR-REVIEW-OPTIMIZATION.md`.

38. The installer (`spw install`) must add `.spec-workflow/specs/** linguist-generated=true` to the project's `.gitattributes`. The rule is idempotent -- if it already exists, do not duplicate.
39. Only files under `.spec-workflow/specs/` are marked. Config (`.spec-workflow/spw-config.toml`) and templates (`.spec-workflow/user-templates/`) are not affected.

## File-first comms (do not break)

For commands that require file-based handoff, ensure the following files are present:

- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

Absence of these files must result in `BLOCKED`.

## Minimum validation checklist

- `bash -n bin/spw`
- `bash -n scripts/bootstrap.sh`
- `bash -n scripts/install-spw-bin.sh`
- `bash -n scripts/validate-thin-orchestrator.sh`
- `scripts/validate-thin-orchestrator.sh`
- `bash -n copy-ready/install.sh`
- `echo '{"workspace":{"current_dir":"'"$(pwd)"'"}}' | spw hook statusline`
- `echo '{"prompt":"/spw:plan"}' | spw hook guard-prompt`
- `echo '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":"README.md"}}' | spw hook guard-paths`
- `echo '{}' | spw hook guard-stop`
- `echo '{}' | spw hook session-start`

## Documentation sync

When changing behavior, defaults, or guardrails, update in the same patch:

- `README.md`
- `AGENTS.md`
- `docs/SPW-WORKFLOW.md` (pointer)
- `hooks/README.md` (pointer)
- `copy-ready/README.md` (pointer)
