---
name: oraculo:design-research
description: Subagent-driven technical research to prepare design.md
argument-hint: "<spec-name> [--focus <topic>] [--web-depth low|medium|high]"
---

<dispatch_pattern>
category: pipeline
subcategory: research
phase: design
comms_path: design/_comms/design-research
policy: @.claude/workflows/oraculo/shared/dispatch-pipeline.md
</dispatch_pattern>

<shared_policies>
- @.claude/workflows/oraculo/shared/config-resolution.md
- @.claude/workflows/oraculo/shared/file-handoff.md
- @.claude/workflows/oraculo/shared/resume-policy.md
- @.claude/workflows/oraculo/shared/skills-policy.md
- @.claude/workflows/oraculo/shared/approval-reconciliation.md
</shared_policies>

<objective>
Generate architecture and implementation research inputs for the spec design.
Output: `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md`.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/steering/*.md` (if present)
- post-mortem memory entries (if enabled)

output:
- `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md`

comms:
- `.spec-workflow/specs/<spec-name>/design/_comms/design-research/run-NNN/`

Forbidden output locations for generated research:
- `docs/*`
- project root
- `.spec-workflow/steering/*`
- `.spec-workflow/user-templates/*`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `codebase-pattern-scanner` (model: implementation)
  - Finds reusable patterns, boundaries, integration points.
  - Investigates cross-boundary integrations: verify whether framework features (e.g., LiveView navigation, event handlers) work inside third-party-managed DOM (e.g., Leaflet popups, phx-update="ignore" zones). Factual questions must be resolved here, not deferred to design-draft.
- `web-pattern-scout-*` (model: web_research, parallel)
  - Performs external web/library/pattern scans.
- `risk-analyst` (model: complex_reasoning)
  - Identifies architecture/operational risks and mitigations.
- `research-synthesizer` (model: complex_reasoning)
  - Produces final consolidated recommendation set.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the pipeline dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: SPA/prototype URL detection, skills, resume .... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Apply skills policy: run design skills preflight and write `SKILLS-DESIGN-RESEARCH.md`.
3. Verify preconditions: `requirements.md` exists and is approved (MCP `spec-status`).
4. Load post-mortem memory inputs via `<post_mortem_memory>`.
5. Inspect existing design-research run dirs and apply resume decision gate.
</pre_pipeline>

<!-- pre_dispatch: Playwright MCP fallback for SPA scouts ......... -->
<pre_dispatch subagent="web-pattern-scout-*">
Apply `<prototype_url_policy>`: if scout target is an SPA or known prototype domain, use Playwright MCP. If unavailable, warn and continue with WebFetch.
</pre_dispatch>

<!-- post_pipeline: artifact generation + guidance ................. -->
<post_pipeline>
1. Write `<run-dir>/_handoff.md` with recommendation summary, unresolved risks, and subagent report references.
2. Ensure final sections include:
   - primary recommendations
   - alternatives and trade-offs
   - references/patterns to adopt
   - technical risks and mitigations
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

<preconditions>
- The spec has `requirements.md` and it is approved.
- Also read steering docs when present (`product.md`, `tech.md`, `structure.md`).
- Approval check must come from MCP `spec-status`; never ask approval in chat.
</preconditions>

<model_policy>
Resolve models from `.spec-workflow/oraculo.toml` `[models]`:
- web_research -> default `haiku`
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`

Multimodal override: when a `web-pattern-scout-*` must analyze images (screenshots, prototype visuals), use `implementation` instead of `web_research`. Record the override reason in `status.json` `model_override_reason` field.
</model_policy>

<post_mortem_memory>
Resolve from `.spec-workflow/oraculo.toml` `[post_mortem_memory]`:
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

<prototype_url_policy>
When a web scout fetches a URL that returns an SPA shell (minimal HTML with only JS bundle references, no meaningful text content), or the URL matches a known prototype/deploy-preview domain (`*.lovable.app`, `*.vercel.app`, `*.netlify.app`, `*.framer.app`, `*.webflow.io`, `*.stackblitz.com`):

1. Use Playwright MCP to navigate the URL, take screenshots, and extract visible content.
   - Playwright MCP is a pre-configured MCP server; discover its tools at runtime. Never invoke `npx` or Node scripts directly for browser automation.
2. If Playwright MCP tools are not available in the current session:
   - Warn the user: "Playwright MCP is not configured — prototype content may be incomplete."
   - Continue with whatever `WebFetch` returned.
3. Include extracted prototype content in the scout's `report.md`.
</prototype_url_policy>

<skills_policy>
Resolve skill policy from `.spec-workflow/oraculo.toml`:
- `[skills].enabled`
- `[skills.design].required`
- `[skills.design].optional`
- `[skills.design].enforce_required` (boolean)

Skill gate (mandatory when `skills.enabled=true`):
1. Run availability preflight and write:
   - `.spec-workflow/specs/<spec-name>/design/SKILLS-DESIGN-RESEARCH.md`
2. Avoid loading full skill content in main context (subagent-first).
3. Require each subagent `status.json` to include `skills_used`/`skills_missing`.
4. If any required skill is missing/not used where required:
   - `enforce_required=true` -> BLOCKED
   - `enforce_required=false` -> warn and continue
</skills_policy>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
@.claude/workflows/oraculo/overlays/active/design-research.md
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] Every relevant functional requirement has at least one technical recommendation.
- [ ] Existing-code reuse section is included.
- [ ] Risks and recommended decisions section is included.
- [ ] Web-only findings came from web_research model (or implementation with multimodal override documented in status.json).
- [ ] File-based handoff exists under `design/_comms/design-research/run-NNN/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm output path: `.spec-workflow/specs/<spec-name>/design/DESIGN-RESEARCH.md`.
- Recommend next command: `oraculo:design-draft <spec-name>`.

If blocked:
- List missing inputs (requirements approval, steering docs, source failures).
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide rerun command: `oraculo:design-research <spec-name>`.
</completion_guidance>
