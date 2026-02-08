<purpose>
Agent Teams overlay for `spw:qa`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (fallback legado `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `use_for_phases`

When `enabled=true` and `qa` is included in `use_for_phases`:
- create a team and set `teammate_mode`
- map QA roles to teammates (`qa-scope-analyst`, `browser-test-designer`, `api-test-designer`, `qa-plan-synthesizer`) (do not exceed `max_teammates`)
- each teammate must still write `brief.md`, `report.md`, `status.json`
</agent_teams_policy>

<workflow_overlay>
Apply these additions to base workflow:
- Keep one-question focus selection behavior unchanged when focus is missing.
- Before subagent dispatch, create team and assign only active roles for selected tool (`playwright|bruno|hybrid`).
</workflow_overlay>
