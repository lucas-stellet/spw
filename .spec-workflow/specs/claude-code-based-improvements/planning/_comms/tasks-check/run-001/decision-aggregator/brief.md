# Brief: decision-aggregator

## Inputs (auditor report paths — read these directly from disk)
- `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/traceability-auditor/report.md`
- `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/traceability-auditor/status.json`
- `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/dag-validator/report.md`
- `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/dag-validator/status.json`
- `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/test-policy-auditor/report.md`
- `.spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/test-policy-auditor/status.json`

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
You are the decision aggregator. Read ALL auditor reports and status files listed above directly from disk.

Synthesize a final PASS/BLOCKED decision:
- If ANY auditor status is `blocked`, the final verdict MUST be BLOCKED.
- If all auditors pass, the final verdict is PASS.

Your report must include:
1. **Verdict**: PASS or BLOCKED
2. **Auditor Summary Table**: One row per auditor with status and one-line summary.
3. **Findings by Severity**: Group any findings as Critical (blocking), Warning (non-blocking advisory), or Info.
4. **Recommended Fixes**: If BLOCKED, list required fixes. If PASS, list any advisories.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/decision-aggregator/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/decision-aggregator/status.json

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": [],
  "skills_missing": [],
  "model_override_reason": null
}
```
