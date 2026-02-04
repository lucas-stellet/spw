---
name: spw:tasks-check
description: Validate tasks.md quality (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<objective>
Validate whether `tasks.md` is ready for subagent execution.
</objective>

<checks>
1. Traceability:
   - every task references `Requirements`
   - every requirement is covered by at least one task
2. Dependencies:
   - no cycles
   - wave ordering is compatible with `Depends On`
3. Parallelism:
   - same-wave tasks do not conflict on critical files
4. Testing:
   - every task has `Test Plan` + `Verification Command`
   - exceptions include explicit justification
5. Definition of done:
   - objective completion criteria per task
</checks>

<output>
Generate `.spec-workflow/specs/<spec-name>/TASKS-CHECK.md` containing:
- PASS/BLOCKED
- findings by severity
- recommended `tasks.md` fixes
</output>
