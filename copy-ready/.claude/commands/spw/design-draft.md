---
name: spw:design-draft
description: Create or update design.md from requirements + DESIGN-RESEARCH
argument-hint: "<spec-name>"
---

<objective>
Generate `.spec-workflow/specs/<spec-name>/design.md` with strong traceability back to requirements.
</objective>

<workflow>
1. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md` (if present)
   - `.spec-workflow/user-templates/design-template.md` (preferred)
   - fallback: `.spec-workflow/templates/design-template.md`
2. Fill design with focus on:
   - mapping `REQ-ID -> technical decision`
   - architecture and boundaries
   - existing code reuse
   - test strategy (unit, integration, e2e)
3. Save to `.spec-workflow/specs/<spec-name>/design.md`.
4. Request approval (standard spec-workflow flow).
</workflow>

<acceptance_criteria>
- [ ] Requirements traceability matrix exists.
- [ ] Technical decisions are justified.
- [ ] Test strategy is explicit.
</acceptance_criteria>
