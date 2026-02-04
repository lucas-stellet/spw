# SPW

SPW is a command/template kit that combines:
- `spec-workflow-mcp` as the source of truth for artifacts and approvals
- stricter agent execution patterns (planning gates, waves, checkpoints)
- subagent-first orchestration with model routing:
  - web scouting -> `haiku`
  - complex reasoning -> `opus`
  - implementation/drafting -> `sonnet`

## Where to start

- Full workflow guide: `spw/docs/SPW-WORKFLOW.md`
- Copy-ready package guide: `spw/copy-ready/README.md`
- Hook setup details: `spw/hooks/README.md`
- Manual planning order + refinement loops: see `spw/docs/SPW-WORKFLOW.md` ("Manual planning order (explicit)")

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
3. Set per-stage skill enforcement as needed:
   - `skills.design.enforce_required = true|false`
   - `skills.implementation.enforce_required = true|false`
4. Start a new session so SessionStart hook can sync the active tasks template.
5. (Optional) Enable SPW statusline from `.claude/settings.json.example`.
6. Default SPW skills are copied into `.claude/skills/` when local sources are found (best effort).
7. (Optional) auto-clean template backups with `safety.cleanup_backups_after_sync=true` in `.spec-workflow/spw-config.toml`.

## Command entry points

- `spw:prd` -> zero-to-PRD requirements flow
- `spw:plan` -> design/tasks planning from existing requirements (with MCP approval gate)
- `spw:exec` -> batch execution with checkpoints
- `spw:checkpoint` -> quality gate report (PASS/BLOCKED)
- `spw:status` -> summarize where workflow stopped + next commands
