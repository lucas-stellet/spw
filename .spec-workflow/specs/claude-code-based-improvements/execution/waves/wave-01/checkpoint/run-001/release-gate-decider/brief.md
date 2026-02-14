# Brief: release-gate-decider

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/evidence-collector/report.md`
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/traceability-judge/report.md`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task

You are the aggregator for the checkpoint audit. Read both auditor reports and produce the final PASS/BLOCKED decision with corrective actions.

1. **Read both auditor reports** from the filesystem:
   - evidence-collector report
   - traceability-judge report

2. **Synthesize findings**:
   - If ANY auditor reports BLOCKED, the final verdict MUST be BLOCKED
   - Summarize the key findings from both auditors

3. **Produce CHECKPOINT-REPORT.md** at:
   `.spec-workflow/specs/claude-code-based-improvements/execution/CHECKPOINT-REPORT.md`

   The report must include:
   - status: PASS | BLOCKED
   - critical issues
   - corrective actions
   - recommended next step
   - implementation log coverage by task ID

4. **Status**: Write status.json with your final decision.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/release-gate-decider/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/release-gate-decider/status.json

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
