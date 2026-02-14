# Brief: traceability-judge

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- `.spec-workflow/specs/claude-code-based-improvements/design.md`
- `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- `.spec-workflow/specs/claude-code-based-improvements/execution/_implementation-logs/`
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/evidence-collector/report.md`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task

Verify requirements/design/tasks alignment and implementation log coverage.

1. Check that each completed task (1, 2, 5) has a corresponding implementation log in _implementation-logs/
2. Verify traceability between requirements, design, tasks, and implementation
3. Verify that implementation logs contain proper artifacts (task ID, summary, files, key artifacts)

Write your findings to report.md and status.json (pass if all logs exist and are complete, blocked otherwise).

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/traceability-judge/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/traceability-judge/status.json

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
