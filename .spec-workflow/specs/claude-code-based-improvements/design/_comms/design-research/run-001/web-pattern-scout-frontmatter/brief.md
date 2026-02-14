# Brief: web-pattern-scout-frontmatter

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md

## Config Context
- tdd_default: off
- max_wave_size: 3

## Task
Research external patterns for YAML frontmatter parsing and validation in Go:
1. Libraries for parsing YAML frontmatter from markdown (goldmark-frontmatter, adrg/frontmatter, manual parsing)
2. Lightweight approaches (no heavy dependencies) vs full goldmark integration
3. Validation schema patterns for required/optional fields with type enforcement

## Output Contract
- Report: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/web-pattern-scout-frontmatter/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/web-pattern-scout-frontmatter/status.json
