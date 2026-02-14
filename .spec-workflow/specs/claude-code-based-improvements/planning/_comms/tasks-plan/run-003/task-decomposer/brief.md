# Brief: task-decomposer

## Context
- **Spec:** claude-code-based-improvements
- **Run:** run-003
- **Mode:** next-wave
- **Planning for:** Wave 3 (tasks 6, 8, 9)

## Inputs
- Requirements: `.spec-workflow/specs/claude-code-based-improvements/requirements.md`
- Design: `.spec-workflow/specs/claude-code-based-improvements/design.md`
- Current tasks: `.spec-workflow/specs/claude-code-based-improvements/tasks.md`
- Execution checkpoint: `.spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/_wave-summary.json`

## Instructions
Analyze the requirements, design, and current task state. Generate atomic tasks for the next executable wave (Wave 3).

### Completed Waves
- Wave 1: tasks 1, 2, 5 (validate package foundation, frontmatter validation, config extensions)
- Wave 2: tasks 3, 4, 7 (mirror validation, status.json validation, iteration limits)

### Next Wave (Wave 3) - Deferred Tasks to Promote
- Task 6: Implement audit confidence gate logic
- Task 8: Wire Cobra validate command group with prompts subcommand
- Task 9: Integrate enhanced status validation into dispatch-read-status

### Constraints
- `max_wave_size`: 3 (from config)
- `tdd_default`: off
- `require_test_per_task`: true

### Requirements Mapping
- REQ-001: Frontmatter validation
- REQ-002: Mirror validation
- REQ-003: status.json enforcement
- REQ-004: High-signal audit gate
- REQ-005: Iteration limits
- REQ-006: Documentation update
- REQ-007: Regression tests

## Output
Return task decomposition with:
- Task IDs and titles
- Dependencies
- Parallel execution opportunities
- Files to modify
- Requirements coverage
- Test strategy per task
