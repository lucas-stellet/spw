---
spw:
  schema: 1
  spec: "<spec-name>"
  doc: "tasks"
  status: "draft"
  source: "spw:tasks-plan"
  updated_at: "YYYY-MM-DD"
  inputs:
    - ".spec-workflow/specs/<spec-name>/requirements.md"
    - ".spec-workflow/specs/<spec-name>/design-draft.md"
  requirements: []
  task_ids: []
  test_required: true
  risk: "low"
  open_questions: []
---

# Tasks Document (TDD OFF by default)

## Execution Constraints
- max_tasks_per_wave: 3
- require_test_per_task: true
- allow_no_test_exception: true
- tdd_default: off

## Wave Plan
- Wave 1:
- Wave 2:

---

- [ ] 1.1 [Task title]
  - Wave: 1
  - Depends On: none
  - Can Run In Parallel With: 1.2, 1.3
  - Files:
    - modify: path/to/file.ex
    - test: test/path/to/file_test.exs
  - _Requirements: REQ-001_
  - TDD: inherit
  - Test Plan:
    - Unit:
    - Integration (if applicable):
  - Verification Command:
    - mix test test/path/to/file_test.exs
  - No-TDD Justification (only for TDD: skip):
    - Reason:
    - Alternative validation:
  - Definition of Done:
    - behavior implemented
    - tests green (or approved exception)
    - no known regression
  - _Prompt: Role: [specialist] | Task: Implement 1.1 according to approved design and requirements | Restrictions: do not expand scope, respect architectural boundaries | Success: full DoD and green verification_

- [ ] 1.2 [Task title]
  - Wave: 1
  - Depends On: none
  - Can Run In Parallel With: 1.1, 1.3
  - Files:
    - modify:
    - test:
  - _Requirements: REQ-002_
  - TDD: inherit
  - Test Plan:
    - Unit:
  - Verification Command:
    -
  - Definition of Done:
    - define measurable outcome

- [ ] 2.1 [Task title]
  - Wave: 2
  - Depends On: 1.1, 1.2
  - Can Run In Parallel With: none
  - Files:
    - modify:
    - test:
  - _Requirements: REQ-001, REQ-002_
  - TDD: inherit
  - Test Plan:
    - Integration:
  - Verification Command:
    -
  - Definition of Done:
    - define measurable outcome
