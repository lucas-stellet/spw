#!/usr/bin/env bash
set -euo pipefail

SPW_REPO="${SPW_REPO:-lucas-stellet/spw}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TARGET_BIN="${INSTALL_DIR}/spw"
RAW_URL="https://raw.githubusercontent.com/${SPW_REPO}/main/bin/spw"
GH_API_PATH="repos/${SPW_REPO}/contents/bin/spw?ref=main"

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "[spw-bootstrap] Required command not found: $cmd" >&2
    exit 1
  fi
}

require_cmd curl

download_with_gh() {
  gh api "$GH_API_PATH" -H 'Accept: application/vnd.github.raw'
}

download_with_curl() {
  curl -fsSL "$RAW_URL"
}

mkdir -p "$INSTALL_DIR"
tmp_bin="$(mktemp)"
trap 'rm -f "$tmp_bin"' EXIT

if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
  download_with_gh > "$tmp_bin"
else
  download_with_curl > "$tmp_bin"
fi

chmod +x "$tmp_bin"
mv "$tmp_bin" "$TARGET_BIN"

echo "[spw-bootstrap] Installed spw to ${TARGET_BIN}"
echo "[spw-bootstrap] Source repo: ${SPW_REPO} (main)"

if ! echo "$PATH" | tr ':' '\n' | grep -Fx "$INSTALL_DIR" >/dev/null 2>&1; then
  echo "[spw-bootstrap] Add ${INSTALL_DIR} to your PATH to run 'spw'."
fi

echo "[spw-bootstrap] Next: run 'spw doctor'"
