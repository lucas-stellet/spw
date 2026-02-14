# Brief: evidence-collector

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- Spec directory: `.spec-workflow/specs/claude-code-based-improvements`
- Tasks file: `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- Requirements file: `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- Design file: `.spec-workflow/specs/claude-code-based-improvements/design.md`
- Wave summary: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/_wave-summary.json`
- Execution run: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/`
- Implementation logs: `.spec-workflow/specs/claude-code-based-improvements/execution/_implementation-logs/`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
You are the **evidence-collector** for a checkpoint audit. Your role is to gather comprehensive evidence about the executed wave's deliverables.

### Your Responsibilities

1. **Task State Verification**
   - Read the wave summary to identify completed tasks (tasks 3, 4, 7 in wave-02)
   - Verify each completed task against tasks.md

2. **Test/Lint/Typecheck Outputs**
   - Run: `go build ./cli/...` (from cli directory)
   - Run: `go test ./cli/internal/validate/ -run "TestValidateMirrors|TestValidateStatus|TestIterationLimit" -v`
   - Capture all outputs

3. **Implementation Log Coverage**
   - Check `.spec-workflow/specs/claude-code-based-improvements/execution/_implementation-logs/` for task-3.md, task-4.md, task-7.md
   - Map each completed task to its implementation log entry
   - Identify any missing implementation logs

4. **Git Status**
   - Run: `git status --porcelain`
   - Include uncommitted changes in your report

5. **Code Quality**
   - Verify the Go code compiles without errors
   - Check test coverage for implemented features

### Output Format

Write your findings to `report.md` in your working directory. Include:

```
## Evidence Summary
- Tasks completed: [list task IDs]
- Implementation logs found: [list]
- Implementation logs missing: [list]
- Build status: pass/fail
- Test status: pass/fail
- Git status: clean/dirty

## Detailed Findings
[Each task with evidence of completion]

## Critical Issues
[Any blocking issues found]
```

Also write `status.json` with either:
- `{"status": "pass", "summary": "Evidence collected successfully"}`
- `{"status": "blocked", "summary": "Critical issue: <description>"}`

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-002/evidence-collector/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-002/evidence-collector/status.json

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
