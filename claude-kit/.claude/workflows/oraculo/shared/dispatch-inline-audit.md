# Inline Audit Dispatch

Pattern for running audit subagents inline within producer commands
(tasks-plan, qa, exec) instead of deferring to standalone check commands.

## When to Use

Apply this pattern when a producer command has finished generating its primary
artifact and you need to validate it before proceeding to approval or the next
phase. The inline audit replaces (but does not remove) standalone check commands,
which remain available for re-validation and CI use.

## Iteration Lifecycle

```
1. audit-iteration start   → initialize state, get iteration 1
2. dispatch-init-audit      → create nested audit dir
3. dispatch auditors        → parallel/sequential via dispatch-setup
4. dispatch-read-status     → read aggregator/decider result
5. IF PASS → done
6. IF BLOCKED:
   a. audit-iteration check → still allowed?
      - YES → audit-iteration advance, re-dispatch writer with feedback, goto 2
      - NO  → STOP, recommend standalone check command
```

## CLI Commands

### Initialize iteration tracking

```
oraculo tools audit-iteration start --run-dir <parent-run> \
  --type <inline-audit|inline-checkpoint> [--max N]
```

Creates `_iteration-state.json` in the audit dir. Max defaults to
`[verification].inline_audit_max_iterations` from config (default 3).

### Create nested audit directory

```
oraculo tools dispatch-init-audit --run-dir <parent-run> \
  --type <inline-audit|inline-checkpoint> [--iteration N]
```

Creates `_inline-audit/iteration-N/` or `_inline-checkpoint/` inside the
parent run dir. Use the returned `audit_dir` as the `--run-dir` for
subsequent `dispatch-setup` calls.

### Check if retry is allowed

```
oraculo tools audit-iteration check --run-dir <parent-run> \
  --type <inline-audit|inline-checkpoint>
```

Returns `allowed: true/false` with a human-readable message:
- `"OK - YOU CAN TRY AGAIN"` — proceed with re-dispatch
- `"BLOCKED - NO MORE TRIES, LET THE MAIN AGENT KNOW"` — stop

### Advance to next iteration

```
oraculo tools audit-iteration advance --run-dir <parent-run> \
  --type <inline-audit|inline-checkpoint> --result blocked
```

Increments the iteration counter. Call this after a BLOCKED result
before re-dispatching the writer.

## Anti-Self-Heal Rule

The orchestrator MUST NOT fix artifacts directly when the audit returns
BLOCKED. Instead:

1. Pass the aggregator/decider report path to the writer subagent via brief.md
2. Re-dispatch the writer subagent to produce a revised artifact
3. Re-run the audit on the revised artifact

This preserves the separation between production and validation.

## Types

| Type | Used by | Scope |
|------|---------|-------|
| `inline-audit` | tasks-plan, qa | Validate producer artifact (tasks.md, QA plan) |
| `inline-checkpoint` | exec | Validate wave completion (evidence, traceability, gate) |

## Inline Checkpoint (exec-specific)

For `exec` waves, the inline checkpoint dispatches the same subagents as
the standalone `/oraculo:checkpoint`:

1. `evidence-collector` — verify impl logs, commits, artifacts
2. `traceability-judge` — trace tasks to requirements
3. `release-gate-decider` — go/no-go decision

On PASS: `oraculo tools wave-update --spec X --wave NN --status pass --tasks ...`
On BLOCKED: max 1 retry (re-dispatch task-implementer to fix missing artifacts),
then recommend standalone `/oraculo:checkpoint`.

## Fallback to Standalone Commands

When inline audit exhausts its iterations:
- `tasks-plan` → recommend `oraculo:tasks-check <spec-name>`
- `qa` → recommend `oraculo:qa-check <spec-name>`
- `exec` → recommend `oraculo:checkpoint <spec-name>`

The standalone commands remain fully functional for re-validation after
manual fixes, CI/CD pipelines, or deeper investigation.
