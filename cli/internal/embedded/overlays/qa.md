<purpose>
Agent Teams overlay for `spw:qa`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (legacy fallback `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)

When `enabled=true` and `qa` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- map QA roles to teammates (`qa-scope-analyst`, `browser-test-designer`, `api-test-designer`, `qa-plan-synthesizer`) (do not exceed `max_teammates`)
- each teammate must still write `brief.md`, `report.md`, `status.json`
</agent_teams_policy>

<extensions_overlay>
Apply these additions to base extensions:
- In `<pre_pipeline>`: keep user intent gate unchanged.
- Before `<pre_dispatch>`: create team and assign only active roles for selected tool (`playwright|bruno|hybrid`).
Dispatch mechanism comes from dispatch-pipeline.md shared policy.
</extensions_overlay>
