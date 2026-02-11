---
spec: test-mixed-deps
task_ids: [1, 2, 3, 4, 5]
generation_strategy: all-at-once
---

# Tasks: test-mixed-deps

## Wave Plan

- Wave 1: Tasks 1, 2
- Wave 2: Tasks 3, 4, 5

## Tasks

- [x] 1 Setup A
  Wave: 1
  Depends On: none
  Files: `a.ts`

- [x] 2 Setup B
  Wave: 1
  Depends On: none
  Files: `b.ts`

- [ ] 3 Depends on A only
  Wave: 2
  Depends On: Task 1
  Files: `c.ts`

- [ ] 4 Depends on both A and B
  Wave: 2
  Depends On: Task 1, Task 2
  Files: `d.ts`

- [ ] 5 Depends on unfinished 3
  Wave: 2
  Depends On: Task 3
  Files: `e.ts`
