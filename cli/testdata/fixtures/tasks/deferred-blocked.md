---
spec: test-deferred-blocked
task_ids: [1, 2, 3]
approval_id: def-456
generation_strategy: rolling-wave
---

# Tasks: test-deferred-blocked

## Wave Plan

- Wave 1: Tasks 1, 2
- Wave 2: Tasks 3

## Tasks

- [x] 1 Initial setup
  Wave: 1
  Depends On: none
  Files: `setup.ts`

- [x] 2 Core feature
  Wave: 1
  Depends On: Task 1
  Files: `core.ts`

- [ ] 3 Main feature
  Wave: 2
  Depends On: Task 2
  Files: `feature.ts`

## Deferred Backlog

- [ ] 5 Enhancement that depends on unfinished task
  Wave: 3
  Depends On: Task 3
  Files: `enhance.ts`
