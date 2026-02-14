<purpose>
Agent Teams overlay for `oraculo:exec`.
</purpose>

<agent_teams_policy>
Resolve Agent Teams config from `.spec-workflow/oraculo.toml` (legacy fallback `.oraculo/oraculo.toml`) `[agent_teams]`:
- `enabled` (default `false`)
- `teammate_mode` (default `"in-process"`)
- `max_teammates`
- `exclude_phases` (default `[]`)
- `require_delegate_mode` (default `true`)

When `enabled=true` and `exec` is NOT listed in `exclude_phases`:
- create a team and set `teammate_mode`
- if `require_delegate_mode=true`, enforce delegate mode for the lead agent
- map wave task IDs to the shared team task list and require teammates to claim tasks
- treat each teammate as the task subagent and still require file-first handoff files
</agent_teams_policy>

<extensions_overlay>
Apply these additions to base extensions:
- In `<pre_pipeline>`: create team and map/claim wave tasks when enabled.
- In strict mode, block when teams are enabled and delegate mode is required but not enforced.
Dispatch mechanism comes from dispatch-wave.md shared policy.
</extensions_overlay>
