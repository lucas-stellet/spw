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

# Tasks Document

## Execution Constraints
- max_tasks_per_wave: 3
- require_test_per_task: true
- allow_no_test_exception: true
- tdd_default: managed-by-config

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
  - Implementation:
    -
  - Test Plan:
    - Unit:
    - Integration (if applicable):
  - Verification Command:
    - mix test test/path/to/file_test.exs
  - _Requirements: REQ-001_
  - TDD: inherit
  - Definition of Done:
    - behavior implemented
    - tests passing
    - no known regression
  - _Prompt: Role: [task specialist] | Task: Implement task 1.1 for spec [spec-name] according to approved requirements and design | Restrictions: preserve architectural boundaries, no scope expansion, follow TDD | Success: DoD criteria met and verification command passes_

- [ ] 1.2 [Task title]
  - Wave: 1
  - Depends On: none
  - Can Run In Parallel With: 1.1, 1.3
  - Files:
    - modify:
    - test:
  - Implementation:
    -
  - Test Plan:
    - Unit:
  - Verification Command:
    -
  - _Requirements: REQ-002_
  - TDD: inherit
  - No-Test Justification (only for exception):
    - Reason:
    - Alternative validation:
  - Definition of Done:
    - define measurable outcome

- [ ] 2.1 [Task title]
  - Wave: 2
  - Depends On: 1.1, 1.2
  - Can Run In Parallel With: none
  - Files:
    - modify:
    - test:
  - Implementation:
    -
  - Test Plan:
    - Integration:
  - Verification Command:
    -
  - _Requirements: REQ-001, REQ-002_
  - TDD: inherit
  - Definition of Done:
    - define measurable outcome
