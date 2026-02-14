# Report: Task 2 - Implement frontmatter validation logic with yaml.v3

## Summary

Successfully implemented the frontmatter validation logic for SPW command files using yaml.v3. The implementation validates all 5 required frontmatter fields (name, description, argument-hint, allowed-tools, model) against a defined schema.

## Implementation Details

### Files Created

1. **`cli/internal/validate/prompts.go`** - Main validation logic
   - `ValidatePrompts(dir string)` - Scans a directory for .md files and validates frontmatter
   - `validatePromptFile(filePath string)` - Validates a single command file
   - `extractFrontmatter(content string)` - Parses YAML frontmatter from markdown content
   - Uses `gopkg.in/yaml.v3` for YAML parsing

2. **`cli/internal/validate/prompts_test.go`** - Comprehensive test coverage
   - Table-driven tests covering all test cases
   - 21 test scenarios including:
     - Valid frontmatter passes
     - Missing each required field produces violation
     - Invalid model enum value
     - No frontmatter delimiter
     - Empty file
     - Malformed YAML
     - Extra unknown fields tolerated
     - Multiple missing fields
     - Wrong type for allowed-tools
     - Empty allowed-tools array
     - All valid model enum values (haiku, sonnet, opus)
   - Golden file test for JSON output format
   - Integration test against real command files

### Field Rules Defined

All 5 required fields are enforced:
- `name` (string, required)
- `description` (string, required)
- `argument-hint` (string, required)
- `allowed-tools` (string_array, required)
- `model` (enum: haiku | sonnet | opus, required)

### Test Results

All tests pass:
- 21 table-driven test cases
- Golden file JSON output validation
- Integration test on real commands (detects missing allowed-tools and model fields - expected per REQ-001)

```
=== RUN   TestValidatePrompts
--- PASS: TestValidatePrompts (0.01s)
=== RUN   TestValidatePromptsGoldenFile
--- PASS: TestValidatePromptsGoldenFile (0.00s)
=== RUN   TestValidatePromptsOnRealCommands
--- PASS: TestValidatePromptsOnRealCommands (0.00s)
PASS
ok  	github.com/lucas-stellet/spw/internal/validate	0.404s
```

### Verification Command

```bash
go test ./cli/internal/validate/ -run TestValidatePrompts -v
```

### Dependencies Added

- `gopkg.in/yaml.v3` - For YAML frontmatter parsing

### Notes

- The validation correctly identifies that existing command files in `commands/spw/` are missing the `allowed-tools` and `model` fields - this is expected behavior as these fields are part of the new contract defined in REQ-001
- The validate package follows the restriction of not importing tools/hook/cli packages
- Uses the schema types from Task 1 (schema.go) as required
