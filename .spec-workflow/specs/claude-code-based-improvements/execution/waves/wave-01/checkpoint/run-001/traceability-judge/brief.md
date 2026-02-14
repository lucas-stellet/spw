# Brief: traceability-judge

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- `.spec-workflow/specs/claude-code-based-improvements/design.md`
- `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/evidence-collector/report.md`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task

Verify requirements/design/tasks alignment for the completed wave-01 changes. You must:

1. **Read the evidence-collector report** to understand what was implemented.

2. **Verify traceability** between:
   - requirements.md (what was requested)
   - design.md (what was designed)
   - tasks.md (what was planned)
   - Implementation (what was delivered)

3. **Check alignment**:
   - Does each completed task (1, 2, 5) trace back to a requirement?
   - Does each completed task match its design specification?
   - Are there any gaps or deviations?

4. **Assess implementation log coverage**:
   - Per the implementation_log_gate policy: for every task marked [x] in the wave, there MUST be a corresponding implementation log entry in `.spec-workflow/specs/claude-code-based-improvements/execution/_implementation-logs/`
   - If implementation logs are missing, this is a BLOCKED condition

5. **Report**: Produce a detailed report with:
   - Traceability matrix (requirement → design → task → implementation)
   - Any gaps or deviations found
   - Implementation log coverage assessment

6. **Status**: Write status.json with PASS if all requirements are met and implementation logs exist, BLOCKED if there are critical gaps.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/traceability-judge/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/traceability-judge/status.json

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
