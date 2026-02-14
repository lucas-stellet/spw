# Codebase Analysis: File I/O Inventory for DB Migration

This document catalogs every filesystem read/write operation across `cli/internal/` packages
to inform the SQLite migration. Each entry includes the function, file path pattern, data format,
and migration priority.

## Priority Legend

| Priority | Meaning |
|----------|---------|
| **P0** | Must migrate -- core runtime data (runs, subagents, waves, task state) |
| **P1** | Should migrate -- artifacts (reports, briefs, logs, summaries) |
| **P2** | Can migrate later -- cache files, temp state |
| **KEEP** | Must remain as files -- user-facing Markdown documents |

---

## 1. Package: `specdir` (`cli/internal/specdir/`)

### 1.1 Path Constants (`paths.go`)

Pure path-building functions -- no I/O. These define the canonical paths used by all other packages.
Migration impact: these path constants become the mapping between filesystem layout and DB table/column names.

**Key path functions:**

| Function | Returns | Used By |
|----------|---------|---------|
| `SpecDir(specName)` | `.spec-workflow/specs/<name>` | everywhere |
| `SpecDirAbs(cwd, specName)` | absolute spec dir | everywhere |
| `TasksPath(specDir)` | `<specDir>/tasks.md` | tasks, tools |
| `ImplLogPath(specDir, taskID)` | `<specDir>/execution/_implementation-logs/task-<id>.md` | tools, hook |
| `WavePath(specDir, waveNum)` | `<specDir>/execution/waves/wave-NN` | wave, tools |
| `WaveExecPath(specDir, waveNum)` | `<specDir>/execution/waves/wave-NN/execution` | wave |
| `WaveCheckpointPath(specDir, waveNum)` | `<specDir>/execution/waves/wave-NN/checkpoint` | wave, tasks |
| `WaveSummaryPath(specDir, waveNum)` | `<specDir>/execution/waves/wave-NN/_wave-summary.json` | wave, tools |
| `WaveLatestPath(specDir, waveNum)` | `<specDir>/execution/waves/wave-NN/_latest.json` | wave, tools, tasks |
| `CommsPath(specDir, command, waveNum)` | `<specDir>/<phase>/_comms/<command>` | tools |
| `CheckpointRunPath(specDir, waveNum, runNum)` | `<specDir>/execution/waves/wave-NN/checkpoint/run-NNN` | tools |

### 1.2 I/O Functions (`specdir.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `Resolve(cwd, specName)` | READ (`os.Stat`) | `.spec-workflow/specs/<name>/` | Directory existence | P0 |
| `ListWaveDirs(specDir)` | READ (`os.ReadDir`) | `<specDir>/execution/waves/` | Directory listing, filtered by `wave-NN` pattern | P0 |
| `ReadStatusJSON(path)` | READ (`os.ReadFile`) | `*/status.json` | JSON: `{status, summary}` | P0 |
| `ReadLatestJSON(path)` | READ (`os.ReadFile`) | `*/_latest.json` | JSON: `{run_id, run_dir, status, summary, execution, checkpoint}` | P0 |
| `LatestRunDir(dir)` | READ (`os.ReadDir`) | `*/run-NNN/` | Directory listing, highest `run-NNN` | P0 |
| `FileExists(path)` | READ (`os.Stat`) | any | Existence check | -- |
| `DirExists(path)` | READ (`os.Stat`) | any | Existence check | -- |

---

## 2. Package: `config` (`cli/internal/config/`)

### 2.1 Config Loading (`config.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ResolveConfigPath(root)` | READ (`os.Stat`) | `.spec-workflow/spw-config.toml` or `.spw/spw-config.toml` | Existence check | KEEP |
| `Load(root)` | READ (`os.ReadFile`) | `<ResolveConfigPath>` | TOML | KEEP |
| `LoadFromPath(path)` | READ (`os.ReadFile`) | explicit path | TOML | KEEP |

