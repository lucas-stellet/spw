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
spw
```

Optional:

```bash
spw status
spw skills
```

`spw status` prints a quick kit/skills summary.  
`spw skills` installs default SPW skills only.

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
8. (Optional) enable SPW enforcement hooks with `hooks.enforcement_mode=warn|block`.

Optional: Agent Teams (disabled by default)
- Enable via installer: `spw install --enable-teams`
- The installer overlays team command variants from `.claude/commands/spw-teams/` into `.claude/commands/spw/`.
- Or manually:
  - set `[agent_teams].enabled = true` in `.spec-workflow/spw-config.toml`
  - add `env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1"` in `.claude/settings.json`
  - set `teammateMode = "in-process"` (change to `"tmux"` manually if desired)
  - copy team command variants from `.claude/commands/spw-teams/` into `.claude/commands/spw/`
- When enabled and the phase is listed in `[agent_teams].use_for_phases`, SPW creates a team.
- `spw:exec` enforces delegate mode when `[agent_teams].require_delegate_mode = true`.

## Command entry points

- `spw:prd` -> zero-to-PRD requirements flow
- `spw:plan` -> design/tasks planning from existing requirements (with MCP approval gate)
- `spw:tasks-plan --mode initial|next-wave` -> rolling-wave task generation
- `spw:exec` -> batch execution with checkpoints
- `spw:checkpoint` -> quality gate report (PASS/BLOCKED)
- `spw:status` -> summarize where workflow stopped + next commands

File-first subagent communication is enabled for planning/validation flows and
stored under:
- planning/research: `.spec-workflow/specs/<spec-name>/agent-comms/<command>/<run-id>/`
- execution/checkpoint by wave: `.spec-workflow/specs/<spec-name>/agent-comms/waves/wave-<NN>/<stage>/<run-id>/`

YAML frontmatter (optional metadata) is included in spec templates under the
`spw` key to help subagents classify documents. It does not replace MCP
approvals or status.
- `schema`, `spec`, `doc`, `status`, `source`, `updated_at`
- `inputs`, `requirements`, `decisions`, `task_ids`, `test_required`
- `risk`, `open_questions`

Skills are configured to be `subagent-first` by default to reduce main-context
growth (`skills.load_mode = "subagent-first"`).

Hook enforcement:
- `warn` -> diagnostics only
- `block` -> deny violating actions
- details: `spw/hooks/README.md`
