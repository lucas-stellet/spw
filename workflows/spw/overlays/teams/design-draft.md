<purpose>
Agent Teams overlay for `spw:design-draft`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (fallback legado `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `use_for_phases`

When `enabled=true` and `design-draft` is included in `use_for_phases`:
- create a team and set `teammate_mode`
- map subagent roles to teammates (`traceability-mapper`, `design-writer`, `design-critic`) (do not exceed `max_teammates`)
- each teammate must still produce expected outputs for downstream approval gate
</agent_teams_policy>

<workflow_overlay>
Apply these additions to base workflow:
- Create team and assign draft roles before subagent dispatch when enabled for phase.
- Keep approval reconciliation and markdown profile gates unchanged.
</workflow_overlay>
