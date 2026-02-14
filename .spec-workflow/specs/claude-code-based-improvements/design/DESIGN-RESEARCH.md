# Design Research: claude-code-based-improvements

## Overview

This document consolidates research findings for moving SPW validations and guardrails from prompt markdown into Go CLI commands. It covers seven functional requirements (REQ-001 through REQ-007) with technical recommendations, alternatives, and risk mitigations.

---

## 1. Primary Recommendations

### REQ-001: Mandatory Frontmatter Validation (`spw validate prompts`)

**Recommendation: New `validate` command group with `prompts` subcommand.**

- **Location:** `cli/internal/cli/validate_cmd.go` (Cobra wiring), `cli/internal/validate/prompts.go` (logic).
- **Parser:** Manual YAML frontmatter parser reusing `parseKeyValue` pattern from `registry.go`. No new dependencies needed since command frontmatter is flat key-value pairs. If `allowed-tools` requires YAML array syntax, add `gopkg.in/yaml.v3` (small, well-established dependency).
- **Schema:** Schema-driven validation with a `[]FieldRule` slice defining required fields, types, and allowed values. Adding new fields is a one-line change.
- **Required fields:** `name`, `description`, `argument-hint`, `allowed-tools`, `model`.
- **Output modes:** Human-readable table (default) and JSON (`--json` flag). JSON schema: `{ ok, summary, violations[], stats }`.
- **Exit codes:** 0 = pass, 1 = violations found, 2 = runtime error.
- **File scanning:** Read all `commands/spw/*.md` files. Use `go:embed` or filesystem reads depending on context (embedded for self-validation, filesystem for project validation).
- **Registration:** Add `cmd.AddCommand(newValidateCmd())` in `root.go`.

### REQ-002: Mirror and Embedded Asset Validation (`--strict`)

**Recommendation: Two-phase implementation within `spw validate prompts --strict`.**

- **Phase 1 — Filesystem mirrors:** Port the bash logic from `validate-thin-orchestrator.sh` to Go. Compare `commands/spw/` vs `copy-ready/.claude/commands/spw/`, `workflows/spw/` vs `copy-ready/.claude/workflows/spw/`. Use `os.ReadDir` + content hash comparison.
- **Phase 2 — Embedded asset validation:** Compare `embedded.Workflows.ReadFile` vs filesystem `workflows/spw/*.md` content. Define an explicit mapping table for path translation between embedded FS and source filesystem.
- **Symlink validation:** Verify overlay symlinks point to `../noop.md` or `../teams/<name>.md`.
- **Output:** Divergent pairs listed with source and mirror paths.

### REQ-003: Full status.json Contract Enforcement

**Recommendation: Extend `DispatchReadStatus` with graduated validation.**

- **Default mode (current behavior preserved):** Validate `status` and `summary` as required. Treat `skills_used`, `skills_missing`, `model_override_reason` as optional — warn if absent, validate types if present.
- **Strict mode:** All 5 fields required. Activated via `--strict` flag on `dispatch-read-status` or by a new config key.
- **Type validation:** `status` = enum("pass", "blocked"), `summary` = non-empty string, `skills_used` = string array, `skills_missing` = string array, `model_override_reason` = string or null.
- **Error reporting:** Return `valid=false` with `errors[]` containing field name, expected type, actual value.
- **Update `dispatch-setup`:** The brief skeleton already documents all 5 fields. No change needed there, only in validation.
- **Backward compatibility:** Old status.json files (with only 2 fields) remain valid in default mode. This avoids breaking existing runs.

### REQ-004: High-Signal Gate for Audits

**Recommendation: New `[audit]` config section + confidence-aware status.json.**

- **Config addition:**
  ```toml
  [audit]
  audit_min_confidence = 0.8
  ```
- **Config struct addition:** `AuditConfig` with `AuditMinConfidence float64` field in `config.go`.
- **status.json extension:** Audit subagents include `confidence` (float, 0.0-1.0) and `validated` (bool) fields alongside their findings.
- **Gate logic:** In `dispatch-read-status` (or a new `dispatch-read-audit-status`), a finding with `status=blocked` is downgraded to warning if `confidence < audit_min_confidence` or `validated=false`.
- **Logging:** All downgraded findings logged in `_handoff.md` under a `## Downgraded Findings` section.
- **Documentation:** Include examples of confidence calibration: syntax error = 1.0, potential logic issue = 0.7, style concern = 0.4.

### REQ-005: Iteration Limits

**Recommendation: Config-driven limits with human decision escalation.**

- **Config addition:**
  ```toml
  [execution]
  max_revision_attempts = 3
  max_replan_attempts = 2
  ```
- **Config struct addition:** Add `MaxRevisionAttempts int` and `MaxReplanAttempts int` to `ExecutionConfig`.
- **Enforcement location:** In workflow orchestration logic. When a revision cycle starts, increment a counter. When counter exceeds limit, set status to `WAITING_FOR_HUMAN_DECISION` with explicit options.
- **Counter storage:** In `_handoff.md` metadata or in a new `_iteration_state.json` file in the run directory.
- **Note:** This requirement is "Should" priority, meaning it can be deferred if the implementation is complex. The core mechanism is simple (counter + threshold), but integration with all revision-capable commands needs care.

### REQ-006: Documentation Update in Same Patch

**Recommendation: Checklist enforcement, not automated validation.**

