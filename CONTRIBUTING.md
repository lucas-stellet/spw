# Contributing to SPW

Thanks for your interest in contributing to SPW.

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
bash -n bin/spw
bash -n scripts/bootstrap.sh
bash -n scripts/install-spw-bin.sh
bash -n scripts/validate-thin-orchestrator.sh
bash -n copy-ready/install.sh

# Validate thin-orchestrator contract (wrapper sizes, workflow refs, mirror sync)
scripts/validate-thin-orchestrator.sh

# Smoke-test Go hooks (build first: cd cli && go build -o /tmp/spw ./cmd/spw && PATH="/tmp:$PATH")
echo '{"workspace":{"current_dir":"'"$(pwd)"'"}}' | spw hook statusline
echo '{"prompt":"/spw:plan"}' | spw hook guard-prompt
echo '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":"README.md"}}' | spw hook guard-paths
echo '{}' | spw hook guard-stop
echo '{}' | spw hook session-start
```

## Mirror system

Source files must stay in sync with their `copy-ready/` counterparts. Always update both sides in the same patch. The validation script checks this:

```bash
scripts/validate-thin-orchestrator.sh
```

See the mirror table in `CLAUDE.md` for the full mapping.

## Documentation updates

When modifying behavior, defaults, or guardrails, update these files in the same patch:

- `README.md`
- `AGENTS.md`
- `docs/SPW-WORKFLOW.md`
- `hooks/README.md`
- `copy-ready/README.md` (mirror sync)

## Code style

- Shell scripts: validate with `bash -n` before committing.
- Go hooks: implemented in `cli/internal/hook/`, invoked via `spw hook <event>`.
- Workflows/commands: follow the thin-orchestrator pattern (max 60 lines for command wrappers).

## Submitting changes

1. Create a branch from `main`.
2. Make your changes, keeping the mirror in sync.
3. Run the full validation suite (see above).
4. Open a pull request with a clear description of the change.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
