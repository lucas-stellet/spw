# SPW Refactor Plan

Three-phase refactoring to standardize dispatch patterns, directory structure, and PR review experience across all SPW commands.

**Reference docs (already created):**
- Phase 1: `docs/DISPATCH-PATTERNS.md`
- Phase 2: `docs/SPEC-DIRECTORY-STRUCTURE.md`
- Phase 3: `docs/PR-REVIEW-OPTIMIZATION.md`
- Workflow example: `.entire/examples/qa-workflow-example.md`
- Shared policy example: `.entire/examples/shared-dispatch-pipeline.md`

---

## Phase 1: Thin-Dispatch + Workflow Refactor

**Goal:** Every workflow references a category-level shared dispatch policy. Orchestrators never accumulate subagent output. Boilerplate moves to shared policies; workflows become declarations with extension points.

### Step 1.1 — Create shared dispatch policies

Create 3 new files in `workflows/spw/shared/`:

| File | Category | Defines |
|------|----------|---------|
| `dispatch-pipeline.md` | Pipeline | Sequential dispatch, status-only reads, path-based briefs, synthesizer-reads-from-fs, pipeline resume, extension points (pre_pipeline, pre_dispatch, post_dispatch, post_pipeline) |
| `dispatch-audit.md` | Audit | Parallel/sequential auditor dispatch, status-only reads, aggregator-reads-from-fs, audit resume (skip passed auditors, always rerun aggregator) |
| `dispatch-wave.md` | Wave Execution | Scout dispatch, wave splitting, per-wave sequential loop, wave-summary.json, status-only reads, synthesizer-reads-from-fs, wave-level resume |

Each policy includes the 5 core thin-dispatch rules from `docs/DISPATCH-PATTERNS.md`.

**Mirror:** Copy all 3 to `copy-ready/.claude/workflows/spw/shared/`.

### Step 1.2 — Refactor Pipeline commands (6 workflows)

Refactor in order:

| # | Command | File | Subcategory | Key extensions |
|---|---------|------|-------------|----------------|
| 1 | `prd` | `workflows/spw/prd.md` | Research | pre_pipeline: user intent gate, revision loop; pre_dispatch: conditional scout branching |
| 2 | `design-research` | `workflows/spw/design-research.md` | Research | pre_pipeline: SPA/prototype URL detection; pre_dispatch: Playwright MCP fallback |
| 3 | `design-draft` | `workflows/spw/design-draft.md` | Synthesis | pre_pipeline: approval reconciliation; post_pipeline: Mermaid diagram validation |
| 4 | `tasks-plan` | `workflows/spw/tasks-plan.md` | Synthesis | pre_pipeline: mode selection (initial/next-wave/config); post_pipeline: dashboard compatibility check |
| 5 | `qa` | `workflows/spw/qa.md` | Synthesis | pre_pipeline: user intent + tool selection; pre_dispatch: conditional designer (playwright/bruno/hybrid) |
| 6 | `post-mortem` | `workflows/spw/post-mortem.md` | Synthesis | post_pipeline: memory index update |

**Per workflow:**
1. Add `<dispatch_pattern>` tag referencing `dispatch-pipeline.md`
2. Extract dispatch boilerplate → replaced by shared policy reference
3. Keep `<subagents>` (concrete, with model roles and dependencies)
4. Keep command-specific policies (user_intent_gate, tool_selection, etc.)
5. Add `<extensions>` with command-specific logic at named extension points
6. Keep `<acceptance_criteria>` — add thin-dispatch verification line
7. Mirror to `copy-ready/`

### Step 1.3 — Refactor Audit commands (3 workflows)

| # | Command | File | Subcategory | Key extensions |
|---|---------|------|-------------|----------------|
| 7 | `tasks-check` | `workflows/spw/tasks-check.md` | Artifact | pre_pipeline: verify tasks.md exists |
| 8 | `qa-check` | `workflows/spw/qa-check.md` | Code | pre_pipeline: verify QA-TEST-PLAN.md exists |
| 9 | `checkpoint` | `workflows/spw/checkpoint.md` | Code | pre_pipeline: wave-awareness (reads current wave context) |

Same pattern: `<dispatch_pattern>` → `dispatch-audit.md`, concrete auditors, extensions, mirror.

### Step 1.4 — Refactor Wave Execution commands (2 workflows)

| # | Command | File | Subcategory | Key extensions |
|---|---------|------|-------------|----------------|
| 10 | `exec` | `workflows/spw/exec.md` | Implementation | inter_wave: checkpoint + user authorization; per_task: git hygiene, commit policy |
| 11 | `qa-exec` | `workflows/spw/qa-exec.md` | Validation | pre_pipeline: verify QA-CHECK.md PASS; per_wave: re-auth (--isolated); post_pipeline: selector drift reporting |

