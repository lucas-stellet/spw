# Wave 3 Task Decomposition Report

## Run: run-003
## Planning for: Wave 3 (tasks 6, 8, 9)
## Mode: next-wave

---

## Executive Summary

Wave 3 contains **3 tasks** that are ready for execution:
- Task 6: Audit confidence gate logic
- Task 8: CLI command wiring for `spw validate`
- Task 9: Integration with dispatch-read-status

All tasks fit within the `max_wave_size: 3` constraint.

---

## Completed Prerequisites

The following tasks have been completed in prior waves and satisfy all dependencies:

| Task | Wave | Status | Description |
|------|------|--------|-------------|
| 1 | 1 | Complete | Validate package foundation with schema types |
| 2 | 1 | Complete | Frontmatter validation with yaml.v3 |
| 3 | 2 | Complete | Mirror and embedded asset validation |
| 4 | 2 | Complete | Enhanced status.json validation |
| 5 | 1 | Complete | Config extensions (AuditConfig, iteration limits) |
| 7 | 2 | Complete | Iteration limit logic |

---

## Wave 3 Task Breakdown

### Task 6: Implement audit confidence gate logic

| Attribute | Value |
|-----------|-------|
| **Task ID** | 6 |
| **Title** | Audit confidence gate logic |
| **Dependencies** | Task 1 (schema.go), Task 5 (config.go) - both COMPLETE |
| **Can Run In Parallel With** | Task 8 |
| **Files to Modify** | `cli/internal/validate/audit.go`, `cli/internal/validate/audit_test.go` |
| **Requirements Coverage** | REQ-004 (high-signal gate), REQ-007 (test coverage) |
| **Test Strategy** | Table-driven unit tests covering boundary conditions |
| **TDD Mode** | inherit (off) |

**Detailed Test Matrix:**
- Confidence exactly at threshold (0.8) → stays blocked
- Confidence below threshold (0.79) → downgraded to warning
- Confidence above threshold (0.81) → stays blocked
- validated=false → always downgraded to warning
- Missing confidence field → treated as 0 (downgraded)
- Custom threshold from config parameter

**Implementation Notes:**
- Pure validation function: `ApplyAuditGate(statusJSON, minConfidence) AuditGateResult`
- Receives `minConfidence` as parameter (does not read config directly)
- Returns struct with `OriginalStatus`, `EffectiveStatus`, `WasDowngraded`, `Reason`
- All test cases must pass for completion

---

### Task 8: Wire Cobra validate command group with prompts subcommand

| Attribute | Value |
|-----------|-------|
| **Task ID** | 8 |
| **Title** | CLI command wiring for spw validate |
| **Dependencies** | Task 2 (prompts.go), Task 3 (mirror.go) - both COMPLETE |
| **Can Run In Parallel With** | Task 6 |
| **Files to Modify** | `cli/internal/cli/validate_cmd.go`, `cli/internal/cli/root.go` |
| **Requirements Coverage** | REQ-001 (validate prompts), REQ-002 (mirror --strict) |
| **Test Strategy** | Integration test: build CLI, run validate prompts, verify exit codes |
| **TDD Mode** | inherit (off) |

**Detailed Test Matrix:**
- `spw validate prompts` runs without error
- `spw validate prompts --json` produces valid JSON
- `spw validate prompts --strict` includes mirror validation
- Exit code 0: no violations found
- Exit code 1: violations found
- Exit code 2: error (file not found, etc.)

**Implementation Notes:**
- Follow existing patterns in `hook.go` and `tools.go`
- Create `newValidateCmd()` function
- Add `prompts` subcommand with `--json` and `--strict` flags
- Delegate to `validate.ValidatePrompts()` and `validate.ValidateMirrors()`
- Output: human-readable table by default, JSON with `--json`

---

### Task 9: Integrate enhanced status validation into dispatch-read-status

