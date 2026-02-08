#!/usr/bin/env bash
set -euo pipefail

# Install or inspect the copy-ready kit in the current project root.
# Usage:
#   ./install.sh [install|skills|status]
#
# Behavior:
# - help (default): prints usage
# - install: copies kit files into current project
# - install --enable-teams: enables Agent Teams in config/settings and overlays team command pack
# - skills: installs default SPW skills into .claude/skills (best effort)
# - status: prints a quick summary of kit presence + default skills
# - Does not overwrite .claude/settings.json (prints merge instruction instead)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_ROOT="$(pwd)"
CONFIG_PATH_CANONICAL="${TARGET_ROOT}/.spec-workflow/spw-config.toml"
CONFIG_PATH_LEGACY="${TARGET_ROOT}/.spw/spw-config.toml"
CONFIG_PATH="${CONFIG_PATH_CANONICAL}"
# Resolve repository root robustly.
# Typical cached layout used by `spw` wrapper:
#   <cache>/repos/<repo>/copy-ready/install.sh
# where repo root is one level above SCRIPT_DIR.
SPW_REPO_ROOT_CANDIDATE_1="$(cd "${SCRIPT_DIR}/.." && pwd)"
SPW_REPO_ROOT_CANDIDATE_2="$(cd "${SCRIPT_DIR}/../.." && pwd)"
if [ -d "${SPW_REPO_ROOT_CANDIDATE_1}/skills" ] || [ -d "${SPW_REPO_ROOT_CANDIDATE_1}/.git" ]; then
  SPW_REPO_ROOT="${SPW_REPO_ROOT_CANDIDATE_1}"
else
  SPW_REPO_ROOT="${SPW_REPO_ROOT_CANDIDATE_2}"
fi
SUPERPOWERS_SKILLS_DIR="${SPW_SUPERPOWERS_SKILLS_DIR:-${SPW_REPO_ROOT}/superpowers/skills}"

resolve_config_path() {
  if [ -f "$CONFIG_PATH_CANONICAL" ]; then
    CONFIG_PATH="$CONFIG_PATH_CANONICAL"
  elif [ -f "$CONFIG_PATH_LEGACY" ]; then
    CONFIG_PATH="$CONFIG_PATH_LEGACY"
  else
    CONFIG_PATH="$CONFIG_PATH_CANONICAL"
  fi
}

DEFAULT_SKILLS=(
  "using-elixir-skills"
  "elixir-thinking"
  "elixir-anti-patterns"
  "phoenix-thinking"
  "ecto-thinking"
  "otp-thinking"
  "oban-thinking"
  "mermaid-architecture"
  "qa-validation-planning"
  "conventional-commits"
  "test-driven-development"
)

