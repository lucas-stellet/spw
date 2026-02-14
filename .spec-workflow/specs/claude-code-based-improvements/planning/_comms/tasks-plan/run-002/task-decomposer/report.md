# Task Decomposer Report: Wave 2 Planning

## Wave 1 Completion Verification

All 3 Wave 1 tasks are confirmed complete per checkpoint report (PASS) and git history:

| Task | Files | Status |
|------|-------|--------|
| 1 - Validate package foundation | `cli/internal/validate/schema.go`, `schema_test.go` | Complete |
| 2 - Frontmatter validation | `cli/internal/validate/prompts.go`, `prompts_test.go` | Complete |
| 5 - Config extensions | `cli/internal/config/config.go`, `config_test.go` | Complete |

## Wave 2 Confirmed Task Set: Tasks 3, 4, 7

### Dependency Validation

| Task | Depends On | Dependency Status | Eligible |
|------|-----------|-------------------|----------|
| 3 - Mirror validation | Task 1 | Complete | Yes |
| 4 - Status.json validation | Task 1 | Complete | Yes |
| 7 - Iteration limits | Task 1 | Complete | Yes |

All three tasks depend only on Task 1 (validate package foundation), which is complete. No cross-dependencies exist among tasks 3, 4, and 7 -- they can run in parallel within the wave. The set fits within `max_wave_size=3`.

### Wave 2 Parallel Compatibility

Tasks.md declares:
- Task 3: `Can Run In Parallel With: 4, 7`
- Task 4: `Can Run In Parallel With: 3, 7`
- Task 7: `Can Run In Parallel With: 3, 4`

All three are mutually parallel-compatible. Each creates new files in `cli/internal/validate/` with no overlap.

## Scope Alignment Check Against Wave 1 Implementation

### Task 3 (Mirror Validation) -- No Adjustments Needed

The task calls for `ValidateMirrors(rootDir)` in `cli/internal/validate/mirror.go`. Wave 1 established:
- `Violation`, `ValidationResult`, `ValidationStats` types in `schema.go` -- directly usable by mirror validation
- `BuildValidationResult()` helper -- can be used for mirror results
- No conflicts with existing code

**Observation:** Task 3 references using `embedded.Workflows.ReadFile` for embedded asset comparison. The task implementor should verify that `cli/internal/embedded/embed.go` exports the necessary functions before implementation.

### Task 4 (Status.json Validation) -- Minor Scope Note

The task calls for `ValidateStatus(data, strict)` in `cli/internal/validate/status.go`. Wave 1 types are suitable:
- `Violation` struct has `File`, `Field`, `Rule`, `Message` fields -- works for status field-level errors
- `ValidateField()` in `schema.go` handles string, string_array, and enum types -- covers `status` (enum: pass/blocked), `summary` (string), `skills_used`/`skills_missing` (string arrays)

**Scope note:** The task definition says "Pure validation logic only" and "Do not modify dispatch_status.go" -- this is correct. Integration with dispatch is deferred to Task 9 (Wave 3). The `StatusValidationResult` mentioned in the DoD should extend or wrap `ValidationResult` with field-level error detail and a `warnings` list for default-mode optional field messages.

### Task 7 (Iteration Limits) -- No Adjustments Needed

The task calls for `CheckIterationLimit` in `cli/internal/validate/iteration.go` with `_iteration_state.json` persistence. Wave 1 config provides:
- `MaxRevisionAttempts = 3` and `MaxReplanAttempts = 2` defaults in `ExecutionConfig`
- Task 7 correctly states "receive max as parameter" (no direct config dependency)

The `_iteration_state.json` schema from design (`revision_count`, `replan_count`, `last_updated`) is straightforward. The `WAITING_FOR_HUMAN_DECISION` return is a string status, not a Go error -- implementor should define a result struct for this.

## Risks and Concerns

1. **Embedded package availability (Task 3):** The mirror task references `cli/internal/embedded/embed.go` for embedded asset comparison. If this package does not yet embed the necessary workflow files, the embedded comparison portion of Task 3 may need to be stubbed or conditionally skipped. Recommend the implementor check `embedded/embed.go` imports early.

2. **No existing test infrastructure for temp filesystem fixtures:** Tasks 3 and 4 both need filesystem-based test fixtures (mirror directories, status.json files). Wave 1 tests (prompts_test.go) established the pattern of using `t.TempDir()` which should be followed consistently.

3. **Wave 2 tasks are all pure validation logic:** None of them wire into CLI commands or modify existing dispatch code. This is intentional -- integration is Wave 3. No risk of breaking existing functionality.

## Recommendation

**Proceed with Wave 2 as planned: Tasks 3, 4, 7.** No task definition changes required. All dependencies are satisfied, all three fit within max_wave_size=3, and Wave 1 implementation aligns with what Wave 2 tasks expect to consume.
