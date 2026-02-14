# Design Critic Brief

## Objective
Evaluate the design document for consistency, completeness, and correctness. Verify all requirements are traced, all technical decisions are justified, diagrams are valid, and markdown is UI-safe.

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design Draft: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-draft/run-001/design-writer/report.md
- Traceability Mapping: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-draft/run-001/traceability-mapper/report.md

## Checklist
1. All REQ-001 through REQ-007 traced in the requirement mapping table
2. Technical decisions are justified with pros/cons
3. Architecture section contains at least one valid Mermaid diagram
4. Mermaid diagrams use fenced lowercase `mermaid` language markers
5. No unescaped angle brackets outside fenced code blocks (MDX safety)
6. ATX headings with consistent hierarchy
7. Valid tables with header separator rows
8. Fenced code blocks balanced and language-tagged
9. Test strategy covers unit, integration, and E2E
10. Risk mitigations are actionable
11. Config additions are backward compatible
12. Exit code convention is consistent (0/1/2)

## Output
Your `report.md` must contain:
- PASS or BLOCKED verdict
- Checklist results (each item pass/fail with notes)
- If BLOCKED: specific issues that must be fixed with line-level guidance

## Model
Use complex_reasoning (opus) for this task.
