# Wave Execution Dispatch Pattern

Iterative dispatch over a set of work items, split into waves. Each wave dispatches
subagents for a bounded group of items, completes, then the next wave starts.
A synthesizer at the end consolidates everything into the command's final artifact.

## Thin-Dispatch Rules

These rules are mandatory for all wave execution commands. They override any command-specific behavior.

### 1. Status-Only Reads

After dispatching any subagent, read ONLY `<subagent>/status.json`.
- If `status=pass`: proceed to next subagent or wave step.
- If `status=blocked`: read `<subagent>/report.md` to decide action (log + skip, or stop BLOCKED).
- Never read `report.md` when status is pass.

### 2. Path-Based Briefs

When dispatching subsequent subagents or the synthesizer:
- Write **filesystem paths** to previous report files in the brief.
- Never copy or summarize report content into the brief.

### 3. Synthesizer Reads From Filesystem

The final synthesizer receives a brief listing ALL wave summary paths and report paths.
It reads them directly from disk — the orchestrator does not relay content.

### 4. Run Structure

```
<phase>/waves/wave-NN/
  <stage>/run-NNN/
    <subagent-1>/brief.md, report.md, status.json
    <subagent-2>/brief.md, report.md, status.json
    _handoff.md
  _wave-summary.json
  _latest.json
```

### 5. Resume Policy

On `continue-unfinished`:
- Scout inspects `_wave-summary.json` per wave.
- Skip completed waves entirely.
- Resume from first incomplete wave.
- Always rerun synthesizer.

## Wave Lifecycle

```
orchestrator:
  dispatch state-scout → resume state (compact)
  resolve waves from work items + wave size config
  for each wave-NN:
    for each subagent in wave:
      dispatch subagent → read status.json only
      on blocked: read report.md, decide action
    write wave-NN/_wave-summary.json (from status.json data)
  dispatch synthesizer (brief includes paths to all wave summaries + reports)
  synthesizer reads from fs → final artifact
```

## Extension Points

Wave execution commands may inject logic at these points:

- **`pre_pipeline`**: After resolving spec dir, before scout dispatch. Use for precondition checks, skill loading.
- **`inter_wave`**: Between waves, after wave summary is written. Use for quality gates (checkpoint), user authorization, re-authentication.
- **`per_task`**: Within a wave, around each task/scenario dispatch. Use for git hygiene, commit policy, per-item gates.
- **`post_pipeline`**: After all waves complete and synthesizer runs, before final _handoff.md. Use for artifact generation, drift reporting, next-step guidance.

## Implementation Procedure

For step-by-step CLI commands to execute this pattern, see `dispatch-implementation.md`.
The CLI handles directory creation, brief skeleton generation, status validation, and handoff generation.
