# AGENTS.md

## Project intent

SPW is a command/template kit for `spec-workflow-mcp` with stricter planning/execution gates and subagent-first orchestration.

## Canonical locations

- `commands/spw/*.md`: command entry points (`spw:prd`, `spw:plan`, `spw:exec`, etc.)
- `templates/user-templates/*.md`: default user templates and TDD variants
- `config/spw-config.toml`: default SPW runtime config
- `hooks/*`: runtime hook scripts (SessionStart sync + statusline)
- `docs/SPW-WORKFLOW.md`: workflow reference
- `copy-ready/`: distributable package copied into target projects

## Working rules for agents

1. Keep runtime files mirrored into `copy-ready/` whenever changing canonical files in `commands/`, `templates/`, `config/`, or `hooks/`.
2. Preserve command names and lifecycle contract unless the same change also updates docs and copy-ready assets.
3. Keep generated planning/research artifacts scoped to `.spec-workflow/specs/<spec-name>/` (never in generic folders like `docs/`).
4. Prefer focused changes; avoid unrelated refactors in the same patch.

## Minimal validation checklist

- `bash -n hooks/session-start-sync-tasks-template.sh`
- `bash -n copy-ready/install.sh`
- `node hooks/spw-statusline.js <<< '{"workspace":{"current_dir":"'"$(pwd)"'"}}'`

## Documentation sync

If behavior or defaults change, update in the same patch:

- `README.md`
- `docs/SPW-WORKFLOW.md`
- `copy-ready/README.md` (for install/runtime-facing changes)
