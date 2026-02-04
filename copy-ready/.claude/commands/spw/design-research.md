---
name: spw:design-research
description: Technical research to prepare design.md from approved requirements
argument-hint: "<spec-name> [--focus <topic>] [--web-depth low|medium|high]"
---

<objective>
Generate architecture and implementation research inputs for the spec design.
Output: `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md`.
</objective>

<preconditions>
- The spec has `requirements.md` and it is approved.
- Also read steering docs when present (`product.md`, `tech.md`, `structure.md`).
</preconditions>

<workflow>
1. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/steering/*.md` (if present)
2. Analyze the codebase for:
   - existing patterns
   - reusable components/utilities
   - integration points and risks
3. Run external web research for relevant libraries/patterns.
4. For Elixir/Phoenix projects, explicitly check:
   - context boundaries (Ecto)
   - LiveView/Phoenix conventions
   - real process needs (OTP)
5. Write `DESIGN-RESEARCH.md` with:
   - primary recommendations
   - alternatives and trade-offs
   - references/patterns to adopt
   - technical risks and mitigations
</workflow>

<acceptance_criteria>
- [ ] Every relevant functional requirement has at least one technical recommendation.
- [ ] Existing-code reuse section is included.
- [ ] Risks and recommended decisions section is included.
</acceptance_criteria>