### 2.2 Config Merge (`merge.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `Merge(template, user, output)` | READ (`os.Open`, `toml.DecodeFile`) + WRITE (`os.WriteFile`) | template TOML, user TOML, output TOML | TOML | KEEP |
| `readLines(path)` | READ (`os.Open`) | any | line-by-line text | KEEP |
| `normalizeToml(path)` | READ (`toml.DecodeFile`) | any TOML | TOML round-trip | KEEP |

**Note:** Config files KEEP as files -- they are user-editable TOML, not runtime state.

---

## 3. Package: `spec` (`cli/internal/spec/`)

### 3.1 Stage Classification (`stage.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ClassifyStage(specDir)` | READ (`specdir.FileExists`, `specdir.DirExists`) | `<specDir>/post-mortem/report.md`, `qa/QA-TEST-PLAN.md`, `qa/QA-CHECK.md`, `qa/QA-EXECUTION-REPORT.md`, `execution/waves/`, `tasks.md`, `design.md`, `design/DESIGN-RESEARCH.md`, `requirements.md` | Existence checks | P0 |
| `allTasksComplete(tasksPath)` | READ (`os.ReadFile`) | `<specDir>/tasks.md` | Markdown: scans for `- [x]` vs `- [ ]` patterns | P0 |

### 3.2 Artifacts (`artifacts.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `CheckArtifacts(specDir)` | READ (`specdir.FileExists`) | 20 canonical artifact paths | Existence map | P1 |
| `detectDeviations(specDir)` | READ (`os.Stat`) | Known deviation paths (6 entries) | Existence check | P2 |

### 3.3 Approval (`approval.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `CheckApproval(cwd, specName, docType)` | READ (`os.ReadDir` + `os.Stat` + `os.ReadFile`) | `.spec-workflow/approvals/<spec>/approval_*.json` | JSON: `{approvalId, filePath, approval.id}` | P0 |

### 3.4 List (`list.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `List(cwd)` | READ (`os.ReadDir`) | `.spec-workflow/specs/` | Directory listing | P0 |

### 3.5 Prerequisites (`prereqs.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `CheckPrereqs(specDir, command)` | READ (`specdir.FileExists`, `specdir.DirExists`) | Various artifact paths per command | Existence checks | P0 |

---

## 4. Package: `tasks` (`cli/internal/tasks/`)

### 4.1 Parser (`parser.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ParseFile(path)` | READ (`os.ReadFile`) | `<specDir>/tasks.md` | Markdown: frontmatter (YAML-like) + checkbox task lines + wave plan | KEEP (reads) / P0 (parsed data) |
| `Parse(content)` | NONE (in-memory) | -- | -- | -- |

**Note:** `tasks.md` itself is KEEP (user-facing), but the parsed task state (status, wave, deps) is P0 migration target.

### 4.2 Mark (`mark.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `MarkTaskInFile(filePath, taskID, newStatus, requireImplLog, specDir)` | READ+WRITE (`os.ReadFile` + `os.WriteFile`) | `<specDir>/tasks.md` | Markdown: surgical checkbox replacement | KEEP (file) / P0 (state change) |
| `checkImplLog(specDir, taskID)` | READ (`specdir.FileExists` + `os.ReadDir`) | `<specDir>/execution/_implementation-logs/task-<id>.md` | Existence check + directory scan | P1 |

### 4.3 Next Wave (`next.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ResolveNextWave(doc, specDir)` | READ (indirect via `resolveCheckpointStatus`) | -- | In-memory from parsed `Document` | P0 |
| `resolveCheckpointStatus(specDir, waveNum)` | READ (`specdir.ReadLatestJSON`, `specdir.ReadStatusJSON`, `specdir.LatestRunDir`) | `_latest.json`, `_wave-summary.json`, `checkpoint/run-NNN/release-gate-decider/status.json` | JSON | P0 |

### 4.4 Validate (`validate.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `Validate(content)` | NONE (in-memory) | -- | -- | -- |

### 4.5 Complexity (`complexity.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ScoreComplexity(task)` | NONE (in-memory) | -- | -- | -- |

---

## 5. Package: `wave` (`cli/internal/wave/`)

