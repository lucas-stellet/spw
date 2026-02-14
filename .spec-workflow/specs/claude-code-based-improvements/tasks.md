---
spw:
  schema: 1
  spec: "claude-code-based-improvements"
  doc: "tasks"
  status: "draft"
  source: "spw:tasks-plan"
  updated_at: "2026-02-14"
  inputs:
    - ".spec-workflow/specs/claude-code-based-improvements/requirements.md"
    - ".spec-workflow/specs/claude-code-based-improvements/design.md"
  requirements:
    - "REQ-001"
    - "REQ-002"
    - "REQ-003"
    - "REQ-004"
    - "REQ-005"
    - "REQ-006"
    - "REQ-007"
  task_ids: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
  test_required: true
  risk: "medium"
  open_questions: []
---

# Tasks Document (TDD OFF by default)

## Execution Constraints
- max_tasks_per_wave: 3
- require_test_per_task: true
- allow_no_test_exception: true
- tdd_default: off
- required_skills: effective-go (all Go implementation tasks)

## Wave Plan
- Wave 1 (complete): Foundation — validate package types + frontmatter validator + config extensions
- Wave 2 (complete): Core validators — mirror, status.json, iteration limits
- Wave 3 (current): Integration — audit gate, CLI wiring, dispatch integration
- Wave 4 (deferred): Finalization — documentation, mirror sync

---

- [x] 1. Create validate package foundation with shared schema types
  - Wave: 1
  - Depends On: none
  - Can Run In Parallel With: 5
  - Files: cli/internal/validate/schema.go
  - _Requirements: REQ-001, REQ-007_
  - TDD: inherit
  - Test Plan:
    - Unit: FieldRule validation helpers — type checking for string, string_array, enum. Enum match/mismatch. Required vs optional field logic.
  - Verification Command:
    - go build ./cli/...
  - Definition of Done:
    - FieldRule, Violation, ValidationResult, ValidationStats types exported
    - Helper functions for field type validation work correctly
    - Package compiles with no errors
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Create cli/internal/validate/schema.go with shared types (FieldRule, Violation, ValidationResult, ValidationStats) and helper functions for field validation per design schema | Restrictions: No dependencies on tools, hook, or cli packages. Pure validation types only. | Success: Package compiles, types match design contract, helpers pass unit tests_

