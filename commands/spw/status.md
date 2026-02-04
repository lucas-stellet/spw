---
name: spw:status
description: Summarize current spec stage, blockers, and exact next commands
argument-hint: "[<spec-name>] [--all false|true]"
---

<objective>
Show where the workflow stopped and what to run next, with explicit approval/execution blockers.
</objective>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<subagents>
- `state-inspector` (model: implementation)
  - Inspects artifacts, task progress, and wave state.
- `approval-auditor` (model: implementation)
  - Reads approval status through MCP only.
- `next-step-planner` (model: complex_reasoning)
  - Produces ordered, minimal next actions.
</subagents>

<scope_resolution>
1. If `<spec-name>` is provided, use it.
2. Otherwise, inspect `.spec-workflow/specs/*`:
   - if only one spec exists, use it.
   - if multiple specs exist and `--all=true`, summarize all.
   - if multiple specs exist and `--all=false`, ask user to choose one via AskUserQuestion.
</scope_resolution>

<workflow>
1. Resolve target spec(s) from `.spec-workflow/specs/`.
2. For each spec, dispatch `state-inspector` to collect:
   - artifact presence: `requirements.md`, `DESIGN-RESEARCH.md`, `design.md`, `tasks.md`
   - tasks progress counts: `[ ]`, `[-]`, `[x]`
   - active wave/blocked/manual markers when present
3. Dispatch `approval-auditor`:
   - call `spec-status`
   - read document approval state (requirements/design/tasks) from boolean + status fields
4. Dispatch `next-step-planner` to classify stage:
   - product discovery
   - design research
   - design approval
   - tasks planning/approval
   - execution (wave/batch)
   - done
5. Produce:
   - console summary (current stage, blockers, next command)
   - `.spec-workflow/specs/<spec-name>/STATUS-SUMMARY.md` for each inspected spec
</workflow>

<output_contract>
Always include:
- Current stage (single label)
- Why this stage was inferred (artifacts + approval status)
- Blocking reason (if any)
- Exact next command(s), in order

When execution is in progress, also include:
- task counters (`todo/in-progress/done`)
- whether human authorization is required between waves
</output_contract>

<completion_guidance>
If status is clear:
- Recommend exactly one immediate next command.

If blocked:
- Show exact unblock action and rerun command (`spw:status <spec-name>`).

If multiple specs:
- Recommend explicit spec target in next command (`<spec-name>`).
</completion_guidance>
