#!/usr/bin/env bash
#
# Fail-open SessionStart hook:
# - Never blocks Claude startup due to hook/runtime errors.
# - Logs problems and exits 0.
set -uo pipefail

# Sync tasks-template based on .spec-workflow/spw-config.toml
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
CONFIG_PATH="${WORKSPACE_ROOT}/.spec-workflow/spw-config.toml"

if [ ! -f "$CONFIG_PATH" ]; then
  log "Config not found at ${CONFIG_PATH}. Nothing to sync."
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

to_non_negative_int() {
  local raw="${1:-0}"
  if [[ "$raw" =~ ^[0-9]+$ ]]; then
    printf '%s' "$raw"
  else
    printf '0'
  fi
}

cleanup_template_backups() {
  local target_path="$1"
  local keep_count="$2"
  local target_dir target_base
  local backups=()
  local idx

  target_dir="$(dirname "$target_path")"
  target_base="$(basename "$target_path")"

  while IFS= read -r file; do
    backups+=("$file")
  done < <(find "$target_dir" -maxdepth 1 -type f -name "${target_base}.bak-*" 2>/dev/null | sort -r)

  if [ "${#backups[@]}" -eq 0 ]; then
    return
  fi

  idx=0
  for file in "${backups[@]}"; do
    idx=$((idx + 1))
    if [ "$idx" -le "$keep_count" ]; then
      continue
    fi
    rm -f -- "$file"
    log "Removed old backup: ${file}"
  done
}

SYNC_ENABLED="$(to_bool "$(get_toml_value templates sync_tasks_template_on_session_start true)")"
if [ "$SYNC_ENABLED" != "true" ]; then
  log "Sync disabled by configuration."
  safe_exit
fi

TDD_DEFAULT="$(to_bool "$(get_toml_value execution tdd_default false)")"
TEMPLATE_MODE="$(get_toml_value templates tasks_template_mode auto)"
VARIANTS_DIR_REL="$(get_toml_value templates variants_dir .spec-workflow/user-templates/variants)"
TARGET_REL="$(get_toml_value templates active_tasks_template_path .spec-workflow/user-templates/tasks-template.md)"
FILE_ON="$(get_toml_value templates tasks_template_tdd_on_file tasks-template.tdd-on.md)"
FILE_OFF="$(get_toml_value templates tasks_template_tdd_off_file tasks-template.tdd-off.md)"
BACKUP_ENABLED="$(to_bool "$(get_toml_value safety backup_before_overwrite true)")"
DRY_RUN="$(to_bool "$(get_toml_value safety dry_run false)")"
FAIL_HARD="$(to_bool "$(get_toml_value safety fail_hard_on_missing_template false)")"
CLEANUP_BACKUPS="$(to_bool "$(get_toml_value safety cleanup_backups_after_sync false)")"
BACKUP_RETENTION_COUNT="$(to_non_negative_int "$(get_toml_value safety backup_retention_count 5)")"

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
    if [ "$FAIL_HARD" = "true" ]; then
      log "fail_hard_on_missing_template=true is ignored in hook mode (fail-open)."
    fi
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
  if [ "$FAIL_HARD" = "true" ]; then
    log "fail_hard_on_missing_template=true is ignored in hook mode (fail-open)."
  fi
  safe_exit
fi

mkdir -p "$(dirname "$TARGET_PATH")"

if [ -f "$TARGET_PATH" ] && cmp -s "$SOURCE_PATH" "$TARGET_PATH"; then
  log "Template is already synchronized (${MODE_NORMALIZED})."
  if [ "$CLEANUP_BACKUPS" = "true" ]; then
    cleanup_template_backups "$TARGET_PATH" "$BACKUP_RETENTION_COUNT"
  fi
  safe_exit
fi

if [ "$DRY_RUN" = "true" ]; then
  log "[dry-run] Would copy ${SOURCE_PATH} -> ${TARGET_PATH}"
  safe_exit
fi

if [ "$BACKUP_ENABLED" = "true" ] && [ -f "$TARGET_PATH" ]; then
  ts="$(date +%Y%m%d%H%M%S)"
  cp "$TARGET_PATH" "${TARGET_PATH}.bak-${ts}"
  log "Backup created: ${TARGET_PATH}.bak-${ts}"
fi

cp "$SOURCE_PATH" "$TARGET_PATH"
log "Template synchronized (${MODE_NORMALIZED}): ${TARGET_PATH}"

if [ "$CLEANUP_BACKUPS" = "true" ]; then
  cleanup_template_backups "$TARGET_PATH" "$BACKUP_RETENTION_COUNT"
fi

safe_exit
