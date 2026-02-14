# Traceability Mapper Brief

## Objective
Map each REQ-ID from requirements.md to concrete technical decisions, affected files/modules, and test strategies. Produce a structured traceability matrix that the design-writer will consume.

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design Research: .spec-workflow/specs/claude-code-based-improvements/design/DESIGN-RESEARCH.md
- Design Template: .spec-workflow/user-templates/design-template.md

## Output Format
Your `report.md` must contain:
1. A traceability matrix table mapping REQ-ID -> Technical Decision -> Components/Files -> Test Strategy
2. Key architectural decisions with justification
3. Dependency and risk notes per requirement
4. Identified code reuse opportunities in the existing Go CLI codebase

## Context
This spec is about moving SPW validations and guardrails from prompt markdown into Go CLI commands. The requirements cover:
- REQ-001: Mandatory frontmatter validation (`spw validate prompts`) with --json output
- REQ-002: Mirror/embedded asset validation (`--strict` mode)
- REQ-003: Full status.json contract enforcement (extended fields)
- REQ-004: High-signal gate for audits (configurable `audit_min_confidence` in TOML)
- REQ-005: Iteration limits for revision/replanning cycles
- REQ-006: Documentation update in same patch
- REQ-007: Regression test coverage

## Constraints
- All new validations are Go CLI commands under `cli/`
- Must integrate with existing hook architecture (`cli/internal/hook/`)
- Config reads from `.spec-workflow/spw-config.toml`
- Exit code contract: 0 = ok, 2 = block (matching existing hooks)

## Model
Use complex_reasoning (opus) for this task.
