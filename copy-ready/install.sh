#!/usr/bin/env bash
set -euo pipefail

# Install or inspect the copy-ready kit in the current project root.
# Usage:
#   ./install.sh [install|skills|status]
#
# Behavior:
# - help (default): prints usage
# - install: copies kit files into current project (preserves user config across upgrades)
# - skills: installs default Oraculo skills into .claude/skills (best effort)
# - status: prints a quick summary of kit presence + default skills
# - Merges Oraculo hooks into .claude/settings.json (preserves non-Oraculo entries)
# - Agent Teams activation is driven by [agent_teams].enabled in oraculo.toml

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_ROOT="$(pwd)"
CONFIG_PATH_CANONICAL="${TARGET_ROOT}/.spec-workflow/oraculo.toml"
CONFIG_PATH_LEGACY="${TARGET_ROOT}/.oraculo/oraculo.toml"
CONFIG_PATH="${CONFIG_PATH_CANONICAL}"
# Resolve repository root robustly.
# Typical cached layout used by `oraculo` wrapper:
#   <cache>/repos/<repo>/copy-ready/install.sh
# where repo root is one level above SCRIPT_DIR.
ORACULO_REPO_ROOT_CANDIDATE_1="$(cd "${SCRIPT_DIR}/.." && pwd)"
ORACULO_REPO_ROOT_CANDIDATE_2="$(cd "${SCRIPT_DIR}/../.." && pwd)"
if [ -d "${ORACULO_REPO_ROOT_CANDIDATE_1}/skills" ] || [ -d "${ORACULO_REPO_ROOT_CANDIDATE_1}/.git" ]; then
  ORACULO_REPO_ROOT="${ORACULO_REPO_ROOT_CANDIDATE_1}"
else
  ORACULO_REPO_ROOT="${ORACULO_REPO_ROOT_CANDIDATE_2}"
fi
SUPERPOWERS_SKILLS_DIR="${ORACULO_SUPERPOWERS_SKILLS_DIR:-${ORACULO_REPO_ROOT}/superpowers/skills}"

resolve_config_path() {
  if [ -f "$CONFIG_PATH_CANONICAL" ]; then
    CONFIG_PATH="$CONFIG_PATH_CANONICAL"
  elif [ -f "$CONFIG_PATH_LEGACY" ]; then
    CONFIG_PATH="$CONFIG_PATH_LEGACY"
  else
    CONFIG_PATH="$CONFIG_PATH_CANONICAL"
  fi
}

GENERAL_SKILLS=(
  "mermaid-architecture"
  "qa-validation-planning"
  "conventional-commits"
  "test-driven-development"
)

show_help() {
  cat <<'USAGE'
oraculo - install or inspect the Oraculo kit in the current project

Usage:
  oraculo
  oraculo install [--global]
  oraculo init
  oraculo skills
  oraculo skills install
  oraculo status

Behavior:
- help (default): prints this help output.
- install: copies commands, hooks, templates, and config into cwd (full local install).
  --global: installs commands, workflows, hooks, and skills to ~/.claude/ for all projects.
- init: initializes project-specific config, templates, snippets, and .gitattributes.
  Use with a global install — does NOT copy commands/workflows locally.
- skills: shows installed/available/missing status for all skill sets.
- skills install: installs general skills into .claude/skills.
- status: prints a quick summary of kit presence + default skills.

Notes:
- Oraculo hooks are auto-merged into .claude/settings.json (non-Oraculo entries preserved).
- User config (.spec-workflow/oraculo.toml) is preserved across installs via smart merge.
- Agent Teams activation is driven by [agent_teams].enabled in oraculo.toml.
- Global + init coexist: local installs take precedence over global (Claude Code native behavior).
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
    echo "[oraculo-kit] Added .gitattributes rule for PR review optimization."
  fi
}

