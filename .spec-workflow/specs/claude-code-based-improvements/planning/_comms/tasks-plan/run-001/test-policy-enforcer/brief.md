# Test Policy Enforcer Brief

## Objective
Verify every task has tests or a documented no-test exception with valid justification.

## Inputs
- Task decomposition: .spec-workflow/specs/claude-code-based-improvements/planning/_comms/tasks-plan/run-001/task-decomposer/report.md
- Config: tdd_default=false, require_test_per_task=true, allow_no_test_exception=true

## Rules
- Every task must include a test plan with at least one test category (unit/integration)
- Every task must include a verification command
- No-test exceptions allowed only with explicit justification and alternative validation
- TDD is off by default (tdd_default=false), tasks use TDD: inherit -> TDD off

## Output
Write `report.md` with pass/fail per task for test policy compliance, and any exceptions with justification evaluation.
