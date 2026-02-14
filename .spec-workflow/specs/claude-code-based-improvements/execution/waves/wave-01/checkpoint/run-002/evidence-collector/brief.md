# Brief: evidence-collector

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- `.spec-workflow/specs/claude-code-based-improvements/execution/_implementation-logs/`
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/_wave-summary.json`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task

Collect evidence for the checkpoint audit of wave-01.

1. Verify task completion status from tasks.md
2. Check implementation logs in _implementation-logs/ directory - each completed task should have a corresponding log
3. Report on implementation log coverage

Write your findings to report.md and status.json.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/evidence-collector/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/evidence-collector/status.json

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