- [x] 2. Implement frontmatter validation logic with yaml.v3
  - Wave: 1
  - Depends On: 1
  - Can Run In Parallel With: 5
  - Files: cli/internal/validate/prompts.go, cli/internal/validate/prompts_test.go
  - _Requirements: REQ-001, REQ-007_
  - TDD: inherit
  - Test Plan:
    - Unit: Valid frontmatter passes. Missing each required field (name, description, argument-hint, allowed-tools, model) produces violation. Invalid model enum value. No frontmatter delimiter. Empty file. Malformed YAML. Extra unknown fields tolerated. Golden file test for JSON output format.
  - Verification Command:
    - go test ./cli/internal/validate/ -run TestValidatePrompts -v
  - Definition of Done:
    - ValidatePrompts(dir) scans commands/spw/*.md and returns structured ValidationResult
    - All 5 required fields enforced (name, description, argument-hint, allowed-tools, model)
    - yaml.v3 used for frontmatter parsing
    - Table-driven tests cover all cases in test matrix
    - Golden file test for JSON stability
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Implement ValidatePrompts in cli/internal/validate/prompts.go using yaml.v3 for frontmatter extraction and FieldRule schema validation per design. Write comprehensive table-driven tests in prompts_test.go. | Restrictions: Do not wire Cobra command (Task 8). Do not modify registry.go parseKeyValue. Validate package must not import tools/hook/cli. | Success: All test cases pass, function returns correct ValidationResult for valid and invalid inputs_

- [x] 5. Extend config with AuditConfig struct and iteration limit fields
  - Wave: 1
  - Depends On: none
  - Can Run In Parallel With: 1, 2
  - Files: cli/internal/config/config.go, cli/internal/config/config_test.go
  - _Requirements: REQ-004, REQ-005, REQ-007_
  - TDD: inherit
  - Test Plan:
    - Unit: Config parses new [audit] section with audit_min_confidence. Config parses max_revision_attempts and max_replan_attempts from [execution]. Defaults are correct (0.8, 3, 2). Override from TOML works. Missing section falls back to defaults.
  - Verification Command:
    - go test ./cli/internal/config/ -v
  - Definition of Done:
    - AuditConfig struct with AuditMinConfidence float64 added to Config
    - MaxRevisionAttempts and MaxReplanAttempts added to ExecutionConfig
    - Defaults() updated with sensible values (0.8, 3, 2)
    - Existing config tests still pass
    - New tests cover parsing and defaults for added fields
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Add AuditConfig struct to config.go with AuditMinConfidence field. Add MaxRevisionAttempts and MaxReplanAttempts to ExecutionConfig. Update Defaults(). Write tests in config_test.go. | Restrictions: Do not change existing field names or defaults. Follow HooksConfig pattern for the new struct. | Success: go test ./cli/internal/config/ passes, new fields parse correctly from TOML_

- [x] 3. Implement mirror and embedded asset validation
  - Wave: 2
  - Depends On: 1
  - Can Run In Parallel With: 4, 7
  - Files: cli/internal/validate/mirror.go, cli/internal/validate/mirror_test.go
  - _Requirements: REQ-002, REQ-007_
  - TDD: inherit
  - Test Plan:
    - Unit: Matching content hash. Divergent content detected. Missing mirror file. Extra files in mirror. Broken symlinks. Symlink target validation (noop.md or teams/*.md). Embedded vs filesystem comparison.
  - Verification Command:
    - go test ./cli/internal/validate/ -run TestValidateMirrors -v
  - Definition of Done:
    - ValidateMirrors(rootDir) compares source vs copy-ready using SHA-256
    - Overlay symlink targets validated
    - Embedded asset comparison via embedded.Workflows.ReadFile
    - All test cases pass
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Implement ValidateMirrors in cli/internal/validate/mirror.go. Port mirror logic from validate-thin-orchestrator.sh to Go with SHA-256 content hashing. Add embedded asset comparison. | Restrictions: Follow mapping table from design. Do not modify embedded/embed.go. | Success: Mirror divergences detected, symlink targets validated, all tests pass_

- [x] 4. Implement enhanced status.json validation with graduated enforcement
  - Wave: 2
  - Depends On: 1
  - Can Run In Parallel With: 3, 7
  - Files: cli/internal/validate/status.go, cli/internal/validate/status_test.go
  - _Requirements: REQ-003, REQ-007_
  - TDD: inherit
  - Test Plan:
    - Unit: All 5 fields present and valid. Missing optional fields in default mode (warn). Missing optional fields in strict mode (error). Wrong types per field. Invalid status enum. Null vs missing distinction for model_override_reason. Empty skills arrays valid.
  - Verification Command:
    - go test ./cli/internal/validate/ -run TestValidateStatus -v
  - Definition of Done:
    - ValidateStatus(data, strict) validates all 5 fields with graduated enforcement
    - Default mode: 2 required + 3 optional (warn)
    - Strict mode: all 5 required
    - StatusValidationResult with field-level errors
    - All test cases pass
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Implement ValidateStatus in cli/internal/validate/status.go with graduated validation (default vs strict mode) per design contract. | Restrictions: Do not modify dispatch_status.go (Task 9). Pure validation logic only. | Success: Default mode backward compatible, strict mode enforces all fields, all tests pass_

- [x] 7. Implement iteration limit logic with state persistence
  - Wave: 2
  - Depends On: 1
  - Can Run In Parallel With: 3, 4
  - Files: cli/internal/validate/iteration.go, cli/internal/validate/iteration_test.go
  - _Requirements: REQ-005, REQ-007_
  - TDD: inherit
  - Test Plan:
    - Unit: State file creation on first call. Counter increment on subsequent calls. Threshold trigger (count > max). State persistence (read back after write). Config override for limits. WAITING_FOR_HUMAN_DECISION returned when exceeded.
  - Verification Command:
    - go test ./cli/internal/validate/ -run TestIterationLimit -v
  - Definition of Done:
    - CheckIterationLimit reads/creates _iteration_state.json
    - Counters increment correctly
    - Limit exceeded returns WAITING_FOR_HUMAN_DECISION
    - State file persisted in run directory
    - All test cases pass
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Implement CheckIterationLimit in cli/internal/validate/iteration.go with _iteration_state.json persistence per design iteration flow. | Restrictions: Use run-dir for state storage. Do not depend on config package (receive max as parameter). | Success: Counter logic correct, state persists, threshold triggers human decision halt_

- [ ] 6. Implement audit confidence gate logic
  - Wave: 3
  - Depends On: 1, 5
  - Can Run In Parallel With: 8
  - Files: cli/internal/validate/audit.go, cli/internal/validate/audit_test.go
  - _Requirements: REQ-004, REQ-007_
  - TDD: inherit
  - Test Plan:
    - Unit: Confidence exactly at threshold (0.8) stays blocked. Below threshold (0.79) downgraded to warning. Above threshold (0.81) stays blocked. validated=false always downgraded. Missing confidence field treated as 0 (downgraded). Custom threshold from config.
  - Verification Command:
    - go test ./cli/internal/validate/ -run TestAuditGate -v
  - Definition of Done:
    - ApplyAuditGate checks confidence and validated fields
    - Findings below threshold downgraded to warnings
    - AuditGateResult includes original and effective status
    - Boundary tests pass at exactly audit_min_confidence
    - All test cases pass
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Implement ApplyAuditGate in cli/internal/validate/audit.go per design audit gate flow. | Restrictions: Receive minConfidence as parameter. Do not read config directly. | Success: Boundary behavior correct at exactly 0.8, downgrade logic works, all tests pass_

- [ ] 8. Wire Cobra validate command group with prompts subcommand
  - Wave: 3
  - Depends On: 2, 3
  - Can Run In Parallel With: 6
  - Files: cli/internal/cli/validate_cmd.go, cli/internal/cli/root.go
  - _Requirements: REQ-001, REQ-002_
  - TDD: inherit
  - Test Plan:
    - Integration: spw validate prompts on actual repository completes without error. --json produces valid JSON. --strict includes mirror validation results. Exit code 0 when no violations. Exit code 1 when violations found.
  - Verification Command:
    - go build -o /tmp/spw ./cli/cmd/spw && /tmp/spw validate prompts --json
  - Definition of Done:
    - newValidateCmd() registered in root.go
    - prompts subcommand with --json and --strict flags
    - Delegates to validate.ValidatePrompts and validate.ValidateMirrors
    - Exit codes: 0=pass, 1=violations, 2=error
    - Produces human-readable table by default, JSON with --json
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Create cli/internal/cli/validate_cmd.go with Cobra command group. Wire prompts subcommand with --json and --strict flags. Register in root.go. | Restrictions: Follow existing hook.go and tools.go patterns. Command only delegates to validate package. | Success: spw validate prompts runs end-to-end, exit codes correct, JSON output valid_

- [ ] 9. Integrate enhanced status validation into dispatch-read-status
  - Wave: 3
  - Depends On: 4, 6
  - Can Run In Parallel With: none
  - Files: cli/internal/tools/dispatch_status.go
  - _Requirements: REQ-003, REQ-004_
  - TDD: inherit
  - Test Plan:
    - Unit: Default mode backward compatible (existing 2-field status.json still valid). Strict mode enforces all 5 fields. Audit context applies confidence gate. Extended dispatch_test.go patterns.
  - Verification Command:
    - go test ./cli/internal/tools/ -run TestDispatchReadStatus -v
  - Definition of Done:
    - DispatchReadStatus calls validate.ValidateStatus for graduated validation
    - --strict flag supported
    - Audit confidence gate applied when audit context detected
    - Existing behavior unchanged in default mode
    - Tests pass
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Extend DispatchReadStatus in dispatch_status.go to use validate.ValidateStatus for graduated validation. Add audit confidence gate integration. | Restrictions: Do not break existing default behavior. 2-field status.json must remain valid in default mode. | Success: Backward compatible, strict mode works, audit gate applied, tests pass_

---

## Deferred Tasks (not executable until next planning wave)

- [ ] 10. Update documentation for new CLI commands and config sections
  - Wave: 4 (deferred)
  - Depends On: 8, 9
  - Can Run In Parallel With: none
  - Files: README.md, AGENTS.md, docs/SPW-WORKFLOW.md, hooks/README.md, copy-ready/README.md
  - _Requirements: REQ-006_
  - TDD: skip
  - No-TDD Justification:
    - Reason: Documentation-only task with no Go code changes
    - Alternative validation: grep verification that new commands and config sections are documented
  - Verification Command:
    - grep -l "validate prompts" README.md docs/SPW-WORKFLOW.md copy-ready/README.md
  - Definition of Done:
    - spw validate prompts command documented with flags and examples
    - audit config section documented with audit_min_confidence
    - max_revision_attempts and max_replan_attempts documented in [execution]
    - Enhanced status.json contract documented
    - All 5 doc files updated
  - _Prompt: Role: Technical writer | Task: Update README.md, AGENTS.md, docs/SPW-WORKFLOW.md, hooks/README.md, copy-ready/README.md with documentation for spw validate prompts, [audit] config, iteration limits, and enhanced status.json contract. | Restrictions: Only add documentation for implemented features. Do not change code. Match existing doc style. | Success: All new features documented in all 5 files, grep verification passes_

- [ ] 11. Sync mirror copies and validate embedded asset integrity
  - Wave: 4 (deferred)
  - Depends On: 8, 10
  - Can Run In Parallel With: none
  - Files: copy-ready/.spec-workflow/spw-config.toml, copy-ready/.claude/commands/spw/*, copy-ready/.claude/workflows/spw/*
  - _Requirements: REQ-002, REQ-006_
  - TDD: inherit
  - Test Plan:
    - Integration: spw validate prompts --strict self-validates with no divergences. scripts/validate-thin-orchestrator.sh passes.
  - Verification Command:
    - scripts/validate-thin-orchestrator.sh && /tmp/spw validate prompts --strict
  - Definition of Done:
    - All modified source files synced to copy-ready counterparts
    - Embedded assets updated if needed
    - validate-thin-orchestrator.sh passes
    - spw validate prompts --strict passes (self-validation)
  - _Prompt: Role: Go developer | Skills: effective-go | Task: Sync all modified source files to copy-ready mirrors. Update embedded assets. Run validation scripts. | Restrictions: Follow mirror mapping table from CLAUDE.md. Do not modify source files in this task. | Success: Both validation scripts pass with zero divergences_
