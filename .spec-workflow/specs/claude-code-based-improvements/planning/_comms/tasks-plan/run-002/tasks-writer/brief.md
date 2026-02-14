# Brief: tasks-writer (next-wave synthesizer)

## Inputs
- Task decomposer report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/task-decomposer/report.md
- Dependency graph report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/dependency-graph-builder/report.md
- Parallel conflict checker report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/parallel-conflict-checker/report.md
- Test policy enforcer report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/test-policy-enforcer/report.md
- Current tasks.md: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md
- Tasks template: .spec-workflow/user-templates/tasks-template.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true
- effective mode: next-wave
- required_skills: effective-go (all Go implementation tasks)

## Task
You are the synthesizer. Read ALL previous subagent reports from disk and produce the updated tasks.md.

**Mode: next-wave.** You must:
1. Keep all completed tasks (1, 2, 5) with their [x] markers unchanged
2. Promote tasks 3, 4, 7 from "Deferred Tasks" to executable Wave 2 tasks with [ ] markers
3. Keep remaining tasks (6, 8, 9, 10, 11) in the "Deferred Tasks" section
4. Update the Wave Plan section to reflect Wave 2 is now executable
5. Update frontmatter: task_ids, updated_at

**CRITICAL Dashboard Markdown Profile Rules:**
- Checkbox markers only on real task lines: `- [ ] <id>. <description>`
- Use `-` as list marker (never `*`)
- Every task line starts with numeric ID
- Never use nested checkboxes inside metadata
- Metadata lines as regular bullets
- `Files:` must be single-line CSV
- `_Prompt:` must include `Role: ... | Task: ... | Restrictions: ... | Success: ...` and close with `_`
- `_Requirements:_` in underscore delimiters

**IMPORTANT:** Write the COMPLETE tasks.md content as your report. Do not abbreviate or truncate any task definitions. The report will be copied directly to tasks.md.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/tasks-writer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-002/tasks-writer/status.json

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