`qa-exec` gains wave-based dispatch (scenarios split into waves by `[qa].max_scenarios_per_wave`).

**New config key:** Add `[qa].max_scenarios_per_wave = 5` to `config/spw-config.toml` + mirror.

### Step 1.5 — Consolidate old shared policies

| Old file | Action |
|----------|--------|
| `shared/file-handoff.md` | Keep (handoff contract is still universal). Update to reference thin-dispatch rules. |
| `shared/resume-policy.md` | Keep (run-level resume is universal). Category policies add category-specific resume on top. |
| `shared/config-resolution.md` | Keep unchanged. |
| `shared/skills-policy.md` | Keep unchanged. |
| `shared/approval-reconciliation.md` | Keep unchanged. |

No files deleted — old policies are still referenced and complemented by the new category policies.

### Step 1.6 — Update teams overlays

All 11 `workflows/spw/overlays/teams/*.md` files need minor updates to reference the new dispatch pattern format. Changes are minimal — the overlay just adds team creation/role mapping; the dispatch mechanism comes from the shared policy.

Mirror all to `copy-ready/`.

### Step 1.7 — Validation

```bash
# All scripts parse
bash -n bin/spw && bash -n scripts/*.sh && bash -n copy-ready/install.sh

# Mirror integrity + wrapper size
scripts/validate-thin-orchestrator.sh

# Hook smoke tests
node hooks/spw-guard-stop.js <<< '{}'
node hooks/spw-statusline.js <<< '{"workspace":{"current_dir":"'"$(pwd)"'"}}'
node hooks/spw-guard-user-prompt.js <<< '{"prompt":"/spw:qa-exec my-spec"}'

# Wrapper size < 60 lines (all commands)
wc -l commands/spw/*.md

# Verify all workflows reference a dispatch policy
grep -L "dispatch_pattern" workflows/spw/{prd,design-research,design-draft,tasks-plan,tasks-check,qa,qa-check,qa-exec,exec,checkpoint,post-mortem}.md
# Expected: no output (all files should match)
```

### Deliverables

- 3 new shared policies
- 11 refactored workflows + 11 mirrors
- 11 updated teams overlays + 11 mirrors
- 1 config update + mirror
- Updated `file-handoff.md` + mirror
- Docs: `README.md`, `AGENTS.md`, `copy-ready/README.md`

---

## Phase 2: Spec Directory Restructure

**Goal:** Artifacts organized by workflow phase. No more `_generated/` and `_agent-comms/` flat dumps. Each phase owns its outputs and agent comms.

**Depends on:** Phase 1 (workflows must reference new paths).

### Step 2.1 — Update artifact paths in all workflows

Every workflow has `<artifact_boundary>` and `<file_handoff_protocol>` sections with hardcoded paths. Update all 11:

| Old path pattern | New path pattern |
|-----------------|-----------------|
| `_generated/PRD.md` | `prd/PRD.md` |
| `_generated/DESIGN-RESEARCH.md` | `design/DESIGN-RESEARCH.md` |
| `_generated/TASKS-CHECK.md` | `planning/TASKS-CHECK.md` |
| `_generated/CHECKPOINT-REPORT.md` | `execution/CHECKPOINT-REPORT.md` |
| `_generated/QA-*.md` | `qa/QA-*.md` |
| `_agent-comms/prd/run-NNN/` | `prd/_comms/run-NNN/` |
| `_agent-comms/design-research/run-NNN/` | `design/_comms/design-research/run-NNN/` |
| `_agent-comms/tasks-plan/run-NNN/` | `planning/_comms/tasks-plan/run-NNN/` |
| `_agent-comms/tasks-check/run-NNN/` | `planning/_comms/tasks-check/run-NNN/` |
| `_agent-comms/waves/wave-NN/` | `execution/waves/wave-NN/` |
| `_agent-comms/post-mortem/run-NNN/` | `post-mortem/_comms/run-NNN/` |
| `_agent-comms/qa/run-NNN/` | `qa/_comms/qa/run-NNN/` |
| `_agent-comms/qa-check/run-NNN/` | `qa/_comms/qa-check/run-NNN/` |
| `_agent-comms/qa-exec/run-NNN/` | `qa/_comms/qa-exec/waves/wave-NN/run-NNN/` |

Mirror all to `copy-ready/`.

### Step 2.2 — Update hooks

| Hook | Change needed |
|------|--------------|
| `spw-statusline.js` | Update path patterns for spec detection (reads `_generated/` → phase dirs) |
| `spw-guard-paths.js` | Update allowed write paths (add phase dirs, remove `_generated/`) |
| `spw-guard-stop.js` | Update `collectRunDirs()` to scan `<phase>/_comms/` instead of `_agent-comms/` |
| `spw-hook-lib.js` | Update `simpleCommands` array (add missing qa/post-mortem commands). Update any path constants. |

