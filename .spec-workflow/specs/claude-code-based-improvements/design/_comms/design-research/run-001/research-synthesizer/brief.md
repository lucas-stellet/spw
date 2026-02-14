# Brief: research-synthesizer

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Codebase patterns: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/codebase-pattern-scanner/report.md
- CLI patterns: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/web-pattern-scout-cli/report.md
- Frontmatter patterns: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/web-pattern-scout-frontmatter/report.md
- Risk analysis: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/risk-analyst/report.md

## Config Context
- tdd_default: off
- max_wave_size: 3

## Task
Consolidate all research into DESIGN-RESEARCH.md with:
1. Primary recommendations per requirement
2. Alternatives and trade-offs
3. Existing code reuse opportunities
4. Technical risks and mitigations
5. References and patterns to adopt

## Output Contract
- Report: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/research-synthesizer/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/design/_comms/design-research/run-001/research-synthesizer/status.json
