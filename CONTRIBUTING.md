# Contributing to Oraculo

Thanks for your interest in contributing to Oraculo.

## Prerequisites

- [Claude Code](https://claude.ai/code) CLI installed
- [spec-workflow-mcp](https://github.com/lucas-stellet/spec-workflow-mcp) for approval gates
- Node.js (for hook scripts)
- Bash (for CLI and install scripts)

## Getting started

1. Fork and clone the repository.
2. Run validation to make sure everything passes before making changes:

```bash
# Validate all shell scripts parse correctly
bash -n bin/oraculo
bash -n scripts/bootstrap.sh
bash -n scripts/install-oraculo-bin.sh
bash -n scripts/validate-kit.sh
bash -n claude-kit/install.sh

# Validate kit structure (wrapper sizes, workflow refs, overlays)
scripts/validate-kit.sh

# Smoke-test Go hooks (build first: cd cli && go build -o /tmp/oraculo ./cmd/oraculo && PATH="/tmp:$PATH")
echo '{"workspace":{"current_dir":"'"$(pwd)"'"}}' | oraculo hook statusline
echo '{"prompt":"/oraculo:plan"}' | oraculo hook guard-prompt
echo '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":"README.md"}}' | oraculo hook guard-paths
echo '{}' | oraculo hook guard-stop
echo '{}' | oraculo hook session-start
```

## Source of truth

`claude-kit/` is the single source of truth for all user-facing content (commands, workflows, config, templates, skills). The validation script checks structural integrity:

```bash
scripts/validate-kit.sh
```

## Documentation updates

When modifying behavior, defaults, or guardrails, update these files in the same patch:

- `README.md`
- `AGENTS.md`
- `docs/ORACULO-WORKFLOW.md`
- `claude-kit/README.md`

## Code style

- Shell scripts: validate with `bash -n` before committing.
- Go hooks: implemented in `cli/internal/hook/`, invoked via `oraculo hook <event>`.
- Workflows/commands: follow the thin-orchestrator pattern (max 60 lines for command wrappers).

## Submitting changes

1. Create a branch from `main`.
2. Make your changes.
3. Run the full validation suite (see above).
4. Open a pull request with a clear description of the change.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