### 5.1 Scanner (`scanner.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ScanWaves(specDir)` | READ (`specdir.ListWaveDirs` + `specdir.WaveExecPath` + `specdir.WaveCheckpointPath` + `specdir.LatestRunDir` + `os.ReadDir`) | `execution/waves/wave-NN/{execution,checkpoint}/run-NNN/` | Directory scan + status resolution | P0 |
| `countRunDirs(dir)` | READ (`specdir.DirExists` + `specdir.LatestRunDir`) | any run dir parent | Directory scan | P0 |
| `countRunEntries(dir)` | READ (`os.ReadDir`) | any run dir parent | Count `run-NNN` dirs | P0 |

### 5.2 Checkpoint (`checkpoint.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ResolveCheckpoint(specDir, waveNum)` | READ (`specdir.FileExists`, `specdir.ReadLatestJSON`, `specdir.ReadStatusJSON`, `specdir.LatestRunDir`, `os.ReadDir`) | `_latest.json`, `checkpoint/run-NNN/release-gate-decider/status.json`, `_wave-summary.json` | JSON | P0 |
| `readCheckpointRunStatus(checkDir, runID)` | READ (`specdir.DirExists`, `specdir.FileExists`, `specdir.ReadStatusJSON`) | `<checkDir>/<runID>/release-gate-decider/status.json` | JSON | P0 |
| `readRunSubagentStatus(runDir)` | READ (`os.ReadDir`, `specdir.FileExists`, `specdir.ReadStatusJSON`) | `<runDir>/*/status.json` | JSON | P0 |

### 5.3 Summary (`summary.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `GenerateSummary(specDir, waveNum)` | READ (`specdir.DirExists`, `specdir.FileExists`, `specdir.ReadStatusJSON`, `specdir.ReadLatestJSON`, `specdir.LatestRunDir`, `os.ReadDir`) | `_wave-summary.json`, `_latest.json`, `execution/run-NNN/*/status.json`, `checkpoint/run-NNN/*/status.json` | JSON | P0 |
| `scanSubagentStatus(runDir)` | READ (`os.ReadDir`, `specdir.FileExists`, `specdir.ReadStatusJSON`) | `<runDir>/*/status.json` | JSON | P0 |

### 5.4 Resume (`resume.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ComputeResume(specDir)` | READ (via `ScanWaves`) | All wave dirs | Aggregated wave state | P0 |

---

## 6. Package: `tools` (`cli/internal/tools/`)

### 6.1 Dispatch Init (`dispatch_init.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `DispatchInit(cwd, command, specName, wave, raw)` | READ (`os.ReadDir`) + WRITE (`os.MkdirAll`) | `.spec-workflow/specs/<name>/<phase>/_comms/<cmd>/run-NNN/`, artifact dirs | Directory creation, run numbering | P0 |

**Critical migration point:** This creates run directories and determines run IDs. In the DB model, this becomes an INSERT into a `runs` table.

### 6.2 Dispatch Init Audit (`dispatch_init_audit.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `dispatchInitAuditCore(cwd, runDir, auditType, iteration)` | WRITE (`os.MkdirAll`) | `<runDir>/_inline-audit/iteration-N/` or `<runDir>/_inline-checkpoint/` | Directory creation | P1 |

### 6.3 Dispatch Setup (`dispatch_setup.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `DispatchSetup(cwd, subagentName, runDir, modelAlias, raw)` | WRITE (`os.MkdirAll` + `os.WriteFile`) | `<runDir>/<subagent>/brief.md` | Markdown (brief template) | P0 |

**Critical migration point:** This creates subagent directories and initial brief.md. In the DB model, this becomes an INSERT into a `subagents` table.

### 6.4 Dispatch Handoff (`dispatch_handoff.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `DispatchHandoff(cwd, runDir, command, raw)` | READ (`os.Stat`, `os.ReadDir`, `os.ReadFile`) + WRITE (`os.WriteFile`) | `<runDir>/*/status.json` (read), `<runDir>/_handoff.md` (write) | JSON (read) + Markdown (write) | P0 |

