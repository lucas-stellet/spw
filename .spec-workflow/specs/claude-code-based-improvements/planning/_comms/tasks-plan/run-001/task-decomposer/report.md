# Task Decomposition Report

## Task List

### Task 1: Create validate package foundation (schema types + FieldRule)
- **Description:** Create `cli/internal/validate/schema.go` with shared types: `FieldRule`, `Violation`, `ValidationResult`, `ValidationStats`. These are the building blocks for all validators.
- **Files:** cli/internal/validate/schema.go (create)
- **Requirements:** REQ-001, REQ-007
- **Dependencies:** none
- **Test plan:** Unit tests for FieldRule validation helper functions (type checking, enum matching).
- **Verification:** `go build ./cli/...`

### Task 2: Implement frontmatter validation logic
- **Description:** Create `cli/internal/validate/prompts.go` with `ValidatePrompts(dir string) (*ValidationResult, error)` that scans `commands/spw/*.md`, extracts YAML frontmatter via yaml.v3, validates against the FieldRule schema (name, description, argument-hint, allowed-tools, model). Returns structured result with violations and stats.
- **Files:** cli/internal/validate/prompts.go (create), cli/internal/validate/prompts_test.go (create)
- **Requirements:** REQ-001, REQ-007
- **Dependencies:** Task 1
- **Test plan:**
  - Unit: Valid frontmatter, missing each required field, invalid model enum, no frontmatter, empty file, extra unknown fields, malformed YAML.
  - Golden file: JSON output format stability.
- **Verification:** `go test ./cli/internal/validate/ -run TestValidatePrompts -v`

### Task 3: Implement mirror and embedded asset validation
- **Description:** Create `cli/internal/validate/mirror.go` with `ValidateMirrors(rootDir string) (*ValidationResult, error)` that compares source directories with copy-ready counterparts using SHA-256 content hashing. Also validates overlay symlink targets and embedded asset consistency.
- **Files:** cli/internal/validate/mirror.go (create), cli/internal/validate/mirror_test.go (create)
- **Requirements:** REQ-002, REQ-007
- **Dependencies:** Task 1
- **Test plan:**
  - Unit: Matching content, divergent content, missing mirror file, extra files, broken symlinks, symlink target validation.
  - Integration: Embedded vs filesystem comparison.
- **Verification:** `go test ./cli/internal/validate/ -run TestValidateMirrors -v`

### Task 4: Implement enhanced status.json validation
- **Description:** Create `cli/internal/validate/status.go` with `ValidateStatus(data []byte, strict bool) (*StatusValidationResult, error)` that validates all 5 status.json fields. Default mode: 2 required + 3 optional (warn). Strict mode: all 5 required with type checking.
- **Files:** cli/internal/validate/status.go (create), cli/internal/validate/status_test.go (create)
- **Requirements:** REQ-003, REQ-007
- **Dependencies:** Task 1
- **Test plan:**
  - Unit: All 5 present, missing optional (default=warn, strict=error), wrong types, invalid status enum, null vs missing distinction.
- **Verification:** `go test ./cli/internal/validate/ -run TestValidateStatus -v`

### Task 5: Extend config with AuditConfig and iteration limit fields
- **Description:** Add `AuditConfig` struct with `AuditMinConfidence float64` to `Config`. Add `MaxRevisionAttempts int` and `MaxReplanAttempts int` to `ExecutionConfig`. Update `Defaults()` with sensible values (0.8, 3, 2). Update config_test.go.
- **Files:** cli/internal/config/config.go (modify), cli/internal/config/config_test.go (modify)
- **Requirements:** REQ-004, REQ-005, REQ-007
- **Dependencies:** none
- **Test plan:**
  - Unit: Config parsing with new sections, default values, override from TOML.
- **Verification:** `go test ./cli/internal/config/ -v`

### Task 6: Implement audit confidence gate logic
- **Description:** Create `cli/internal/validate/audit.go` with `ApplyAuditGate(status map[string]any, minConfidence float64) *AuditGateResult` that checks confidence >= threshold and validated==true. Returns whether finding stays blocked or is downgraded to warning.
- **Files:** cli/internal/validate/audit.go (create), cli/internal/validate/audit_test.go (create)
- **Requirements:** REQ-004, REQ-007
- **Dependencies:** Task 1, Task 5
- **Test plan:**
  - Unit: Confidence boundary (exactly 0.8, 0.79, 0.81), downgrade behavior, validated=false, missing confidence field.
