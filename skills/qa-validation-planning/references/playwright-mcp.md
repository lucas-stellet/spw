# Playwright MCP Reference

Source of truth:
- https://github.com/microsoft/playwright-mcp

## Install/Configure

Basic server command:
- `npx @playwright/mcp@latest`

Common args (official README):
- `--browser <chrome|firefox|webkit|msedge>`
- `--headless`
- `--isolated`
- `--output-dir <path>`
- `--save-trace`
- `--save-session`
- `--caps <vision,pdf,devtools>`

## SPW QA Runtime Policy

For `spw:qa` browser validations:
- always run with `--headless`
- do not switch to headed mode

## What It Is Good For

- browser-native flow validation
- realistic user actions and navigation
- multi-tab/session checks
- UI-state evidence via snapshots/traces
- network-aware debugging in browser context

## QA Planning Guidance

Prefer Playwright MCP when confidence depends on:
- client rendering and interaction state
- end-to-end paths through UI
- timing/async behavior visible in browser

Define in plan:
- scenario sequence
- test data per scenario
- expected evidence (snapshot, trace, console/network signals)
- failure triage notes
