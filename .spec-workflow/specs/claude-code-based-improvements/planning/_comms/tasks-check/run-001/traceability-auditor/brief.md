# Brief: traceability-auditor

## Inputs
- `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- `.spec-workflow/specs/claude-code-based-improvements/design.md`

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
Verify bidirectional traceability between tasks.md and requirements.md:

1. **Every task references at least one requirement.** Verify each task (1-11) has a `_Requirements:` line with valid REQ-NNN identifiers.
2. **Every requirement maps to at least one task.** Verify REQ-001 through REQ-007 each appear in at least one task's `_Requirements:` line.
3. **No orphan references.** Verify no task references a requirement ID not in requirements.md.
4. **Frontmatter consistency.** Verify the `requirements` list in tasks.md frontmatter matches the set of requirements actually referenced in task bodies.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/traceability-auditor/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-check/run-001/traceability-auditor/status.json

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
