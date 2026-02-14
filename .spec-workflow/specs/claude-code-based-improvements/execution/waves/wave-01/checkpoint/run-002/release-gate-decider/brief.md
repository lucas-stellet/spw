# Brief: release-gate-decider

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/evidence-collector/report.md`
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/traceability-judge/report.md`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task

You are the aggregator. Read both auditor reports and produce the final PASS/BLOCKED decision.

1. Read both auditor reports from the filesystem
2. If ANY auditor reports BLOCKED, final verdict MUST be BLOCKED
3. Produce CHECKPOINT-REPORT.md at `.spec-workflow/specs/claude-code-based-improvements/execution/CHECKPOINT-REPORT.md` with:
   - status: PASS | BLOCKED
   - critical issues
   - corrective actions
   - recommended next step
   - implementation log coverage by task ID

Write status.json with your final decision.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/release-gate-decider/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-002/release-gate-decider/status.json

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
