#!/usr/bin/env bash
set -euo pipefail

# Install or inspect the copy-ready kit in the current project root.
# Usage:
#   ./install.sh [install|skills|status]
#
# Behavior:
# - install (default): copies kit files into current project
# - skills: installs default SPW skills into .claude/skills (best effort)
# - status: prints a quick summary of kit presence + default skills
# - Does not overwrite .claude/settings.json (prints merge instruction instead)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_ROOT="$(pwd)"
CONFIG_PATH="${TARGET_ROOT}/.spec-workflow/spw-config.toml"
SPW_REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SUPERPOWERS_SKILLS_DIR="${SPW_SUPERPOWERS_SKILLS_DIR:-${SPW_REPO_ROOT}/superpowers/skills}"

DEFAULT_SKILLS=(
  "using-elixir-skills"
  "elixir-thinking"
  "elixir-anti-patterns"
  "phoenix-thinking"
  "ecto-thinking"
  "otp-thinking"
  "oban-thinking"
  "conventional-commits"
  "test-driven-development"
  "requesting-code-review"
)

show_help() {
  cat <<'USAGE'
spw - install or inspect the SPW kit in the current project

Usage:
  spw
  spw install
  spw skills
  spw status

Behavior:
- install (default): copies commands, hooks, templates, and config into cwd.
- skills: installs default SPW skills into .claude/skills (best effort).
- status: prints a quick summary of kit presence + default skills.

Notes:
- If .claude/settings.json already exists, it is not overwritten (manual merge required).
USAGE
}

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

  local installed=0
  local skipped_existing=0
  local missing=()

  local skill src_dir
  for skill in "${DEFAULT_SKILLS[@]}"; do
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

status_default_skills() {
  local target_skills_dir="${TARGET_ROOT}/.claude/skills"
  local installed=0
  local missing=()

  local skill
  for skill in "${DEFAULT_SKILLS[@]}"; do
    if [ -f "${target_skills_dir}/${skill}/SKILL.md" ]; then
      installed=$((installed + 1))
    else
      missing+=("$skill")
    fi
  done

  echo "[spw-kit] Default skills: installed=${installed}, missing=${#missing[@]}"
  if [ "${#missing[@]}" -gt 0 ]; then
    echo "[spw-kit] Missing in .claude/skills: ${missing[*]}"
  fi
}

cmd_install() {
  if [ "$#" -gt 0 ]; then
    echo "[spw-kit] Unexpected arguments for install: $*" >&2
    exit 1
  fi

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
}

cmd_skills() {
  if [ "$#" -gt 0 ]; then
    echo "[spw-kit] Unexpected arguments for skills: $*" >&2
    exit 1
  fi

  echo "[spw-kit] Installing default skills into: ${TARGET_ROOT}/.claude/skills"
  install_default_skills
}

cmd_status() {
  if [ "$#" -gt 0 ]; then
    echo "[spw-kit] Unexpected arguments for status: $*" >&2
    exit 1
  fi

  echo "[spw-kit] Status for project: ${TARGET_ROOT}"
  echo "[spw-kit] Kit dir: ${SCRIPT_DIR}"

  if [ -d "${TARGET_ROOT}/.claude" ]; then
    echo "[spw-kit] .claude: present"
  else
    echo "[spw-kit] .claude: missing"
  fi

  if [ -d "${TARGET_ROOT}/.spec-workflow" ]; then
    echo "[spw-kit] .spec-workflow: present"
  else
    echo "[spw-kit] .spec-workflow: missing"
  fi

  if [ -f "${CONFIG_PATH}" ]; then
    echo "[spw-kit] .spec-workflow/spw-config.toml: present"
    echo "[spw-kit] auto_install_defaults_on_spw_install=$(toml_bool_value skills auto_install_defaults_on_spw_install true)"
  else
    echo "[spw-kit] .spec-workflow/spw-config.toml: missing"
  fi

  if [ -f "${TARGET_ROOT}/.claude/settings.json" ]; then
    echo "[spw-kit] .claude/settings.json: present"
  else
    echo "[spw-kit] .claude/settings.json: missing"
  fi

  if [ -x "${TARGET_ROOT}/.claude/hooks/session-start-sync-tasks-template.sh" ]; then
    echo "[spw-kit] hook session-start-sync-tasks-template.sh: present"
  else
    echo "[spw-kit] hook session-start-sync-tasks-template.sh: missing"
  fi

  if [ -x "${TARGET_ROOT}/.claude/hooks/spw-statusline.js" ]; then
    echo "[spw-kit] hook spw-statusline.js: present"
  else
    echo "[spw-kit] hook spw-statusline.js: missing"
  fi

  status_default_skills
}

cmd="${1:-install}"
if [ "$#" -gt 0 ]; then
  shift
fi

case "$cmd" in
  -h|--help|help)
    show_help
    ;;
  install)
    cmd_install "$@"
    ;;
  skills)
    cmd_skills "$@"
    ;;
  status)
    cmd_status "$@"
    ;;
  *)
    echo "[spw-kit] Unknown command: $cmd" >&2
    show_help
    exit 1
    ;;
esac
