# Brief: spec-compliance-reviewer-8

## Inputs
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/task-implementer-8/report.md
- cli/internal/cli/validate_cmd.go
- cli/internal/cli/root.go

## Task
Verify Task 8 implementation (wire Cobra validate command) against tasks.md Definition of Done:

1. newValidateCmd() registered in root.go - verify registration exists
2. prompts subcommand with --json and --strict flags - verify flags defined
3. Delegates to validate.ValidatePrompts and validate.ValidateMirrors - verify delegation
4. Exit codes: 0=pass, 1=violations, 2=error - verify exit code handling
5. Build and run: go build -o /tmp/spw ./cli/cmd/spw && /tmp/spw validate prompts --json

Report PASS if all criteria met, BLOCKED with specific issues if not.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/spec-compliance-reviewer-8/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-03/execution/run-001/spec-compliance-reviewer-8/status.json

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
