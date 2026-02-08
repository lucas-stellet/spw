<purpose>
Agent Teams overlay for `spw:tasks-plan`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/spw-config.toml` (fallback legado `.spw/spw-config.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `use_for_phases`

When `enabled=true` and `tasks-plan` is included in `use_for_phases`:
- create a team and set `teammate_mode`
- map planning roles to teammates (`task-decomposer`, `dependency-graph-builder`, `parallel-conflict-checker`, `test-policy-enforcer`, `tasks-writer`) (do not exceed `max_teammates`)
- each teammate must still write `brief.md`, `report.md`, `status.json` in the run dir
</agent_teams_policy>

<workflow_overlay>
Apply these additions to base workflow:
- In resume decision gate, do not create a new run-id or team before user decision.
- Before tasks-plan subagent dispatch, create team and assign roles when enabled for phase.
- Keep mode precedence and dashboard markdown compatibility rules unchanged.
</workflow_overlay>
