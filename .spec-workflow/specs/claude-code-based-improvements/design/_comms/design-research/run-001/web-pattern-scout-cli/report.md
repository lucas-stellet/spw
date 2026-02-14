# Web Pattern Scout: CLI Validation Patterns

## 1. Cobra Nested Subcommand Pattern

The standard Cobra pattern for validation subcommands uses a parent command group with child commands:

```go
// validate_cmd.go
func newValidateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "validate",
        Short: "Validate SPW artifacts and contracts",
    }
    cmd.AddCommand(newValidatePromptsCmd())
    cmd.AddCommand(newValidateStatusCmd())
    return cmd
}
```

This is exactly how `tools`, `hook`, `spec`, `tasks`, and `wave` are structured in the existing codebase. The pattern is well-established.

## 2. Dual Output Mode (JSON vs Human-Readable)

Two established patterns:

### Pattern A: --json flag (used by kubectl, terraform, gh)
```go
cmd.Flags().Bool("json", false, "Output in JSON format")
```
Output switches between a formatted table/list and structured JSON. This is the pattern specified in REQ-001 (`spw validate prompts --json`).

### Pattern B: --raw flag (used by existing SPW tools)
The existing SPW tools use `--raw` for non-JSON output. The `Output(result, summary, raw)` function handles this.

**Recommendation:** Use `--json` as specified in REQ-001, but internally map it to the existing `Output()` infrastructure by inverting the flag (json=true means raw=false). The JSON output schema should include `ok`, `summary`, `violations[]`, and `stats` fields.

## 3. Exit Code Conventions

Standard conventions for validation CLIs:

| Exit Code | Meaning | Examples |
|-----------|---------|---------|
| 0 | All validations pass | shellcheck, golangci-lint |
| 1 | Validation violations found | shellcheck, eslint |
| 2 | Tool/runtime error (bad args, file not found) | Most tools |

The existing SPW hooks use 0 (ok) and 2 (block). For `spw validate`, the recommendation is:
- **0** = all pass
- **1** = violations found (deterministic, expected)
- **2** = runtime error (unexpected)

This distinction matters for CI integration where exit code 1 means "fixable issues" and 2 means "tool broken".

## 4. Table-Driven Test Patterns for Validators

The Go community standard for validator tests:

```go
func TestValidateFrontmatter(t *testing.T) {
    tests := []struct {
        name       string
        content    string
        wantOk     bool
        wantErrors []string
    }{
        {
            name:    "valid with all fields",
            content: "---\nname: spw:exec\ndescription: ...\nargument-hint: ...\nallowed-tools: [...]\nmodel: sonnet\n---\n",
            wantOk:  true,
        },
        {
            name:       "missing model",
            content:    "---\nname: spw:exec\ndescription: ...\n---\n",
            wantOk:     false,
            wantErrors: []string{"missing required field: model"},
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // write temp file, run validator, check result
        })
    }
}
```

This matches the existing SPW test patterns (see `dispatch_test.go`, `hook_test.go`).

## 5. CI Integration Patterns

For `--json` output suitable for CI, the standard structure is:

```json
{
  "ok": false,
  "summary": "3 violations in 2 files",
  "violations": [
    {
      "file": "commands/spw/exec.md",
      "field": "model",
      "message": "missing required field",
      "severity": "error"
    }
  ],
  "stats": {
    "files_checked": 13,
    "files_passed": 11,
    "files_failed": 2,
    "violations_total": 3
  }
}
```
