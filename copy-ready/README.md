# SPW Copy-Ready Kit

This file is intentionally lightweight.

Copy-ready package summary:
- Kit payload is under `copy-ready/.` and is installed by `copy-ready/install.sh`
- Source of truth for usage/workflow: `README.md`
- Agent/contributor rules: `AGENTS.md`

SPW behavior updates (CLI default/help + cache/update behavior, planning defaults such as `[planning].tasks_generation_strategy` and `[planning].max_wave_size`, command guardrails such as unfinished-run handling in long subagent commands (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`), dashboard markdown compatibility for `tasks.md` including unique IDs + single-line parseable `Files` + structured `_Prompt`, and Mermaid architecture guidance for `design.md` with fenced lowercase `mermaid`) are documented in `README.md`.

If you are installing SPW in another project, follow the steps in `README.md`.
