# SPW Copy-Ready Kit

Ready-to-copy package for any project using `spec-workflow-mcp`.

## What the kit includes

- `.claude/commands/spw/*.md` (planning/execution/checkpoint commands)
- `.claude/hooks/session-start-sync-tasks-template.sh` (sync hook)
- `.claude/hooks/spw-statusline.js` (statusline: project/spec/context)
- `.claude/settings.json.example` (hook config snippet)
- `.spec-workflow/spw-config.toml` (central config with extensive comments)
- `.spec-workflow/user-templates/*.md` (custom templates)
- `.spec-workflow/user-templates/prd-template.md` (PRD template)
- `.spec-workflow/user-templates/variants/tasks-template.tdd-*.md` (ON/OFF variants)

## How to install in the target project

From the target project root:

```bash
cp -R /PATH/TO/spw/copy-ready/. .
```

Then:

1. Merge `.claude/settings.json.example` into your `.claude/settings.json` (SessionStart + statusLine).
2. Adjust `.spec-workflow/spw-config.toml` (especially `execution.tdd_default`, `skills.design.enforce_required`, and `skills.implementation.enforce_required`).
3. Start a new session so the hook syncs `tasks-template.md`.
4. `spw-install` also tries to install default SPW skills into `.claude/skills/` (best effort, non-blocking).

If you do not want `.bak` accumulation for `tasks-template.md`, set:
- `safety.cleanup_backups_after_sync = true`
- `safety.backup_retention_count = 0` (delete all backups after sync)

Hook enforcement levels:
- `hooks.enforcement_mode = "warn"` (recommended rollout)
- `hooks.enforcement_mode = "block"` (strict mode)
- see `.claude/hooks/*` and `spw/hooks/README.md` for guard behavior

## spec-workflow compatibility

This kit only uses:
- `.spec-workflow/user-templates/` to override custom templates
- `.spec-workflow/spw-config.toml` for runtime workflow config

It does not modify default templates under `.spec-workflow/templates/`.

Artifact locality rule:
- Generated planning/research files must stay under `.spec-workflow/specs/<spec-name>/`.
- Supporting research files must stay under `.spec-workflow/specs/<spec-name>/research/`.
- Avoid generating research files in generic folders like `docs/`.

YAML frontmatter (optional metadata) is included in spec templates under the
`spw` key to help subagents classify documents. This does not replace MCP
approvals or status.
- `schema`, `spec`, `doc`, `status`, `source`, `updated_at`
- `inputs`, `requirements`, `decisions`, `task_ids`, `test_required`
- `risk`, `open_questions`

## Default subagent/model policy

- Subagent-first workflows across product, planning, execution, and checkpoints.
- Model routing comes from `.spec-workflow/spw-config.toml`:
  - web-only research/scouting -> `haiku`
  - complex synthesis/validation gates -> `opus`
  - implementation/drafting -> `sonnet`

## Default skills (installed by `spw-install` when found locally)

- Elixir defaults:
  - `using-elixir-skills`
  - `elixir-thinking`
  - `elixir-anti-patterns`
  - `phoenix-thinking`
  - `ecto-thinking`
  - `otp-thinking`
  - `oban-thinking`
- Git hygiene:
  - `conventional-commits`
- Optional quality/TDD:
  - `test-driven-development`
  - `requesting-code-review`

The installer searches common local skill directories (`~/.claude/skills`, `~/.codex/skills`, `~/.codex/superpowers/skills`) and also checks the local `superpowers/skills` folder (or `SPW_SUPERPOWERS_SKILLS_DIR`) when available.

## Available commands

- `/spw:prd` (zero-to-PRD: generates requirements)
- `/spw:plan` (from existing requirements: generates design/tasks; validates approval via MCP)
- `/spw:design-research`
- `/spw:design-draft`
- `/spw:tasks-plan`
- `/spw:tasks-check`
- `/spw:exec`
- `/spw:checkpoint`
- `/spw:status`

All commands include end-of-command guidance: next-step command, blocked remediation path, and context reset suggestion when appropriate.

Subagent coverage:
- Subagent-driven commands: `/spw:prd`, `/spw:plan`, `/spw:design-research`, `/spw:design-draft`, `/spw:tasks-plan`, `/spw:tasks-check`, `/spw:exec`, `/spw:checkpoint`.
- Orchestrator-only parts (non-subagent): MCP approval checks, AskUserQuestion prompts, wait/block states, hooks/install scripts.

File-first communication (GSD-style):
- `/spw:design-research`, `/spw:prd`, `/spw:tasks-plan`, `/spw:tasks-check`, and `/spw:checkpoint` persist subagent communication under:
  - planning/research: `.spec-workflow/specs/<spec-name>/agent-comms/<command>/<run-id>/`
  - execution/checkpoint by wave: `.spec-workflow/specs/<spec-name>/agent-comms/waves/wave-<NN>/<stage>/<run-id>/`
- each subagent writes `brief.md`, `report.md`, `status.json`
- each run writes `_handoff.md`
- missing communication files cause BLOCKED

Runtime guard hooks (optional, configurable in `[hooks]`):
- prompt guard: requires `<spec-name>` on critical `/spw:*` commands
- path guard: enforces SPW artifact paths + wave folder layout
- stop guard: enforces recent run handoff file completeness

Skill load strategy:
- default is `skills.load_mode = "subagent-first"` to avoid filling main context early
- set `principal-first` only if you intentionally want skill loading in the orchestrator

`/spw:exec` guardrails:
- mandatory subagent dispatch per task (including single-task sequential waves)
- orchestrator-only main context (no direct implementation edits)
- out-of-scope fixes are blocked and must be reported explicitly
- default human gate between waves (no auto-continue without explicit authorization)
- default atomic commit per completed task
- Conventional Commits for task-level commits (`<type>(<spec>): task <id> - <title>`)
- default clean-worktree requirement before wave progression
- manual/human-gated tasks stop in handoff mode (no automatic in-progress mark)
- missing implementation logs are enforced at `/spw:checkpoint` (BLOCKED if absent)

## Manual planning/refinement flow

If you want to run planning stages manually (instead of `/spw:plan`):

```bash
/spw:design-research <spec-name>
/spw:design-draft <spec-name>
/spw:tasks-plan <spec-name> --mode initial --max-wave-size 3
/spw:tasks-check <spec-name>
```

Rolling-wave follow-up after a wave is executed:

```bash
/spw:tasks-plan <spec-name> --mode next-wave --max-wave-size 3
/spw:tasks-check <spec-name>
```

Approval-related outputs:
- `/spw:design-draft` -> `design.md` (approval requested)
- `/spw:tasks-plan` -> `tasks.md` (approval requested)
- `/spw:design-research` -> `DESIGN-RESEARCH.md` (input/report)
- `/spw:tasks-check` -> `TASKS-CHECK.md` (validation report)

Note:
- `/spw:design-draft` requires `DESIGN-RESEARCH.md`; run `/spw:design-research <spec-name>` first.
