# SPW Hook: SessionStart Template Sync

This hook automatically synchronizes the active tasks template based on `.spec-workflow/spw-config.toml`.

## Files

- Script: `spw/hooks/session-start-sync-tasks-template.sh`
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

3. Register the SessionStart hook in your `.claude/settings.json` (or equivalent config), pointing to:

```text
<repo>/spw/hooks/session-start-sync-tasks-template.sh
```

See JSON structure example in `spw/hooks/claude-hooks.snippet.json`.

## Quick manual test

```bash
./spw/hooks/session-start-sync-tasks-template.sh
```

Expected output:
- reports when template was synchronized
- reports when template was already synchronized
- reports missing config/template (without breaking the session by default)
