---
name: spw:post-mortem
description: Analyze post-spec commits and generate reusable process learnings
argument-hint: "<spec-name> [--since-commit <sha>] [--until-ref <ref>] [--tags <tag1,tag2>] [--topic <short-subject>]"
---

<objective>
Generate a structured post-mortem for a finished spec by analyzing commits made after execution completion, then store reusable learnings for future planning/design phases.
</objective>

<shared_policies>
- @.claude/workflows/spw/shared/config-resolution.md
- @.claude/workflows/spw/shared/file-handoff.md
- @.claude/workflows/spw/shared/resume-policy.md
- @.claude/workflows/spw/shared/skills-policy.md
- @.claude/workflows/spw/shared/approval-reconciliation.md
</shared_policies>

<when_to_use>
- Use after `spw:exec` + `spw:checkpoint` are done and follow-up commits changed behavior.
- Use when final delivery required manual corrections that were not captured by PRD/design/tasks/review/tests.
</when_to_use>

<model_policy>
Resolve models from `.spec-workflow/spw-config.toml` `[models]`:
- complex_reasoning -> default `opus`
- implementation -> default `sonnet`
</model_policy>

<file_handoff_protocol>
Subagent communication must be file-first (no implicit-only handoff).

Create a run folder:
- `.spec-workflow/specs/<spec-name>/_agent-comms/post-mortem/<run-id>/`

For each subagent, use:
- `<run-dir>/<subagent>/brief.md` (written by orchestrator before dispatch)
- `<run-dir>/<subagent>/report.md` (written by subagent after execution)
- `<run-dir>/<subagent>/status.json` (written by subagent)

Status schema (minimum):
- `status`: `pass|blocked`
- `summary`: short result
- `inputs`: key files/commits used
- `outputs`: generated artifacts
- `open_questions`: unresolved items

After synthesis, write:
- `<run-dir>/_handoff.md` (final reasoning + links to reports)

If a required `report.md` or `status.json` is missing, stop BLOCKED.
</file_handoff_protocol>

<resume_policy>
Before creating a new run, inspect existing post-mortem run folders:
- `.spec-workflow/specs/<spec-name>/_agent-comms/post-mortem/<run-id>/`

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
- Reuse completed subagent outputs (`report.md` + `status.json` with `status=pass`).
- Redispatch only missing/blocked subagents.
- Always rerun `lessons-synthesizer` and `memory-indexer` before final output.

If user chooses `delete-and-restart`:
- Delete the selected unfinished run dir.
- Continue workflow with a fresh run-id.
- Record deleted path in final output.
</resume_policy>

<range_resolution>
Post-mortem commit range:
- `from` (exclusive):
  - if `--since-commit` is provided, use it.
  - otherwise auto-detect from latest commit related to execution completion for `<spec-name>`:
    - latest commit touching `.spec-workflow/specs/<spec-name>/_generated/CHECKPOINT-REPORT.md` with PASS semantics, or
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

<workflow>
1. Inspect existing post-mortem run dirs and apply `<resume_policy>` decision gate.
2. Determine active run directory:
   - `continue-unfinished` -> reuse latest unfinished run dir
   - `delete-and-restart` or no unfinished run -> create:
     `.spec-workflow/specs/<spec-name>/_agent-comms/post-mortem/<run-id>/`
3. Resolve analysis commit range with `<range_resolution>`.
4. Read baseline/head artifacts when present:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/specs/<spec-name>/tasks.md`
   - `.spec-workflow/specs/<spec-name>/_generated/TASKS-CHECK.md`
   - `.spec-workflow/specs/<spec-name>/_generated/CHECKPOINT-REPORT.md`
5. Dispatch `commit-diff-analyzer` with commit range evidence.
   - if resuming, redispatch only when output is missing/blocked
6. Dispatch `artifact-gap-analyzer` and `review-test-failure-analyzer` using step 5 output + artifacts.
   - if resuming, redispatch only when output is missing/blocked
7. Require analyzer `report.md` + `status.json`; BLOCKED if missing.
8. Dispatch `lessons-synthesizer` to produce final lessons and prevention checklist.
   - if resuming, always rerun `lessons-synthesizer`
9. Write post-mortem report:
   - `.spec-workflow/post-mortems/<spec-name>/<timestamp>-post-mortem.md`
   - include YAML frontmatter:
     - `schema: spw-post-mortem/v1`
     - `spec`, `topic`, `tags`, `created_at`, `branch`
     - `range_from`, `range_to`, `commit_count`
10. Dispatch `memory-indexer` to update:
    - `.spec-workflow/post-mortems/INDEX.md`
    - append one-line entry with date/spec/tags/range/path
11. Write `<run-dir>/_handoff.md` with final root-cause mapping and prevention actions.
</workflow>

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

<acceptance_criteria>
- [ ] Commit range is explicit and traceable (`from..to`).
- [ ] Every major follow-up change has at least one root-cause bucket.
- [ ] Report explains why review and/or tests did not catch the issue.
- [ ] Preventive actions are concrete and command-stage specific.
- [ ] YAML frontmatter exists with topic/tags/range metadata.
- [ ] Shared index `.spec-workflow/post-mortems/INDEX.md` was updated with report path.
- [ ] File-based handoff exists under `.spec-workflow/specs/<spec-name>/_agent-comms/post-mortem/<run-id>/`.
- [ ] If unfinished run exists, explicit user decision (`continue-unfinished` or `delete-and-restart`) was respected.
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
