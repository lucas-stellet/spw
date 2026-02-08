#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail() {
  echo "[thin-orchestrator] $*" >&2
  exit 1
}

base_cmd_dir="commands/spw"
team_cmd_dir="commands/spw-teams"
base_wf_dir="workflows/spw"
team_overlay_dir="workflows/spw/overlays/teams"

for file in "$base_cmd_dir"/*.md "$team_cmd_dir"/*.md; do
  [ -f "$file" ] || continue

  lines="$(wc -l < "$file" | tr -d ' ')"
  if [ "$lines" -gt 60 ]; then
    fail "Wrapper too large (>60 lines): $file ($lines)"
  fi

  if ! rg -q "<execution_context>" "$file"; then
    fail "Missing <execution_context> in wrapper: $file"
  fi

  if ! rg -q "@\\.claude/workflows/spw/" "$file"; then
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

for team in design-research.md tasks-check.md exec.md checkpoint.md; do
  [ -f "$team_overlay_dir/$team" ] || fail "Missing team overlay: $team_overlay_dir/$team"
done

# Mirror checks
for dir in \
  "commands/spw copy-ready/.claude/commands/spw" \
  "commands/spw-teams copy-ready/.claude/commands/spw-teams" \
  "workflows/spw copy-ready/.claude/workflows/spw"; do
  src="${dir%% *}"
  dst="${dir##* }"
  diff -rq "$src" "$dst" >/dev/null || fail "Mirror mismatch: $src <-> $dst"
done

echo "[thin-orchestrator] OK"