**Critical migration point:** Reads all subagent status.json files and generates _handoff.md. In DB model, this reads from `subagents` table and writes to `runs.handoff_md`.

### 6.5 Dispatch Read Status (`dispatch_status.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `DispatchReadStatus(cwd, subagentName, runDir, raw)` | READ (`os.ReadFile`) | `<runDir>/<subagent>/status.json` | JSON: `{status, summary}` | P0 |

### 6.6 Runs (`runs.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `RunsLatestUnfinished(cwd, phaseDir, raw)` | READ (`os.Stat`, `os.ReadDir`, `os.ReadFile`) | `<phaseDir>/run-NNN/`, `*/brief.md`, `*/report.md`, `*/status.json`, `*/_handoff.md` | Directory scan + existence checks + JSON | P0 |
| `inspectRunDir(runDir)` | READ (`os.Stat`, `os.ReadDir`, `os.ReadFile`) | `<runDir>/_handoff.md`, `<runDir>/*/brief.md`, `*/report.md`, `*/status.json` | Existence + JSON content scan | P0 |

### 6.7 Spec Resolve (`spec_resolve.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `SpecResolveDir(cwd, specName, raw)` | READ (`os.Stat`) | `.spec-workflow/specs/<name>/` | Existence check | P0 |

### 6.8 Wave Resolve (`wave_resolve.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `WaveResolveCurrent(cwd, specName, raw)` | READ (`os.Stat`, `os.ReadDir`) | `execution/waves/wave-NN/` (with legacy `_agent-comms/waves/` fallback) | Directory scan | P0 |

### 6.9 Wave Status (`wave_status.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `waveStatusResult(cwd, specName)` | READ (`specdir.ListWaveDirs`, `os.ReadFile`, `specdir.ReadLatestJSON`, `os.ReadFile` for tasks.md) | `_wave-summary.json`, `_latest.json`, `tasks.md` | JSON + Markdown | P0 |
| `parseTasks(specDirAbs)` | READ (`os.ReadFile`) | `<specDir>/tasks.md` | Markdown: checkbox parsing | P0 |

### 6.10 Wave Update (`wave_update.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `waveUpdateResult(...)` | WRITE (`os.MkdirAll`, `os.WriteFile` x2) | `execution/waves/wave-NN/`, `_wave-summary.json`, `_latest.json` | JSON | P0 |
| `writeJSONFile(path, v)` | WRITE (`os.WriteFile`) | any | JSON | -- |

**Critical migration point:** Writes wave summary and latest JSON. In DB model, this becomes UPDATE on `waves` table.

### 6.11 Task Mark (`task_mark.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `taskMarkResult(cwd, specName, taskID, status)` | READ+WRITE (`os.ReadFile` + `os.WriteFile`) | `<specDir>/tasks.md` | Markdown: checkbox replacement (bold-format `**ID.**`) | KEEP (file) / P0 (state change) |

### 6.12 Impl Log (`impl_log.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `implLogRegisterResult(...)` | WRITE (`os.MkdirAll` + `os.WriteFile`) | `<specDir>/execution/_implementation-logs/task-<id>.md` | Markdown (implementation log) | P1 |
| `implLogCheckResult(...)` | READ (`specdir.FileExists`) | `<specDir>/execution/_implementation-logs/task-<id>.md` | Existence check | P1 |

### 6.13 Verify Task (`verify_task.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `verifyTaskResult(cwd, specName, taskID, checkCommit)` | READ (`specdir.FileExists`) + EXEC (`git log`) | `<specDir>/execution/_implementation-logs/task-<id>.md` | Existence + git output | P1 |

### 6.14 Approval (`approval.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ApprovalFallbackID(cwd, specName, docType, raw)` | READ (`os.ReadDir`, `os.Stat`, `os.ReadFile`) | `.spec-workflow/approvals/<spec>/approval_*.json` | JSON | P0 |

