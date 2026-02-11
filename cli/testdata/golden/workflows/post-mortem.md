---
name: spw:post-mortem
description: Analyze post-spec commits and generate reusable process learnings
argument-hint: "<spec-name> [--since-commit <sha>] [--until-ref <ref>] [--tags <tag1,tag2>] [--topic <short-subject>]"
---

<dispatch_pattern>
category: pipeline
subcategory: synthesis
policy: (inlined below)

# Pipeline Dispatch Pattern

Sequential chain of subagents where each produces output that feeds the next.
A synthesizer at the end consolidates everything into the command's final artifact.

## Thin-Dispatch Rules

These rules are mandatory for all pipeline commands. They override any command-specific behavior.

### 1. Status-Only Reads

After dispatching any subagent, read ONLY `<subagent>/status.json`.
- If `status=pass`: proceed to next step.
- If `status=blocked`: read `<subagent>/report.md` to decide action (log + skip, or stop BLOCKED).
- Never read `report.md` when status is pass.

### 2. Path-Based Briefs

When subagent-B depends on output from subagent-A:
- Write the **filesystem path** to `subagent-A/report.md` in `subagent-B/brief.md`.
- Never copy or summarize report content into the brief.

Example brief content:
```
## Inputs
- Scope analysis: <run-dir>/qa-scope-analyst/report.md
- Requirements: .spec-workflow/specs/<spec-name>/requirements.md
```

### 3. Synthesizer Reads From Filesystem

The last subagent (synthesizer/writer) receives a brief listing ALL previous report paths.
It reads them directly from disk — the orchestrator does not relay content.

### 4. Run Structure

```
<phase>/_comms/<command>/run-NNN/
  <subagent-1>/brief.md, report.md, status.json
  <subagent-2>/brief.md, report.md, status.json
  <synthesizer>/brief.md, report.md, status.json
  _handoff.md
```

### 5. Resume Policy

On `continue-unfinished`:
- Skip subagents where `status.json` exists with `status=pass`.
- Redispatch missing or blocked subagents.
- Always rerun synthesizer.

## Extension Points

Pipeline commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir and reading config, before first dispatch. Use for user intent gates, preflight checks, skill loading.
- **`pre_dispatch(<subagent>)`**: Before writing a specific subagent's brief. Use for conditional dispatch (e.g., selecting which designer to run based on a gate decision).
- **`post_dispatch(<subagent>)`**: After reading a subagent's status.json. Use for mid-pipeline decisions that affect subsequent dispatches.
- **`post_pipeline`**: After synthesizer completes, before writing _handoff.md. Use for artifact generation, approval reconciliation, completion guidance.

</dispatch_pattern>

<shared_policies>
# Config Resolution

Canonical runtime config path is `.spec-workflow/spw-config.toml`.

Transitional compatibility:
- If `.spec-workflow/spw-config.toml` is missing, fallback to `.spw/spw-config.toml`.

When shell logic is required, prefer:
- `spw tools config-get <section.key> --default <value> [--raw]`

This keeps workflow behavior stable and avoids hardcoded path drift.

# File-First Handoff Contract

Required files for each dispatched subagent:
- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

If any required handoff file is missing, return `BLOCKED`.

**CRITICAL — Run-id format**: MUST be `run-NNN` (zero-padded 3-digit sequential).
Examples: `run-001`, `run-002`, `run-003`.
NEVER use dates, timestamps, or any other format (e.g. `run-20260209-1` is WRONG).
To create a new run, scan existing sibling directories, extract the highest NNN, and increment by 1.

## Thin-Dispatch Integration

This contract defines the **file structure**. The category-level dispatch policies define **how the orchestrator interacts** with these files:

- `dispatch-pipeline.md` — sequential chain, status-only reads, path-based briefs
- `dispatch-audit.md` — parallel auditors, aggregator reads from filesystem
- `dispatch-wave.md` — wave iteration, wave summaries, scout-based resume

The 5 core thin-dispatch rules apply on top of this contract:
1. Orchestrator reads only `status.json` after dispatch (never `report.md` on pass).
2. Briefs contain filesystem paths to prior reports (never content).
3. Synthesizers/aggregators read from disk directly.
4. Run structure follows category layout.
5. Resume skips completed subagents, always reruns final stage.

# Resume Policy

For commands with run folders:
- Detect the latest unfinished run before creating a new run.
- Ask user explicitly: `continue-unfinished` or `delete-and-restart`.
- Never auto-restart without explicit user decision.

# Skills Policy Canonical Notes

- Skill loading is always subagent-first.
- Enforce per stage via `skills.<stage>.enforce_required` (default: `true`).

# MCP Approval Reconciliation

Approval source of truth is MCP.

When `spec-status` is incomplete or ambiguous:
1. Resolve `approvalId` from `spec-status` fields.
2. If missing, inspect `.spec-workflow/approvals/<spec-name>/approval_*.json`.
3. If `approvalId` exists, call MCP `approvals status`.
4. Never infer approval from phase labels or `overallStatus` alone.

</shared_policies>

<objective>
Generate a structured post-mortem for a finished spec by analyzing commits made after execution completion, then store reusable learnings for future planning/design phases.
</objective>

<artifact_boundary>
inputs:
- `.spec-workflow/specs/<spec-name>/requirements.md`
- `.spec-workflow/specs/<spec-name>/design.md`
- `.spec-workflow/specs/<spec-name>/tasks.md`
- `.spec-workflow/specs/<spec-name>/planning/TASKS-CHECK.md`
- `.spec-workflow/specs/<spec-name>/execution/CHECKPOINT-REPORT.md`
- Git commit history (range)

