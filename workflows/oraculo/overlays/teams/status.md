<purpose>
Agent Teams overlay for `oraculo:status`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/oraculo.toml` (legacy fallback `.oraculo/oraculo.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)

When `enabled=true` and `status` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- map status roles to teammates (`state-inspector`, `approval-auditor`, `next-step-planner`) (do not exceed `max_teammates`)
</agent_teams_policy>

<workflow_overlay>
Apply these additions to base workflow:
- Create team before status-role dispatch when enabled for phase.
- When inspecting multiple specs, reuse the same team across the command run.
</workflow_overlay>
