# Brief: spec-compliance-reviewer-5

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/design.md
- .spec-workflow/specs/claude-code-based-improvements/requirements.md
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer-5/report.md
- cli/internal/config/config.go
- cli/internal/config/config_test.go

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
<!-- Describe what this subagent must do -->

### Review Task 5: Extend config with AuditConfig struct and iteration limit fields

Verify implementation against:
- **Requirements:** REQ-004, REQ-005, REQ-007
- **Definition of Done from tasks.md:**
  - AuditConfig struct with AuditMinConfidence float64 added to Config
  - MaxRevisionAttempts and MaxReplanAttempts added to ExecutionConfig
  - Defaults() updated with sensible values (0.8, 3, 2)
  - Existing config tests still pass
  - New tests cover parsing and defaults for added fields
- **Test Plan:** Config parses new [audit] section with audit_min_confidence. Config parses max_revision_attempts and max_replan_attempts from [execution]. Defaults are correct (0.8, 3, 2). Override from TOML works. Missing section falls back to defaults.

**Compliance Check:**
1. Read the implementation files (config.go, config_test.go)
2. Verify AuditConfig struct exists with AuditMinConfidence field
3. Verify MaxRevisionAttempts and MaxReplanAttempts in ExecutionConfig
4. Verify Defaults() has correct values (0.8, 3, 2)
5. Run `go test ./cli/internal/config/ -v` to confirm tests pass

**Restrictions:**
- Do not change existing field names or defaults
- Follow HooksConfig pattern for the new struct

Report pass/blocked with specific evidence of compliance or gaps.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/spec-compliance-reviewer-5/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/spec-compliance-reviewer-5/status.json

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["skill-name"],
  "skills_missing": [],
  "model_override_reason": null
}
```
