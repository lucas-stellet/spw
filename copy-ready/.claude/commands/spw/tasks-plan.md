---
name: spw:tasks-plan
description: Create tasks.md optimized for waves, parallelism, and per-task TDD
argument-hint: "<spec-name> [--max-wave-size 3] [--allow-no-test-exception true|false]"
---

<objective>
Generate `.spec-workflow/specs/<spec-name>/tasks.md` for predictable parallel execution.
</objective>

<rules>
- Each task must be self-contained.
- Each task must include tests and a verification command.
- No-test exception is allowed only with explicit justification.
- Plan dependencies to maximize safe parallelism per wave.
</rules>

<workflow>
1. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/user-templates/tasks-template.md` (preferred)
   - fallback: `.spec-workflow/templates/tasks-template.md`
2. Build a dependency DAG across tasks.
3. Assign `Wave: N` respecting `max-wave-size`.
4. For each task, fill:
   - `Depends On`
   - `Files`
   - `Test Plan`
   - `Verification Command`
   - `Requirements`
   - `No-Test Justification` (when needed)
5. Save to `.spec-workflow/specs/<spec-name>/tasks.md`.
6. Request approval.
</workflow>

<acceptance_criteria>
- [ ] All tasks have `Requirements`.
- [ ] All tasks have tests or a documented exception.
- [ ] Waves respect the configured limit.
</acceptance_criteria>
