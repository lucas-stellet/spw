<purpose>
Agent Teams overlay for `spw:design-draft`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (fallback legado `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)

When `enabled=true` and `design-draft` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- map subagent roles to teammates (`traceability-mapper`, `design-writer`, `design-critic`) (do not exceed `max_teammates`)
- each teammate must still produce expected outputs for downstream approval gate
</agent_teams_policy>

<extensions_overlay>
Apply these additions to base extensions:
- In `<pre_pipeline>`: create team and assign draft roles before dispatch when enabled for phase.
- Keep approval reconciliation and markdown profile gates unchanged.
Dispatch mechanism comes from dispatch-pipeline.md shared policy.
</extensions_overlay>
