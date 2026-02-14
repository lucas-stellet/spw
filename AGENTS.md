# AGENTS.md

## Project purpose

Oraculo is a command/template kit for `spec-workflow-mcp`, with subagent-first execution, explicit approval gates, and per-wave checkpoints.

## Canonical sources (read order)

1. `README.md` (primary source for installation/usage/workflow)
2. `AGENTS.md` (operational rules for agents and contributors)
3. `config/oraculo.toml` (operational defaults)

Note:
- `docs/ORACULO-WORKFLOW.md`, `hooks/README.md`, and `copy-ready/README.md` should remain lean and point to `README.md`.

## Mirror file map

- `commands/oraculo/*.md` <-> `copy-ready/.claude/commands/oraculo/*.md`
- `workflows/oraculo/*.md` <-> `copy-ready/.claude/workflows/oraculo/*.md`
- `workflows/oraculo/overlays/teams/*.md` <-> `copy-ready/.claude/workflows/oraculo/overlays/teams/*.md`
- `workflows/oraculo/overlays/noop.md` <-> `copy-ready/.claude/workflows/oraculo/overlays/noop.md`
- `workflows/oraculo/overlays/active/*.md` <-> `copy-ready/.claude/workflows/oraculo/overlays/active/*.md` (symlinks)
- `templates/user-templates/**` <-> `copy-ready/.spec-workflow/user-templates/**`
- `config/oraculo.toml` <-> `copy-ready/.spec-workflow/oraculo.toml`
- `hooks/*.js|*.sh` <-> `copy-ready/.claude/hooks/*`
- `hooks/claude-hooks.snippet.json` aligned with `copy-ready/.claude/settings.json.example`

## Mandatory operational rules

1. Respect canonical SPW paths: use `.spec-workflow/specs/<spec-name>/` (never `.specs/`).
2. Canonical runtime config: `.spec-workflow/oraculo.toml` (with legacy fallback to `.spw/spw-config.toml`).
3. Keep artifact locality: research/planning artifacts stay inside the active spec; supporting material goes in `.spec-workflow/specs/<spec-name>/research/`.
4. Approval is MCP-only: check status via MCP; never substitute with manual chat approval.
5. Preserve command contracts (`oraculo:prd`, `oraculo:plan`, `oraculo:tasks-plan`, `oraculo:exec`, `oraculo:checkpoint`, `oraculo:status`, `oraculo:post-mortem`, `oraculo:qa`, `oraculo:qa-check`, `oraculo:qa-exec`, `oraculo finalizar`, `oraculo view`, `oraculo search`, `oraculo summary`) and update docs if behavior changes.
6. Thin-orchestrator pattern is mandatory: `commands/` are thin wrappers (max 60 lines) and detailed logic lives in `workflows/`.
7. In `oraculo:tasks-plan`, maintain semantics + precedence:
   - `--mode initial`: generates only the initial executable wave
   - `--mode next-wave`: adds only the next executable wave
   - without `--mode`, use `[planning].tasks_generation_strategy`:
     - `rolling-wave`: generates one executable wave per cycle
     - `all-at-once`: generates all executable waves in a single run
   - `--max-wave-size` overrides `[planning].max_wave_size`; without the argument, use config
