#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail() {
  echo "[validate-kit] $*" >&2
  exit 1
}

base_cmd_dir="claude-kit/.claude/commands/oraculo"
base_wf_dir="claude-kit/.claude/workflows/oraculo"
team_overlay_dir="claude-kit/.claude/workflows/oraculo/overlays/teams"
active_overlay_dir="claude-kit/.claude/workflows/oraculo/overlays/active"

# Validate command wrappers: max 60 lines, must reference execution_context and workflow path
for file in "$base_cmd_dir"/*.md; do
  [ -f "$file" ] || continue

  lines="$(wc -l < "$file" | tr -d ' ')"
  if [ "$lines" -gt 60 ]; then
    fail "Wrapper too large (>60 lines): $file ($lines)"
  fi

  if ! rg -q "<execution_context>" "$file"; then
    fail "Missing <execution_context> in wrapper: $file"
  fi

  if ! rg -q "@\.claude/workflows/oraculo/" "$file"; then
    fail "Wrapper does not reference workflow path: $file"
  fi

  if rg -q "<workflow>|<subagents>|<approval_protocol>|<file_handoff_protocol>" "$file"; then
    fail "Wrapper still contains detailed orchestration blocks: $file"
  fi

done

# Validate that each command has a corresponding base workflow
for base in "$base_cmd_dir"/*.md; do
  [ -f "$base" ] || continue
  name="$(basename "$base")"
  [ -f "$base_wf_dir/$name" ] || fail "Missing base workflow: $base_wf_dir/$name"
done

# Validate team overlays exist for expected commands
for team in design-research.md tasks-check.md exec.md checkpoint.md qa-check.md qa-exec.md; do
  [ -f "$team_overlay_dir/$team" ] || fail "Missing team overlay: $team_overlay_dir/$team"
done

# Validate overlay symlinks: each must point to ../noop.md or ../teams/<name>.md
for link in "$active_overlay_dir"/*.md; do
  [ -L "$link" ] || fail "Not a symlink: $link"
  target="$(readlink "$link")"
  case "$target" in
    ../noop.md|../teams/*.md) ;;
    *) fail "Invalid symlink target in $link -> $target (expected ../noop.md or ../teams/<name>.md)" ;;
  esac
done

# Verify key structural files exist
[ -f "claude-kit/.claude/workflows/oraculo/shared/dispatch-implementation.md" ] || fail "Missing: claude-kit/.claude/workflows/oraculo/shared/dispatch-implementation.md"
[ -f "claude-kit/.claude.md.snippet" ] || fail "Missing: claude-kit/.claude.md.snippet"
[ -f "claude-kit/.agents.md.snippet" ] || fail "Missing: claude-kit/.agents.md.snippet"
[ -f "claude-kit/install.sh" ] || fail "Missing: claude-kit/install.sh"

echo "[validate-kit] OK"
