#!/usr/bin/env bash
set -euo pipefail

# Install or inspect the copy-ready kit in the current project root.
# Usage:
#   ./install.sh [install|skills|status]
#
# Behavior:
# - help (default): prints usage
# - install: copies kit files into current project (preserves user config across upgrades)
# - skills: installs default SPW skills into .claude/skills (best effort)
# - status: prints a quick summary of kit presence + default skills
# - Merges SPW hooks into .claude/settings.json (preserves non-SPW entries)
# - Agent Teams activation is driven by [agent_teams].enabled in spw-config.toml

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

ELIXIR_SKILLS=(
  "using-elixir-skills"
  "elixir-thinking"
  "elixir-anti-patterns"
  "phoenix-thinking"
  "ecto-thinking"
  "otp-thinking"
  "oban-thinking"
)

GENERAL_SKILLS=(
  "mermaid-architecture"
  "qa-validation-planning"
  "conventional-commits"
  "test-driven-development"
)

# All skills combined (backward compat for spw install)
DEFAULT_SKILLS=("${GENERAL_SKILLS[@]}" "${ELIXIR_SKILLS[@]}")

show_help() {
  cat <<'USAGE'
spw - install or inspect the SPW kit in the current project

Usage:
  spw
  spw install
  spw skills
  spw skills install [--elixir]
  spw status

Behavior:
- help (default): prints this help output.
- install: copies commands, hooks, templates, and config into cwd.
- skills: shows installed/available/missing status for all skill sets.
- skills install: installs general skills into .claude/skills (best effort).
  --elixir: installs Elixir-specific skills and patches config required lists.
- status: prints a quick summary of kit presence + default skills.

Notes:
- SPW hooks are auto-merged into .claude/settings.json (non-SPW entries preserved).
- User config (.spec-workflow/spw-config.toml) is preserved across installs via smart merge.
- Agent Teams activation is driven by [agent_teams].enabled in spw-config.toml.
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

setup_gitattributes() {
  local rule='.spec-workflow/specs/** linguist-generated=true'
  local gitattributes="${TARGET_ROOT}/.gitattributes"
  if [ ! -f "$gitattributes" ] || ! grep -qF "$rule" "$gitattributes"; then
    echo "$rule" >> "$gitattributes"
    echo "[spw-kit] Added .gitattributes rule for PR review optimization."
  fi
}

activate_teams_overlay_symlinks() {
  local active_dir="${TARGET_ROOT}/.claude/workflows/spw/overlays/active"
  local teams_dir="${TARGET_ROOT}/.claude/workflows/spw/overlays/teams"
  [ -d "$teams_dir" ] || { echo "[spw-kit] Team overlays not found; skipping." >&2; return 0; }
  mkdir -p "$active_dir"
  for overlay in "$teams_dir"/*.md; do
    [ -f "$overlay" ] || continue
    local name; name="$(basename "$overlay")"
    rm -f "${active_dir}/${name}"
    ln -s "../teams/${name}" "${active_dir}/${name}"
  done
  echo "[spw-kit] Activated team overlays via symlinks in overlays/active/."
}

deactivate_teams_overlay_symlinks() {
  local active_dir="${TARGET_ROOT}/.claude/workflows/spw/overlays/active"
  [ -d "$active_dir" ] || return 0
  for link in "$active_dir"/*.md; do
    [ -L "$link" ] || continue
    local name; name="$(basename "$link")"
    rm -f "$link"
    ln -s "../noop.md" "${active_dir}/${name}"
  done
  echo "[spw-kit] Deactivated team overlays (all symlinks → noop.md)."
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
  install_skill_set "all" "${DEFAULT_SKILLS[@]}"
}

install_skill_set() {
  local label="$1"
  shift
  local skills=("$@")
  local target_skills_dir="${TARGET_ROOT}/.claude/skills"
  mkdir -p "$target_skills_dir"

  local installed=0
  local skipped_existing=0
  local missing=()

  local skill src_dir
  for skill in "${skills[@]}"; do
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

  echo "[spw-kit] Skills (${label}): installed=${installed}, existing=${skipped_existing}, missing=${#missing[@]}"
  if [ "${#missing[@]}" -gt 0 ]; then
    echo "[spw-kit] Missing local skill sources (non-blocking): ${missing[*]}"
  fi
}

patch_config_elixir_skills() {
  resolve_config_path
  if [ ! -f "$CONFIG_PATH" ]; then
    echo "[spw-kit] No config file found; skipping Elixir config patch."
    return 0
  fi

  local elixir_required=("using-elixir-skills" "elixir-anti-patterns")
  local patched=false

  for skill in "${elixir_required[@]}"; do
    if ! grep -q "\"${skill}\"" "$CONFIG_PATH" 2>/dev/null; then
      patched=true
      break
    fi
  done

  if [ "$patched" = "false" ]; then
    echo "[spw-kit] Elixir skills already in config required lists."
    return 0
  fi

  # Patch required arrays in [skills.design] and [skills.implementation]
  local tmp; tmp="$(mktemp)"
  awk '
    BEGIN { in_target=0; in_array=0 }
    /^\[skills\.design\]/ || /^\[skills\.implementation\]/ { in_target=1 }
    /^\[/ && !/^\[skills\.design\]/ && !/^\[skills\.implementation\]/ { in_target=0 }

    in_target && /^required[[:space:]]*=/ {
      in_array=1
      print
      next
    }

    in_array && /\]/ {
      # Insert missing skills before closing bracket
      needs_using=1; needs_anti=1
      # Check what is already present by scanning backwards in output
    }

    { print }
  ' "$CONFIG_PATH" > "$tmp"

  # Simpler approach: use sed to insert entries before the ] in each target section
  rm -f "$tmp"
  tmp="$(mktemp)"
  cp "$CONFIG_PATH" "$tmp"

  for skill in "${elixir_required[@]}"; do
    if grep -q "\"${skill}\"" "$tmp"; then
      continue
    fi
    # Insert skill into required arrays in [skills.design] and [skills.implementation].
    # Handles both single-line (required = []) and multiline arrays.
    awk -v skill="$skill" '
      BEGIN { in_design=0; in_impl=0; in_req=0; done_design=0; done_impl=0 }
      /^\[skills\.design\]/ { in_design=1; in_impl=0 }
      /^\[skills\.implementation\]/ { in_impl=1; in_design=0 }
      /^\[/ && !/^\[skills\.design\]/ && !/^\[skills\.implementation\]/ { in_design=0; in_impl=0 }

      (in_design || in_impl) && /^required[[:space:]]*=/ {
        # Single-line array: required = [] or required = ["existing"]
        if ($0 ~ /\[.*\]/) {
          if ((in_design && !done_design) || (in_impl && !done_impl)) {
            # Replace closing ] with skill entry + ]
            line = $0
            sub(/\]/, "", line)
            # Check if array has existing entries
            if (line ~ /\[[[:space:]]*$/) {
              # Empty array: required = [ → insert skill
              printf "%s\"%s\"]\n", line, skill
            } else {
              # Has entries: required = ["x" → append with comma
              printf "%s, \"%s\"]\n", line, skill
            }
            if (in_design) done_design=1
            if (in_impl) done_impl=1
            next
          }
        } else {
          in_req=1
        }
      }

      in_req && /\]/ {
        if ((in_design && !done_design) || (in_impl && !done_impl)) {
          printf "  \"%s\",\n", skill
          if (in_design) done_design=1
          if (in_impl) done_impl=1
        }
        in_req=0
      }

      { print }
    ' "$tmp" > "${tmp}.2"
    mv "${tmp}.2" "$tmp"
  done

  mv "$tmp" "$CONFIG_PATH"
  echo "[spw-kit] Patched config: added using-elixir-skills and elixir-anti-patterns to required lists."
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

inject_snippet() {
  local target_file="$1"
  local snippet_file="$2"
  local marker_start="<!-- SPW-KIT-START"
  local marker_end="<!-- SPW-KIT-END -->"

  [ -f "$snippet_file" ] || return 0

  if [ -f "$target_file" ] && grep -qF "$marker_start" "$target_file"; then
    # Replace existing block (idempotent reinstall)
    local tmp; tmp="$(mktemp)"
    awk -v start="$marker_start" -v end="$marker_end" -v snippet="$snippet_file" '
      $0 ~ start { skip=1; while((getline line < snippet) > 0) print line; next }
      $0 ~ end { skip=0; next }
      !skip { print }
    ' "$target_file" > "$tmp"
    mv "$tmp" "$target_file"
  else
    # Append (first install or file doesn't exist)
    echo "" >> "$target_file"
    cat "$snippet_file" >> "$target_file"
  fi
}

cmd_install() {
  if [ "$#" -gt 0 ]; then
    echo "[spw-kit] Unexpected arguments for install: $*" >&2
    exit 1
  fi

  echo "[spw-kit] Installing into project: ${TARGET_ROOT}"

  # Backup user config before rsync overwrites it
  local config_backup=""
  resolve_config_path
  if [ -f "$CONFIG_PATH" ]; then
    config_backup="$(mktemp)"
    cp "$CONFIG_PATH" "$config_backup"
  fi

  # Copy only SPW runtime assets (avoid touching project root files like README.md)
  rsync -a "${SCRIPT_DIR}/.claude/" "${TARGET_ROOT}/.claude/"
  rsync -a "${SCRIPT_DIR}/.spec-workflow/" "${TARGET_ROOT}/.spec-workflow/"

  # PR review optimization: collapse spec-workflow files in GitHub diffs
  setup_gitattributes

  # Inject SPW dispatch instructions into project CLAUDE.md and AGENTS.md
  inject_snippet "${TARGET_ROOT}/CLAUDE.md" "${SCRIPT_DIR}/.claude.md.snippet"
  inject_snippet "${TARGET_ROOT}/AGENTS.md" "${SCRIPT_DIR}/.agents.md.snippet"
  echo "[spw-kit] Updated CLAUDE.md and AGENTS.md with SPW dispatch instructions."

  # Smart merge: preserve user config values with new template structure
  if [ -n "$config_backup" ]; then
    # Try Go binary's merge-config (suppress output in case the bash wrapper is
    # the only `spw` in PATH — it doesn't support `tools` and would abort).
    if command -v spw >/dev/null 2>&1 && spw tools merge-config "$CONFIG_PATH" "$config_backup" "$CONFIG_PATH" >/dev/null 2>&1; then
      echo "[spw-kit] Config merged: user values preserved, new keys added."
    else
      # Fallback: restore user backup as-is (keeps values, misses new keys)
      cp "$config_backup" "$CONFIG_PATH"
      echo "[spw-kit] spw Go binary not available; restored user config as-is (new template keys may be missing)."
    fi
    rm -f "$config_backup"
  fi

  if [ ! -f "${TARGET_ROOT}/.claude/settings.json" ]; then
    mkdir -p "${TARGET_ROOT}/.claude"
    cp "${SCRIPT_DIR}/.claude/settings.json.example" "${TARGET_ROOT}/.claude/settings.json"
    echo "[spw-kit] Created .claude/settings.json with SPW hooks."
  else
    if command -v spw >/dev/null 2>&1 && spw tools merge-settings >/dev/null 2>&1; then
      echo "[spw-kit] Hooks merged into .claude/settings.json."
    else
      echo "[spw-kit] spw Go binary not available; manually merge hooks from ${SCRIPT_DIR}/.claude/settings.json.example"
    fi
  fi

  resolve_config_path
  AUTO_INSTALL_SKILLS="$(toml_bool_value skills auto_install_defaults_on_spw_install true)"
  if [ "$AUTO_INSTALL_SKILLS" = "true" ]; then
    install_default_skills
  else
    echo "[spw-kit] Skipping default skills install (auto_install_defaults_on_spw_install=false)."
  fi

  # Agent Teams: activate/deactivate overlay symlinks based on config
  local teams_enabled
  teams_enabled="$(toml_bool_value agent_teams enabled false)"
  if [ "$teams_enabled" = "true" ]; then
    activate_teams_overlay_symlinks
    # Inject Agent Teams env into settings.json
    if command -v node >/dev/null 2>&1; then
      node -e '
        const fs = require("fs");
        const path = process.argv[1];
        const data = JSON.parse(fs.readFileSync(path, "utf8"));
        data.env = data.env || {};
        data.env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1";
        data.teammateMode = "in-process";
        fs.writeFileSync(path, JSON.stringify(data, null, 2));
      ' "${TARGET_ROOT}/.claude/settings.json"
      echo "[spw-kit] Enabled Agent Teams in settings.json (teammateMode=in-process)."
    else
      echo "[spw-kit] Agent Teams enabled in config but cannot inject settings."
      echo "[spw-kit] Add to .claude/settings.json manually:"
      echo "[spw-kit] - env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = \"1\""
      echo "[spw-kit] - teammateMode = \"in-process\" (or \"tmux\")"
    fi
  else
    deactivate_teams_overlay_symlinks
  fi

  echo "[spw-kit] Installation complete."
  echo "[spw-kit] Next step: adjust .spec-workflow/spw-config.toml"
}

diagnose_skill_set() {
  local label="$1"
  shift
  local skills=("$@")
  local target_skills_dir="${TARGET_ROOT}/.claude/skills"

  local installed=0
  local available=0
  local missing=0

  local skill
  for skill in "${skills[@]}"; do
    if [ -f "${target_skills_dir}/${skill}/SKILL.md" ]; then
      echo "    ✓ ${skill}"
      installed=$((installed + 1))
    elif src_dir="$(find_skill_source_dir "$skill" 2>/dev/null)"; then
      echo "    ○ ${skill} (available)"
      available=$((available + 1))
    else
      echo "    ✗ ${skill} (no source found)"
      missing=$((missing + 1))
    fi
  done

  echo "  ${label}: ${installed} installed, ${available} available, ${missing} missing"
}

cmd_skills() {
  local subcmd="${1:-}"

  case "$subcmd" in
    install)
      shift
      cmd_skills_install "$@"
      ;;
    "")
      echo "[spw-kit] Skills diagnosis:"
      diagnose_skill_set "General" "${GENERAL_SKILLS[@]}"
      diagnose_skill_set "Elixir" "${ELIXIR_SKILLS[@]}"
      ;;
    *)
      echo "[spw-kit] Unknown skills subcommand: $subcmd" >&2
      echo "Usage: spw skills | spw skills install [--elixir]" >&2
      exit 1
      ;;
  esac
}

cmd_skills_install() {
  local mode="general"
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --elixir) mode="elixir"; shift ;;
      *)
        echo "[spw-kit] Unknown flag for skills install: $1" >&2
        echo "Usage: spw skills install [--elixir]" >&2
        exit 1
        ;;
    esac
  done

  case "$mode" in
    elixir)
      echo "[spw-kit] Installing Elixir skills into: ${TARGET_ROOT}/.claude/skills"
      install_skill_set "elixir" "${ELIXIR_SKILLS[@]}"
      patch_config_elixir_skills
      ;;
    *)
      echo "[spw-kit] Installing general skills into: ${TARGET_ROOT}/.claude/skills"
      install_skill_set "general" "${GENERAL_SKILLS[@]}"
      ;;
  esac
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

  if command -v spw >/dev/null 2>&1; then
    echo "[spw-kit] spw binary: present ($(spw --version 2>/dev/null || echo 'unknown version'))"
  else
    echo "[spw-kit] spw binary: missing (hooks require spw in PATH)"
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