8. In `oraculo:exec`, execution is via per-task subagents (including sequential waves of 1 task); the orchestrator never implements code directly.
8b. In `oraculo:checkpoint` (and all audit commands), the orchestrator MUST NOT create, modify, or delete artifacts outside the comms directory to resolve a BLOCKED auditor. If BLOCKED, propagate to final verdict and stop.
9. If `execution.require_user_approval_between_waves=true`, do not advance to the next wave without explicit user authorization.
10. If `execution.commit_per_task="auto"` or `"manual"`, require atomic commit per task; if `"manual"`, stop with explicit git commands; if `"none"`, skip per-task commit enforcement. Respect the clean worktree gate when enabled.
11. `oraculo update` must first update the binary itself (`oraculo`) and then clear the local kit cache before updating, to avoid stale templates/commands.
12. In long-running commands with subagents (`oraculo:prd`, `oraculo:design-research`, `oraculo:tasks-plan`, `oraculo:tasks-check`, `oraculo:checkpoint`, `oraculo:post-mortem`, `oraculo:qa`, `oraculo:qa-check`, `oraculo:qa-exec`), if an incomplete run exists, AskUserQuestion is mandatory (`continue-unfinished` or `delete-and-restart`); the agent must not choose to restart on its own.
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
15. CLI UX: `oraculo` must show help by default; installation is explicit via `oraculo install`.
16. In approval gates (`oraculo:prd`, `oraculo:status`, `oraculo:plan`, `oraculo:design-draft`, `oraculo:tasks-plan`), when `spec-status` returns incomplete/ambiguous, reconcile via MCP `approvals status` (resolving `approvalId` from `spec-status` and, if needed, from `.spec-workflow/approvals/<spec-name>/`); never decide based on `overallStatus`/phases alone and never use `STATUS-SUMMARY.md` as source of truth.
17. In `oraculo:post-mortem`, save reports to `.spec-workflow/post-mortems/<spec-name>/` with YAML front matter (`spec`, `topic`, `tags`, `range_from`, `range_to`) and update `.spec-workflow/post-mortems/INDEX.md`.
18. When `[post_mortem_memory].enabled=true`, design/planning commands (`oraculo:prd`, `oraculo:design-research`, `oraculo:design-draft`, `oraculo:tasks-plan`, `oraculo:tasks-check`) must consult the post-mortems index and apply at most `[post_mortem_memory].max_entries_for_design` relevant entries.
19. Default skill catalog: do not include `requesting-code-review`; keep alignment between `copy-ready/install.sh`, `config/oraculo.toml`, and `copy-ready/.spec-workflow/oraculo.toml`.
20. `test-driven-development` belongs to the common catalog; in `oraculo:exec`/`oraculo:checkpoint`, it only becomes mandatory when `[execution].tdd_default=true`.
21. In `oraculo:exec` (normal and teams), before broad reading the orchestrator must dispatch `execution-state-scout` (implementation/sonnet model by default) to consolidate checkpoint, in-progress `[-]` task, next executable task(s), and resume action; the main agent must consume only the compact summary and then read context per task.
22. In `oraculo:qa`, when the focus is not provided, explicitly ask the user for the validation target and choose `playwright|bruno|hybrid` with risk/scope justification. The plan must include concrete selectors/endpoints per scenario (CSS, `data-testid`, routes, HTTP methods).
23. In Playwright validations within `oraculo:qa`/`oraculo:qa-exec`, use pre-configured Playwright MCP server tools; never invoke npx or Node scripts directly for browser automation.
24. In `oraculo:prd` and `oraculo:design-research`, when a URL returns an SPA shell (minimal HTML with only JS bundle refs) or belongs to a prototype domain (`*.lovable.app`, `*.vercel.app`, etc.), use Playwright MCP to navigate and extract visible content; if unavailable, warn the user and continue with the WebFetch result.
25. Agent Teams coverage for subagent-first commands uses symlinks in `workflows/oraculo/overlays/active/` (pointing to `../noop.md` when disabled or `../teams/<cmd>.md` when enabled); by default all phases are eligible (`[agent_teams].exclude_phases = []`); phases can be excluded by adding them to `exclude_phases`.
26. In `oraculo:qa-check`, validate plan selectors/endpoints against actual source code (the only QA command that reads implementation files); produce a verified map in `QA-CHECK.md`.
27. In `oraculo:qa-exec`, never read implementation source files; use only verified selectors from `QA-CHECK.md`. If a selector fails at runtime, log it as a "selector drift" defect and recommend `oraculo:qa-check`.

## Thin-dispatch pattern (mandatory)

All Oraculo commands follow the thin-dispatch model (`docs/DISPATCH-PATTERNS.md`):

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
37. Phase directories are created on demand. If `oraculo:qa` has never run, `qa/` does not exist.

## PR review optimization (mandatory)

Spec-workflow files are marked as `linguist-generated` so GitHub collapses them by default in PR diffs. Full reference: `docs/PR-REVIEW-OPTIMIZATION.md`.

38. The installer (`oraculo install`) must add `.spec-workflow/specs/** linguist-generated=true` to the project's `.gitattributes`. The rule is idempotent -- if it already exists, do not duplicate.
39. Only files under `.spec-workflow/specs/` are marked. Config (`.spec-workflow/oraculo.toml`) and templates (`.spec-workflow/user-templates/`) are not affected.

## Inline verification (mandatory)

