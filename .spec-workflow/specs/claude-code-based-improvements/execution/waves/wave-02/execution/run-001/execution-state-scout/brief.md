# Brief: execution-state-scout

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- Tasks: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Previous Wave Summary: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/_wave-summary.json
- Previous Checkpoint: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/release-gate-decider/status.json
- Wave Plan: Review tasks.md Wave Plan section for wave-02 executable tasks

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
You are the execution-state-scout for wave-02 of spec "claude-code-based-improvements".

Wave 1 completed with checkpoint PASS. Wave 2 contains tasks 3, 4, 7.

Your job is to:
1. Read the tasks.md to identify wave-02 executable tasks (3, 4, 7)
2. Verify wave-01 checkpoint passed
3. Return a compact handoff with:
   - `current_wave`: wave-02
   - `next_executable_tasks`: [3, 4, 7] in dependency order
   - `resume_action`: start-next-task (since checkpoint passed)
   - `reason`: brief explanation

Output budget: max 12 bullets plus one machine-readable JSON block.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/execution-state-scout/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/execution-state-scout/status.json

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