activate_teams_overlay_symlinks() {
  local active_dir="${TARGET_ROOT}/.claude/workflows/oraculo/overlays/active"
  local teams_dir="${TARGET_ROOT}/.claude/workflows/oraculo/overlays/teams"
  [ -d "$teams_dir" ] || { echo "[oraculo-kit] Team overlays not found; skipping." >&2; return 0; }
  mkdir -p "$active_dir"
  for overlay in "$teams_dir"/*.md; do
    [ -f "$overlay" ] || continue
    local name; name="$(basename "$overlay")"
    rm -f "${active_dir}/${name}"
    ln -s "../teams/${name}" "${active_dir}/${name}"
  done
  echo "[oraculo-kit] Activated team overlays via symlinks in overlays/active/."
}

deactivate_teams_overlay_symlinks() {
  local active_dir="${TARGET_ROOT}/.claude/workflows/oraculo/overlays/active"
  [ -d "$active_dir" ] || return 0
  for link in "$active_dir"/*.md; do
    [ -L "$link" ] || continue
    local name; name="$(basename "$link")"
    rm -f "$link"
    ln -s "../noop.md" "${active_dir}/${name}"
  done
  echo "[oraculo-kit] Deactivated team overlays (all symlinks → noop.md)."
}

find_skill_source_dir() {
  local skill="$1"
  local candidates=(
    "${ORACULO_REPO_ROOT}/skills/${skill}"
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

  echo "[oraculo-kit] Skills (${label}): installed=${installed}, existing=${skipped_existing}, missing=${#missing[@]}"
  if [ "${#missing[@]}" -gt 0 ]; then
    echo "[oraculo-kit] Missing local skill sources (non-blocking): ${missing[*]}"
  fi
}

status_default_skills() {
  local target_skills_dir="${TARGET_ROOT}/.claude/skills"
  local installed=0
  local missing=()

  local skill
  for skill in "${GENERAL_SKILLS[@]}"; do
    if [ -f "${target_skills_dir}/${skill}/SKILL.md" ]; then
      installed=$((installed + 1))
    else
      missing+=("$skill")
    fi
  done

  echo "[oraculo-kit] General skills: installed=${installed}, missing=${#missing[@]}"
  if [ "${#missing[@]}" -gt 0 ]; then
    echo "[oraculo-kit] Missing in .claude/skills: ${missing[*]}"
  fi
}

inject_snippet() {
  local target_file="$1"
  local snippet_file="$2"
  local marker_start="<!-- ORACULO-KIT-START"
  local marker_end="<!-- ORACULO-KIT-END -->"

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

cmd_install_global() {
  local global_root="${HOME}"
  echo "[oraculo-kit] Installing globally into: ${global_root}"

  # 1. Command stubs → ~/.claude/commands/oraculo/
  rsync -a "${SCRIPT_DIR}/.claude/commands/" "${global_root}/.claude/commands/"
  echo "[oraculo-kit] Copied command stubs to ~/.claude/commands/oraculo/"

  # 2. Workflows → ~/.claude/workflows/oraculo/
  rsync -a "${SCRIPT_DIR}/.claude/workflows/" "${global_root}/.claude/workflows/"
  echo "[oraculo-kit] Copied workflows to ~/.claude/workflows/oraculo/"

  # 3. Settings.json → ~/.claude/settings.json
  if [ ! -f "${global_root}/.claude/settings.json" ]; then
    mkdir -p "${global_root}/.claude"
    cp "${SCRIPT_DIR}/.claude/settings.json.example" "${global_root}/.claude/settings.json"
    echo "[oraculo-kit] Created ~/.claude/settings.json with Oraculo hooks."
  else
    if command -v oraculo >/dev/null 2>&1 && oraculo tools merge-settings --global >/dev/null 2>&1; then
      echo "[oraculo-kit] Hooks merged into ~/.claude/settings.json."
    else
      echo "[oraculo-kit] oraculo Go binary not available; manually merge hooks from ${SCRIPT_DIR}/.claude/settings.json.example"
    fi
  fi

  # 4. General skills only → ~/.claude/skills/
  local saved_target_root="${TARGET_ROOT}"
  TARGET_ROOT="${global_root}"
  install_skill_set "general" "${GENERAL_SKILLS[@]}"
  TARGET_ROOT="${saved_target_root}"

  # 5. Overlay symlinks → noop by default
  local active_dir="${global_root}/.claude/workflows/oraculo/overlays/active"
  mkdir -p "$active_dir"
  # Ensure noop.md exists
  local noop_path="${global_root}/.claude/workflows/oraculo/overlays/noop.md"
  if [ ! -f "$noop_path" ]; then
    mkdir -p "$(dirname "$noop_path")"
    echo "<!-- noop overlay -->" > "$noop_path"
  fi
  deactivate_teams_overlay_symlinks_at "${global_root}"

  echo "[oraculo-kit] Global installation complete."
  echo "[oraculo-kit] Use 'oraculo init' in each project to set up project-specific config."
}

deactivate_teams_overlay_symlinks_at() {
  local root="$1"
  local active_dir="${root}/.claude/workflows/oraculo/overlays/active"
  [ -d "$active_dir" ] || return 0
  # Create noop symlinks for all commands
  local commands=("discover" "plan" "design-research" "design-draft" "tasks-plan" "tasks-check" "exec" "checkpoint" "post-mortem" "qa" "qa-check" "qa-exec" "status")
  local cmd_name
  for cmd_name in "${commands[@]}"; do
    rm -f "${active_dir}/${cmd_name}.md"
    ln -s "../noop.md" "${active_dir}/${cmd_name}.md"
  done
  echo "[oraculo-kit] Overlay symlinks set to noop (teams disabled)."
}

cmd_init() {
  if [ "$#" -gt 0 ]; then
    echo "[oraculo-kit] Unexpected arguments for init: $*" >&2
    exit 1
  fi

  echo "[oraculo-kit] Initializing project: ${TARGET_ROOT}"

  # 1. Write defaults (.spec-workflow/oraculo.toml + user-templates/)
  rsync -a "${SCRIPT_DIR}/.spec-workflow/" "${TARGET_ROOT}/.spec-workflow/"
  echo "[oraculo-kit] Copied config and templates to .spec-workflow/"

  # 2. Inject snippets (CLAUDE.md, AGENTS.md)
  inject_snippet "${TARGET_ROOT}/CLAUDE.md" "${SCRIPT_DIR}/.claude.md.snippet"
  inject_snippet "${TARGET_ROOT}/AGENTS.md" "${SCRIPT_DIR}/.agents.md.snippet"
  echo "[oraculo-kit] Updated CLAUDE.md and AGENTS.md with Oraculo dispatch instructions."

  # 3. Setup .gitattributes
  setup_gitattributes

  # 4. Diagnose global install presence
  local global_cmds="${HOME}/.claude/commands/oraculo"
  if [ -d "$global_cmds" ] && [ "$(ls -1 "$global_cmds" 2>/dev/null | wc -l)" -gt 0 ]; then
    echo "[oraculo-kit] Global install detected: commands found in ~/.claude/commands/oraculo/"
  else
    echo "[oraculo-kit] No global install detected. Run 'oraculo install --global' or 'oraculo install' for a full local install."
  fi

  echo "[oraculo-kit] Project initialized."
  echo "[oraculo-kit] Next step: adjust .spec-workflow/oraculo.toml"
}

cmd_install() {
  if [ "$#" -gt 0 ]; then
    echo "[oraculo-kit] Unexpected arguments for install: $*" >&2
    exit 1
  fi

  echo "[oraculo-kit] Installing into project: ${TARGET_ROOT}"

  # Backup user config before rsync overwrites it
  local config_backup=""
  resolve_config_path
  if [ -f "$CONFIG_PATH" ]; then
    config_backup="$(mktemp)"
    cp "$CONFIG_PATH" "$config_backup"
  fi

  # Copy only Oraculo runtime assets (avoid touching project root files like README.md)
  rsync -a "${SCRIPT_DIR}/.claude/" "${TARGET_ROOT}/.claude/"
  rsync -a "${SCRIPT_DIR}/.spec-workflow/" "${TARGET_ROOT}/.spec-workflow/"

  # PR review optimization: collapse spec-workflow files in GitHub diffs
  setup_gitattributes

  # Inject Oraculo dispatch instructions into project CLAUDE.md and AGENTS.md
  inject_snippet "${TARGET_ROOT}/CLAUDE.md" "${SCRIPT_DIR}/.claude.md.snippet"
  inject_snippet "${TARGET_ROOT}/AGENTS.md" "${SCRIPT_DIR}/.agents.md.snippet"
  echo "[oraculo-kit] Updated CLAUDE.md and AGENTS.md with Oraculo dispatch instructions."

  # Smart merge: preserve user config values with new template structure
  if [ -n "$config_backup" ]; then
    # Try Go binary's merge-config (suppress output in case the bash wrapper is
    # the only `oraculo` in PATH — it doesn't support `tools` and would abort).
    if command -v oraculo >/dev/null 2>&1 && oraculo tools merge-config "$CONFIG_PATH" "$config_backup" "$CONFIG_PATH" >/dev/null 2>&1; then
      echo "[oraculo-kit] Config merged: user values preserved, new keys added."
    else
      # Fallback: restore user backup as-is (keeps values, misses new keys)
      cp "$config_backup" "$CONFIG_PATH"
      echo "[oraculo-kit] oraculo Go binary not available; restored user config as-is (new template keys may be missing)."
    fi
    rm -f "$config_backup"
  fi

  if [ ! -f "${TARGET_ROOT}/.claude/settings.json" ]; then
    mkdir -p "${TARGET_ROOT}/.claude"
    cp "${SCRIPT_DIR}/.claude/settings.json.example" "${TARGET_ROOT}/.claude/settings.json"
    echo "[oraculo-kit] Created .claude/settings.json with Oraculo hooks."
  else
    if command -v oraculo >/dev/null 2>&1 && oraculo tools merge-settings >/dev/null 2>&1; then
      echo "[oraculo-kit] Hooks merged into .claude/settings.json."
    else
      echo "[oraculo-kit] oraculo Go binary not available; manually merge hooks from ${SCRIPT_DIR}/.claude/settings.json.example"
    fi
  fi

  resolve_config_path
  AUTO_INSTALL_SKILLS="$(toml_bool_value skills auto_install_defaults_on_oraculo_install true)"
  if [ "$AUTO_INSTALL_SKILLS" = "true" ]; then
    install_skill_set "general" "${GENERAL_SKILLS[@]}"
  else
    echo "[oraculo-kit] Skipping default skills install (auto_install_defaults_on_oraculo_install=false)."
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
      echo "[oraculo-kit] Enabled Agent Teams in settings.json (teammateMode=in-process)."
    else
      echo "[oraculo-kit] Agent Teams enabled in config but cannot inject settings."
      echo "[oraculo-kit] Add to .claude/settings.json manually:"
      echo "[oraculo-kit] - env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = \"1\""
      echo "[oraculo-kit] - teammateMode = \"in-process\" (or \"tmux\")"
    fi
  else
    deactivate_teams_overlay_symlinks
  fi

  echo "[oraculo-kit] Installation complete."
  echo "[oraculo-kit] Next step: adjust .spec-workflow/oraculo.toml"
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
      echo "[oraculo-kit] Skills diagnosis:"
      diagnose_skill_set "General" "${GENERAL_SKILLS[@]}"
      ;;
    *)
      echo "[oraculo-kit] Unknown skills subcommand: $subcmd" >&2
      echo "Usage: oraculo skills | oraculo skills install" >&2
      exit 1
      ;;
  esac
}

cmd_skills_install() {
  if [ "$#" -gt 0 ]; then
    echo "[oraculo-kit] Unknown flag for skills install: $1" >&2
    echo "Usage: oraculo skills install" >&2
    exit 1
  fi

  echo "[oraculo-kit] Installing general skills into: ${TARGET_ROOT}/.claude/skills"
  install_skill_set "general" "${GENERAL_SKILLS[@]}"
}

cmd_status() {
  if [ "$#" -gt 0 ]; then
    echo "[oraculo-kit] Unexpected arguments for status: $*" >&2
    exit 1
  fi

  echo "[oraculo-kit] Status for project: ${TARGET_ROOT}"
  echo "[oraculo-kit] Kit dir: ${SCRIPT_DIR}"
  resolve_config_path

  if [ -d "${TARGET_ROOT}/.claude" ]; then
    echo "[oraculo-kit] .claude: present"
  else
    echo "[oraculo-kit] .claude: missing"
  fi

  if [ -d "${TARGET_ROOT}/.spec-workflow" ]; then
    echo "[oraculo-kit] .spec-workflow: present"
  else
    echo "[oraculo-kit] .spec-workflow: missing"
  fi

  if [ -f "${CONFIG_PATH}" ]; then
    if [ "${CONFIG_PATH}" = "${CONFIG_PATH_CANONICAL}" ]; then
      echo "[oraculo-kit] .spec-workflow/oraculo.toml: present"
    else
      echo "[oraculo-kit] .spec-workflow/oraculo.toml: missing (using legacy .oraculo/oraculo.toml)"
    fi
    echo "[oraculo-kit] auto_install_defaults_on_oraculo_install=$(toml_bool_value skills auto_install_defaults_on_oraculo_install true)"
  else
    echo "[oraculo-kit] .spec-workflow/oraculo.toml: missing"
  fi

  if [ -f "${TARGET_ROOT}/.claude/settings.json" ]; then
    echo "[oraculo-kit] .claude/settings.json: present"
  else
    echo "[oraculo-kit] .claude/settings.json: missing"
  fi

  if command -v oraculo >/dev/null 2>&1; then
    echo "[oraculo-kit] oraculo binary: present ($(oraculo --version 2>/dev/null || echo 'unknown version'))"
  else
    echo "[oraculo-kit] oraculo binary: missing (hooks require oraculo in PATH)"
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
    if [ "${1:-}" = "--global" ]; then
      shift
      cmd_install_global "$@"
    else
      cmd_install "$@"
    fi
    ;;
  init)
    cmd_init "$@"
    ;;
  skills)
    cmd_skills "$@"
    ;;
  status)
    cmd_status "$@"
    ;;
  *)
    echo "[oraculo-kit] Unknown command: $cmd" >&2
    show_help
    exit 1
    ;;
esac
