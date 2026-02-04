---
name: spw:design-draft
description: Subagent-driven design.md drafting from requirements + DESIGN-RESEARCH
argument-hint: "<spec-name>"
---

<objective>
Generate `.spec-workflow/specs/<spec-name>/design.md` with strong traceability back to requirements.
</objective>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<subagents>
- `traceability-mapper` (model: complex_reasoning)
  - Maps REQ-IDs to technical decisions, files, and tests.
- `design-writer` (model: implementation)
  - Produces design draft from mapped decisions.
- `design-critic` (model: complex_reasoning)
  - Runs consistency and completeness gate.
</subagents>

<workflow>
1. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md` (if present)
   - `.spec-workflow/user-templates/design-template.md` (preferred)
   - fallback: `.spec-workflow/templates/design-template.md`
2. Dispatch `traceability-mapper`.
3. Dispatch `design-writer` using mapper output.
4. Dispatch `design-critic`.
5. If critic returns BLOCKED:
   - revise with `design-writer`
   - re-run `design-critic`
6. Save to `.spec-workflow/specs/<spec-name>/design.md`.
7. Handle approval via MCP only:
   - call `spec-status`
   - if already approved, continue without re-requesting
   - if not approved, call `request-approval` then `get-approval-status` once
   - if pending, stop with `WAITING_FOR_APPROVAL` and instruct UI approval + rerun
   - if rejected/changes-requested, stop BLOCKED
   - never ask for approval in chat
</workflow>

<acceptance_criteria>
- [ ] Requirements traceability matrix exists.
- [ ] Technical decisions are justified.
- [ ] Test strategy is explicit.
- [ ] Critic gate returned PASS before approval request.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/design.md`.
- Confirm approval request status for design.
- Recommend next command: `spw:tasks-plan <spec-name> --max-wave-size <N>`.

If blocked:
- Show critic/review failures with required fixes.
- Provide rerun command: `spw:design-draft <spec-name>`.
</completion_guidance>
