---
name: spw:prd
description: Zero-to-PRD discovery flow to generate requirements.md
argument-hint: "<spec-name> [--source <url-or-file.md>]"
---

<objective>
Create or update `.spec-workflow/specs/<spec-name>/requirements.md` in PRD format.

This command combines:
- GSD strengths: v1/v2/out-of-scope scoping, REQ-IDs, testable criteria, traceability.
- superpowers strengths: one-question-at-a-time discovery, recommendation + trade-off framing, incremental section validation.
</objective>

<when_to_use>
- Use when the spec does NOT have approved requirements yet (zero-to-PRD).
- Use when requirements need to be revisited using new product sources.
</when_to_use>

<out_of_scope>
- This command does not create `design.md`.
- This command does not create `tasks.md`.
- Next step after PRD approval: `spw:plan <spec-name>`.
</out_of_scope>

<inputs>
- `spec-name` (required)
- `--source` (optional): URL (GitHub/Linear/ClickUp/etc.) or markdown file.
</inputs>

<source_handling>
If `--source` is provided and looks like a URL (`http://` or `https://`) or markdown (`.md`), run a source-reading gate:

1. Ask with AskUserQuestion:
   - header: "Source"
   - question: "I detected an external source. Do you want to use a specific MCP to read it?"
   - options:
     - "Yes, choose MCP (Recommended)" — Explicit connector selection
     - "Auto" — Try compatible MCP first, fallback to direct read
     - "No" — Read without MCP

2. If user selects "Yes, choose MCP", ask:
   - header: "MCP"
   - question: "Which MCP should be used for this source?"
   - options:
     - "GitHub" — Issues/PRs/repos
     - "Linear" — Linear issues/projects
     - "ClickUp" — ClickUp tasks/lists
     - "Web/Browser" — generic web fetch
     - "Local markdown file" — direct local file read

3. If selected MCP is unavailable, clearly report and ask fallback:
   - "Read without MCP"
   - "Choose another MCP"
</source_handling>

<workflow>
1. Read existing context:
   - `.spec-workflow/specs/<spec-name>/requirements.md` (if present)
   - `.spec-workflow/specs/<spec-name>/design.md` (if present)
   - `.spec-workflow/steering/*.md` (if present)
2. If `--source` is present, process source via the MCP gate above.
3. Run one-question-at-a-time discovery (not a giant form):
   - problem, audience, and context of use
   - primary outcome and success conditions
   - v1 scope, v2 scope, and out-of-scope boundaries
   - constraints and dependencies
   - risks and open questions
4. For ambiguity, propose 2-3 options with an explicit recommendation.
5. Draft the PRD in 200-300 word sections and validate section-by-section.
6. Fill PRD template using priority order:
   - `.spec-workflow/user-templates/prd-template.md` (preferred)
   - fallback: `.spec-workflow/templates/prd-template.md`
   - fallback: built-in structure from this command
7. Save artifacts:
   - Canonical: `.spec-workflow/specs/<spec-name>/requirements.md`
   - Product mirror: `.spec-workflow/specs/<spec-name>/PRD.md`
8. Confirm design-readiness:
   - functional requirements with REQ-IDs and testable acceptance criteria
   - NFRs and success metrics
   - explicit v2 and out-of-scope sections
</workflow>

<acceptance_criteria>
- [ ] Final document is PRD format and remains compatible with spec-workflow requirements flow.
- [ ] Every functional requirement has REQ-ID, priority, and verifiable acceptance criteria.
- [ ] Explicit separation exists for v1, v2, and out-of-scope.
- [ ] If `--source` was provided, MCP usage was explicitly asked.
- [ ] PRD is approved before moving to design/tasks.
</acceptance_criteria>
