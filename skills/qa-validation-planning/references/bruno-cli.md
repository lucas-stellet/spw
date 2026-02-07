# Bruno CLI Reference

Source of truth:
- https://docs.usebruno.com/bru-cli/commandOptions
- https://docs.usebruno.com/bru-cli/runCollection
- https://docs.usebruno.com/bru-cli/generate-reports

## Core Command

- `bru run`

## Key Options For QA

Environment/data:
- `--env`
- `--env-file`
- `--env-var key=value`
- `--csv-file-path`
- `--iteration-count`

Scope control:
- `--folder`
- `--exclude-folder`
- `--tags`
- `--exclude-tags`
- `--tests-only`

Execution behavior:
- `--bail`
- `--no-fail`
- `--delay`
- `--insecure`
- `--parallel`
- `--parallelism`
- `--sandbox [safe|developer]`

Reports:
- `--reporter-junit`
- `--reporter-json`
- `--reporter-html`
- `--junit-filename`
- `--json-filename`
- `--html-filename`
- `--output`

## Important Behavior

As documented for CLI v2+, safe mode is the default and local environment variables are restricted unless explicitly passed (`--env`, `--env-file`, `--env-var`) or sandbox mode is changed.

## QA Planning Guidance

Prefer Bruno CLI when confidence depends on:
- API contract/status/body validation
- auth and permission matrices
- deterministic collection-based regression
- CI-friendly machine-readable reports

Define in plan:
- env matrix and secrets policy
- collection/folder/tag selection
- pass/fail policy (`--bail` or full-run)
- report artifacts and storage path
