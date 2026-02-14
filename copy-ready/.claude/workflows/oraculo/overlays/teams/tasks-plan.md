<purpose>
Agent Teams overlay for `oraculo:tasks-plan`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/oraculo.toml` (legacy fallback `.oraculo/oraculo.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)

When `enabled=true` and `tasks-plan` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- map planning roles to teammates (`task-decomposer`, `dependency-graph-builder`, `parallel-conflict-checker`, `test-policy-enforcer`, `tasks-writer`) (do not exceed `max_teammates`)
- each teammate must still write `brief.md`, `report.md`, `status.json` in the run dir
</agent_teams_policy>

<extensions_overlay>
Apply these additions to base extensions:
- In `<pre_pipeline>`: do not create a new run-id or team before resume decision.
- Before tasks-plan dispatch: create team and assign roles when enabled for phase.
- Keep mode precedence and dashboard markdown compatibility rules unchanged.
Dispatch mechanism comes from dispatch-pipeline.md shared policy.
</extensions_overlay>
