# Handoff Summary

## Run Information
- **Command:** design-research
- **Spec:** claude-code-based-improvements
- **Run:** run-001
- **Output:** `.spec-workflow/specs/claude-code-based-improvements/design/DESIGN-RESEARCH.md`

## Subagent Results

| Subagent | Status | Summary |
|----------|--------|---------|
| codebase-pattern-scanner | pass | Identified 7 reusable patterns across CLI, registry, hooks, config, and test infrastructure |
| web-pattern-scout-cli | pass | Documented Cobra subcommand patterns, dual output modes, exit codes, and CI-ready JSON schema |
| web-pattern-scout-frontmatter | pass | Evaluated 4 frontmatter parsing options; recommended manual parser with schema-driven validation |
| risk-analyst | pass | Identified 7 risks across backward compatibility, migration, mirror scope, confidence calibration, iteration limits, docs, and testing |
| research-synthesizer | pass | Consolidated research into recommendations for all 7 requirements with code reuse, alternatives, and risk mitigations |

**All pass:** true

## Key Recommendations Summary

1. **REQ-001 (Frontmatter validation):** New `spw validate prompts` command with schema-driven validation, dual output modes (human/JSON), exit codes 0/1/2.
2. **REQ-002 (Mirror validation):** `--strict` flag ports bash mirror checks to Go, adds embedded asset comparison.
3. **REQ-003 (status.json):** Graduated validation — default tolerates missing optional fields, strict enforces all 5.
4. **REQ-004 (Audit gate):** New `[audit]` config section with `audit_min_confidence` threshold; findings below threshold become warnings.
5. **REQ-005 (Iteration limits):** Config-driven `max_revision_attempts` / `max_replan_attempts` with human escalation.
6. **REQ-006 (Docs):** Explicit doc-update sub-tasks per implementation task.
7. **REQ-007 (Tests):** Table-driven tests with minimum case matrices per validator.

## Unresolved Decisions (for design-draft)

1. **`allowed-tools` field format:** YAML array `[Read, Grep]` vs comma-separated string `"Read, Grep"`. Determines whether yaml.v3 dependency is needed.
2. **Per-command vs global audit threshold:** Start global, consider per-command overrides later.
3. **Iteration counter storage:** `_handoff.md` metadata vs separate `_iteration_state.json`.

## Subagent Reports
- `run-001/codebase-pattern-scanner/report.md`
- `run-001/web-pattern-scout-cli/report.md`
- `run-001/web-pattern-scout-frontmatter/report.md`
- `run-001/risk-analyst/report.md`
- `run-001/research-synthesizer/report.md`
