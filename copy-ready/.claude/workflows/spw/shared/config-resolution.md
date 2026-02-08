# Config Resolution

Canonical runtime config path is `.spec-workflow/spw-config.toml`.

Transitional compatibility:
- If `.spec-workflow/spw-config.toml` is missing, fallback to `.spw/spw-config.toml`.

When shell logic is required, prefer:
- `node .claude/hooks/spw-tools.js config get <section.key> --default <value> [--raw]`

This keeps workflow behavior stable during migration and avoids hardcoded path drift.
