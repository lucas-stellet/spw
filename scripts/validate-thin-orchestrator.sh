#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail() {
  echo "[thin-orchestrator] $*" >&2
  exit 1
}

base_cmd_dir="commands/oraculo"
base_wf_dir="workflows/oraculo"
team_overlay_dir="workflows/oraculo/overlays/teams"
active_overlay_dir="workflows/oraculo/overlays/active"

for file in "$base_cmd_dir"/*.md; do
  [ -f "$file" ] || continue

  lines="$(wc -l < "$file" | tr -d ' ')"
  if [ "$lines" -gt 60 ]; then
    fail "Wrapper too large (>60 lines): $file ($lines)"
  fi

  if ! rg -q "<execution_context>" "$file"; then
    fail "Missing <execution_context> in wrapper: $file"
  fi

  if ! rg -q "@\\.claude/workflows/oraculo/" "$file"; then
    fail "Wrapper does not reference workflow path: $file"
  fi

  if rg -q "<workflow>|<subagents>|<approval_protocol>|<file_handoff_protocol>" "$file"; then
    fail "Wrapper still contains detailed orchestration blocks: $file"
  fi

done

for base in "$base_cmd_dir"/*.md; do
  [ -f "$base" ] || continue
  name="$(basename "$base")"
  [ -f "$base_wf_dir/$name" ] || fail "Missing base workflow: $base_wf_dir/$name"
done

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

# Mirror checks
for dir in \
  "commands/oraculo copy-ready/.claude/commands/oraculo"; do
  src="${dir%% *}"
  dst="${dir##* }"
  diff -rq "$src" "$dst" >/dev/null || fail "Mirror mismatch: $src <-> $dst"
done

# Workflow mirror check (exclude overlays/active since symlink targets may differ)
diff -rq --exclude=active "$base_wf_dir" "copy-ready/.claude/workflows/oraculo" >/dev/null \
  || fail "Mirror mismatch: $base_wf_dir <-> copy-ready/.claude/workflows/oraculo"

# Overlay symlink mirror: verify copy-ready active symlinks match source
for link in "$active_overlay_dir"/*.md; do
  name="$(basename "$link")"
  mirror_link="copy-ready/.claude/workflows/oraculo/overlays/active/$name"
  [ -L "$mirror_link" ] || fail "Mirror missing symlink: $mirror_link"
  src_target="$(readlink "$link")"
  dst_target="$(readlink "$mirror_link")"
  [ "$src_target" = "$dst_target" ] || fail "Symlink target mismatch: $link ($src_target) vs $mirror_link ($dst_target)"
done

# Snippet template → copy-ready mirror checks
for pair in \
  "templates/claude-md-snippet.md copy-ready/.claude.md.snippet" \
  "templates/agents-md-snippet.md copy-ready/.agents.md.snippet"; do
  src="${pair%% *}"
  dst="${pair##* }"
  [ -f "$src" ] || fail "Missing snippet source: $src"
  [ -f "$dst" ] || fail "Missing snippet mirror: $dst"
  diff -q "$src" "$dst" >/dev/null || fail "Snippet mirror mismatch: $src <-> $dst"
done

# Verify dispatch-implementation.md exists in both locations
[ -f "workflows/oraculo/shared/dispatch-implementation.md" ] || fail "Missing: workflows/oraculo/shared/dispatch-implementation.md"
[ -f "copy-ready/.claude/workflows/oraculo/shared/dispatch-implementation.md" ] || fail "Missing: copy-ready/.claude/workflows/oraculo/shared/dispatch-implementation.md"

echo "[thin-orchestrator] OK"
