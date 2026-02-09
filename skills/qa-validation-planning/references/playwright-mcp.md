# Playwright MCP Reference

Source of truth:
- https://github.com/microsoft/playwright-mcp

## Setup

Playwright MCP is a pre-configured MCP server. Register it once before starting a session:

```bash
claude mcp add playwright -- npx @playwright/mcp@latest --headless --isolated
```

After registration, the `playwright` server exposes browser automation tools that the agent uses directly — no `npx` or shell commands needed during execution.

## How to Use

The server provides tools for browser automation (navigation, clicking, typing, screenshots, etc.). The exact set of available tools may change across versions.

Rules:
- Discover available tools from the `playwright` server at runtime — do not hardcode or assume specific tool names
- Use the server's tools for all browser interactions
- Never invoke `npx`, `node`, or shell scripts for browser automation

## What It Is Good For

- browser-native flow validation
- realistic user actions and navigation
- multi-tab/session checks
- UI-state evidence via screenshots and console messages
- network-aware debugging in browser context

## Evidence Collection

Collect evidence during test execution using the server's tools:
- Take screenshots after key assertions and at failure points
- Capture console messages for runtime error logs
- Generate reproducible test scripts if the server provides a script generation tool

## QA Planning Guidance

Prefer Playwright MCP when confidence depends on:
- client rendering and interaction state
- end-to-end paths through UI
- timing/async behavior visible in browser

Define in plan:
- scenario sequence
- test data per scenario
- expected evidence (screenshots, console messages)
- failure triage notes
