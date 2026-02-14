#!/usr/bin/env bash
set -euo pipefail

# Install the Oraculo Go binary from the latest GitHub Release.
# Usage:
#   gh api repos/lucas-stellet/oraculo/contents/scripts/bootstrap.sh?ref=main -H 'Accept: application/vnd.github.raw' | bash
#   curl -fsSL https://raw.githubusercontent.com/lucas-stellet/oraculo/main/scripts/bootstrap.sh | bash

ORACULO_REPO="${ORACULO_REPO:-lucas-stellet/oraculo}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TARGET_BIN="${INSTALL_DIR}/oraculo"

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "[oraculo-bootstrap] Required command not found: $cmd" >&2
    exit 1
  fi
}

require_cmd gh
require_cmd tar

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) echo "[oraculo-bootstrap] Unsupported architecture: $arch" >&2; exit 1 ;;
  esac

  case "$os" in
    darwin|linux) ;;
    *) echo "[oraculo-bootstrap] Unsupported OS: $os" >&2; exit 1 ;;
  esac

  printf '%s_%s' "$os" "$arch"
}

platform="$(detect_platform)"

# Fetch latest release tag
echo "[oraculo-bootstrap] Fetching latest release from ${ORACULO_REPO}..."
tag_name="$(gh api "repos/${ORACULO_REPO}/releases/latest" --jq '.tag_name')"
if [ -z "$tag_name" ]; then
  echo "[oraculo-bootstrap] Could not determine latest release tag." >&2
  exit 1
fi

version="${tag_name#v}"
asset_pattern="oraculo_${version}_${platform}.tar.gz"

# Download and extract
echo "[oraculo-bootstrap] Downloading ${asset_pattern} (${tag_name})..."
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

gh release download "$tag_name" \
  --repo "$ORACULO_REPO" \
  --pattern "$asset_pattern" \
  --dir "$tmp_dir"

tar -xzf "${tmp_dir}/${asset_pattern}" -C "$tmp_dir"

if [ ! -f "$tmp_dir/oraculo" ]; then
  echo "[oraculo-bootstrap] Binary 'oraculo' not found in tarball." >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
mv "$tmp_dir/oraculo" "$TARGET_BIN"
chmod +x "$TARGET_BIN"

echo "[oraculo-bootstrap] Installed oraculo ${tag_name} to ${TARGET_BIN}"

if ! echo "$PATH" | tr ':' '\n' | grep -Fx "$INSTALL_DIR" >/dev/null 2>&1; then
  echo "[oraculo-bootstrap] Add ${INSTALL_DIR} to your PATH to run 'oraculo'."
fi

echo "[oraculo-bootstrap] Next: run 'oraculo doctor'"