### 6.15 Skills (`skills.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `SkillsEffectiveSet(cwd, stage, raw)` | READ (via `config.ResolveConfigPath` + `config.LoadFromPath`) | `spw-config.toml` | TOML | KEEP |

### 6.16 Config Get (`config_get.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `ConfigGet(cwd, key, defaultValue, raw)` | READ (via `config.ResolveConfigPath` + `config.LoadFromPath`, `os.Stat`) | `spw-config.toml` | TOML | KEEP |

### 6.17 Audit Iteration (`audit_iteration.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `auditIterationStartCore(...)` | WRITE (`os.MkdirAll` + `os.WriteFile`) | `<runDir>/_inline-audit/_iteration-state.json` or `<runDir>/_inline-checkpoint/_iteration-state.json` | JSON: `{type, current_iteration, max_iterations, history[]}` | P1 |
| `auditIterationCheckCore(...)` | READ (`os.ReadFile`) | `_iteration-state.json` | JSON | P1 |
| `auditIterationAdvanceCore(...)` | READ+WRITE (`os.ReadFile` + `os.WriteFile`) | `_iteration-state.json` | JSON | P1 |

### 6.18 Other Tools (no file I/O)

- `ResolveModel` -- reads config only (KEEP)
- `MergeConfig` -- delegates to `config.Merge` (KEEP)
- `MergeSettings` -- delegates to `install.MergeSettings` (KEEP)
- `HandoffValidate` -- delegates to `inspectRunDir` (covered in 6.6)
- `Output`, `Fail` -- stdout/stderr output only

---

## 7. Package: `hook` (`cli/internal/hook/`)

### 7.1 Statusline (`statusline.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `detectCurrentTask(sessionID)` | READ (`os.ReadDir`, `os.ReadFile`) | `~/.claude/todos/<sessionID>-agent-*.json` | JSON: `[{status, activeForm, content}]` | KEEP (external) |
| `detectActiveSpec(dir)` | READ (`os.Stat`, `os.ReadDir`, cache read/write) | `.spec-workflow/specs/`, spec dirs, cache | Directory scan + JSON cache | P2 |
| `detectSpecFromGit(repoRoot, baseBranches)` | NONE (delegates to `git` package) | -- | -- | KEEP |
| `detectSpecByMtime(specsRoot)` | READ (`os.ReadDir`, `os.Stat`) | `.spec-workflow/specs/*/requirements.md`, `*/design.md`, `*/tasks.md` | Stat mtime | P2 |

### 7.2 Cache (`cache.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `writeStatuslineCache(root, spec, meta)` | WRITE (`os.MkdirAll` + `os.WriteFile`) | `.spec-workflow/.spw-cache/statusline.json` | JSON: `{ts, spec, ...meta}` | P2 |
| `readStatuslineCache(root, ttl, ignoreTTL)` | READ (`os.ReadFile`) | `.spec-workflow/.spw-cache/statusline.json` | JSON | P2 |
| `clearStatuslineCache(root)` | DELETE (`os.Remove`) | `.spec-workflow/.spw-cache/statusline.json` | -- | P2 |

### 7.3 Guard Prompt (`guard_prompt.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `HandleGuardPrompt()` | READ (via `newHookContext`), WRITE (via `writeStatuslineCache`) | `spw-config.toml` (read), `.spw-cache/statusline.json` (write) | TOML (read) + JSON (write) | P2 |

### 7.4 Guard Paths (`guard_paths.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `HandleGuardPaths()` | READ (via `newHookContext`) | `spw-config.toml` | TOML (config) | KEEP |

No spec-data I/O -- it only validates path patterns from the tool input.

### 7.5 Guard Stop (`guard_stop.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `HandleGuardStop()` | READ (`os.Stat`, `os.ReadDir`, `os.Stat` per file) | All `_comms/run-NNN/` across all phases for all specs | Directory traversal + existence checks for `_handoff.md`, `brief.md`, `report.md`, `status.json` | P0 |
| `collectRunDirs(specDir)` | READ (`os.ReadDir`) | 6 phase patterns of `_comms/run-NNN/` | Directory listing | P0 |
| `checkRunCompleteness(runDir)` | READ (`os.Stat`, `os.ReadDir`, `os.ReadFile`) | `<runDir>/_handoff.md`, `<runDir>/*/brief.md`, `*/report.md`, `*/status.json` | Existence + JSON content | P0 |

