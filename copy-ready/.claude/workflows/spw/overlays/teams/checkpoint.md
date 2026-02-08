<purpose>
Agent Teams overlay for `spw:checkpoint`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spw/spw-config.toml` (fallback `.spec-workflow/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `use_for_phases`

When `enabled=true` and `checkpoint` is included in `use_for_phases`:
- create a team and set `teammate_mode`
- map subagent roles to teammates (do not exceed `max_teammates`)
- each teammate must still write `brief.md`, `report.md`, `status.json` in the run dir
</agent_teams_policy>

<workflow_overlay>
Apply these additions to base workflow:
- Create team and assign checkpoint roles before subagent dispatch when enabled for phase.
</workflow_overlay>