40. Subagents that produce artifacts MUST run a self-check before reporting `pass`. Implementation subagents use `oraculo tools impl-log register` and `oraculo tools verify-task --check-commit`. Self-check results go in `status.json` (`self_check` field).
41. After reading a subagent status via `dispatch-read-status`, the orchestrator runs an independent spot-check with `oraculo tools verify-task`. If spot-check fails, treat as BLOCKED.
42. Producer commands (`oraculo:tasks-plan`, `oraculo:qa`) run inline audit after artifact generation, using `oraculo tools audit-iteration start/check/advance` and `oraculo tools dispatch-init-audit`. Max iterations from `[verification].inline_audit_max_iterations`.
43. `oraculo:exec` runs an inline checkpoint at the end of each wave (evidence-collector, traceability-judge, release-gate-decider). On PASS: `oraculo tools wave-update`. On BLOCKED: max 1 retry, then recommend standalone `/oraculo:checkpoint`.
44. Task status updates use `oraculo tools task-mark --status <in-progress|done|blocked>` instead of manual edits to `tasks.md`.
45. Wave state resolution uses `oraculo tools wave-status --spec <name>` for deterministic state (replaces AI interpretation of wave directories).
46. Standalone check commands (`oraculo:tasks-check`, `oraculo:checkpoint`, `oraculo:qa-check`) remain available for re-validation after manual fixes, CI/CD pipelines, or when inline audit recommends them.

## Local Storage (mandatory)

47. `spec.db` (per-spec SQLite, WAL mode) is created automatically by `oraculo finalizar`. The dispatch-handoff dual-writes artifacts into the DB when the store is available.
48. `.oraculo-index.db` (global FTS5 index) is updated by `oraculo finalizar` when registering a completed spec. `oraculo search` reads only from this index.
49. The 3 MCP-managed files (`requirements.md`, `design.md`, `tasks.md`) remain on disk as source of truth for the dashboard. The DB is complementary, not a substitute.
50. `oraculo finalizar` must be executed after completion of all waves and QA. It generates a completion summary with YAML frontmatter and indexes the spec in the global index.

## File-first comms (do not break)

For commands that require file-based handoff, ensure the following files are present:

- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

Absence of these files must result in `BLOCKED`.

## Minimum validation checklist

- `bash -n bin/oraculo`
- `bash -n scripts/bootstrap.sh`
- `bash -n scripts/install-oraculo-bin.sh`
- `bash -n scripts/validate-thin-orchestrator.sh`
- `scripts/validate-thin-orchestrator.sh`
- `bash -n copy-ready/install.sh`
- `echo '{"workspace":{"current_dir":"'"$(pwd)"'"}}' | oraculo hook statusline`
- `echo '{"prompt":"/oraculo:plan"}' | oraculo hook guard-prompt`
- `echo '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":"README.md"}}' | oraculo hook guard-paths`
- `echo '{}' | oraculo hook guard-stop`
- `echo '{}' | oraculo hook session-start`
- `oraculo finalizar --help`
- `oraculo view --help`
- `oraculo search --help`
- `oraculo summary --help`
- `oraculo tasks state --help`
- `oraculo wave state --help`
- `oraculo spec list --help`

## Documentation sync

When changing behavior, defaults, or guardrails, update in the same patch:

- `README.md`
- `AGENTS.md`
- `docs/ORACULO-WORKFLOW.md` (pointer)
- `hooks/README.md` (pointer)
- `copy-ready/README.md` (pointer)
- `.claude/docs/oraculo-cli-reference.md` (CLI reference)

<!-- ORACULO-KIT-START — managed by oraculo install, do not edit manually -->
## Oraculo Dispatch Rules

When executing Oraculo workflow commands, follow these rules strictly:

1. **Always use CLI for dispatch.** Never create run dirs or subagent dirs manually. Use `oraculo tools dispatch-init` and `oraculo tools dispatch-setup`.
2. **Status-only reads.** After dispatching a subagent, read ONLY status.json via `oraculo tools dispatch-read-status`. Never read report.md unless status=blocked.
3. **Paths, not content.** When subagent-B depends on subagent-A, write the filesystem PATH to A's report.md in B's brief.md. Never copy content.
4. **Synthesizer reads from disk.** The final subagent (synthesizer/aggregator) receives a brief listing all report paths and reads them directly.
5. **MCP inline exception.** When a subagent needs session-scoped MCP tools (Linear, Playwright), run dispatch-setup normally but execute the work inline — still write report.md and status.json to the subagent directory.
6. **Always finalize.** Call `oraculo tools dispatch-handoff --run-dir <dir>` after all subagents complete.
<!-- ORACULO-KIT-END -->