### 7.6 Session Start (`session_start.go`)

| Function | I/O Type | File Path Pattern | Data Format | Priority |
|----------|----------|-------------------|-------------|----------|
| `syncTasksTemplate(root, cfg)` | READ+WRITE (`os.Stat`, `os.ReadFile`, `os.WriteFile`) | `.spec-workflow/user-templates/variants/tasks-template.tdd-{on,off}.md` -> `.spec-workflow/user-templates/tasks-template.md` | Markdown (template files) | KEEP |
| `reRenderIfStale(root, cfg)` | READ+WRITE (`os.Stat`, `os.ReadDir`, `os.WriteFile`) | `spw-config.toml`, `.claude/workflows/spw/*.md` | TOML (read) + Markdown (write) | KEEP |
| `filesEqual(a, b)` | READ (`os.ReadFile` x2) | any two files | Binary comparison | KEEP |

---

## 8. Data Flow Map

### 8.1 Dispatch Lifecycle (the core flow)

```
DispatchInit                    DispatchSetup                 [Subagent]               DispatchHandoff
  |                                |                              |                        |
  v                                v                              v                        v
CREATE run-NNN dir  --->  CREATE <subagent>/   --->   WRITE brief.md     READ */status.json
SCAN existing runs        WRITE brief.md              WRITE report.md    WRITE _handoff.md
READ registry meta                                    WRITE status.json
READ config
```

**DB equivalent:**
1. `DispatchInit` -> INSERT into `runs` table (run_id, command, spec, phase, wave)
2. `DispatchSetup` -> INSERT into `subagents` table (name, run_id, brief_md, model)
3. Subagent writes -> UPDATE `subagents` SET report_md, status, summary
4. `DispatchHandoff` -> UPDATE `runs` SET handoff_md, all_pass

### 8.2 Wave State Resolution

```
ListWaveDirs  --->  For each wave:
                      |
                      +---> countRunDirs(execution/)
                      +---> countRunDirs(checkpoint/)
                      +---> ResolveCheckpoint:
                              |
                              +---> ReadLatestJSON(_latest.json)   [authoritative]
                              +---> ReadStatusJSON(status.json)    [ground truth]
                              +---> ReadStatusJSON(_wave-summary)  [staleness check]
                              +---> LatestRunDir(checkpoint/)      [fallback scan]
```

**DB equivalent:** Single query `SELECT * FROM waves WHERE spec_id = ? ORDER BY wave_num`.

### 8.3 Task State Resolution

```
ParseFile(tasks.md)  --->  Document{Tasks[], WavePlan[], Frontmatter}
     |
     v
ResolveNextWave(doc, specDir)
     |
     +---> Check in_progress tasks
     +---> Build wave map
     +---> findHighestCompletedWave
     +---> resolveCheckpointStatus (reads _latest.json, status.json)
     +---> findExecutableTasks (deps resolution)
     +---> findDeferredReady
```

**DB equivalent:** tasks.md remains file, but parsed state could be cached in `tasks` table. Checkpoint status comes from `waves` table.

### 8.4 Guard Stop Scan

```
ListSpecDirs  --->  For each spec:
                      collectRunDirs (6 phase patterns)
                        |
                        v
                      For each recent run:
                        checkRunCompleteness
                          |
                          +---> Stat _handoff.md
                          +---> ReadDir (subagent dirs)
                          +---> Stat brief.md, report.md, status.json per subagent
```

**DB equivalent:** `SELECT * FROM runs JOIN subagents WHERE runs.created_at > ? AND (subagents.status IS NULL OR runs.handoff_md IS NULL)`.

---

## 9. Migration Priority Summary

### P0 -- Must Migrate (Core Runtime State)

