# SPW Hooks

This file is intentionally lightweight.

Hooks reference:
- Source of truth: `README.md`
- Agent/contributor rules: `AGENTS.md`
- Runtime hook settings: `.spec-workflow/spw-config.toml` (`[hooks]` section)

SPW behavior updates (CLI cache/update, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, and command guardrails such as unfinished-run handling in long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`)) are tracked in `README.md`.

Hook scripts live in this folder (`hooks/*.js`, `hooks/*.sh`).
