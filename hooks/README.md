# SPW Hooks

Hooks are now implemented in Go and invoked via `spw hook <event>`.

Source code: `cli/internal/hook/`

Available hooks:
- `spw hook statusline` — StatusLine: detects active spec from git diff/cache, shows token usage and cost
- `spw hook guard-prompt` — UserPromptSubmit: validates spec arg presence in SPW commands
- `spw hook guard-paths` — PreToolUse (Write/Edit): prevents writes outside spec-workflow paths
- `spw hook guard-stop` — Stop: checks file-first handoff completeness in recent runs
- `spw hook session-start` — SessionStart: syncs active tasks template variant based on TDD config

Configuration: `.spec-workflow/spw-config.toml` (`[hooks]` section, legacy fallback: `.spw/spw-config.toml`)

## Statusline Token & Cost Display

When Claude Code sends `context_window.total_input_tokens`, `context_window.total_output_tokens`, and `cost.total_cost_usd` in the statusline payload, SPW displays cumulative token usage and cost:

```
Model | Task | Dir | spec:name | 25.3k $0.42 | ████░░░░░░ 50%
```

Token counts are formatted compactly (`847`, `25.3k`, `1.2M`). Input + output tokens are summed.

### `show_token_cost` config (`[statusline]`)

| Value | Behavior |
|-------|----------|
| `"auto"` (default) | Show only when `cost.total_cost_usd > 0` (API-key billing detected). Subscription users see `$0` and the segment is hidden. |
| `"always"` | Show whenever any token or cost data is present in the payload. |
| `"never"` | Never show the token/cost segment. |

The `claude-hooks.snippet.json` file in this directory contains the hook configuration snippet for `.claude/settings.json`.
