# Report: task-implementer-8

## Task
Implemented **Task 8: Wire Cobra validate command group with prompts subcommand**.

## Files Modified
- `/Users/lucas/dev/projects/my-claude-commands/spw/cli/internal/cli/root.go` - Added `newValidateCmd()` registration
- `/Users/lucas/dev/projects/my-claude-commands/spw/cli/internal/cli/validate_cmd.go` - New file implementing the validate command

## Implementation Details

### newValidateCmd()
Creates the parent `validate` command that groups validation subcommands.

### newValidatePromptsCmd()
Creates the `prompts` subcommand with the following features:

1. **Flags**:
   - `--json`: Output results as JSON
   - `--strict`: Also validate mirror integrity (delegates to `validate.ValidateMirrors`)

2. **Delegates to**:
   - `validate.ValidatePrompts()` - validates frontmatter in `.claude/commands/spw/*.md` files
   - `validate.ValidateMirrors()` - validates mirror integrity when `--strict` is used

3. **Exit Codes**:
   - `0`: Pass (no violations)
   - `1`: Violations found
   - `2`: Error (e.g., directory doesn't exist)

4. **Output Modes**:
   - Human-readable table by default (shows summary, stats, and violations)
   - JSON with `--json` flag (includes ok, summary, stats, violations, mirrors, mirror_violations)

## Verification

Build and test commands executed successfully:

```bash
# Build
go build -o /tmp/spw ./cli/cmd/spw

# Test JSON output
/tmp/spw validate prompts --json

# Test human-readable output
/tmp/spw validate prompts

# Test with strict mode (includes mirror validation)
/tmp/spw validate prompts --json --strict

# Verify help works
/tmp/spw validate --help
/tmp/spw validate prompts --help
```

## Definition of Done Checklist

- [x] newValidateCmd() registered in root.go
- [x] prompts subcommand with --json and --strict flags
- [x] Delegates to validate.ValidatePrompts and validate.ValidateMirrors
- [x] Exit codes: 0=pass, 1=violations, 2=error
- [x] Produces human-readable table by default, JSON with --json
