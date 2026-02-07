# SPW Hooks

This file is intentionally lightweight.

Hooks reference:
- Source of truth: `README.md`
- Agent/contributor rules: `AGENTS.md`
- Runtime hook settings: `.spec-workflow/spw-config.toml` (`[hooks]` section)

SPW behavior updates (CLI default/help + cache/update behavior, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, command guardrails such as unfinished-run handling in long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`, `spw:post-mortem`, `spw:qa`), `spw:exec` state-recon delegation via `execution-state-scout`, MCP approval reconciliation fallback for incomplete `spec-status` payloads, shared post-mortem memory/indexing (`.spec-workflow/post-mortems/INDEX.md` + `[post_mortem_memory]`), dashboard markdown compatibility for `tasks.md` including unique IDs + single-line parseable `Files` + structured `_Prompt`, Mermaid architecture guidance for `design.md` with fenced lowercase `mermaid`, QA planning via `spw:qa` (Playwright MCP/Bruno CLI/hybrid + mandatory headless Playwright), current default skills catalog, and conditional TDD-skill enforcement based on `[execution].tdd_default`) are tracked in `README.md`.

Hook scripts live in this folder (`hooks/*.js`, `hooks/*.sh`).
