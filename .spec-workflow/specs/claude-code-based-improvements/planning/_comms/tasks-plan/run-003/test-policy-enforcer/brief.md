# Brief: test-policy-enforcer

## Context
- **Spec:** claude-code-based-improvements
- **Run:** run-003
- **Mode:** next-wave
- **Planning for:** Wave 3 (tasks 6, 8, 9)

## Inputs
- Task decomposition: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/task-decomposer/report.md`

## Instructions
Enforce test-per-task policy for Wave 3:

### Wave 3 Tasks
- Task 6: Audit confidence gate logic - needs unit tests
- Task 8: CLI command wiring - needs integration tests
- Task 9: Dispatch-read-status integration - needs unit tests

### Constraints
- require_test_per_task: true
- allow_no_test_exception: true
- tdd_default: off

## Output
Write to: `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-003/test-policy-enforcer/report.md`

Verify each task has:
- Valid test strategy
- Valid verification command
- Complete test matrix

Then write status.json:
```json
{"status": "pass", "summary": "All 3 Wave 3 tasks have complete test plans"}
```
