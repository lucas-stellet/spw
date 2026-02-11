<purpose>
Agent Teams overlay for `spw:qa-check`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (legacy fallback `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)

When `enabled=true` and `qa-check` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- map subagent roles to teammates (do not exceed `max_teammates`)
- each teammate must still write `brief.md`, `report.md`, `status.json` in the run dir
</agent_teams_policy>

<extensions_overlay>
Apply these additions to base extensions:
- In `<pre_pipeline>`: create team and assign qa-check roles before dispatch when enabled for phase.
Dispatch mechanism comes from dispatch-audit.md shared policy.
</extensions_overlay>
