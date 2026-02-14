# Brief: release-gate-decider

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- Spec directory: `.spec-workflow/specs/claude-code-based-improvements`
- Tasks file: `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- Requirements file: `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- Design file: `.spec-workflow/specs/claude-code-based-improvements/design.md`
- Wave summary: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/_wave-summary.json`
- Evidence collector report: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-001/evidence-collector/report.md`
- Traceability judge report: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-001/traceability-judge/report.md`

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true
- require_clean_worktree_for_wave_pass: true

## Task
You are the **release-gate-decider** for a checkpoint audit. Your role is to produce the final PASS/BLOCKED decision and corrective actions.

### Your Responsibilities

1. **Read Both Auditor Reports**
   - Read evidence-collector report
   - Read traceability-judge report

2. **Synthesize Findings**
   - Combine evidence from both auditors
   - Identify any critical issues

3. **Apply Gates**
   - Implementation log gate: Every completed task must have implementation log
   - Git gate: Must have clean worktree (no uncommitted changes)
   - Build gate: Code must compile
   - Test gate: Tests must pass
   - Traceability gate: Requirements must be fulfilled

4. **Produce Final Decision**
   - If all gates pass: PASS
   - If any gate fails: BLOCKED with corrective actions

5. **Generate CHECKPOINT-REPORT.md**
   - Write to `.spec-workflow/specs/claude-code-based-improvements/execution/CHECKPOINT-REPORT.md`
   - Include: status, critical issues, corrective actions, recommended next step

### Output Format

Write your findings to `report.md` and generate `CHECKPOINT-REPORT.md`.

For report.md:
```
## Gate Results
- Implementation Log Gate: [pass/fail]
- Git Gate: [pass/fail]
- Build Gate: [pass/fail]
- Test Gate: [pass/fail]
- Traceability Gate: [pass/fail]

## Final Decision
[PASS or BLOCKED]

## Critical Issues
[List any blocking issues]

## Corrective Actions
[If blocked, list exact steps to fix]
```

For status.json:
```json
{
  "status": "pass | blocked",
  "summary": "Final decision summary",
  "skills_used": ["skill-name"],
  "skills_missing": [],
  "model_override_reason": null
}
```

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-001/release-gate-decider/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/checkpoint/run-001/release-gate-decider/status.json
- CHECKPOINT-REPORT: .spec-workflow/specs/claude-code-based-improvements/execution/CHECKPOINT-REPORT.md

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["skill-name"],
  "skills_missing": [],
  "model_override_reason": null
}
```
