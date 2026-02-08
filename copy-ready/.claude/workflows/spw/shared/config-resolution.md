# Config Resolution

Canonical runtime config path is `.spw/spw-config.toml`.

Transitional compatibility:
- If `.spw/spw-config.toml` is missing, fallback to `.spec-workflow/spw-config.toml`.

When shell logic is required, prefer:
- `node .claude/hooks/spw-tools.js config get <section.key> --default <value> [--raw]`

This keeps workflow behavior stable during migration and avoids hardcoded path drift.
