---
name: spw:design-research
description: Subagent-driven technical research to prepare design.md
argument-hint: "<spec-name> [--focus <topic>] [--web-depth low|medium|high]"
---

<objective>
Generate architecture and implementation research inputs for the spec design.
Output: `.spec-workflow/specs/<spec-name>/_generated/DESIGN-RESEARCH.md`.
</objective>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
</shared_policies>

<artifact_boundary>
All research outputs must stay inside the spec directory:
- canonical summary: `.spec-workflow/specs/<spec-name>/_generated/DESIGN-RESEARCH.md`
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
- `.spec-workflow/specs/<spec-name>/_agent-comms/design-research/<run-id>/`

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
- `skills_used`: skills actually used by the subagent
- `skills_missing`: required skills not available for the subagent (if any)

After all dispatches, write:
- `<run-dir>/_handoff.md` (orchestrator synthesis of subagent outputs)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<resume_policy>
Before creating a new run, inspect existing design-research run folders:
- `.spec-workflow/specs/<spec-name>/_agent-comms/design-research/<run-id>/`

A run is `unfinished` when any of these is true:
- `_handoff.md` is missing
- any subagent directory is missing `brief.md`, `report.md`, or `status.json`
- any subagent `status.json` reports `status=blocked`

Resume decision gate (mandatory):
1. Find latest unfinished run (if multiple, sort by mtime descending and use the newest).
2. If found, ask user once (AskUserQuestion) with options:
   - `continue-unfinished` (Recommended): continue with that run directory.
   - `delete-and-restart`: delete that unfinished run directory and start a new run.
3. Never choose automatically. Do not infer user intent.
4. If explicit user decision is unavailable, stop with `WAITING_FOR_USER_DECISION`.
5. Do not create a new run-id before this decision.

If user chooses `continue-unfinished`:
- Reuse the same run dir.
- Reuse completed subagent outputs (`report.md` + `status.json` with `status=pass`).
- Redispatch only missing/blocked subagents.
- Always rerun `risk-analyst` and `research-synthesizer` before final synthesis.

If user chooses `delete-and-restart`:
- Delete the selected unfinished run dir.
- Continue workflow with a fresh run-id.
- Record deleted path in final output.
</resume_policy>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- web_research -> default `haiku`
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<post_mortem_memory>
Resolve from `.spec-workflow/spw-config.toml` `[post_mortem_memory]`:
- `enabled` (default `true`)
- `max_entries_for_design` (default `5`)

If enabled and index exists:
1. Read `.spec-workflow/post-mortems/INDEX.md`.
2. Select up to `max_entries_for_design` relevant entries:
   - same `<spec-name>` first
   - then by tag/topic similarity and recency
3. Load selected post-mortems and derive design constraints:
   - failure patterns to avoid
   - missing decisions to enforce
   - review/test checks to include in recommendations

If index/report files are missing, continue with warning (non-blocking).
</post_mortem_memory>

<skills_policy>
Resolve skill policy from `.spec-workflow/spw-config.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/_generated/SKILLS-DESIGN-RESEARCH.md`
2. Avoid loading full skill content in main context (subagent-first).
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<prototype_url_policy>
When a web scout fetches a URL that returns an SPA shell (minimal HTML with only JS bundle references, no meaningful text content), or the URL matches a known prototype/deploy-preview domain (`*.lovable.app`, `*.vercel.app`, `*.netlify.app`, `*.framer.app`, `*.webflow.io`, `*.stackblitz.com`):

1. Use Playwright MCP to navigate the URL, take screenshots, and extract visible content.
   - Playwright MCP is a pre-configured MCP server; discover its tools at runtime. Never invoke `npx` or Node scripts directly for browser automation.
2. If Playwright MCP tools are not available in the current session:
   - Warn the user: "Playwright MCP is not configured — prototype content may be incomplete."
   - Continue with whatever `WebFetch` returned.
3. Include extracted prototype content in the scout's `report.md`.
</prototype_url_policy>

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
1. Run design skills preflight (availability) and write `SKILLS-DESIGN-RESEARCH.md`.
2. Inspect existing design-research run dirs and apply `<resume_policy>` decision gate.
3. Determine active run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create:
     `.spec-workflow/specs/<spec-name>/_agent-comms/design-research/<run-id>/`
4. Ensure research directory exists:
   - `.spec-workflow/specs/<spec-name>/research/`
5. Read:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/steering/*.md` (if present)
   - post-mortem memory inputs via `<post_mortem_memory>`
6. Write subagent briefs (including required skills for each role) and dispatch:
   - `codebase-pattern-scanner`
   - `web-pattern-scout-*` (2-4 scouts depending on depth)
   - if resuming, redispatch only missing/blocked subagents
7. Require each subagent to write `report.md` + `status.json` (with skill fields); BLOCKED if missing.
8. Dispatch `risk-analyst` using outputs from step 6 reports.
9. Dispatch `research-synthesizer` using all prior reports to produce:
   - `.spec-workflow/specs/<spec-name>/_generated/DESIGN-RESEARCH.md`
   - optional supporting files only under `.spec-workflow/specs/<spec-name>/research/`
10. Write `<run-dir>/_handoff.md` with:
   - recommendation summary
   - unresolved risks/questions
   - references to all subagent report files
   - resume decision taken (`continue-unfinished` or `delete-and-restart`)
11. Ensure final sections include:
   - primary recommendations
   - alternatives and trade-offs
   - references/patterns to adopt
   - technical risks and mitigations
12. If any generated research file is outside the spec directory, move it into
   `.spec-workflow/specs/<spec-name>/research/` and report relocation in output.
</workflow>

<acceptance_criteria>
- [ ] Every relevant functional requirement has at least one technical recommendation.
- [ ] Existing-code reuse section is included.
- [ ] Risks and recommended decisions section is included.
- [ ] Web-only findings came from web_research model.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/_agent-comms/design-research/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/_generated/DESIGN-RESEARCH.md`.
- Confirm supporting artifacts path: `.spec-workflow/specs/<spec-name>/research/`.
- Recommend next command: `spw:design-draft <spec-name>`.

If blocked:
- List missing inputs (requirements approval, steering docs, source failures).
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide rerun command: `spw:design-research <spec-name>`.
</completion_guidance>
