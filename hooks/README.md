# SPW Hooks

Hooks are now implemented in Go and invoked via `spw hook <event>`.

Source code: `cli/internal/hook/`

Available hooks:
- `spw hook statusline` — StatusLine: detects active spec from git diff/cache
- `spw hook guard-prompt` — UserPromptSubmit: validates spec arg presence in SPW commands
- `spw hook guard-paths` — PreToolUse (Write/Edit): prevents writes outside spec-workflow paths
- `spw hook guard-stop` — Stop: checks file-first handoff completeness in recent runs
- `spw hook session-start` — SessionStart: syncs active tasks template variant based on TDD config

Configuration: `.spec-workflow/spw-config.toml` (`[hooks]` section, legacy fallback: `.spw/spw-config.toml`)

The `claude-hooks.snippet.json` file in this directory contains the hook configuration snippet for `.claude/settings.json`.
