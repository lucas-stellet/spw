# Brief: task-implementer-8

## Inputs
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/design.md
- cli/internal/cli/root.go (to see existing patterns)
- cli/internal/cli/hook.go or tools.go (to see existing command patterns)

## Task
Implement **Task 8: Wire Cobra validate command group with prompts subcommand**.

From tasks.md:
- Files: cli/internal/cli/validate_cmd.go, cli/internal/cli/root.go
- Requirements: REQ-001, REQ-002
- Test Plan:
  - Integration: spw validate prompts on actual repository completes without error. --json produces valid JSON. --strict includes mirror validation results. Exit code 0 when no violations. Exit code 1 when violations found.
- Verification Command: go build -o /tmp/spw ./cli/cmd/spw && /tmp/spw validate prompts --json
- Definition of Done:
  - newValidateCmd() registered in root.go
  - prompts subcommand with --json and --strict flags
  - Delegates to validate.ValidatePrompts and validate.ValidateMirrors
  - Exit codes: 0=pass, 1=violations, 2=error
  - Produces human-readable table by default, JSON with --json

Skill required: effective-go

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/task-implementer-8/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/task-implementer-8/status.json

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
