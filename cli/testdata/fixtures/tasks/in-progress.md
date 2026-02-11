---
spec: test-in-progress
task_ids: [1, 2, 3, 4]
generation_strategy: all-at-once
---

# Tasks: test-in-progress

## Wave Plan

- Wave 1: Tasks 1, 2
- Wave 2: Tasks 3, 4

## Tasks

- [x] 1 Setup
  Wave: 1
  Depends On: none
  Files: `setup.ts`

- [x] 2 Base feature
  Wave: 1
  Depends On: Task 1
  Files: `base.ts`

- [-] 3 In progress work
  Wave: 2
  Depends On: Task 2
  Files: `work.ts`

- [ ] 4 Remaining work
  Wave: 2
  Depends On: Task 2
  Files: `remaining.ts`
