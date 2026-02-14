# Brief: task-implementer-2

## Inputs
<!-- Fill file paths here — PATHS ONLY, never paste content -->
- .spec-workflow/specs/claude-code-based-improvements/tasks.md
- .spec-workflow/specs/claude-code-based-improvements/design.md
- .spec-workflow/specs/claude-code-based-improvements/requirements.md
- cli/internal/validate/schema.go (created by Task 1)

## Config Context
<!-- Auto-resolved from spw-config.toml by dispatch-setup -->
- tdd_default: off
- max_wave_size: 3
- require_test_per_task: true
- allow_no_test_exception: true

## Task
<!-- Describe what this subagent must do -->

### Task 2: Implement frontmatter validation logic with yaml.v3

**Files to create:**
- `cli/internal/validate/prompts.go`
- `cli/internal/validate/prompts_test.go`

**Requirements:** REQ-001, REQ-007

**Depends on:** Task 1 (validate package foundation)

**Definition of Done:**
- ValidatePrompts(dir) scans commands/spw/*.md and returns structured ValidationResult
- All 5 required fields enforced (name, description, argument-hint, allowed-tools, model)
- yaml.v3 used for frontmatter parsing
- Table-driven tests cover all cases in test matrix
- Golden file test for JSON stability

**Verification Command:**
```
go test ./cli/internal/validate/ -run TestValidatePrompts -v
```

**Restrictions:**
- Do not wire Cobra command (Task 8)
- Do not modify registry.go parseKeyValue
- Validate package must not import tools/hook/cli
- Use the schema types from Task 1 (schema.go)

**Test Plan (required):**
- Unit: Valid frontmatter passes. Missing each required field (name, description, argument-hint, allowed-tools, model) produces violation. Invalid model enum value. No frontmatter delimiter. Empty file. Malformed YAML. Extra unknown fields tolerated. Golden file test for JSON output format.

Execute the task per the Definition of Done. Write comprehensive table-driven tests. Ensure all test cases pass.

## Output Contract
Write your output to these exact paths:
- Report: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer-2/report.md
- Status: .spec-workflow/specs/claude-code-based-improvements/execution/waves/wave-01/execution/run-002/task-implementer-2/status.json

status.json format:
```json
{
  "status": "pass | blocked",
  "summary": "one-line description",
  "skills_used": ["skill-name"],
  "skills_missing": [],
  "model_override_reason": null
}
```
