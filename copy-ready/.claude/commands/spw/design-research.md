---
name: spw:design-research
description: Subagent-driven technical research to prepare design.md
argument-hint: "<spec-name> [--focus <topic>] [--web-depth low|medium|high]"
---

<objective>
Generate architecture and implementation research inputs for the spec design.
Output: `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md`.
</objective>

<artifact_boundary>
All research outputs must stay inside the spec directory:
- canonical summary: `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md`
- supporting research files: `.spec-workflow/specs/<spec-name>/research/*`

Forbidden output locations for generated research:
- `docs/*`
- project root
- `.spec-workflow/steering/*`
- `.spec-workflow/user-templates/*`
</artifact_boundary>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder:
- `.spec-workflow/specs/<spec-name>/agent-comms/design-research/<run-id>/`

For each subagent, use:
- `<run-dir>/<subagent>/brief.md` (written by orchestrator before dispatch)
- `<run-dir>/<subagent>/report.md` (written by subagent after execution)
- `<run-dir>/<subagent>/status.json` (written by subagent, machine-readable)

Status schema (minimum):
- `status`: `pass|blocked`
- `summary`: short result
- `inputs`: key files/URLs used
- `outputs`: generated artifacts
- `open_questions`: unresolved items

After all dispatches, write:
- `<run-dir>/_handoff.md` (orchestrator synthesis of subagent outputs)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- web_research -> default `haiku`
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Backward compatibility:
- if `[skills.design].enforce_required` is absent, map `[skills].enforcement`:
  - `"strict"` -> `true`
  - any other value -> `false`

Skill loading gate (mandatory when `skills.enabled=true`):
1. Explicitly invoke every required design skill before running subagents.
2. Record loaded/missing skills in:
   - `.spec-workflow/specs/<spec-name>/SKILLS-DESIGN-RESEARCH.md`
3. If any required skill is missing/not invoked:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<subagents>
- `codebase-pattern-scanner` (model: implementation)
  - Finds reusable patterns, boundaries, integration points.
- `web-pattern-scout-*` (model: web_research, parallel)
  - Performs external web/library/pattern scans.
- `risk-analyst` (model: complex_reasoning)
  - Identifies architecture/operational risks and mitigations.
- `research-synthesizer` (model: complex_reasoning)
  - Produces final consolidated recommendation set.
</subagents>

<preconditions>
- The spec has `requirements.md` and it is approved.
- Also read steering docs when present (`product.md`, `tech.md`, `structure.md`).
- Approval check must come from MCP `spec-status`; never ask approval in chat.
</preconditions>

<workflow>
1. Run design skill loading gate and write `SKILLS-DESIGN-RESEARCH.md`.
2. Create communication run directory:
   - `.spec-workflow/specs/<spec-name>/agent-comms/design-research/<run-id>/`
3. Ensure research directory exists:
   - `.spec-workflow/specs/<spec-name>/research/`
4. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/steering/*.md` (if present)
5. Write subagent briefs and dispatch in parallel:
   - `codebase-pattern-scanner`
   - `web-pattern-scout-*` (2-4 scouts depending on depth)
6. Require each subagent to write `report.md` + `status.json`; BLOCKED if missing.
7. Dispatch `risk-analyst` using outputs from step 5 reports.
8. Dispatch `research-synthesizer` using all prior reports to produce:
   - `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md`
   - optional supporting files only under `.spec-workflow/specs/<spec-name>/research/`
9. Write `<run-dir>/_handoff.md` with:
   - recommendation summary
   - unresolved risks/questions
   - references to all subagent report files
10. Ensure final sections include:
   - primary recommendations
   - alternatives and trade-offs
   - references/patterns to adopt
   - technical risks and mitigations
11. If any generated research file is outside the spec directory, move it into
   `.spec-workflow/specs/<spec-name>/research/` and report relocation in output.
</workflow>

<acceptance_criteria>
- [ ] Every relevant functional requirement has at least one technical recommendation.
- [ ] Existing-code reuse section is included.
- [ ] Risks and recommended decisions section is included.
- [ ] Web-only findings came from web_research model.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/agent-comms/design-research/<run-id>/`.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md`.
- Confirm supporting artifacts path: `.spec-workflow/specs/<spec-name>/research/`.
- Recommend next command: `spw:design-draft <spec-name>`.

If blocked:
- List missing inputs (requirements approval, steering docs, source failures).
- Provide rerun command: `spw:design-research <spec-name>`.
</completion_guidance>