| Data | Current Format | Source Functions | Suggested Table |
|------|----------------|-----------------|-----------------|
| Run directories | `run-NNN/` dirs | `DispatchInit`, `LatestRunDir`, `inspectRunDir` | `runs` |
| Subagent state | `status.json`, `brief.md`, `report.md` per subagent dir | `DispatchSetup`, `DispatchHandoff`, `DispatchReadStatus` | `subagents` |
| Handoff records | `_handoff.md` per run | `DispatchHandoff` | `runs.handoff_md` |
| Wave state | `_wave-summary.json`, `_latest.json` per wave | `WaveUpdate`, `WaveStatus`, `ScanWaves`, `ResolveCheckpoint`, `GenerateSummary` | `waves` |
| Spec list | directories in `.spec-workflow/specs/` | `List`, `Resolve`, `SpecResolveDir` | `specs` |
| Stage classification | derived from file existence | `ClassifyStage` | `specs.stage` (computed) |
| Approval records | `approval_*.json` in `.spec-workflow/approvals/` | `CheckApproval`, `ApprovalFallbackID` | `approvals` |
| Run completeness | derived from subagent files | `HandleGuardStop`, `RunsLatestUnfinished` | query on `runs`/`subagents` |

### P1 -- Should Migrate (Artifacts)

| Data | Current Format | Source Functions | Suggested Table |
|------|----------------|-----------------|-----------------|
| Implementation logs | `task-<id>.md` in `_implementation-logs/` | `ImplLogRegister`, `ImplLogCheck`, `checkImplLog` | `impl_logs` |
| Audit iterations | `_iteration-state.json` | `AuditIterationStart/Check/Advance` | `audit_iterations` |
| Audit directories | `_inline-audit/iteration-N/`, `_inline-checkpoint/` | `DispatchInitAudit` | `audit_runs` |

### P2 -- Can Migrate Later (Cache/Temp)

| Data | Current Format | Source Functions | Suggested Table |
|------|----------------|-----------------|-----------------|
| Statusline cache | `.spw-cache/statusline.json` | `writeStatuslineCache`, `readStatuslineCache`, `clearStatuslineCache` | `cache` |
| Spec detection by mtime | `os.Stat` on dashboard files | `detectSpecByMtime` | Computed from `specs` |
| Artifact existence map | `specdir.FileExists` per artifact | `CheckArtifacts` | Computed from `subagents` |

### KEEP -- Must Remain as Files

| Data | Reason |
|------|--------|
| `spw-config.toml` | User-editable config, human-readable TOML |
| `requirements.md`, `design.md`, `tasks.md` | Dashboard files, consumed by MCP and humans |
| `STATUS-SUMMARY.md` | Output-only human-readable summary |
| User templates (`tasks-template.md`, variants) | User-editable templates |
| Rendered workflows (`.claude/workflows/spw/*.md`) | Claude Code reads these as slash commands |
| Implementation log `.md` files | Human-readable per-task records (but metadata in DB) |
| `brief.md`, `report.md` (subagent content) | Human-readable artifacts (but metadata/status in DB) |

---

## 10. Integration Points (Dual-Write Hooks)

These are the specific locations where dual-write should be inserted:

### 10.1 Run Creation
- **File:** `cli/internal/tools/dispatch_init.go:88-93`
- **What:** After `os.MkdirAll(runDir, ...)` succeeds, INSERT into `runs` table
- **Data:** `{run_id, spec_name, command, phase, category, subcategory, wave_num, created_at}`

### 10.2 Subagent Creation
- **File:** `cli/internal/tools/dispatch_setup.go:28-91`
- **What:** After `os.WriteFile(briefFullPath, ...)` succeeds, INSERT into `subagents` table
- **Data:** `{name, run_id, model, brief_path}`

### 10.3 Handoff Generation
- **File:** `cli/internal/tools/dispatch_handoff.go:83-86`
- **What:** After `os.WriteFile(handoffPath, ...)` succeeds, UPDATE `runs` SET handoff_md, all_pass; UPDATE `subagents` status/summary per agent
- **Data:** `{handoff_md, all_pass, subagent_statuses[]}`

