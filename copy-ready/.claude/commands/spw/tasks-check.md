---
name: spw:tasks-check
description: Subagent-driven tasks.md validation (traceability, dependencies, tests)
argument-hint: "<spec-name>"
---

<objective>
Validate whether `tasks.md` is ready for subagent execution.
</objective>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills]`
- `[skills.design]`

Before validation, attempt to load required design/check skills.
If required skills are missing:
- `enforcement = "strict"` -> BLOCKED
- `enforcement = "advisory"` -> warn and continue
</skills_policy>

<subagents>
- `traceability-auditor` (model: complex_reasoning)
- `dag-validator` (model: implementation)
- `test-policy-auditor` (model: complex_reasoning)
- `decision-aggregator` (model: complex_reasoning)
</subagents>

<workflow>
1. Read `.spec-workflow/specs/<spec-name>/tasks.md` + requirements/design docs.
2. Dispatch in parallel:
   - `traceability-auditor`
   - `dag-validator`
   - `test-policy-auditor`
3. Dispatch `decision-aggregator` to produce PASS/BLOCKED decision.
4. Generate `.spec-workflow/specs/<spec-name>/TASKS-CHECK.md` containing:
   - PASS/BLOCKED
   - findings by severity
   - recommended fixes
</workflow>

<acceptance_criteria>
- [ ] Every task references at least one requirement.
- [ ] Every requirement maps to at least one task.
- [ ] DAG has no cycles and wave order is valid.
- [ ] Test policy gate is satisfied.
</acceptance_criteria>

<completion_guidance>
On PASS:
- Confirm output path: `.spec-workflow/specs/<spec-name>/TASKS-CHECK.md`.
- Recommend next command: `spw:exec <spec-name> --batch-size <N>`.
- Recommend running `/clear` before execution.

On BLOCKED:
- Show findings by severity and required fixes.
- Recommend fix path: update `tasks.md`, then rerun `spw:tasks-check <spec-name>`.
</completion_guidance>
