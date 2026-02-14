# Brief: risk-analyst

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Codebase scan: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/codebase-pattern-scanner/report.md

## Config Context
- tdd_default: off
- max_wave_size: 3

## Task
Identify architecture and operational risks for the planned changes:
1. Backward compatibility risks when adding required fields to status.json
2. Risks from adding new frontmatter fields (allowed-tools, model) to all 13 commands
3. Mirror system complexity when adding embedded asset validation
4. Configuration schema evolution risks (new [audit] section)
5. Risk of iteration limits being too aggressive or too lenient
6. Impact on existing hooks and dispatch infrastructure

## Output Contract
- Report: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/risk-analyst/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/risk-analyst/status.json
