# Brief: execution-state-scout

## Inputs
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/_wave-summary.json
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/_wave-summary.json
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-002/_handoff.md

## Task
Read the tasks.md to identify tasks in wave-03 (tasks 6, 8, 9). Check wave summaries to confirm waves 1 and 2 are complete with passing checkpoints. Report the compact resume decision indicating tasks ready for execution in wave-03.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/execution-state-scout/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/execution-state-scout/status.json

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
