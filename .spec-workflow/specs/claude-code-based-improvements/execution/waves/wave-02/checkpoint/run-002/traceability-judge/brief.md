# Brief: traceability-judge

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- Spec directory: `.spec-workflow/specs/claude-code-based-improvements`
- Tasks file: `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- Requirements file: `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- Design file: `.spec-workflow/specs/claude-code-based-improvements/design.md`
- Wave summary: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/_wave-summary.json`
- Evidence collector report: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-002/evidence-collector/report.md`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
You are the **traceability-judge** for a checkpoint audit. Your role is to verify requirements/design/tasks alignment for delivered changes.

### Your Responsibilities

1. **Requirements Alignment**
   - Verify each completed task (3, 4, 7) fulfills its declared requirements from tasks.md
   - Check REQ-002, REQ-003, REQ-005 coverage as listed in task definitions

2. **Design Alignment**
   - Verify implementation matches design.md specifications
   - For task 3 (mirror validation): SHA-256 hashing, embedded asset comparison, symlink target validation
   - For task 4 (status validation): graduated enforcement (default vs strict), 5-field validation
   - For task 7 (iteration limits): _iteration_state.json persistence, counter logic, threshold triggers

3. **Tasks Alignment**
   - Verify all task definition criteria are met
   - Check "Definition of Done" sections from tasks.md

4. **Evidence Review**
   - Read the evidence-collector report
   - Verify build and test outputs are captured
   - Verify implementation log coverage

### Output Format

Write your findings to `report.md` in your working directory. Include:

```
## Traceability Summary
- Task 3: REQ-002, REQ-007 fulfillment: [pass/fail]
- Task 4: REQ-003, REQ-007 fulfillment: [pass/fail]
- Task 7: REQ-005, REQ-007 fulfillment: [pass/fail]

## Design Alignment
[For each task, verify implementation matches design contract]

## Tasks Alignment
[For each task, verify Definition of Done criteria met]

## Issues Found
[Any alignment issues or gaps]
```

Also write `status.json` with either:
- `{"status": "pass", "summary": "Traceability verified"}`
- `{"status": "blocked", "summary": "Alignment issue: <description>"}`

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-002/traceability-judge/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-002/traceability-judge/status.json

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
