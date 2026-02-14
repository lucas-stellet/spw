# Traceability Mapping Report

## Requirement Traceability Matrix

| Requirement | Technical Decision | Components/Files | Test Strategy |
|---|---|---|---|
| REQ-001: Frontmatter validation | New `spw validate prompts` command group. Schema-driven `[]FieldRule` validation. `gopkg.in/yaml.v3` for frontmatter parsing (needed for `allowed-tools` array syntax). Exit codes: 0=pass, 1=violations, 2=error. JSON output via `--json` flag. | `cli/internal/validate/prompts.go` (logic), `cli/internal/validate/schema.go` (field rules), `cli/internal/cli/validate_cmd.go` (Cobra wiring). Reuses `Output()` from `cli/internal/tools/output.go`. | Table-driven: valid file, missing each required field, invalid model value, no frontmatter, empty file, extra fields. Golden file test for `--json` output format. |
| REQ-002: Mirror/asset validation | `--strict` flag on `spw validate prompts`. Two-phase: (1) filesystem mirror diff via content-hash, (2) embedded asset comparison via `embedded.Workflows.ReadFile`. Symlink target validation for overlays. | `cli/internal/validate/mirror.go` (mirror logic), `cli/internal/validate/mirror_test.go`. Ports logic from `scripts/validate-thin-orchestrator.sh`. Uses `embedded.Workflows` from `cli/internal/embedded/embed.go`. | Mirror: matching files, divergent content, missing mirror, extra files, broken symlinks. Embedded: matching vs divergent. Integration: end-to-end on real repo. |
| REQ-003: status.json enforcement | Extend `DispatchReadStatus` in `cli/internal/tools/dispatch_status.go` with graduated validation. Default mode: 2 core fields required, 3 extended optional with warnings. Strict mode (`--strict`): all 5 required. | `cli/internal/validate/status.go` (validation logic extracted), `cli/internal/tools/dispatch_status.go` (call site). New types: `StatusValidationResult{Valid bool, Errors []FieldError, Warnings []FieldError}`. | All 5 fields present, missing optional (default=warn, strict=error), wrong types, invalid status enum, null vs missing, backward compat with old 2-field files. |
| REQ-004: High-signal audit gate | New `[audit]` config section with `audit_min_confidence` (float64, default 0.8). New `AuditConfig` struct in config.go. Gate logic: blocked finding with `confidence < threshold` or `validated=false` becomes warning. Downgraded findings logged in `_handoff.md`. | `cli/internal/config/config.go` (add `AuditConfig` struct + `Audit` field to `Config`), `cli/internal/validate/audit.go` (gate logic), `cli/internal/tools/dispatch_status.go` (integrate confidence check). | Boundary tests: exactly at threshold, above, below. Downgrade behavior. Config default. Missing confidence field handling. |
| REQ-005: Iteration limits | Config-driven: `max_revision_attempts=3`, `max_replan_attempts=2` in `[execution]`. Counter stored in `_iteration_state.json` per run-dir. Limit exceeded -> `WAITING_FOR_HUMAN_DECISION`. | `cli/internal/config/config.go` (add fields to `ExecutionConfig`), `cli/internal/validate/iteration.go` (counter logic). New file: `_iteration_state.json` in run dirs. | Counter increment, threshold trigger, state file persistence, config override. |
| REQ-006: Doc updates | Enforcement via task structure (each implementation task includes doc sub-tasks). Future: `spw validate docs` via git diff analysis. | Affects: `README.md`, `AGENTS.md`, `docs/SPW-WORKFLOW.md`, `hooks/README.md`, `copy-ready/README.md`. | Manual checklist per task. |
| REQ-007: Regression tests | Table-driven Go tests following `dispatch_test.go` and `hook_test.go` patterns. Each new validator has a `_test.go` companion. | `cli/internal/validate/prompts_test.go`, `mirror_test.go`, `status_test.go`, `audit_test.go`, `iteration_test.go`. | Minimum test matrix defined per validator. Golden file tests for JSON output stability. |

## Key Architectural Decisions

### D-001: New `validate` package under `cli/internal/validate/`
**Justification:** Separates validation logic from existing `tools` and `hook` packages. Validation is a distinct concern (static analysis of artifacts) vs runtime dispatch tools or session hooks. Follows the existing pattern where each concern has its own package (`hook/`, `tools/`, `config/`, etc.).

### D-002: `gopkg.in/yaml.v3` for frontmatter parsing
**Justification:** Command frontmatter uses YAML syntax including arrays (e.g., `allowed-tools`). The manual `parseKeyValue` parser in `registry.go` does not handle arrays. yaml.v3 is a small, well-established Go dependency with no transitive deps. This is the only new dependency required.

### D-003: Graduated validation for status.json (default/strict modes)
**Justification:** Breaking change avoidance. Old runs have 2-field status.json files that must remain valid. Default mode warns on missing extended fields; strict mode (opt-in) requires all 5. This matches the requirements statement that "no version bump needed, just enforcing the full existing contract."

### D-004: `spw validate` as top-level command group
**Justification:** Discoverable (vs buried under `tools`). Follows conventions of shellcheck, golangci-lint. Room for future subcommands (`spw validate docs`, `spw validate config`). Registered in `root.go` alongside existing commands.

### D-005: Content-hash comparison for mirror validation
**Justification:** More reliable than byte-for-byte comparison when line endings or trailing whitespace differ across environments. Uses SHA-256 on normalized content. Replaces the bash `diff -rq` approach with something cross-platform.

### D-006: Iteration state file (`_iteration_state.json`)
**Justification:** Counters must survive across orchestrator re-dispatches. Filesystem is the only shared state mechanism in the file-first architecture. Stored alongside `_handoff.md` in the run directory.

## Dependency and Risk Notes

| Requirement | Dependencies | Risk |
|---|---|---|
| REQ-001 | `gopkg.in/yaml.v3` (new dep) | Low — well-established, no transitive deps |
| REQ-002 | `embedded.Workflows` internal package | Medium — path mapping between embedded FS and source filesystem needs explicit table |
| REQ-003 | Existing `dispatch_status.go` modification | Low — additive change, backward compatible |
| REQ-004 | Config struct extension + dispatch_status integration | Medium — confidence calibration guidelines need documentation to avoid subjective thresholds |
| REQ-005 | Workflow markdown changes (counter reads) | Low-Medium — workflows must check `_iteration_state.json`, adding a coupling point |
| REQ-006 | Human discipline | Low — checklist-based |
| REQ-007 | All above implementations | Low — table-driven tests are mechanical |

## Code Reuse Opportunities

| Existing Code | Reuse For |
|---|---|
| `cli/internal/tools/output.go` `Output()` | JSON/plain output formatting for `spw validate` |
| `cli/internal/config/config.go` `Config` struct + `Defaults()` + `Load()` | Adding `AuditConfig`, iteration limit fields |
| `cli/internal/config/config.go` `GetValue()` (reflection) | Auto-support for `spw tools config-get audit.audit_min_confidence` |
| `cli/internal/embedded/embed.go` `Workflows.ReadFile` + `AllWorkflowNames` | Embedded asset validation in `--strict` mode |
| `cli/internal/registry/registry.go` | Reference for frontmatter field patterns |
| `cli/internal/tools/dispatch_test.go` patterns | Test structure for new validators |
| `scripts/validate-thin-orchestrator.sh` mirror logic | Port to Go for `--strict` mode |
| `cli/internal/cli/root.go` command registration | Pattern for adding `newValidateCmd()` |
