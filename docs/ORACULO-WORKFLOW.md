# Oraculo Workflow

This file is intentionally lightweight.

The canonical workflow/usage documentation is centralized in:
- `README.md`
- `AGENTS.md`

Latest Oráculo behavior updates (CLI default/help + cache/update behavior, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, command guardrails such as unfinished-run handling in long subagent commands (`oraculo:prd`, `oraculo:design-research`, `oraculo:tasks-plan`, `oraculo:tasks-check`, `oraculo:checkpoint`, `oraculo:post-mortem`, `oraculo:qa`, `oraculo:qa-check`, `oraculo:qa-exec`), `oraculo:exec` state-recon delegation via `execution-state-scout`, MCP approval reconciliation fallback for incomplete `spec-status` payloads, shared post-mortem memory/indexing (`.spec-workflow/post-mortems/INDEX.md` + `[post_mortem_memory]`), expanded Agent Teams overlays/command-pack coverage for subagent-first commands, dashboard markdown compatibility for `tasks.md` including unique IDs + single-line parseable `Files` + structured `_Prompt`, Mermaid architecture guidance for `design.md` with fenced lowercase `mermaid`, 3-phase QA via `oraculo:qa` (planning with concrete selectors) → `oraculo:qa-check` (selector/traceability validation) → `oraculo:qa-exec` (execution without source reads, Playwright MCP/Bruno CLI/hybrid + mandatory headless Playwright), current default skills catalog, conditional TDD-skill enforcement based on `[execution].tdd_default`, and checkpoint guardrails (anti-self-heal, handoff consistency, session isolation between exec and checkpoint)) are documented in `README.md`.

Use this file only as an entry point when someone opens `docs/` first.

## Go to the source of truth

- Main guide: `README.md`
- Agent/contributor rules: `AGENTS.md`
- Runtime defaults: `config/oraculo.toml`
- Runtime installed path: `.spec-workflow/oraculo.toml` (legacy fallback: `.spw/spw-config.toml`)
