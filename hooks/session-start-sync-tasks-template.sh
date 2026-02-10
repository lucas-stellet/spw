#!/usr/bin/env bash
#
# Fail-open SessionStart hook:
# - Never blocks Claude startup due to hook/runtime errors.
# - Logs problems and exits 0.
set -uo pipefail

# Sync tasks-template based on runtime config
# (canonical .spec-workflow/spw-config.toml, fallback legacy .spw/spw-config.toml)
# Intended usage: Claude/Codex SessionStart hook.

log() {
  printf '[spw-hook] %s\n' "$*"
}

safe_exit() {
  exit 0
}

trap 'rc=$?; if [ "$rc" -ne 0 ]; then log "Unexpected hook error (fail-open)."; fi; exit 0' EXIT

# Resolve workspace root
resolve_workspace_root() {
  if [ -n "${CLAUDE_PROJECT_DIR:-}" ] && [ -d "${CLAUDE_PROJECT_DIR}" ]; then
    printf '%s' "${CLAUDE_PROJECT_DIR}"
    return
  fi

  if git_root=$(git rev-parse --show-toplevel 2>/dev/null); then
    printf '%s' "$git_root"
    return
  fi

  pwd
}

WORKSPACE_ROOT="$(resolve_workspace_root)"
CONFIG_PATH_CANONICAL="${WORKSPACE_ROOT}/.spec-workflow/spw-config.toml"
CONFIG_PATH_LEGACY="${WORKSPACE_ROOT}/.spw/spw-config.toml"

if [ -f "$CONFIG_PATH_CANONICAL" ]; then
  CONFIG_PATH="$CONFIG_PATH_CANONICAL"
elif [ -f "$CONFIG_PATH_LEGACY" ]; then
  CONFIG_PATH="$CONFIG_PATH_LEGACY"
else
  log "Config not found at ${CONFIG_PATH_CANONICAL} (fallback: ${CONFIG_PATH_LEGACY}). Nothing to sync."
  safe_exit
fi

# Read a TOML key value from a section.
# Supported format: key = value (with or without quotes)
get_toml_value() {
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
        # ignore comments and empty lines
        if ($0 ~ /^[ \t]*#/ || $0 ~ /^[ \t]*$/) {
          next
        }

        if ($0 ~ "^[ \\t]*" wanted_key "[ \\t]*=") {
          line = $0
          sub(/^[^=]*=[ \t]*/, "", line)
          sub(/[ \t]*#.*/, "", line)
          line = trim(line)

          # remove single/double quotes if present
          if ((line ~ /^".*"$/) || (line ~ /^\047.*\047$/)) {
            line = substr(line, 2, length(line) - 2)
          }

          print line
          exit
        }
      }
    ' "$CONFIG_PATH"
  } || true)"

  if [ -z "$value" ]; then
    printf '%s' "$default_value"
  else
    printf '%s' "$value"
  fi
}

to_bool() {
  local raw="${1:-}"
  local raw_lower
  raw_lower="$(printf '%s' "$raw" | tr '[:upper:]' '[:lower:]')"
  case "$raw_lower" in
    true|1|yes|on) printf 'true' ;;
    false|0|no|off) printf 'false' ;;
    *) printf 'false' ;;
  esac
}

SYNC_ENABLED="$(to_bool "$(get_toml_value templates sync_tasks_template_on_session_start true)")"
if [ "$SYNC_ENABLED" != "true" ]; then
  log "Sync disabled by configuration."
  safe_exit
fi

TDD_DEFAULT="$(to_bool "$(get_toml_value execution tdd_default false)")"
TEMPLATE_MODE="$(get_toml_value templates tasks_template_mode auto)"
VARIANTS_DIR_REL=".spec-workflow/user-templates/variants"
TARGET_REL=".spec-workflow/user-templates/tasks-template.md"
FILE_ON="tasks-template.tdd-on.md"
FILE_OFF="tasks-template.tdd-off.md"
BACKUP_ENABLED="$(to_bool "$(get_toml_value safety backup_before_overwrite true)")"
DRY_RUN="$(to_bool "${SPW_DRY_RUN:-false}")"

MODE_NORMALIZED="$(printf '%s' "$TEMPLATE_MODE" | tr '[:upper:]' '[:lower:]')"
if [ "$MODE_NORMALIZED" = "auto" ]; then
  if [ "$TDD_DEFAULT" = "true" ]; then
    MODE_NORMALIZED="on"
  else
    MODE_NORMALIZED="off"
  fi
fi

case "$MODE_NORMALIZED" in
  on|off) ;;
  *)
    log "Invalid tasks_template_mode: '${TEMPLATE_MODE}'. Use auto|on|off."
    safe_exit
    ;;
esac

VARIANTS_DIR="${WORKSPACE_ROOT}/${VARIANTS_DIR_REL}"
TARGET_PATH="${WORKSPACE_ROOT}/${TARGET_REL}"

if [ "$MODE_NORMALIZED" = "on" ]; then
  SOURCE_PATH="${VARIANTS_DIR}/${FILE_ON}"
else
  SOURCE_PATH="${VARIANTS_DIR}/${FILE_OFF}"
fi

if [ ! -f "$SOURCE_PATH" ]; then
  log "Source template not found: ${SOURCE_PATH}"
  safe_exit
fi

mkdir -p "$(dirname "$TARGET_PATH")"

if [ -f "$TARGET_PATH" ] && cmp -s "$SOURCE_PATH" "$TARGET_PATH"; then
  log "Template is already synchronized (${MODE_NORMALIZED})."
  safe_exit
fi

if [ "$DRY_RUN" = "true" ]; then
  log "[dry-run] Would copy ${SOURCE_PATH} -> ${TARGET_PATH}"
  safe_exit
fi

if [ "$BACKUP_ENABLED" = "true" ] && [ -f "$TARGET_PATH" ]; then
  cp "$TARGET_PATH" "${TARGET_PATH}.bak"
  log "Backup created: ${TARGET_PATH}.bak"
fi

cp "$SOURCE_PATH" "$TARGET_PATH"
log "Template synchronized (${MODE_NORMALIZED}): ${TARGET_PATH}"

safe_exit
