# SPW Workflow

This file is intentionally lightweight.

The canonical workflow/usage documentation is centralized in:
- `README.md`
- `AGENTS.md`

Latest SPW behavior updates (CLI default/help + cache/update behavior, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, command guardrails such as unfinished-run handling in long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`, `spw:post-mortem`), shared post-mortem memory/indexing (`.spec-workflow/post-mortems/INDEX.md` + `[post_mortem_memory]`), dashboard markdown compatibility for `tasks.md` including unique IDs + single-line parseable `Files` + structured `_Prompt`, and Mermaid architecture guidance for `design.md` with fenced lowercase `mermaid`) are documented in `README.md`.

Use this file only as an entry point when someone opens `docs/` first.

## Go to the source of truth

- Main guide: `README.md`
- Agent/contributor rules: `AGENTS.md`
- Runtime defaults: `config/spw-config.toml`
