# SPW Hooks

SPW includes session hooks plus optional enforcement hooks.

Core hooks:
- `session-start-sync-tasks-template.sh` (SessionStart)
- `spw-statusline.js` (status line)

Enforcement hooks:
- `spw-guard-user-prompt.js` (UserPromptSubmit)
- `spw-guard-paths.js` (PreToolUse for Write/Edit/MultiEdit)
- `spw-guard-stop.js` (Stop)
- `spw-hook-lib.js` (shared helpers/config parser)

## Enforcement levels (`warn` / `block`)

Hooks read `.spec-workflow/spw-config.toml` `[hooks]`.

- `enforcement_mode = "warn"`
  - prints diagnostics
  - does not block
  - best mode for rollout and tuning

- `enforcement_mode = "block"`
  - denies violating action
  - use after validating warnings

Recommended rollout:
1. Start with `warn`
2. Review warnings with team
3. Switch to `block`

## Hook rules

### UserPromptSubmit (`spw-guard-user-prompt.js`)
Purpose:
- prevent critical `/spw:*` commands without `<spec-name>`

Current guarded commands:
- `spw:prd`, `spw:plan`, `spw:design-research`, `spw:design-draft`
- `spw:tasks-plan`, `spw:tasks-check`
- `spw:exec`, `spw:checkpoint`

### PreToolUse (`spw-guard-paths.js`)
Purpose:
- keep SPW artifacts in canonical spec folders
- enforce wave folder layout

Current rules:
- managed SPW artifacts must be under `.spec-workflow/specs/<spec-name>/`
- deny legacy `agent-comms/checkpoint/...` paths
- enforce `agent-comms/waves/wave-<NN>/...` (zero-padded wave id)

### Stop (`spw-guard-stop.js`)
Purpose:
- ensure recent file-first run folders are complete

Current rules (recent runs only):
- require `_handoff.md`
- require subagent files: `brief.md`, `report.md`, `status.json`

Recent window is configurable with `recent_run_window_minutes`.

## `[hooks]` config reference

```toml
[hooks]
enabled = true
enforcement_mode = "warn"      # warn | block
verbose = true
recent_run_window_minutes = 30

guard_prompt_require_spec = true
guard_paths = true
guard_wave_layout = true
guard_stop_handoff = true
```

Notes:
- `enabled=false` disables all SPW guard hooks.
- each `guard_*` can be toggled independently.

## Backup cleanup options (SessionStart sync hook)

In `.spec-workflow/spw-config.toml` under `[safety]`:
- `backup_before_overwrite`
- `cleanup_backups_after_sync`
- `backup_retention_count`

Example:

```toml
[safety]
backup_before_overwrite = true
cleanup_backups_after_sync = true
backup_retention_count = 0
```

## Project installation

1) Copy config:

```bash
mkdir -p .spec-workflow
cp spw/config/spw-config.toml .spec-workflow/spw-config.toml
```

2) Copy template variants:

```bash
mkdir -p .spec-workflow/user-templates/variants
cp spw/templates/user-templates/variants/tasks-template.tdd-on.md .spec-workflow/user-templates/variants/
cp spw/templates/user-templates/variants/tasks-template.tdd-off.md .spec-workflow/user-templates/variants/
```

3) Merge hook blocks into `.claude/settings.json`:
- use `spw/copy-ready/.claude/settings.json.example`
- or `spw/hooks/claude-hooks.snippet.json`

## Quick manual tests

```bash
./spw/hooks/session-start-sync-tasks-template.sh
echo '{"workspace":{"current_dir":"'$(pwd)'"}}' | node ./spw/hooks/spw-statusline.js
echo '{"prompt":"/spw:plan"}' | node ./spw/hooks/spw-guard-user-prompt.js
echo '{"cwd":"'$(pwd)'","tool_input":{"file_path":"docs/DESIGN-RESEARCH.md"}}' | node ./spw/hooks/spw-guard-paths.js
echo '{}' | node ./spw/hooks/spw-guard-stop.js
```

Expected:
- session hook synchronizes tasks template
- status line renders
- guard hooks warn or block according to `[hooks].enforcement_mode`
