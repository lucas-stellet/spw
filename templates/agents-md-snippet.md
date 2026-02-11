<!-- SPW-KIT-START — managed by spw install, do not edit manually -->
## SPW Dispatch Rules

When executing SPW workflow commands, follow these rules strictly:

1. **Always use CLI for dispatch.** Never create run dirs or subagent dirs manually. Use `spw tools dispatch-init` and `spw tools dispatch-setup`.
2. **Status-only reads.** After dispatching a subagent, read ONLY status.json via `spw tools dispatch-read-status`. Never read report.md unless status=blocked.
3. **Paths, not content.** When subagent-B depends on subagent-A, write the filesystem PATH to A's report.md in B's brief.md. Never copy content.
4. **Synthesizer reads from disk.** The final subagent (synthesizer/aggregator) receives a brief listing all report paths and reads them directly.
5. **MCP inline exception.** When a subagent needs session-scoped MCP tools (Linear, Playwright), run dispatch-setup normally but execute the work inline — still write report.md and status.json to the subagent directory.
6. **Always finalize.** Call `spw tools dispatch-handoff --run-dir <dir>` after all subagents complete.
<!-- SPW-KIT-END -->
