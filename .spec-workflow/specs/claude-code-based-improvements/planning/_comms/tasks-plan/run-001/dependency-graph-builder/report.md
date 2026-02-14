# Dependency Graph Report

## DAG

```
Task 1 (schema types)     -> no deps
Task 5 (config extension) -> no deps

Task 2 (prompts.go)       -> Task 1
Task 3 (mirror.go)        -> Task 1
Task 4 (status.go)        -> Task 1
Task 6 (audit.go)         -> Task 1, Task 5
Task 7 (iteration.go)     -> Task 1

Task 8 (validate_cmd.go)  -> Task 2, Task 3
Task 9 (dispatch_status)  -> Task 4, Task 6

Task 10 (documentation)   -> Task 8, Task 9
Task 11 (mirror sync)     -> Task 8, Task 10
```

## Wave Assignment

### Wave 1 (executable, max_wave_size=3)
- **Task 1**: Create validate package foundation (schema.go)
- **Task 2**: Implement frontmatter validation logic (prompts.go + tests)
- **Task 5**: Extend config with AuditConfig and iteration limits

**Rationale:** Tasks 1 and 5 are the two foundation tasks with zero dependencies. Task 2 is the highest-value validator (REQ-001, the core deliverable) and depends only on Task 1. Since Task 1 is small (types only), Tasks 1+2 can be treated as a unit where Task 2 starts after Task 1 completes within the same wave. Task 5 is fully independent. This gives us the maximum unblocking: after Wave 1, Tasks 3, 4, 6, 7 are all unblocked.

**Parallelism within Wave 1:**
- Task 1 and Task 5: fully parallel (no shared files)
- Task 2: starts after Task 1 completes (depends on schema types)

### Wave 2 (deferred)
- Task 3: Mirror validation (depends on Task 1)
- Task 4: Enhanced status.json validation (depends on Task 1)
- Task 7: Iteration limit logic (depends on Task 1)

**Note:** Task 6 (audit gate) depends on both Task 1 and Task 5, both completed in Wave 1. Could fit in Wave 2 but max_wave_size=3 is already filled by 3, 4, 7.

### Wave 3 (deferred)
- Task 6: Audit confidence gate (depends on Task 1, Task 5)
- Task 8: Cobra validate command wiring (depends on Task 2, Task 3)
- Task 9: Dispatch status integration (depends on Task 4, Task 6)

**Note:** Task 8 depends on Wave 2 (Task 3). Task 9 depends on Task 6 from this same wave. This means Task 9 cannot start until Task 6 finishes. Replan at next-wave may split these.

### Wave 4 (deferred)
- Task 10: Documentation updates
- Task 11: Mirror sync and embedded assets

## Critical Path
Task 1 -> Task 4 -> Task 9 -> Task 10 -> Task 11 (longest chain through status validation + dispatch integration + docs + mirror sync)

Alternative: Task 1 -> Task 2 -> Task 8 -> Task 10 -> Task 11 (through prompts + CLI wiring)
