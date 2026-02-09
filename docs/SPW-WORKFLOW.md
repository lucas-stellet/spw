# SPW Workflow

This file is intentionally lightweight.

The canonical workflow/usage documentation is centralized in:
- `README.md`
- `AGENTS.md`

Latest SPW behavior updates (CLI default/help + cache/update behavior, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, command guardrails such as unfinished-run handling in long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`, `spw:post-mortem`, `spw:qa`, `spw:qa-check`, `spw:qa-exec`), `spw:exec` state-recon delegation via `execution-state-scout`, MCP approval reconciliation fallback for incomplete `spec-status` payloads, shared post-mortem memory/indexing (`.spec-workflow/post-mortems/INDEX.md` + `[post_mortem_memory]`), expanded Agent Teams overlays/command-pack coverage for subagent-first commands, dashboard markdown compatibility for `tasks.md` including unique IDs + single-line parseable `Files` + structured `_Prompt`, Mermaid architecture guidance for `design.md` with fenced lowercase `mermaid`, 3-phase QA via `spw:qa` (planning with concrete selectors) → `spw:qa-check` (selector/traceability validation) → `spw:qa-exec` (execution without source reads, Playwright MCP/Bruno CLI/hybrid + mandatory headless Playwright), current default skills catalog, and conditional TDD-skill enforcement based on `[execution].tdd_default`) are documented in `README.md`.

Use this file only as an entry point when someone opens `docs/` first.

## Go to the source of truth

- Main guide: `README.md`
- Agent/contributor rules: `AGENTS.md`
- Runtime defaults: `config/spw-config.toml`
- Runtime installed path: `.spec-workflow/spw-config.toml` (fallback legado: `.spw/spw-config.toml`)
