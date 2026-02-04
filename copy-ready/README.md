# SPW Copy-Ready Kit

Ready-to-copy package for any project using `spec-workflow-mcp`.

## What the kit includes

- `.claude/commands/spw/*.md` (planning/execution/checkpoint commands)
- `.claude/hooks/session-start-sync-tasks-template.sh` (sync hook)
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

1. Merge `.claude/settings.json.example` into your `.claude/settings.json`.
2. Adjust `.spec-workflow/spw-config.toml` (especially `execution.tdd_default`).
3. Start a new session so the hook syncs `tasks-template.md`.

## spec-workflow compatibility

This kit only uses:
- `.spec-workflow/user-templates/` to override custom templates
- `.spec-workflow/spw-config.toml` for runtime workflow config

It does not modify default templates under `.spec-workflow/templates/`.

## Default subagent/model policy

- Subagent-first workflows across product, planning, execution, and checkpoints.
- Model routing comes from `.spec-workflow/spw-config.toml`:
  - web-only research/scouting -> `haiku`
  - complex synthesis/validation gates -> `opus`
  - implementation/drafting -> `sonnet`

## Available commands

- `/spw:prd` (zero-to-PRD: generates requirements)
- `/spw:plan` (from existing requirements: generates design/tasks; validates approval via MCP)
- `/spw:design-research`
- `/spw:design-draft`
- `/spw:tasks-plan`
- `/spw:tasks-check`
- `/spw:exec`
- `/spw:checkpoint`

All commands include end-of-command guidance: next-step command, blocked remediation path, and context reset suggestion when appropriate.
