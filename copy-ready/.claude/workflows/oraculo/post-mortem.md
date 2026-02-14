---
name: oraculo:post-mortem
description: Analyze post-spec commits and generate reusable process learnings
argument-hint: "<spec-name> [--since-commit <sha>] [--until-ref <ref>] [--tags <tag1,tag2>] [--topic <short-subject>]"
---

<dispatch_pattern>
category: pipeline
subcategory: synthesis
phase: post-mortem
comms_path: post-mortem/_comms
policy: @.claude/workflows/oraculo/shared/dispatch-pipeline.md
</dispatch_pattern>

<shared_policies>
- @.claude/workflows/oraculo/shared/config-resolution.md
- @.claude/workflows/oraculo/shared/file-handoff.md
- @.claude/workflows/oraculo/shared/resume-policy.md
- @.claude/workflows/oraculo/shared/skills-policy.md
- @.claude/workflows/oraculo/shared/approval-reconciliation.md
- @.claude/workflows/oraculo/shared/dispatch-implementation.md
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

<subagent_artifact_map>
| Subagent | Artifact | Dispatch | Model |
|----------|----------|----------|-------|
| commit-diff-analyzer | (report.md only) | task | implementation |
| artifact-gap-analyzer | (report.md only) | task | complex_reasoning |
| review-test-failure-analyzer | (report.md only) | task | complex_reasoning |
| lessons-synthesizer | post-mortem report | task | complex_reasoning |
| memory-indexer | INDEX.md update | task | implementation |
</subagent_artifact_map>

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
- Use after `oraculo:exec` + `oraculo:checkpoint` are done and follow-up commits changed behavior.
- Use when final delivery required manual corrections that were not captured by PRD/design/tasks/review/tests.
</when_to_use>

<model_policy>
Resolve models from `.spec-workflow/oraculo.toml` `[models]`:
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
- `discover-gap`: missing/ambiguous requirement intent, scope, or acceptance criteria.
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
- Gaps by taxonomy (`discover-gap`, `design-gap`, `tasks-gap`, `review-gap`, `test-gap`, `execution-gap`)
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
@.claude/workflows/oraculo/overlays/active/post-mortem.md
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
  - if planning a new cycle: `oraculo:discover <spec-name>` or `oraculo:plan <spec-name>`
  - if only sharing learning: `oraculo:status <spec-name>`

If blocked:
- Show missing evidence (commit range ambiguity, missing artifacts, or missing subagent outputs).
- If waiting on resume decision, ask user to choose `continue-unfinished` or `delete-and-restart`, then rerun.
- Provide rerun command: `oraculo:post-mortem <spec-name> [--since-commit <sha>]`.
</completion_guidance>
