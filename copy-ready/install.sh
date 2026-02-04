#!/usr/bin/env bash
set -euo pipefail

# Install the copy-ready kit into the current project root.
# Usage:
#   ./install.sh
#
# Behavior:
# - Copies kit files into current project
# - Does not overwrite .claude/settings.json (prints merge instruction instead)
# - Best-effort installs default SPW skills into .claude/skills (if available locally)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_ROOT="$(pwd)"
CONFIG_PATH="${TARGET_ROOT}/.spec-workflow/spw-config.toml"
SPW_REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SUPERPOWERS_SKILLS_DIR="${SPW_SUPERPOWERS_SKILLS_DIR:-${SPW_REPO_ROOT}/superpowers/skills}"

toml_bool_value() {
  local section="$1"
  local key="$2"
  local default_value="$3"
  local value

  value="$({
    awk -v wanted_section="[$section]" -v wanted_key="$key" '
      function trim(s) {
        sub(/^[ \t\r\n]+/, "", s)
        sub(/[ \t\r\n]+$/, "", s)
        return s
      }

      /^[ \t]*\[/ {
        in_section = ($0 == wanted_section)
      }

      in_section {
        if ($0 ~ /^[ \t]*#/ || $0 ~ /^[ \t]*$/) {
          next
        }

        if ($0 ~ "^[ \\t]*" wanted_key "[ \\t]*=") {
          line = $0
          sub(/^[^=]*=[ \t]*/, "", line)
          sub(/[ \t]*#.*/, "", line)
          line = trim(line)
          print line
          exit
        }
      }
    ' "$CONFIG_PATH"
  } || true)"

  value="$(printf '%s' "${value:-$default_value}" | tr '[:upper:]' '[:lower:]')"
  case "$value" in
    true|1|yes|on) printf 'true' ;;
    false|0|no|off) printf 'false' ;;
    *) printf '%s' "$default_value" ;;
  esac
}

find_skill_source_dir() {
  local skill="$1"
  local candidates=(
    "${HOME}/.claude/skills/${skill}"
    "${HOME}/.codex/skills/${skill}"
    "${HOME}/.codex/superpowers/skills/${skill}"
    "${HOME}/.config/opencode/skills/${skill}"
  )

  if [ -d "${SUPERPOWERS_SKILLS_DIR}" ]; then
    candidates+=("${SUPERPOWERS_SKILLS_DIR}/${skill}")
  fi

  local dir
  for dir in "${candidates[@]}"; do
    if [ -f "${dir}/SKILL.md" ]; then
      printf '%s' "$dir"
      return 0
    fi
  done
  return 1
}

install_default_skills() {
  local target_skills_dir="${TARGET_ROOT}/.claude/skills"
  mkdir -p "$target_skills_dir"

  local default_skills=(
    "using-elixir-skills"
    "elixir-thinking"
    "elixir-anti-patterns"
    "phoenix-thinking"
    "ecto-thinking"
    "otp-thinking"
    "oban-thinking"
    "test-driven-development"
    "requesting-code-review"
  )

  local installed=0
  local skipped_existing=0
  local missing=()

  local skill src_dir
  for skill in "${default_skills[@]}"; do
    if [ -e "${target_skills_dir}/${skill}" ]; then
      skipped_existing=$((skipped_existing + 1))
      continue
    fi

    if src_dir="$(find_skill_source_dir "$skill")"; then
      rsync -a "${src_dir}/" "${target_skills_dir}/${skill}/"
      installed=$((installed + 1))
    else
      missing+=("$skill")
    fi
  done

  echo "[spw-kit] Default skills install: installed=${installed}, existing=${skipped_existing}, missing=${#missing[@]}"
  if [ "${#missing[@]}" -gt 0 ]; then
    echo "[spw-kit] Missing local skill sources (non-blocking): ${missing[*]}"
  fi
}

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
chmod +x "${TARGET_ROOT}/.claude/hooks/spw-statusline.js" || true

AUTO_INSTALL_SKILLS="$(toml_bool_value skills auto_install_defaults_on_spw_install true)"
if [ "$AUTO_INSTALL_SKILLS" = "true" ]; then
  install_default_skills
else
  echo "[spw-kit] Skipping default skills install (auto_install_defaults_on_spw_install=false)."
fi

echo "[spw-kit] Installation complete."
echo "[spw-kit] Next step: adjust .spec-workflow/spw-config.toml"