- **Verification:** `go test ./cli/internal/validate/ -run TestAuditGate -v`

### Task 7: Implement iteration limit logic
- **Description:** Create `cli/internal/validate/iteration.go` with `CheckIterationLimit(runDir string, counterType string, maxAttempts int) (*IterationResult, error)` that reads/creates `_iteration_state.json`, increments counter, checks against limit. Returns whether to continue or halt with WAITING_FOR_HUMAN_DECISION.
- **Files:** cli/internal/validate/iteration.go (create), cli/internal/validate/iteration_test.go (create)
- **Requirements:** REQ-005, REQ-007
- **Dependencies:** Task 1
- **Test plan:**
  - Unit: Counter creation, increment, threshold trigger, state persistence across calls, config override.
- **Verification:** `go test ./cli/internal/validate/ -run TestIterationLimit -v`

### Task 8: Wire Cobra validate command
- **Description:** Create `cli/internal/cli/validate_cmd.go` with `newValidateCmd()` returning a Cobra command group. Subcommand `prompts` with `--json` and `--strict` flags. Register in root.go. Exit codes: 0=pass, 1=violations, 2=error.
- **Files:** cli/internal/cli/validate_cmd.go (create), cli/internal/cli/root.go (modify)
- **Requirements:** REQ-001, REQ-002
- **Dependencies:** Task 2, Task 3
- **Test plan:**
  - Integration: End-to-end `spw validate prompts` on actual repository. `--json` produces valid JSON. `--strict` includes mirror results. Exit codes correct.
- **Verification:** `go build -o /tmp/spw ./cli/cmd/spw && /tmp/spw validate prompts --json`

### Task 9: Integrate enhanced status validation into dispatch-read-status
- **Description:** Modify `cli/internal/tools/dispatch_status.go` to call `validate.ValidateStatus()` for graduated validation. Add `--strict` flag support. Integrate audit confidence gate when audit context is detected.
- **Files:** cli/internal/tools/dispatch_status.go (modify)
- **Requirements:** REQ-003, REQ-004
- **Dependencies:** Task 4, Task 6
- **Test plan:**
  - Unit: Extended via existing dispatch_test.go patterns. Default mode backward compatible. Strict mode enforces all fields.
- **Verification:** `go test ./cli/internal/tools/ -run TestDispatchReadStatus -v`

### Task 10: Update documentation
- **Description:** Update all required documentation files to reflect new CLI commands, config sections, and validation behavior. Must be in the same patch as behavior changes per REQ-006.
- **Files:** README.md (modify), AGENTS.md (modify), docs/SPW-WORKFLOW.md (modify), hooks/README.md (modify), copy-ready/README.md (modify)
- **Requirements:** REQ-006
- **Dependencies:** Task 8, Task 9
- **Test plan:**
  - Manual review: all new CLI commands documented, config sections documented, validation behavior described.
  - No-test exception: documentation-only task.
- **Verification:** `grep -l "validate prompts" README.md docs/SPW-WORKFLOW.md`

### Task 11: Update mirror copies and embedded assets
- **Description:** Sync all modified source files to their copy-ready counterparts. Update embedded assets. Run mirror validation to confirm integrity.
- **Files:** copy-ready/.spec-workflow/spw-config.toml (modify), copy-ready/.claude/commands/spw/* (sync), copy-ready/.claude/workflows/spw/* (sync)
- **Requirements:** REQ-002, REQ-006
- **Dependencies:** Task 8, Task 10
- **Test plan:**
  - Integration: `spw validate prompts --strict` passes (self-validation).
  - Script: `scripts/validate-thin-orchestrator.sh` passes.
- **Verification:** `scripts/validate-thin-orchestrator.sh && /tmp/spw validate prompts --strict`

## Summary
- 11 total tasks identified
- Tasks 1, 5 are foundation (no dependencies)
- Tasks 2, 3, 4, 6, 7 are core validators (depend on foundation)
- Tasks 8, 9 are integration/wiring (depend on validators)
- Tasks 10, 11 are finalization (depend on integration)
