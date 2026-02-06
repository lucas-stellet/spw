---
name: spw:design-draft
description: Subagent-driven design.md drafting from requirements + DESIGN-RESEARCH
argument-hint: "<spec-name>"
---

<objective>
Generate `.spec-workflow/specs/<spec-name>/design.md` with strong traceability back to requirements.
</objective>

<preconditions>
- `.spec-workflow/specs/<spec-name>/requirements.md` exists.
- `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md` exists (mandatory intermediate artifact).
- If `DESIGN-RESEARCH.md` is missing, stop BLOCKED and instruct:
  - `spw:design-research <spec-name>`
</preconditions>

<artifact_boundary>
Use only spec-local research artifacts:
- `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md`
- `.spec-workflow/specs/<spec-name>/research/*` (optional supporting notes)

Do not consume generated research from generic locations (for example `docs/*`).
</artifact_boundary>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills].load_mode` (`subagent-first|principal-first`)
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Backward compatibility:
- if `[skills.design].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Load modes:
- `subagent-first` (default): orchestrator does availability preflight only and
  delegates skill loading/use to subagents.
- `principal-first` (legacy): orchestrator loads required skills before dispatch.

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/SKILLS-DESIGN-DRAFT.md`
2. If `load_mode=subagent-first`, avoid loading full skill content in main context.
3. Require subagent outputs to explicitly mention skills used/missing.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<subagents>
- `traceability-mapper` (model: complex_reasoning)
  - Maps REQ-IDs to technical decisions, files, and tests.
- `design-writer` (model: implementation)
  - Produces design draft from mapped decisions.
- `design-critic` (model: complex_reasoning)
  - Runs consistency and completeness gate.
</subagents>

<diagram_policy>
For `design.md` output:
- Include at least one valid Mermaid diagram in `## Architecture` main flow.
- Prefer diagrams that represent real boundaries and data/control flow.
- If `mermaid-architecture` skill is available, use it for diagram type selection and syntax quality.
- Keep diagram terms consistent with requirement IDs and section vocabulary.
</diagram_policy>

<workflow>
1. Run design skills preflight (availability + load mode) and write `SKILLS-DESIGN-DRAFT.md`.
2. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md` (required)
   - `.spec-workflow/specs/<spec-name>/research/*` (if present)
   - `.spec-workflow/user-templates/design-template.md` (preferred)
   - fallback: `.spec-workflow/templates/design-template.md`
3. Dispatch `traceability-mapper`.
4. Dispatch `design-writer` using mapper output and apply `<diagram_policy>`.
5. Dispatch `design-critic`.
6. If critic returns BLOCKED:
   - revise with `design-writer`
   - re-run `design-critic`
7. Save to `.spec-workflow/specs/<spec-name>/design.md`.
8. Handle approval via MCP only:
   - call `spec-status`
   - resolve design status from:
     - `documents.design.approved`
     - `documents.design.status`
     - `approvals.design.status`
   - if approved, continue without re-requesting
   - if `needs-revision`/`changes-requested`/`rejected`, stop BLOCKED
   - if pending, stop with `WAITING_FOR_APPROVAL` and instruct UI approval + rerun
   - only if approval was never requested (missing/empty/unknown status):
     - call `request-approval` then `get-approval-status` once
     - if pending, stop with `WAITING_FOR_APPROVAL`
     - if needs revision, stop BLOCKED
   - never ask for approval in chat
</workflow>

<acceptance_criteria>
- [ ] Requirements traceability matrix exists.
- [ ] Technical decisions are justified.
- [ ] Test strategy is explicit.
- [ ] Architecture section contains at least one valid Mermaid diagram.
- [ ] Critic gate returned PASS before approval request.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/design.md`.
- Confirm approval request status for design.
- Recommend next command: `spw:tasks-plan <spec-name>` (use config defaults, or override with `--mode` / `--max-wave-size` when needed).

If blocked:
- Show precondition/critic/review failures with required fixes.
- Provide rerun command: `spw:design-draft <spec-name>`.
</completion_guidance>
