# Tasks Document (TDD ON)

## Execution Constraints
- max_tasks_per_wave: 3
- require_test_per_task: true
- allow_no_test_exception: true
- tdd_default: on

## Wave Plan
- Wave 1:
- Wave 2:

---

- [ ] 1.1 [Titulo da tarefa]
  - Wave: 1
  - Depends On: none
  - Can Run In Parallel With: 1.2, 1.3
  - Files:
    - modify: path/to/file.ex
    - test: test/path/to/file_test.exs
  - Requirements: REQ-001
  - TDD: required
  - Test Plan:
    - Unit:
    - Integration (se aplicavel):
  - Verification Command:
    - mix test test/path/to/file_test.exs
  - No-TDD Justification (somente se TDD: skip):
    - Reason:
    - Alternative validation:
  - Definition of Done:
    - [ ] comportamento implementado
    - [ ] evidencias RED->GREEN registradas
    - [ ] testes em verde
    - [ ] sem regressao conhecida
  - _Prompt: Role: [especialista] | Task: Implementar 1.1 conforme design e requirements aprovados | Restrictions: seguir TDD estrito (RED->GREEN->REFACTOR), nao aumentar escopo | Success: DoD completo e verificacao em verde_

- [ ] 1.2 [Titulo da tarefa]
  - Wave: 1
  - Depends On: none
  - Can Run In Parallel With: 1.1, 1.3
  - Files:
    - modify:
    - test:
  - Requirements: REQ-002
  - TDD: required
  - Test Plan:
    - Unit:
  - Verification Command:
    - 
  - Definition of Done:
    - [ ]

- [ ] 2.1 [Titulo da tarefa]
  - Wave: 2
  - Depends On: 1.1, 1.2
  - Can Run In Parallel With: none
  - Files:
    - modify:
    - test:
  - Requirements: REQ-001, REQ-002
  - TDD: required
  - Test Plan:
    - Integration:
  - Verification Command:
    - 
  - Definition of Done:
    - [ ]
