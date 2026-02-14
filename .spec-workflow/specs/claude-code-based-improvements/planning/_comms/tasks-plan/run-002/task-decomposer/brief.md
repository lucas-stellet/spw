# Brief: task-decomposer (next-wave mode)

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md
- Current tasks.md: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Checkpoint report: .spec-workflow/specs/claude-code-based-improvements/execution/CHECKPOINT-REPORT.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true
- effective mode: next-wave (rolling-wave, tasks.md exists, wave-01 complete)

## Task
This is Wave 2 planning in next-wave mode. Wave 1 (tasks 1, 2, 5) is complete and checkpointed.

Analyze the deferred backlog from tasks.md and confirm which tasks are eligible for Wave 2:
- A task is eligible if all dependencies are satisfied (completed in Wave 1)
- Must fit within max_wave_size=3

Current plan groups tasks 3, 4, 7 as Wave 2. Validate:
- Task 3 (mirror validation) depends on 1 (complete) - check eligibility
- Task 4 (status.json validation) depends on 1 (complete) - check eligibility
- Task 7 (iteration limits) depends on 1 (complete) - check eligibility

Check whether existing task definitions remain accurate given Wave 1 implementation. Read the actual implemented code to verify scope alignment. Flag any adjustments needed.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/task-decomposer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/task-decomposer/status.json

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
