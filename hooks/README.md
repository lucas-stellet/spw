# SPW Hooks

This file is intentionally lightweight.

Hooks reference:
- Source of truth: `README.md`
- Agent/contributor rules: `AGENTS.md`
- Runtime hook settings: `.spec-workflow/spw-config.toml` (`[hooks]` section)

SPW behavior updates (CLI default/help + cache/update behavior, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, command guardrails such as unfinished-run handling in long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`), dashboard markdown compatibility for `tasks.md` including unique IDs + single-line parseable `Files` + structured `_Prompt`, and Mermaid architecture guidance for `design.md` with fenced lowercase `mermaid`) are tracked in `README.md`.

Hook scripts live in this folder (`hooks/*.js`, `hooks/*.sh`).
