# PR Review Optimization

Oraculo generates dozens of spec-workflow files alongside feature code (PRDs, design research, QA plans, agent communications, etc.). Without mitigation, these files bury the actual code changes in pull request diffs, making review harder.

## Problem

A typical feature PR after a full Oraculo pipeline (`discover → design → tasks → exec → qa`) might show:

```
Files changed (47)  +2,340 -180
```

Where only 3-5 files are actual code changes and 40+ are spec-workflow artifacts. Reviewers must scroll past walls of generated markdown to find the code that matters.

## Solution: `.gitattributes` with `linguist-generated`

GitHub supports marking files as generated via `.gitattributes`. Generated files are:

- **Collapsed by default** in PR diff views
- **Excluded from the "Files changed" count**
- **Excluded from repository language statistics**
- **Still committed, tracked, and searchable** — nothing is lost
- **Expandable on demand** — reviewers click to show if they want to inspect

### What reviewers see

**Before (no `.gitattributes`):**
```
Files changed (47)  +2,340 -180

  lib/accounts/user.ex                                    +45 -12
  lib/accounts_web/live/user_live.ex                      +78 -23
  lib/accounts_web/templates/user.heex                    +34 -8
  .spec-workflow/specs/user-auth/requirements.md          +120 -0
  .spec-workflow/specs/user-auth/discover/PRD.md               +340 -0
  .spec-workflow/specs/user-auth/discover/PRD-SOURCE-NOTES.md  +89 -0
  .spec-workflow/specs/user-auth/design/DESIGN-RESEARCH.md +210 -0
  ... 40 more spec files ...
```

**After (with `.gitattributes`):**
```
Files changed (3)  +157 -43

  lib/accounts/user.ex                  +45 -12
  lib/accounts_web/live/user_live.ex    +78 -23
  lib/accounts_web/templates/user.heex  +34 -8

  44 generated files not shown  [Click to expand]
```

The reviewer sees only the code changes. Spec artifacts are one click away if needed.

## Implementation

### The rule

A single line in `.gitattributes` at the project root:

```gitattributes
.spec-workflow/specs/** linguist-generated=true
```

This marks all files under any spec directory as generated. Dashboard files (`requirements.md`, `design.md`, `tasks.md`), phase outputs (`PRD.md`, `QA-CHECK.md`), and agent comms (`_comms/`) are all covered.

### Installer integration

The `claude-kit/install.sh` script adds this rule automatically during `oraculo install`:

```bash
setup_gitattributes() {
  local rule='.spec-workflow/specs/** linguist-generated=true'
  local gitattributes="${TARGET_ROOT}/.gitattributes"
  if [ ! -f "$gitattributes" ] || ! grep -qF "$rule" "$gitattributes"; then
    echo "$rule" >> "$gitattributes"
    echo "[oraculo-kit] Added .gitattributes rule for PR review optimization."
  fi
}
```

The function:
- Creates `.gitattributes` if it doesn't exist
- Appends the rule only if not already present
- Is idempotent — safe to run multiple times

### What is NOT marked as generated

- Feature code (`lib/`, `src/`, `test/`, etc.) — always visible in PR diffs
- Oraculo config (`.spec-workflow/oraculo.toml`) — not under `specs/`, so not affected
- Oraculo templates (`.spec-workflow/user-templates/`) — not under `specs/`, so not affected

## Scope

This optimization applies only to **GitHub PR diffs** (and compatible platforms like GitLab that support `linguist-generated`). It has zero effect on:

- Git operations (add, commit, diff, log, blame)
- Local development workflows
- CI/CD pipelines
- File content or accessibility