output:
- `.spec-workflow/post-mortems/<spec-name>/<timestamp>-post-mortem.md`
- `.spec-workflow/post-mortems/INDEX.md` (updated)

comms:
- `.spec-workflow/specs/<spec-name>/post-mortem/_comms/run-NNN/`
</artifact_boundary>

<!-- ============================================================
     SUBAGENTS — who does what, in what order, with which model
     ============================================================ -->

<subagents>
- `commit-diff-analyzer` (model: implementation)
  - Builds commit/file-level change map and intent summary.
- `artifact-gap-analyzer` (model: complex_reasoning)
  - Maps changes to PRD/design/tasks/checkpoint/test gaps.
- `review-test-failure-analyzer` (model: complex_reasoning)
  - Explains why review/test gates did not catch the issue.
- `lessons-synthesizer` (model: complex_reasoning)
  - Produces actionable prevention rules for future runs.
- `memory-indexer` (model: implementation)
  - Writes report metadata + updates shared post-mortem index.
</subagents>

<!-- ============================================================
     EXTENSION POINTS — command-specific logic injected into
     the pipeline dispatch pattern
     ============================================================ -->

<extensions>

<!-- pre_pipeline: resume, range resolution ....................... -->
<pre_pipeline>
1. Resolve `SPEC_DIR=.spec-workflow/specs/<spec-name>`.
2. Inspect existing post-mortem run dirs and apply resume decision gate.
3. Resolve analysis commit range with `<range_resolution>`.
4. Read baseline/head artifacts when present.
</pre_pipeline>

<!-- post_pipeline: memory index update ........................... -->
<post_pipeline>
1. Write post-mortem report:
   - `.spec-workflow/post-mortems/<spec-name>/<timestamp>-post-mortem.md`
   - include YAML frontmatter: `schema`, `spec`, `topic`, `tags`, `created_at`, `branch`, `range_from`, `range_to`, `commit_count`
2. Dispatch `memory-indexer` to update:
   - `.spec-workflow/post-mortems/INDEX.md`
   - append one-line entry with date/spec/tags/range/path
3. Write `<run-dir>/_handoff.md` with final root-cause mapping and prevention actions.
</post_pipeline>

</extensions>

<!-- ============================================================
     COMMAND-SPECIFIC POLICIES — referenced by extensions above
     ============================================================ -->

<when_to_use>
- Use after `spw:exec` + `spw:checkpoint` are done and follow-up commits changed behavior.
- Use when final delivery required manual corrections that were not captured by PRD/design/tasks/review/tests.
</when_to_use>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<range_resolution>
Post-mortem commit range:
- `from` (exclusive):
  - if `--since-commit` is provided, use it.
  - otherwise auto-detect from latest commit related to execution completion for `<spec-name>`:
    - latest commit touching `.spec-workflow/specs/<spec-name>/execution/CHECKPOINT-REPORT.md` with PASS semantics, or
    - latest commit touching `.spec-workflow/specs/<spec-name>/tasks.md` where execution appears completed.
- `to` (inclusive):
  - if `--until-ref` is provided, use it.
  - otherwise use `HEAD`.

If auto-detection is ambiguous or no safe anchor is found:
- ask for explicit baseline commit (AskUserQuestion), or
- stop with `WAITING_FOR_USER_DECISION` and instruct rerun with `--since-commit`.
</range_resolution>

<taxonomy>
Classify each relevant change into one or more root-cause buckets:
- `prd-gap`: missing/ambiguous requirement intent, scope, or acceptance criteria.
- `design-gap`: architecture/flow/integration decisions missing or weak.
- `tasks-gap`: decomposition/dependency/wave granularity missing.
- `review-gap`: validation checkpoints missed issue despite evidence.
- `test-gap`: tests missing, weak, flaky, or misaligned with behavior.
- `execution-gap`: implementation sequence or operational guardrail failed.
</taxonomy>

<output_contract>
Post-mortem report must include:
- What changed after spec completion (commit clusters)
- Why each change was missed earlier (by artifact/gate)
- Gaps by taxonomy (`prd-gap`, `design-gap`, `tasks-gap`, `review-gap`, `test-gap`, `execution-gap`)
- Concrete command/process improvements:
  - PRD prompt checks
  - design/research checks
  - tasks decomposition checks
  - review/checkpoint checks
  - test policy checks
- Reusable "Design Agent Guardrails" section with concise bullets
</output_contract>

<!-- ============================================================
     AGENT TEAMS OVERLAY
     ============================================================ -->

<agent_teams_policy>
</agent_teams_policy>

<!-- ============================================================
     ACCEPTANCE CRITERIA
     ============================================================ -->

<acceptance_criteria>
- [ ] Commit range is explicit and traceable (`from..to`).
- [ ] Every major follow-up change has at least one root-cause bucket.
- [ ] Report explains why review and/or tests did not catch the issue.
- [ ] Preventive actions are concrete and command-stage specific.
- [ ] YAML frontmatter exists with topic/tags/range metadata.
- [ ] Shared index `.spec-workflow/post-mortems/INDEX.md` was updated with report path.
- [ ] File-based handoff exists under `post-mortem/_comms/run-NNN/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
- [ ] Orchestrator never read report.md from any subagent (thin-dispatch).
</acceptance_criteria>

<completion_guidance>
On success:
- Confirm report path and analyzed commit range.
- Confirm index update path.
- Recommend next command:
  - if planning a new cycle: `spw:prd <spec-name>` or `spw:plan <spec-name>`
  - if only sharing learning: `spw:status <spec-name>`

If blocked:
- Show missing evidence (commit range ambiguity, missing artifacts, or missing subagent outputs).
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide rerun command: `spw:post-mortem <spec-name> [--since-commit <sha>]`.
</completion_guidance>
