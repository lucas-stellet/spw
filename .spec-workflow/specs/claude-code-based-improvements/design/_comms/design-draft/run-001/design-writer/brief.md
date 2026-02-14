# Design Writer Brief

## Objective
Produce a complete `design.md` document for the `claude-code-based-improvements` spec following the design template. The document must have full requirement traceability, justified technical decisions, architecture diagrams, and test strategy.

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design Research: .spec-workflow/specs/claude-code-based-improvements/design/DESIGN-RESEARCH.md
- Traceability Mapping: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-draft/run-001/traceability-mapper/report.md
- Design Template: .spec-workflow/user-templates/design-template.md

## Output Format
Your `report.md` must be a complete design.md document that:
1. Follows the design template structure exactly
2. Includes the traceability matrix from the mapper
3. Contains at least one valid Mermaid diagram in the Architecture section
4. Uses fenced lowercase `mermaid` language markers
5. Has no unescaped angle brackets outside fenced code blocks (MDX safety)
6. Uses ATX headings with consistent hierarchy
7. Has valid tables with header separator rows

## Key Design Decisions (from research + mapping)
- New `cli/internal/validate/` package for all validation logic
- `spw validate prompts` as top-level command group (not under `tools`)
- `gopkg.in/yaml.v3` for frontmatter parsing
- Graduated default/strict validation for status.json
- `[audit]` config section with `audit_min_confidence` float
- `_iteration_state.json` for counter persistence
- Content-hash comparison for mirror validation

## Constraints
- All code changes are Go CLI under `cli/`
- Must integrate with existing Cobra command structure in `cli/internal/cli/root.go`
- Exit code contract: 0=pass, 1=violations, 2=error
- Config follows `config.go` patterns (struct + Defaults + Load)
- No unescaped `<...>` outside code fences (MDX requirement)

## Model
Use implementation (sonnet) for this task.
