<purpose>
Agent Teams overlay for `spw:plan`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (legacy fallback `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)

When `enabled=true` and `plan` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- map planner roles (`requirements-approval-gate`, `planning-stage-orchestrator`) to teammates (do not exceed `max_teammates`)
</agent_teams_policy>

<workflow_overlay>
Apply these additions to base workflow:
- Create team before dispatching approval/planning orchestration roles when enabled for phase.
- Keep MCP-only approval behavior unchanged.
</workflow_overlay>
