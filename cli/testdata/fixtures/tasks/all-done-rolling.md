---
spec: test-all-done-rolling
task_ids: [1, 2]
generation_strategy: rolling-wave
---

# Tasks: test-all-done-rolling

## Wave Plan

- Wave 1: Tasks 1, 2

## Tasks

- [x] 1 Only wave task A
  Wave: 1
  Depends On: none
  Files: `a.ts`

- [x] 2 Only wave task B
  Wave: 1
  Depends On: none
  Files: `b.ts`
