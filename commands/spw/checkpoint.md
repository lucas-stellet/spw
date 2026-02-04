---
name: spw:checkpoint
description: Quality gate between execution batches/waves
argument-hint: "<spec-name> [--scope batch|wave|phase]"
---

<objective>
Validate that the executed batch truly meets spec intent, code quality, and integration safety before moving forward.
</objective>

<checks>
1. Task state in `tasks.md`: consistency of `[ ]/[-]/[x]`.
2. Project checks: tests, lint, and typecheck.
3. Spec compliance review (requirements + design + task).
4. Code quality review.
5. Traceability: implemented changes are linked to `Requirements`.
</checks>

<output>
Generate `.spec-workflow/specs/<spec-name>/CHECKPOINT-REPORT.md` with:
- status: PASS | BLOCKED
- critical issues
- corrective actions
- recommended next step
</output>

<gate_rule>
If status is BLOCKED, do not proceed to the next batch/wave.
</gate_rule>
