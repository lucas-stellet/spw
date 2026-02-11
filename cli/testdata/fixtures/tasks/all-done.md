---
spec: test-all-done
task_ids: [1, 2, 3]
generation_strategy: all-at-once
---

# Tasks: test-all-done

## Wave Plan

- Wave 1: Tasks 1, 2
- Wave 2: Task 3

## Tasks

- [x] 1 First task
  Wave: 1
  Depends On: none
  Files: `first.ts`

- [x] 2 Second task
  Wave: 1
  Depends On: Task 1
  Files: `second.ts`

- [x] 3 Third task
  Wave: 2
  Depends On: Task 2
  Files: `third.ts`
