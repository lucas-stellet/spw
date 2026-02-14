# Brief: spec-compliance-reviewer-3

## Inputs
- Implementation: cli/internal/validate/mirror.go
- Tests: cli/internal/validate/mirror_test.go
- Task Definition: .spec-workflow/specs/claude-code-based-improvements/tasks.md (Task 3)
- Design: .spec-workflow/specs/claude-code-based-improvements/design.md

## Config Context
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
**Review Task 3: Mirror and embedded asset validation**

Verify exact adherence to requirements:
- REQ-002: Mirror validation
- REQ-007: Validation infrastructure

### Review Criteria
1. ValidateMirrors(rootDir) compares source vs copy-ready using SHA-256
2. Overlay symlink targets validated (noop.md or teams/*.md)
3. Embedded asset comparison via embedded.Workflows.ReadFile
4. All test cases from tasks.md covered:
   - Matching content hash
   - Divergent content detected
   - Missing mirror file
   - Extra files in mirror
   - Broken symlinks
   - Symlink target validation
   - Embedded vs filesystem comparison

### Verification Command
go test ./cli/internal/validate/ -run TestValidateMirrors -v

## Output Contract
Write your output to:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/spec-compliance-reviewer-3/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-02/execution/run-001/spec-compliance-reviewer-3/status.json

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": [],
  "skills_missing": [],
  "model_override_reason": null
}
```
