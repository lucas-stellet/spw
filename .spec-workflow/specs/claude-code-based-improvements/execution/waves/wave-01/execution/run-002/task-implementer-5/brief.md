# Brief: task-implementer-5

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/design.md
- .spec-workflow/specs/claude-code-based-improvements/requirements.md

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
<!-- Describe what this subagent must do -->

### Task 5: Extend config with AuditConfig struct and iteration limit fields

**Files to modify:**
- `cli/internal/config/config.go`
- `cli/internal/config/config_test.go`

**Requirements:** REQ-004, REQ-005, REQ-007

**Definition of Done:**
- AuditConfig struct with AuditMinConfidence float64 added to Config
- MaxRevisionAttempts and MaxReplanAttempts added to ExecutionConfig
- Defaults() updated with sensible values (0.8, 3, 2)
- Existing config tests still pass
- New tests cover parsing and defaults for added fields

**Verification Command:**
```
go test ./cli/internal/config/ -v
```

**Restrictions:**
- Do not change existing field names or defaults
- Follow HooksConfig pattern for the new struct

**Test Plan (required):**
- Unit: Config parses new [audit] section with audit_min_confidence. Config parses max_revision_attempts and max_replan_attempts from [execution]. Defaults are correct (0.8, 3, 2). Override from TOML works. Missing section falls back to defaults.

Execute the task per the Definition of Done. Write tests to verify config parsing works correctly. Ensure all tests pass.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer-5/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer-5/status.json

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