Mirror all to `copy-ready/.claude/hooks/`.

### Step 2.3 — Update templates

Check `templates/user-templates/` for any hardcoded `_generated/` or `_agent-comms/` references. Update to new phase-based paths.

Mirror to `copy-ready/.spec-workflow/user-templates/`.

### Step 2.4 — Update shared policies

Update `shared/file-handoff.md` path examples from `_agent-comms/<command>/run-NNN/` to `<phase>/_comms/<command>/run-NNN/`.

### Step 2.5 — Update status command

`workflows/spw/status.md` reads artifacts from multiple locations to produce STATUS-SUMMARY.md. Update all artifact read paths.

### Step 2.6 — Validation

```bash
# Mirror integrity
scripts/validate-thin-orchestrator.sh

# Hook smoke tests (all hooks, since paths changed)
node hooks/spw-statusline.js <<< '{"workspace":{"current_dir":"'"$(pwd)"'"}}'
node hooks/spw-guard-user-prompt.js <<< '{"prompt":"/spw:qa-exec my-spec"}'
node hooks/spw-guard-paths.js <<< '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":".spec-workflow/specs/test/qa/QA-CHECK.md"}}'
node hooks/spw-guard-stop.js <<< '{}'

# Grep for stale paths (should return no matches in workflows/)
grep -r "_generated/" workflows/spw/ --include="*.md"
grep -r "_agent-comms/" workflows/spw/ --include="*.md"
```

### Deliverables

- 11 workflows updated (artifact paths) + mirrors
- 4 hooks updated + mirrors
- Templates updated + mirrors
- Shared policies updated + mirrors
- Docs: `README.md`, `AGENTS.md`, `copy-ready/README.md`

---

## Phase 3: PR Review Optimization

**Goal:** Spec-workflow files collapse by default in GitHub PR diffs via `.gitattributes`.

**Independent of:** Phases 1 and 2 (can be done in any order).

### Step 3.1 — Add gitattributes setup to installer

Add `setup_gitattributes()` function to `copy-ready/install.sh`:

```bash
setup_gitattributes() {
  local rule='.spec-workflow/specs/** linguist-generated=true'
  local gitattributes="${TARGET_ROOT}/.gitattributes"
  if [ ! -f "$gitattributes" ] || ! grep -qF "$rule" "$gitattributes"; then
    echo "$rule" >> "$gitattributes"
    echo "[spw-kit] Added .gitattributes rule for PR review optimization."
  fi
}
```

Call it in the install flow after file copy.

### Step 3.2 — Validation

```bash
# Installer parses
bash -n copy-ready/install.sh

# Test in a temp dir
tmpdir=$(mktemp -d)
cd "$tmpdir" && git init
# Run install, verify .gitattributes exists with correct rule
grep "linguist-generated" .gitattributes
# Run install again, verify no duplicate
wc -l .gitattributes  # should be 1
```

### Deliverables

- `copy-ready/install.sh` updated
- Docs already created (`docs/PR-REVIEW-OPTIMIZATION.md`, CLAUDE.md, AGENTS.md sections)

---

## Execution Order

```
Phase 3 ──────────────────────────────── (independent, small, quick win)

Phase 1 ──────────────────────────────── (foundation: dispatch patterns)
  1.1 shared policies
  1.2 pipeline commands (6)
  1.3 audit commands (3)
  1.4 wave commands (2)
  1.5 consolidate old policies
  1.6 teams overlays
  1.7 validation

Phase 2 ──────────────────────────────── (depends on Phase 1: new paths)
  2.1 workflow artifact paths
  2.2 hooks
  2.3 templates
  2.4 shared policies
  2.5 status command
  2.6 validation
```

Phase 3 can run first or in parallel since it only touches the installer. Phases 1 and 2 are sequential — Phase 2 builds on the refactored workflows from Phase 1.

---

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Refactored workflow changes agent behavior | Structural-only changes: subagents, gates, policies stay identical. Only dispatch boilerplate moves to shared policy. |
| Broken mirrors after mass edit | Run `scripts/validate-thin-orchestrator.sh` after every batch of changes. |
| Hooks break on new paths (Phase 2) | Smoke-test every hook after path updates. |
| Existing specs in old structure | Phase 2 does NOT migrate existing spec data. Old specs continue working until manually migrated or a new spec is created. A future `spw migrate-spec` command could handle this. |
| Teams overlays out of sync | Update overlays in same batch as base workflows. |

## Total Scope

| Item | Count |
|------|-------|
| New shared policies | 3 |
| Workflows refactored | 11 |
| Teams overlays updated | 11 |
| Hooks updated | 4 |
| Config updated | 1 |
| Installer updated | 1 |
| Mirrors | ~30 files |
| Docs updated | README.md, AGENTS.md, copy-ready/README.md, docs/SPW-WORKFLOW.md |
