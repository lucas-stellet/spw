# File-First Handoff Contract

Required files for each dispatched subagent:
- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

If any required handoff file is missing, return `BLOCKED`.

Run-id format: `run-NNN` (zero-padded 3-digit sequential, e.g. `run-001`, `run-002`).
To create a new run, scan existing sibling directories, extract the highest NNN, and increment by 1.
