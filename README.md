# SPW

SPW is a command/template kit that combines:
- `spec-workflow-mcp` as the source of truth for artifacts and approvals
- stricter agent execution patterns (planning gates, waves, checkpoints)

## Where to start

- Full workflow guide: `spw/docs/SPW-WORKFLOW.md`
- Copy-ready package guide: `spw/copy-ready/README.md`
- Hook setup details: `spw/hooks/README.md`

## Quick install in another project

Option 1 (recommended, from target project root):

```bash
spw-install
```

Option 2 (manual copy):

```bash
cp -R /path/to/spw/copy-ready/. .
```

After install:
1. Merge `.claude/settings.json.example` into your `.claude/settings.json` (if needed).
2. Review `.spec-workflow/spw-config.toml`.
3. Start a new session so SessionStart hook can sync the active tasks template.

## Command entry points

- `spw:prd` -> zero-to-PRD requirements flow
- `spw:plan` -> design/tasks planning from existing requirements (with MCP approval gate)
- `spw:exec` -> batch execution with checkpoints
- `spw:checkpoint` -> quality gate report (PASS/BLOCKED)

