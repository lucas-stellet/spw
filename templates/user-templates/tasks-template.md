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

- [ ] 1.1 [Título da tarefa]
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
    - Integration (se aplicável):
  - Verification Command:
    - mix test test/path/to/file_test.exs
  - Requirements: REQ-001
  - TDD: inherit
  - Definition of Done:
    - [ ] comportamento implementado
    - [ ] testes passando
    - [ ] sem regressão conhecida
  - _Prompt: Role: [especialista da tarefa] | Task: Implementar tarefa 1.1 para a spec [spec-name] conforme requirements e design aprovados | Restrictions: manter boundaries arquiteturais, não criar escopo extra, seguir TDD | Success: critérios de DoD atendidos e comando de verificação em verde_

- [ ] 1.2 [Título da tarefa]
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
  - Requirements: REQ-002
  - TDD: inherit
  - No-Test Justification (somente se exceção):
    - Reason:
    - Alternative validation:
  - Definition of Done:
    - [ ]

- [ ] 2.1 [Título da tarefa]
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
  - Requirements: REQ-001, REQ-002
  - TDD: inherit
  - Definition of Done:
    - [ ]
