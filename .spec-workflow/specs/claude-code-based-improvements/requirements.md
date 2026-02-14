---
spw:
  schema: 1
  spec: "claude-code-based-improvements"
  doc: "requirements"
  status: "draft"
  source: "spw:prd"
  updated_at: "2026-02-13"
  inputs:
    - "https://github.com/anthropics/claude-code"
  open_questions: []
  risk: "medium"
---

# Requirements Document

## Product Context
- Business problem: Critical rules still rely too heavily on prompt markdown and have only partial enforcement.
- Feature objective: Move validations and guardrails into `spw` CLI commands (Go), making the workflow more deterministic.
- Target audience: SPW maintainers and operators of `spw:*` commands.

## Scope
### In Scope
- New CLI command for prompt/workflow validation (`spw validate prompts`).
- Full status.json contract enforcement by the CLI (closing the gap between what `dispatch-setup` documents and what `dispatch-read-status` validates).
- "High-signal only" gate for audit commands with configurable confidence threshold.
- Explicit revision/replanning attempt limits.
- Go test coverage for new contracts.

### Out of Scope
- Rewriting dispatch categories.
- Changing the macro workflow `prd -> plan -> exec -> checkpoint -> qa`.
- Creating new MCP artifacts beyond `requirements.md`, `design.md`, `tasks.md`.

## Functional Requirements

### REQ-001 - Mandatory frontmatter validation for commands
- User story: As a maintainer, I want to enforce a minimum contract on `commands/spw/*.md`.
- Acceptance criteria (EARS):
  - WHEN `spw validate prompts` runs THEN each command MUST contain `description`, `argument-hint`, `allowed-tools`, `model`.
  - IF any required field is missing THEN the command fails with an explicit list of violations.
  - WHEN `spw validate prompts --json` runs THEN output is machine-parseable JSON with `ok`, `summary`, `violations[]`, `stats` fields (suitable for CI pipelines).
- Priority: Must

### REQ-002 - Mirror and embedded asset validation
- User story: As a maintainer, I want to detect drift automatically.
- Acceptance criteria (EARS):
  - WHEN `spw validate prompts --strict` runs THEN it must validate consistency between `commands/`, `workflows/`, `copy-ready/` and CLI embedded assets.
  - IF there is a divergence THEN it must fail with pairs of divergent paths.
- Priority: Must

### REQ-003 - Full status.json contract enforcement
- User story: As an orchestrator, I want structured status for reliable decision-making.
- Context: The current status.json has two core fields (`status`, `summary`) validated by the CLI, but `dispatch-setup` already documents extended fields (`skills_used`, `skills_missing`, `model_override_reason`) that are not validated. This requirement closes that gap — no version bump needed, just enforcing the full existing contract.
- Acceptance criteria (EARS):
  - WHEN `spw tools dispatch-setup` generates a brief THEN the output contract section must list all required fields: `status`, `summary`, `skills_used`, `skills_missing`, `model_override_reason`.
  - WHEN `spw tools dispatch-read-status` reads status THEN it validates all fields including extended ones (types, required vs optional).
  - IF invalid THEN it returns `valid=false` and `errors[]` per field.
- Priority: Must

### REQ-004 - High-signal gate for audits
- User story: As a maintainer, I want to reduce false positives in audit commands.
- Context: Currently audit decisions are binary (pass/blocked). This adds a confidence dimension so low-confidence blockers become warnings instead of hard blocks.
- Acceptance criteria (EARS):
  - WHEN a finding is blocking THEN it MUST have `validated=true` and `confidence >= audit_min_confidence`.
  - IF the threshold is not met THEN it becomes a warning, not a blocker.
  - The `audit_min_confidence` value MUST be configurable in `spw-config.toml` under a new `[audit]` section (e.g. `audit_min_confidence = 0.8`).
  - Documentation MUST be updated to describe how `audit_min_confidence` is measured, the criteria auditor subagents use to assign confidence values, the scale (0.0-1.0), and examples of high vs low confidence findings.
- Priority: Must

### REQ-005 - Iteration limits
- User story: As an operator, I want to avoid indefinite loops.
- Acceptance criteria (EARS):
  - WHEN revision/replanning cycles occur THEN they MUST respect `max_revision_attempts` and `max_replan_attempts`.
  - IF the limit is exceeded THEN the result MUST be `WAITING_FOR_HUMAN_DECISION` with an explicit action.
- Priority: Should

### REQ-006 - Documentation update in the same patch
- User story: As a maintainer, I want to avoid doc/behavior drift.
- Acceptance criteria (EARS):
  - WHEN behavior/defaults/guardrails change THEN the patch MUST update `README.md`, `AGENTS.md`, `docs/SPW-WORKFLOW.md`, `hooks/README.md`, `copy-ready/README.md`.
- Priority: Must

### REQ-007 - Regression test coverage
- User story: As a maintainer, I want confidence in safe evolution.
- Acceptance criteria (EARS):
  - WHEN Go tests run THEN new contracts (validator/frontmatter/status enforcement/high-signal) MUST be covered.
- Priority: Must

## Non-Functional Requirements
- Performance: `spw validate prompts` must complete in under 2 seconds on the standard repository.
- Security: Local-only validation, no external calls.
- Observability: Error messages must include rule + path + reason.
- Accessibility (when applicable): N/A (CLI).

## Constraints and Assumptions
- Technical constraints: Implementation focused on Go CLI (Cobra + internal packages).
- External dependencies: No new required dependencies beyond those already in use.
- Assumptions:
  - No backward compatibility needed — the extended fields are already documented in briefs but simply not validated yet.
  - Markdown workflows remain declarative; central enforcement lives in the CLI.

## Success Criteria
- Metric 1: 100% of `commands/spw/*.md` with `allowed-tools` and `model`.
- Metric 2: 100% of `status.json` valid under the full contract in tests.
- Metric 3: Reduction of unvalidated blockers in audits.
