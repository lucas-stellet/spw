# Web Pattern Scout: Frontmatter Parsing in Go

## 1. Library Options

### Option A: Manual YAML Frontmatter Parsing (Recommended)

Since SPW commands use simple `---` delimited YAML frontmatter with flat key-value pairs, a lightweight manual parser avoids new dependencies:

```go
func ParseFrontmatter(content string) (map[string]string, string, error) {
    if !strings.HasPrefix(content, "---\n") {
        return nil, content, fmt.Errorf("no frontmatter delimiter")
    }
    end := strings.Index(content[4:], "\n---")
    if end == -1 {
        return nil, content, fmt.Errorf("unclosed frontmatter")
    }
    fmBlock := content[4 : 4+end]
    body := content[4+end+4:]

    fields := make(map[string]string)
    for _, line := range strings.Split(fmBlock, "\n") {
        k, v, ok := parseKeyValue(line) // reuse existing pattern from registry.go
        if ok {
            fields[k] = v
        }
    }
    return fields, body, nil
}
```

**Pros:** Zero new dependencies, reuses existing `parseKeyValue` pattern from `registry.go`, handles the simple flat YAML used in commands.
**Cons:** Does not handle nested YAML, arrays, or complex types.

### Option B: go-yaml/yaml.v3 (Existing in Go ecosystem)

```go
import "gopkg.in/yaml.v3"

type CommandFrontmatter struct {
    Name         string   `yaml:"name"`
    Description  string   `yaml:"description"`
    ArgumentHint string   `yaml:"argument-hint"`
    AllowedTools []string `yaml:"allowed-tools"`
    Model        string   `yaml:"model"`
}
```

**Pros:** Full YAML support, handles arrays for `allowed-tools`, type-safe struct decoding.
**Cons:** New dependency (yaml.v3). SPW currently only uses `BurntSushi/toml` and `spf13/cobra`.

### Option C: goldmark-frontmatter (go.abhg.dev/goldmark/frontmatter)

Full goldmark integration that parses YAML and TOML frontmatter within a markdown parsing pipeline.

**Pros:** Feature-rich, supports both YAML and TOML.
**Cons:** Heavy dependency (pulls in full goldmark parser), overkill for flat frontmatter validation.

### Option D: adrg/frontmatter (github.com/adrg/frontmatter)

Standalone library that detects and decodes frontmatter with pluggable format support.

**Pros:** Lightweight, supports YAML/TOML/JSON detection.
**Cons:** Still an external dependency.

## 2. Recommendation

**Option A (manual parser)** for initial implementation. The command frontmatter is flat key-value pairs. The `allowed-tools` field could use a comma-separated string format that the manual parser handles:

```yaml
allowed-tools: "Read, Grep, Glob, Bash, WebFetch, WebSearch"
```

Or, if array syntax is needed for future extensibility, adopt **Option B (go-yaml)** which is a small, well-established dependency.

**Decision factor:** If `allowed-tools` must be a YAML array `[Read, Grep, ...]`, then go-yaml is justified. If it can be a comma-separated string, the manual parser suffices.

## 3. Validation Schema Pattern

Regardless of parser choice, the validation logic should be schema-driven:

```go
type FieldRule struct {
    Name     string
    Required bool
    Type     string // "string", "string-list", "enum"
    Enum     []string // valid values for enum type
}

var CommandFrontmatterSchema = []FieldRule{
    {Name: "name", Required: true, Type: "string"},
    {Name: "description", Required: true, Type: "string"},
    {Name: "argument-hint", Required: true, Type: "string"},
    {Name: "allowed-tools", Required: true, Type: "string-list"},
    {Name: "model", Required: true, Type: "enum", Enum: []string{"haiku", "sonnet", "opus"}},
}
```

This schema-driven approach means adding new required fields is a one-line change.
