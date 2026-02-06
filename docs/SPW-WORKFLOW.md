# SPW Workflow

This file is intentionally lightweight.

The canonical workflow/usage documentation is centralized in:
- `README.md`
- `AGENTS.md`

Latest SPW behavior updates (CLI cache/update, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, command guardrails such as unfinished-run handling in long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`), dashboard markdown compatibility for `tasks.md`, and Mermaid architecture guidance for `design.md`) are documented in `README.md`.

Use this file only as an entry point when someone opens `docs/` first.

## Go to the source of truth

- Main guide: `README.md`
- Agent/contributor rules: `AGENTS.md`
- Runtime defaults: `config/spw-config.toml`
