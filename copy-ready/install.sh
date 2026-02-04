#!/usr/bin/env bash
set -euo pipefail

# Install the copy-ready kit into the current project root.
# Usage:
#   ./install.sh
#
# Behavior:
# - Copies kit files into current project
# - Does not overwrite .claude/settings.json (prints merge instruction instead)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_ROOT="$(pwd)"

echo "[spw-kit] Installing into project: ${TARGET_ROOT}"

# Copy only SPW runtime assets (avoid touching project root files like README.md)
rsync -a "${SCRIPT_DIR}/.claude/" "${TARGET_ROOT}/.claude/"
rsync -a "${SCRIPT_DIR}/.spec-workflow/" "${TARGET_ROOT}/.spec-workflow/"

if [ ! -f "${TARGET_ROOT}/.claude/settings.json" ]; then
  mkdir -p "${TARGET_ROOT}/.claude"
  cp "${SCRIPT_DIR}/.claude/settings.json.example" "${TARGET_ROOT}/.claude/settings.json"
  echo "[spw-kit] Created .claude/settings.json with SessionStart hook."
else
  echo "[spw-kit] .claude/settings.json already exists."
  echo "[spw-kit] Manually merge hook block from ${SCRIPT_DIR}/.claude/settings.json.example"
fi

chmod +x "${TARGET_ROOT}/.claude/hooks/session-start-sync-tasks-template.sh" || true

echo "[spw-kit] Installation complete."
echo "[spw-kit] Next step: adjust .spec-workflow/spw-config.toml"
