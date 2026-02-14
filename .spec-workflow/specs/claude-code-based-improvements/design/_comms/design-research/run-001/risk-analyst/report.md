# Risk Analysis Report

## Risk 1: Backward Compatibility of status.json Extended Fields (REQ-003)

**Risk Level: Medium**

The requirements state "no backward compatibility needed" because extended fields are already documented in briefs. However, existing `status.json` files produced by prior runs do NOT contain `skills_used`, `skills_missing`, or `model_override_reason`. If `dispatch-read-status` starts validating these as required, it will retroactively break reading of old status files.

**Mitigation:**
- Make extended fields optional-with-default in validation: `skills_used` defaults to `[]`, `skills_missing` defaults to `[]`, `model_override_reason` defaults to `null`.
- Only enforce presence when `--strict` is passed or when the brief explicitly documented the fields.
- Alternative: use a schema version field (`"schema_version": 2`) to distinguish old vs new status files.

**Recommendation:** Treat extended fields as "validated when present, warned when absent" in default mode. Strict mode enforces all fields. This avoids breaking existing runs while nudging toward full compliance.

## Risk 2: Frontmatter Migration for 13 Commands (REQ-001)

**Risk Level: Low-Medium**

Adding `allowed-tools` and `model` to all 13 `commands/spw/*.md` files requires:
1. Determining the correct `allowed-tools` list for each command (varies by command purpose).
2. Determining the correct `model` for each command.
3. Updating both source and `copy-ready/` mirrors simultaneously.

The risk is getting the `allowed-tools` lists wrong, which could either be too restrictive (breaking commands) or too permissive (undermining the guardrail).

**Mitigation:**
- Derive `allowed-tools` from what each workflow actually needs (read the workflow files to see which tools subagents use).
- Derive `model` from the `<model_policy>` and `<subagents>` sections in each workflow.
- Make the validator cross-check: if a command declares `model: haiku` but its workflow uses `complex_reasoning` subagents, flag a warning.

**Recommendation:** Include an `--auto-fix` mode or a migration script that generates frontmatter from workflow analysis. This reduces manual error.

## Risk 3: Mirror Validation Scope Creep (REQ-002)

**Risk Level: Medium**

The `--strict` mode must validate consistency between:
- `commands/spw/` <-> `copy-ready/.claude/commands/spw/`
- `workflows/spw/` <-> `copy-ready/.claude/workflows/spw/`
- CLI embedded assets <-> filesystem sources

The embedded asset comparison is tricky because `go:embed` strips directory prefixes and the `CompositeFS` routes by prefix. A byte-for-byte comparison requires knowing the exact mapping between embedded paths and filesystem paths.

**Mitigation:**
- Define an explicit mapping table (source path -> embedded path) in the validator.
- Use `embed.FS.ReadFile` + `os.ReadFile` and compare content hashes.
- Start with command/workflow mirrors only; add embedded asset validation as a separate sub-check.

**Recommendation:** Implement in two phases: (1) filesystem mirror validation (ports existing bash script), (2) embedded asset validation (new capability).

## Risk 4: Audit Confidence Threshold Calibration (REQ-004)

**Risk Level: Medium-High**

The `audit_min_confidence` threshold introduces a continuous value (0.0-1.0) into what was previously a binary system. Risks:
- Subagents may not calibrate confidence consistently.
- A threshold of 0.8 might suppress legitimate blockers.
- Different audit commands (tasks-check, qa-check, checkpoint) may need different thresholds.

**Mitigation:**
- Start with a single global threshold (as specified in REQ-004).
- Document clear calibration guidelines: what constitutes 0.9+ vs 0.5-0.8 confidence.
- Log all suppressed blockers (became warnings due to threshold) for observability.
- Allow per-command overrides later (e.g., `[audit.checkpoint]` section).

**Recommendation:** The initial implementation should be conservative (default 0.8). Include a `--verbose` mode that shows confidence scores for all findings, not just suppressed ones.

## Risk 5: Iteration Limits Edge Cases (REQ-005)

**Risk Level: Low**

Iteration limits (`max_revision_attempts`, `max_replan_attempts`) could cause problems if:
- The limits are too low and legitimate complex specs get stuck.
- The `WAITING_FOR_HUMAN_DECISION` state is not properly handled by callers.

**Mitigation:**
- Use sensible defaults (e.g., 3 revision attempts, 2 replan attempts).
- Make limits configurable in `spw-config.toml` under a new section.
- Ensure the human decision prompt includes the iteration count and summary of what was tried.

**Recommendation:** Low risk. Follow the established pattern of config-driven behavior with sensible defaults.

## Risk 6: Documentation Synchronization (REQ-006)

**Risk Level: Low**

Updating 5 documentation files in the same patch is operationally complex but low-risk technically. The main risk is forgetting one of the files.

**Mitigation:**
- Add a CI check (or extend `spw validate`) that verifies documentation files were modified when behavior-changing Go files are modified.
- Or: make this a manual checklist item in the tasks.

**Recommendation:** Include doc updates as explicit sub-tasks in each implementation task rather than a single monolithic doc task.

## Risk 7: Test Coverage Gaps (REQ-007)

**Risk Level: Low**

The existing test infrastructure is solid (table-driven, TempDir, helpers). The risk is mainly about test completeness.

**Mitigation:**
- Define minimum test cases per validator: valid, missing required field, invalid type, empty file, no frontmatter.
- For status.json: test all field combinations (present, missing, wrong type, null).
- For audit confidence: test at threshold boundary (exactly at, above, below).

**Recommendation:** Include golden file tests for JSON output format stability.
