# Brief: codebase-pattern-scanner

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- CLI source: cli/internal/ (Go packages)
- Commands: commands/spw/*.md
- Workflows: workflows/spw/*.md
- Validation script: scripts/validate-thin-orchestrator.sh
- Embedded assets: cli/internal/embedded/embed.go

## Config Context
- tdd_default: off
- max_wave_size: 3

## Task
Scan the SPW codebase for:
1. Existing validation patterns in Go CLI (dispatch_status.go, dispatch_setup.go, hook_test.go)
2. Frontmatter parsing in registry.go and how command metadata is loaded
3. Embedded asset structure and mirror validation in validate-thin-orchestrator.sh
4. Config struct patterns (config.go) for adding new sections like [audit]
5. Reusable patterns for new `validate` subcommand
6. Integration points between hooks, tools, and registry

## Output Contract
- Report: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/codebase-pattern-scanner/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/codebase-pattern-scanner/status.json
