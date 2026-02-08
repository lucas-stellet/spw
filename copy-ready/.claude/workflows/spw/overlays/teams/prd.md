<purpose>
Agent Teams overlay for `spw:prd`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (fallback legado `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `use_for_phases`

When `enabled=true` and `prd` is included in `use_for_phases`:
- create a team and set `teammate_mode`
- map PRD roles (`source-reader-web`, `source-reader-mcp`, `codebase-impact-scanner`, `requirements-structurer`, `prd-editor`, `prd-critic`) to teammates (do not exceed `max_teammates`)
- apply the same mapping in revision protocol roles
- each teammate must still write `brief.md`, `report.md`, `status.json` in the run dir
</agent_teams_policy>

<workflow_overlay>
Apply these additions to base workflow:
- In resume decision gate, do not create a new run-id or team before user decision.
- Before PRD subagent dispatch, create team and assign roles when enabled for phase.
- Before revision protocol subagent dispatch, reuse/create team and assign roles when enabled for phase.
</workflow_overlay>