### 10.4 Wave Summary Update
- **File:** `cli/internal/tools/wave_update.go:47-71`
- **What:** After writing `_wave-summary.json` and `_latest.json`, UPSERT into `waves` table
- **Data:** `{wave_num, status, task_ids[], execution_run, checkpoint_run, updated_at}`

### 10.5 Task Mark
- **File:** `cli/internal/tools/task_mark.go:62-63`
- **What:** After `os.WriteFile(tasksPath, ...)` succeeds, UPDATE `tasks` SET status
- **Data:** `{task_id, spec_name, previous_status, new_status}`

### 10.6 Implementation Log Register
- **File:** `cli/internal/tools/impl_log.go:44`
- **What:** After `os.WriteFile(logPath, ...)` succeeds, INSERT into `impl_logs` table
- **Data:** `{task_id, spec_name, wave, title, files, path, created_at}`

### 10.7 Audit Iteration State
- **File:** `cli/internal/tools/audit_iteration.go:86-88` (start), `:239-241` (advance)
- **What:** After writing `_iteration-state.json`, INSERT/UPDATE `audit_iterations` table
- **Data:** `{run_dir, audit_type, current_iteration, max_iterations, history[]}`

### 10.8 Statusline Cache
- **File:** `cli/internal/hook/cache.go:50`
- **What:** After `os.WriteFile(cacheFile, ...)`, UPSERT into `cache` table
- **Data:** `{key: "statusline", spec, source, timestamp}`

---

## 11. Risk Assessment

### Safe (Additive) Migrations

These are low-risk because they add DB writes alongside existing file writes:

1. **Run creation** (`DispatchInit`) -- additive INSERT, file remains canonical
2. **Subagent creation** (`DispatchSetup`) -- additive INSERT, file remains canonical
3. **Wave update** (`WaveUpdate`) -- additive UPSERT, JSON files remain canonical
4. **Impl log register** -- additive INSERT, .md file remains canonical
5. **Audit iteration state** -- additive INSERT/UPDATE, JSON file remains canonical
6. **Statusline cache** -- can switch to DB-only, cache is ephemeral

### Moderate Risk Migrations

These require careful dual-write because both reads and writes must stay synchronized:

1. **Handoff generation** (`DispatchHandoff`) -- reads subagent status.json, writes _handoff.md; both file and DB must agree
2. **Task mark** (`TaskMark` / `MarkTaskInFile`) -- writes to tasks.md; DB must mirror the state but tasks.md is the KEEP source of truth
3. **Approval records** -- read-only from files; DB should be populated by importing existing JSON files

### Higher Risk Migrations (Read-Path Changes)

These affect the read path. Until all reads are migrated from filesystem to DB, dual-read logic is needed:

1. **`ScanWaves`** -- currently scans directories; migrating reads to DB means queries must return identical results
2. **`ResolveCheckpoint`** -- multi-step resolution with fallbacks; DB query must replicate the priority logic
3. **`ResolveNextWave`** -- depends on checkpoint status + task state; both sources must agree
4. **`HandleGuardStop`** -- scans all spec dirs for recent runs; DB query is a natural fit but must match the window logic
5. **`ClassifyStage`** -- derived from file existence; could be a computed column but must stay consistent with file state during dual-write

### Recommended Migration Sequence

1. **Phase 1 (Write-side, additive):** Add DB writes to `DispatchInit`, `DispatchSetup`, `DispatchHandoff`, `WaveUpdate`, `ImplLogRegister`. No read changes.
2. **Phase 2 (Read-side, optional):** Add DB reads to `WaveStatus`, `ScanWaves`, `RunsLatestUnfinished`, `ResolveCheckpoint`. Validate against file reads.
3. **Phase 3 (DB-primary):** Switch reads to DB-primary with file fallback. Remove file writes for P0 data that is no longer needed on disk.
4. **Phase 4 (New queries):** Enable new query capabilities (search, history, aggregate stats) that only the DB can provide.
