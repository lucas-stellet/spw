# Brief: evidence-collector

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- `.spec-workflow/specs/claude-code-based-improvements/design.md`
- `.spec-workflow/specs/claude-code-based-improvements/execution/_implementation-logs/`
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/_wave-summary.json`
- `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/` (most recent execution run)

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task

Collect evidence for the checkpoint audit of wave-01. You must:

1. **Task State**: Verify which tasks from tasks.md are marked complete ([x]) in the current wave scope.

2. **Implementation Logs**: Check `.spec-workflow/specs/claude-code-based-improvements/execution/_implementation-logs/` for implementation log entries. Map each completed task ID to its corresponding log entry (if any).

3. **Execution Evidence**: Examine the execution run directories to gather:
   - What code was changed/added
   - Any test outputs
   - Any lint/typecheck results

4. **Git Status**: Run `git status --porcelain` to verify worktree state.

5. **Report**: Produce a detailed report.md that includes:
   - Completed task IDs and their summaries
   - Implementation log coverage (which tasks have logs, which are missing)
   - Git status output
   - Any evidence of test/lint results

6. **Status**: Write status.json with your assessment (pass/blocked) based on whether implementation logs exist for all completed tasks.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/evidence-collector/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/checkpoint/run-001/evidence-collector/status.json

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
