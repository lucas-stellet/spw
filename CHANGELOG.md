# Changelog

All notable changes to SPW are documented in this file.

## [2.0.0] - 2025

### Added
- Thin-dispatch architecture: commands as thin wrappers (max 60 lines) delegating to full workflow orchestration.
- Three dispatch categories: Pipeline, Audit, and Wave Execution with shared policies.
- Phase-based spec directory structure replacing flat `_generated/` and `_agent-comms/` dumps.
- 3-phase QA chain: `spw:qa` (plan) -> `spw:qa-check` (validate) -> `spw:qa-exec` (execute).
- Agent Teams via symlink overlays (base + noop/teams toggle).
- Model routing: haiku for scouting, opus for reasoning, sonnet for implementation.
- `spw` CLI with `install`, `update`, `doctor`, `status`, and `skills` commands.
- Rolling-wave planning strategy as default (alternative: all-at-once).
- Post-mortem memory system with indexed lessons for design/planning phases.
- File-first subagent communication with `brief.md`, `report.md`, and `status.json`.
- Execution state scout for compact resume state before broad reads.
- PR review optimization via `.gitattributes` linguist-generated markers.
- Node.js enforcement hooks: statusline, guard-paths, guard-user-prompt, guard-stop.
- SessionStart hook for tasks template sync based on TDD config.
- Skill system with subagent-first loading (TDD, QA validation planning, mermaid-architecture).
- Unfinished-run detection with explicit user decision (continue or restart).
- Approval reconciliation via MCP for gated commands.
- Dashboard markdown compatibility profile for `tasks.md`.
- YAML frontmatter metadata in spec templates.

## [1.0.0] - 2025

### Added
- Initial spec-workflow integration with Claude Code commands.
- Basic command entry points: `spw:prd`, `spw:plan`, `spw:exec`, `spw:checkpoint`, `spw:status`.
- `copy-ready/` installer for project setup.
