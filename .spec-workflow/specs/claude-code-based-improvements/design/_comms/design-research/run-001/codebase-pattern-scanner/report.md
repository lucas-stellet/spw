# Codebase Pattern Scanner Report

## 1. Existing Validation Patterns

### dispatch_status.go — Status Validation
Current `DispatchReadStatus` validates only two fields: `status` (must be "pass" or "blocked") and `summary`. It does NOT validate extended fields (`skills_used`, `skills_missing`, `model_override_reason`) even though `dispatch_setup.go` documents them in the brief skeleton.

**Gap identified:** The brief template in `dispatch_setup.go` (lines 76-82) lists all 5 fields in the output contract, but `dispatch_status.go` only checks the first 2. This is exactly the gap REQ-003 targets.

### registry.go — Frontmatter-like Parsing
The registry already parses `<dispatch_pattern>` blocks from workflow `.md` files using a simple key-value parser (`parseKeyValue`). It extracts: `category`, `subcategory`, `phase`, `comms_path`, `artifacts`, `policy`.

**Reuse opportunity:** The same parsing pattern (scan lines, extract `key: value` pairs) can be adapted for YAML frontmatter, but YAML frontmatter uses `---` delimiters which is a different format from XML-style tags.

### validate-thin-orchestrator.sh — Mirror Validation
The bash script already validates:
- Wrapper size (max 60 lines)
- `<execution_context>` presence
- Workflow path reference
- No detailed orchestration blocks in wrappers
- Mirror integrity via `diff -rq`
- Symlink targets

**Reuse opportunity:** REQ-002 should port this logic to Go for the `--strict` mode, making it part of the same binary. The existing bash script can remain as a compatibility layer or be deprecated.

### hook implementations — Guard Pattern
All hooks follow the same contract: read JSON from stdin, process, exit 0 (ok) or 2 (block). The `HooksConfig` struct shows how to add per-guard toggles.

**Reuse opportunity:** The audit confidence gate (REQ-004) should follow this same pattern. New config fields go in `config.go`, new guard logic in a dedicated file.

## 2. Command Frontmatter Structure

Current frontmatter across all 13 commands has only 3 fields:
```yaml
---
name: spw:<command>
description: <text>
argument-hint: "<args>"
---
```

REQ-001 requires adding `allowed-tools` and `model` fields. These do NOT currently exist in any command file.

**Important:** The `name` field uses the `spw:` prefix. The `argument-hint` field is a string with angle-bracket placeholders. No command currently has `allowed-tools` or `model`.

## 3. Embedded Asset Architecture

`embed.go` uses 7 separate `go:embed` directives for different asset categories (Workflows, Shared, Dispatch, Overlays, Stubs, Defaults, Snippets). The `CompositeFS` type routes reads by path prefix.

**For REQ-002:** Validation of embedded-vs-filesystem consistency requires reading from `embedded.Assets()` and comparing against the filesystem `commands/`, `workflows/`, `copy-ready/` directories. The `CompositeFS.ReadFile` method provides the embedded side; filesystem reads provide the other.

**Key files that must match:**
- `cli/internal/embedded/workflows/*.md` vs `workflows/spw/*.md`
- `cli/internal/embedded/shared/*.md` vs `workflows/spw/shared/*.md`
- `cli/internal/embedded/dispatch/*.md` vs `workflows/spw/shared/dispatch-*.md` (note path difference)
- `cli/internal/embedded/overlays/*.md` vs `workflows/spw/overlays/teams/*.md`

## 4. Config Struct Patterns

Adding a new `[audit]` section follows the established pattern in `config.go`:
1. Define a new struct type (e.g., `AuditConfig`)
2. Add it to the `Config` struct with a `toml:"audit"` tag
3. Add defaults in `Defaults()`
4. Add TOML comments in `spw-config.toml`

The `Config.GetValue` method uses reflection to resolve dot-path keys, so new sections are automatically accessible via `spw tools config-get audit.audit_min_confidence`.

## 5. Test Infrastructure

Existing tests use:
- Standard `testing` package (no testify/assert)
- Table-driven tests extensively (`TestFirstSpwCommand`, `TestCommsPathGeneration`, etc.)
- `t.TempDir()` for filesystem-dependent tests
- Helper functions (`writeJSON`, `contains`, `searchString`)
- Direct file I/O for test fixtures (no mocking frameworks)

**Pattern to follow:** New tests for validators should use the same table-driven approach with `t.TempDir()` for creating temporary command files with various frontmatter configurations.

## 6. CLI Command Registration

New subcommands are added in `root.go` via `cmd.AddCommand(newXxxCmd())`. The `tools.go` file shows the pattern for subcommand groups. A new `validate` command group would follow the same structure as `tools`, `hook`, `spec`, etc.

**Recommended location:** `cli/internal/cli/validate_cmd.go` for the Cobra command wiring, `cli/internal/validate/` for the validation logic.

## 7. Output Formatting

The `tools/output.go` provides `Output(result, summary, raw)` for JSON vs raw output. The `raw` flag controls whether output is JSON-wrapped or plain text. This same pattern should be used for `spw validate prompts [--json]`.
