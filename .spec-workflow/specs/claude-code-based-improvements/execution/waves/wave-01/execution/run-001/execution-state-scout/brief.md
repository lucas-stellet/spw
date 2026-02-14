# Brief: execution-state-scout

## Inputs
- Tasks: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Wave state: .spec-workflow/specs/claude-code-based-improvements/execution/waves/ (no prior waves exist)
- Git status: run `git status --porcelain` to check worktree cleanliness

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
You are the execution-state-scout. Inspect execution state for spec `claude-code-based-improvements`.

1. Read `tasks.md` to identify task statuses ([ ] pending, [-] in-progress, [x] completed).
2. Check if any prior wave execution state exists under `execution/waves/`.
3. Run `git status --porcelain` to check worktree cleanliness.
4. This is a fresh execution — no prior waves exist.

Write a compact handoff in report.md with:
- checkpoint_status: MISSING (no prior checkpoint)
- current_wave: wave-01
- in_progress_tasks: (list any [-] tasks)
- next_executable_tasks: ordered task IDs ready for Wave 1 (tasks 1, 5 can run in parallel, task 2 depends on 1)
- resume_action: start-next-task
- reason and evidence_paths (max 5)

Include a machine-readable JSON block with these fields.

Budget: max 12 bullets + one JSON block. No large excerpts.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-001/execution-state-scout/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-001/execution-state-scout/status.json

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
