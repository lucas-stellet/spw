# SPW Hooks

SPW provides two runtime hooks/scripts:

- `session-start-sync-tasks-template.sh` (SessionStart hook)
- `spw-statusline.js` (status line command)

Both are fail-open by default (errors are logged, startup continues).

## Backup cleanup options

In `.spec-workflow/spw-config.toml` under `[safety]`:

- `backup_before_overwrite` -> create `.bak-<timestamp>` before replacing template
- `cleanup_backups_after_sync` -> prune backup files on SessionStart
- `backup_retention_count` -> number of backups to keep when cleanup is enabled

Example to remove backups after each sync:

```toml
[safety]
backup_before_overwrite = true
cleanup_backups_after_sync = true
backup_retention_count = 0
```

## Files

- SessionStart script: `spw/hooks/session-start-sync-tasks-template.sh`
- Statusline script: `spw/hooks/spw-statusline.js`
- Config: `.spec-workflow/spw-config.toml`
- Expected variants:
  - `.spec-workflow/user-templates/variants/tasks-template.tdd-on.md`
  - `.spec-workflow/user-templates/variants/tasks-template.tdd-off.md`
- Active target:
  - `.spec-workflow/user-templates/tasks-template.md`

## Project installation

1. Copy the sample TOML:

```bash
mkdir -p .spec-workflow
cp spw/config/spw-config.toml .spec-workflow/spw-config.toml
```

2. Copy template variants:

```bash
mkdir -p .spec-workflow/user-templates/variants
cp spw/templates/user-templates/variants/tasks-template.tdd-on.md .spec-workflow/user-templates/variants/
cp spw/templates/user-templates/variants/tasks-template.tdd-off.md .spec-workflow/user-templates/variants/
```

3. Register SessionStart + statusline in `.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "node <repo>/spw/hooks/spw-statusline.js"
  },
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|resume|clear|compact",
        "hooks": [
          {
            "type": "command",
            "command": "<repo>/spw/hooks/session-start-sync-tasks-template.sh"
          }
        ]
      }
    ]
  }
}
```

See snippet: `spw/hooks/claude-hooks.snippet.json`.

## Quick manual test

```bash
./spw/hooks/session-start-sync-tasks-template.sh
echo '{"workspace":{"current_dir":"'$(pwd)'"}}' | node ./spw/hooks/spw-statusline.js
```

Expected output:
- reports when template was synchronized
- reports when template was already synchronized
- reports missing config/template (without breaking the session by default)
- statusline shows model/project/git/spec/context (best effort)
