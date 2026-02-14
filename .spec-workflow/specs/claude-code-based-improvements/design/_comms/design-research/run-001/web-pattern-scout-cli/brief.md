# Brief: web-pattern-scout-cli

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md

## Config Context
- tdd_default: off
- max_wave_size: 3

## Task
Research external patterns for Go CLI validation subcommands:
1. Cobra nested subcommand patterns (validate > prompts, validate > status)
2. JSON vs human-readable dual output modes in CLI validators
3. Exit code conventions for validation CLIs (0=ok, 1=violations, 2=error)
4. Table-driven test patterns for validators in Go

## Output Contract
- Report: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/web-pattern-scout-cli/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/web-pattern-scout-cli/status.json
