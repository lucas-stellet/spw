#!/usr/bin/env bash
set -euo pipefail

SPW_REPO="${SPW_REPO:-lucas-stellet/spw}"
SPW_REF="${SPW_REF:-main}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TARGET_BIN="${INSTALL_DIR}/spw"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
local_wrapper="${script_dir}/../bin/spw"

base64_decode() {
  if printf 'dGVzdA==' | base64 --decode >/dev/null 2>&1; then
    base64 --decode
    return 0
  fi

  if printf 'dGVzdA==' | base64 -D >/dev/null 2>&1; then
    base64 -D
    return 0
  fi

  echo "[spw-install] Could not decode base64 with this system 'base64' command." >&2
  return 1
}

install_local_wrapper() {
  mkdir -p "$INSTALL_DIR"
  cp "$local_wrapper" "$TARGET_BIN"
  chmod +x "$TARGET_BIN"
}

install_remote_wrapper() {
  if ! command -v gh >/dev/null 2>&1; then
    echo "[spw-install] 'gh' is required to download bin/spw from ${SPW_REPO}." >&2
    exit 1
  fi

  mkdir -p "$INSTALL_DIR"
  gh api "repos/${SPW_REPO}/contents/bin/spw?ref=${SPW_REF}" --jq '.content' \
    | tr -d '\n' \
    | base64_decode > "$TARGET_BIN"
  chmod +x "$TARGET_BIN"
}

if [ -f "$local_wrapper" ]; then
  install_local_wrapper
  echo "[spw-install] Installed local wrapper to ${TARGET_BIN}"
else
  install_remote_wrapper
  echo "[spw-install] Downloaded wrapper from ${SPW_REPO}@${SPW_REF} to ${TARGET_BIN}"
fi

if ! echo "$PATH" | tr ':' '\n' | grep -Fx "$INSTALL_DIR" >/dev/null 2>&1; then
  echo "[spw-install] Add ${INSTALL_DIR} to your PATH to use 'spw'."
fi

echo "[spw-install] Run 'spw doctor' to verify configuration."