| Attribute | Value |
|-----------|-------|
| **Task ID** | 9 |
| **Title** | Dispatch-read-status integration |
| **Dependencies** | Task 4 (status.go), Task 6 (audit.go) - Task 6 is in same wave |
| **Can Run In Parallel With** | None (must run after Task 6) |
| **Files to Modify** | `cli/internal/tools/dispatch_status.go` |
| **Requirements Coverage** | REQ-003 (full status.json), REQ-004 (audit gate) |
| **Test Strategy** | Unit tests extending dispatch_test.go patterns |
| **TDD Mode** | inherit (off) |

**Detailed Test Matrix:**
- Default mode: existing.json valid (backward compatible 2-field status)
- Strict mode: all 5 fields required
- Audit context: confidence gate applied when audit fields present
- Extended dispatch patterns from dispatch_test.go

**Implementation Notes:**
- Extend existing `DispatchReadStatus` function
- Add `--strict` flag support (or pass strict parameter)
- Call `validate.ValidateStatus()` for graduated validation
- Apply audit confidence gate when audit fields detected
- CRITICAL: Must not break existing default behavior

---

## Dependency Analysis

```
Wave 3 Execution Order:

[Task 6: audit.go]     [Task 8: validate_cmd.go]
        |                         |
        v                         v (parallel with 6)
        |                         |
        +-----> Task 9 <----------+
        (depends on 6)
```

**Dependency Graph:**
- Task 6 has no blocking dependencies (tasks 1, 5 complete)
- Task 8 has no blocking dependencies (tasks 2, 3 complete)
- Task 9 depends on Task 6 (audit gate needed before integration)

---

## Parallel Execution Opportunities

| Task Pair | Parallel? | Rationale |
|-----------|-----------|-----------|
| 6 + 8 | YES | No shared dependencies; both can start immediately |
| 6 + 9 | NO | Task 9 depends on Task 6 |
| 8 + 9 | NO | Task 9 depends on Task 6 |

**Recommended Execution Order:**
1. Start Task 6 and Task 8 in parallel (both have all prerequisites complete)
2. After Task 6 completes, start Task 9

---

## Requirements Coverage Matrix

| Requirement | Task 6 | Task 8 | Task 9 |
|-------------|--------|--------|--------|
| REQ-001 (frontmatter validation CLI) | | X | |
| REQ-002 (mirror validation --strict) | | X | |
| REQ-003 (full status.json contract) | | | X |
| REQ-004 (audit confidence gate) | X | | X |
| REQ-007 (test coverage) | X | X | X |

---

## Files Summary

| File | New/Modify | Task(s) |
|------|------------|---------|
| `cli/internal/validate/audit.go` | New | 6 |
| `cli/internal/validate/audit_test.go` | New | 6 |
| `cli/internal/cli/validate_cmd.go` | New | 8 |
| `cli/internal/cli/root.go` | Modify | 8 |
| `cli/internal/tools/dispatch_status.go` | Modify | 9 |

---

## Test Strategy Summary

All tasks require tests per `require_test_per_task: true` constraint:

| Task | Test Type | Coverage |
|------|-----------|----------|
| 6 | Unit | Boundary at 0.8, below, above; validated=false; missing field |
| 8 | Integration | CLI build, exit codes, JSON validity |
| 9 | Unit | Default/strict modes, audit context, backward compatibility |

---

## Constraints Compliance

| Constraint | Value | Compliance |
|------------|-------|------------|
| max_wave_size | 3 | Wave 3 has exactly 3 tasks |
| tdd_default | off | All tasks use `inherit` |
| require_test_per_task | true | All tasks include test strategy |

---

## Summary

Wave 3 is ready for execution with 3 tasks:
- **Task 6** and **Task 8** can run in parallel
- **Task 9** must run after Task 6 completes
- All requirements from REQ-001, REQ-002, REQ-003, REQ-004, REQ-007 are covered
- No blocking issues identified
