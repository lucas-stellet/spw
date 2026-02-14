# Task Decomposer Brief

## Objective
Decompose the approved requirements and design into atomic, self-contained implementation tasks for the `claude-code-based-improvements` spec.

## Inputs
- Requirements: .spec-workflow/specs/claude-code-based-improvements/requirements.md
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md
- Design Research: .spec-workflow/specs/claude-code-based-improvements/design/DESIGN-RESEARCH.md

## Context
- Effective mode: `initial` (rolling-wave, no prior tasks.md)
- max_wave_size: 3
- TDD: off by default (tdd_default=false)
- Language: Go (Cobra CLI)
- Existing packages: cli/internal/{config,tools,hook,cli,embedded,registry,tasks,wave,spec,specdir,render,workspace,install,git}
- New package: cli/internal/validate/

## Constraints
- Each task must be atomic (one concern, one deliverable)
- Each task must specify files to modify/create
- Each task must include a test plan and verification command
- Tasks must map to specific requirements (REQ-001 through REQ-007)
- Focus on Wave 1 tasks only (initial mode). Later work as deferred notes only.
- Tasks must respect code boundaries defined in design (validate does NOT depend on tools/hook/cli)

## Requirements to Cover
1. REQ-001: Frontmatter validation (`spw validate prompts`) - schema.go, prompts.go, validate_cmd.go
2. REQ-002: Mirror validation (`--strict`) - mirror.go
3. REQ-003: Full status.json contract enforcement - status.go, dispatch_status.go extension
4. REQ-004: High-signal audit gate - audit.go, config extension (AuditConfig)
5. REQ-005: Iteration limits - iteration.go, config extension (ExecutionConfig)
6. REQ-006: Documentation updates - README.md, AGENTS.md, docs/SPW-WORKFLOW.md, hooks/README.md, copy-ready/README.md
7. REQ-007: Regression test coverage - *_test.go files per validator

## Decomposition Guidelines
- Foundation tasks first (schema types, config extensions)
- Then individual validators (prompts, mirror, status, audit, iteration)
- Then CLI wiring (validate command)
- Then integration (dispatch_status extension)
- Then documentation
- Tests should be co-located with their implementation task (not separate tasks)

## Output
Write `report.md` with a numbered list of atomic tasks, each containing:
- Task ID (numeric, e.g., 1, 2, 3...)
- Title
- Description
- Files to create/modify
- Requirements covered
- Dependencies on other tasks
- Test plan
- Verification command