show_help() {
  cat <<'USAGE'
spw - install or inspect the SPW kit in the current project

Usage:
  spw
  spw install
  spw install --enable-teams
  spw skills
  spw status

Behavior:
- help (default): prints this help output.
- install: copies commands, hooks, templates, and config into cwd.
- install --enable-teams: enables Agent Teams in config/settings and overlays team command pack.
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

enable_agent_teams() {
  if [ ! -f "$CONFIG_PATH" ]; then
    echo "[spw-kit] Unable to enable Agent Teams: missing ${CONFIG_PATH}" >&2
    return 0
  fi

  local tmp_file
  tmp_file="$(mktemp)"

  awk '
    BEGIN { in_section = 0 }
    /^[ \t]*\[/ {
      in_section = ($0 == "[agent_teams]")
    }
    {
      if (in_section && $0 ~ /^[ \t]*enabled[ \t]*=/) {
        print "enabled = true"
        next
      }
      print
    }
  ' "$CONFIG_PATH" > "$tmp_file"

  mv "$tmp_file" "$CONFIG_PATH"
}

apply_teams_settings() {
  local settings_path="${TARGET_ROOT}/.claude/settings.json"
  if ! command -v node >/dev/null 2>&1; then
    echo "[spw-kit] Node not found; add Agent Teams settings manually:"
    echo "[spw-kit] - env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = \"1\""
    echo "[spw-kit] - teammateMode = \"in-process\" (or \"tmux\")"
    return 0
  fi

  node -e '
    const fs = require("fs");
    const path = process.argv[1];
    const data = JSON.parse(fs.readFileSync(path, "utf8"));
    data.env = data.env || {};
    data.env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1";
    data.teammateMode = "in-process";
    fs.writeFileSync(path, JSON.stringify(data, null, 2));
  ' "$settings_path"
  echo "[spw-kit] Enabled Agent Teams in ${settings_path} (teammateMode=in-process)."
}

apply_teams_command_pack() {
  local source_dir="${SCRIPT_DIR}/.claude/commands/spw-teams"
  local target_dir="${TARGET_ROOT}/.claude/commands/spw"

  if [ ! -d "$source_dir" ]; then
    echo "[spw-kit] Team command pack not found at ${source_dir}; skipping overlay."
    return 0
  fi

  mkdir -p "$target_dir"
  rsync -a "${source_dir}/" "${target_dir}/"
  echo "[spw-kit] Applied team command pack from .claude/commands/spw-teams to .claude/commands/spw."
}

find_skill_source_dir() {
  local skill="$1"
  local candidates=(
    "${SPW_REPO_ROOT}/skills/${skill}"
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
  local enable_teams="false"
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --enable-teams)
        enable_teams="true"
        shift
        ;;
      *)
        echo "[spw-kit] Unexpected arguments for install: $*" >&2
        exit 1
        ;;
    esac
  done

  echo "[spw-kit] Installing into project: ${TARGET_ROOT}"

  # Copy only SPW runtime assets (avoid touching project root files like README.md)
  rsync -a "${SCRIPT_DIR}/.claude/" "${TARGET_ROOT}/.claude/"
  rsync -a "${SCRIPT_DIR}/.spec-workflow/" "${TARGET_ROOT}/.spec-workflow/"

  local created_settings="false"
  if [ ! -f "${TARGET_ROOT}/.claude/settings.json" ]; then
    mkdir -p "${TARGET_ROOT}/.claude"
    cp "${SCRIPT_DIR}/.claude/settings.json.example" "${TARGET_ROOT}/.claude/settings.json"
    echo "[spw-kit] Created .claude/settings.json with SessionStart hook."
    created_settings="true"
  else
    echo "[spw-kit] .claude/settings.json already exists."
    echo "[spw-kit] Manually merge hook block from ${SCRIPT_DIR}/.claude/settings.json.example"
  fi

  chmod +x "${TARGET_ROOT}/.claude/hooks/session-start-sync-tasks-template.sh" || true
  chmod +x "${TARGET_ROOT}/.claude/hooks/spw-statusline.js" || true

  resolve_config_path
  AUTO_INSTALL_SKILLS="$(toml_bool_value skills auto_install_defaults_on_spw_install true)"
  if [ "$AUTO_INSTALL_SKILLS" = "true" ]; then
    install_default_skills
  else
    echo "[spw-kit] Skipping default skills install (auto_install_defaults_on_spw_install=false)."
  fi

  if [ "$enable_teams" = "true" ]; then
    enable_agent_teams
    apply_teams_command_pack
    if [ "$created_settings" = "true" ]; then
      apply_teams_settings
    else
      echo "[spw-kit] Agent Teams enabled in config."
      echo "[spw-kit] Add to .claude/settings.json manually:"
      echo "[spw-kit] - env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = \"1\""
      echo "[spw-kit] - teammateMode = \"in-process\" (or \"tmux\")"
    fi
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
  resolve_config_path

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
    if [ "${CONFIG_PATH}" = "${CONFIG_PATH_CANONICAL}" ]; then
      echo "[spw-kit] .spec-workflow/spw-config.toml: present"
    else
      echo "[spw-kit] .spec-workflow/spw-config.toml: missing (using legacy .spw/spw-config.toml)"
    fi
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

cmd="${1:-help}"
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