- **Approach:** Each implementation task that changes behavior includes explicit doc-update sub-tasks for: `README.md`, `AGENTS.md`, `docs/SPW-WORKFLOW.md`, `hooks/README.md`, `copy-ready/README.md`.
- **Future enhancement:** A `spw validate docs` sub-check that verifies doc files were modified when Go source files change (via git diff analysis). This is out of scope for initial implementation.

### REQ-007: Regression Test Coverage

**Recommendation: Table-driven tests following existing patterns.**

- **Frontmatter validation tests:** Valid file, missing each required field, invalid `model` value, no frontmatter, empty file, extra unknown fields.
- **status.json validation tests:** All fields present, missing optional fields, wrong types, invalid `status` value, null vs missing distinction.
- **Audit confidence tests:** At threshold boundary (exactly `audit_min_confidence`), above, below. Downgrade behavior verification.
- **Mirror validation tests:** Matching files, divergent files, missing mirror, extra files.
- **Golden file tests:** JSON output format stability for `--json` mode.
- **Integration test:** End-to-end `spw validate prompts` on the actual repository (as a smoke test).

---

## 2. Existing Code Reuse

| Component | Reuse For | Location |
|-----------|-----------|----------|
| `parseKeyValue` pattern | YAML frontmatter field parsing | `cli/internal/registry/registry.go:128` |
| `Output(result, summary, raw)` | JSON vs plain output | `cli/internal/tools/output.go` |
| `Config` struct + `Defaults()` | Adding `[audit]` section | `cli/internal/config/config.go` |
| `Config.GetValue` (reflection) | Automatic config-get support | `cli/internal/config/config.go:241` |
| `embedded.Workflows.ReadFile` | Embedded asset comparison | `cli/internal/embedded/embed.go` |
| `AllWorkflowNames` slice | Iterating all 13 commands | `cli/internal/embedded/embed.go:109` |
| Table-driven test helpers | All new tests | `cli/internal/tools/dispatch_test.go` |
| `validate-thin-orchestrator.sh` | Mirror validation logic to port | `scripts/validate-thin-orchestrator.sh` |
| `HooksConfig` guard toggle pattern | Per-validation toggles | `cli/internal/config/config.go:95` |

---

## 3. Alternatives and Trade-offs

### Frontmatter Parser Choice

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| Manual parser (reuse `parseKeyValue`) | Zero deps, fast, simple | No YAML arrays, no nested types | Best if `allowed-tools` is comma-separated string |
| `gopkg.in/yaml.v3` | Full YAML, type-safe structs | New dependency | Best if `allowed-tools` needs array syntax |
| `goldmark-frontmatter` | Feature-rich | Heavy dependency, overkill | Reject |
| `adrg/frontmatter` | Lightweight | Still external dep | Consider if yaml.v3 too heavy |

**Trade-off decision:** If `allowed-tools` is defined as a YAML list (e.g., `allowed-tools: [Read, Grep, Bash]`), use yaml.v3. If it can be a quoted string (e.g., `allowed-tools: "Read, Grep, Bash"`), use the manual parser. The design phase should make this decision.

### Validation Command Location

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| Top-level `spw validate` | Discoverable, clean namespace | Adds to root command count | **Recommended** |
| Under `spw tools validate-*` | Groups with existing tools | Less discoverable, verbose | Reject |
| Standalone `spw-validate` binary | Separate concern | Build complexity | Reject |

### status.json Validation Strategy

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| Graduated (warn/strict) | Backward compatible | More complex | **Recommended** |
| Breaking (always strict) | Simpler code | Breaks old runs | Reject |
| Schema version field | Clean evolution | Over-engineering for 5 fields | Consider for future |

---

## 4. Technical Risks and Mitigations

| Risk | Level | Mitigation |
|------|-------|------------|
| Old status.json files break with strict validation | Medium | Default mode tolerates missing optional fields; strict mode opt-in |
| Wrong `allowed-tools` lists in frontmatter migration | Low-Medium | Derive from workflow analysis; include cross-check validation |
| Embedded asset path mapping errors | Medium | Define explicit mapping table; test with actual embedded files |
| Inconsistent audit confidence calibration | Medium-High | Document calibration guidelines; log all suppressed findings; conservative default (0.8) |
| Iteration limits too aggressive | Low | Configurable with sensible defaults (3 revision, 2 replan) |
| Doc updates forgotten | Low | Explicit sub-tasks per implementation task |
| Test coverage gaps | Low | Define minimum test case matrix per validator |

---

## 5. New Package Structure

```
cli/internal/
  validate/
    prompts.go         # Frontmatter validation logic
    prompts_test.go    # Tests for frontmatter validation
    mirror.go          # Mirror/embedded asset validation (--strict)
    mirror_test.go     # Tests for mirror validation
    status.go          # Enhanced status.json validation
    status_test.go     # Tests for status validation
    schema.go          # Shared schema/field-rule types
  cli/
    validate_cmd.go    # Cobra command wiring for `spw validate`
  config/
    config.go          # Add AuditConfig struct
```

---

## 6. References and Patterns to Adopt

- **Cobra subcommand groups:** Follow existing `tools`, `hook`, `spec` patterns in `cli/internal/cli/`.
- **Exit code convention:** 0/1/2 (pass/violations/error) matching shellcheck, golangci-lint.
- **JSON output schema:** `{ ok, summary, violations[], stats }` matching REQ-001 acceptance criteria.
- **Schema-driven validation:** `[]FieldRule` with name, required, type, enum — extensible for future fields.
- **Table-driven Go tests:** Match `dispatch_test.go` and `hook_test.go` patterns.
- **Config evolution:** Follow `HooksConfig` pattern for new `AuditConfig` section.
