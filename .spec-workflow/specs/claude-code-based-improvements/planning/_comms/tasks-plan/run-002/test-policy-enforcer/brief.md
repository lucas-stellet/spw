# Brief: test-policy-enforcer

## Inputs
- Task decomposer report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/task-decomposer/report.md
- Current tasks.md: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
Enforce test policy for Wave 2 tasks (3, 4, 7). Verify each task has:

1. A test plan with specific test cases
2. A verification command
3. TDD mode setting (inherit/required/skip)
4. If TDD: skip, a valid No-TDD Justification

Wave 2 tasks from tasks.md:
- Task 3 (mirror validation): TDD: inherit, has unit test plan and verification command
- Task 4 (status.json validation): TDD: inherit, has unit test plan and verification command
- Task 7 (iteration limits): TDD: inherit, has unit test plan and verification command

Validate that:
- All test plans cover the minimum test matrix from the design document
- Verification commands are valid Go test commands
- No task is missing required test coverage
- TDD mode is compatible with tdd_default=off (inherit means off)

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/test-policy-enforcer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/test-policy-enforcer/status.json

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
