<purpose>
Agent Teams overlay for `spw:prd`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (legacy fallback `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)

When `enabled=true` and `prd` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- map PRD roles (`source-reader-web`, `source-reader-mcp`, `codebase-impact-scanner`, `requirements-structurer`, `prd-editor`, `prd-critic`) to teammates (do not exceed `max_teammates`)
- apply the same mapping in revision protocol roles
- each teammate must still write `brief.md`, `report.md`, `status.json` in the run dir
</agent_teams_policy>

<extensions_overlay>
Apply these additions to base extensions:
- In `<pre_pipeline>`: do not create a new run-id or team before user decision.
- Before `<pre_dispatch>` steps: create team and assign roles when enabled for phase.
- In revision protocol: reuse/create team and assign revision roles.
Dispatch mechanism comes from dispatch-pipeline.md shared policy.
</extensions_overlay>
