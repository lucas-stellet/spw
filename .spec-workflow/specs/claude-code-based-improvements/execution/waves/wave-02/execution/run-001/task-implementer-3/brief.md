# Brief: task-implementer-3

## Inputs
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md
- Tasks: .spec-workflow/specs/claude-code-based-improvements/tasks.md
- Existing Schema: cli/internal/validate/schema.go
- Mirror Mapping: See CLAUDE.md mirror table (commands/spw -> copy-ready/.claude/commands/spw, workflows/spw -> copy-ready/.claude/workflows/spw)

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
**Task 3: Implement mirror and embedded asset validation**

Files to create:
- cli/internal/validate/mirror.go
- cli/internal/validate/mirror_test.go

Requirements: REQ-002, REQ-007

### Test Cases Required
- Matching content hash
- Divergent content detected
- Missing mirror file
- Extra files in mirror
- Broken symlinks
- Symlink target validation (noop.md or teams/*.md)
- Embedded vs filesystem comparison

### Verification Command
go test ./cli/internal/validate/ -run TestValidateMirrors -v

### Definition of Done
- ValidateMirrors(rootDir) compares source vs copy-ready using SHA-256
- Overlay symlink targets validated
- Embedded asset comparison via embedded.Workflows.ReadFile
- All test cases pass

## Output Contract
Write your output to:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/task-implementer-3/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/task-implementer-3/status.json

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["effective-go"],
  "skills_missing": [],
  "model_override_reason": null
}
```
