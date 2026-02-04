#!/usr/bin/env bash
set -euo pipefail

# Sync de tasks-template com base em .spec-workflow/spw-config.toml
# Uso esperado: hook de SessionStart do Claude/Codex.

log() {
  printf '[spw-hook] %s\n' "$*"
}

# Resolve raiz do workspace
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
  log "Config nao encontrada em ${CONFIG_PATH}. Nada a sincronizar."
  exit 0
fi

# Le valor de uma chave TOML em uma secao.
# Suporta formato: key = value  (com ou sem aspas)
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
        # ignora comentarios e linhas vazias
        if ($0 ~ /^[ \t]*#/ || $0 ~ /^[ \t]*$/) {
          next
        }

        if ($0 ~ "^[ \\t]*" wanted_key "[ \\t]*=") {
          line = $0
          sub(/^[^=]*=[ \t]*/, "", line)
          sub(/[ \t]*#.*/, "", line)
          line = trim(line)

          # remove aspas simples ou duplas se houver
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
  case "${raw,,}" in
    true|1|yes|on) printf 'true' ;;
    false|0|no|off) printf 'false' ;;
    *) printf 'false' ;;
  esac
}

SYNC_ENABLED="$(to_bool "$(get_toml_value templates sync_tasks_template_on_session_start true)")"
if [ "$SYNC_ENABLED" != "true" ]; then
  log "Sync desabilitado por configuracao."
  exit 0
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

MODE_NORMALIZED="${TEMPLATE_MODE,,}"
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
    log "tasks_template_mode invalido: '${TEMPLATE_MODE}'. Use auto|on|off."
    if [ "$FAIL_HARD" = "true" ]; then
      exit 1
    fi
    exit 0
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
  log "Template de origem nao encontrado: ${SOURCE_PATH}"
  if [ "$FAIL_HARD" = "true" ]; then
    exit 1
  fi
  exit 0
fi

mkdir -p "$(dirname "$TARGET_PATH")"

if [ -f "$TARGET_PATH" ] && cmp -s "$SOURCE_PATH" "$TARGET_PATH"; then
  log "Template ja esta sincronizado (${MODE_NORMALIZED})."
  exit 0
fi

if [ "$DRY_RUN" = "true" ]; then
  log "[dry-run] Copiaria ${SOURCE_PATH} -> ${TARGET_PATH}"
  exit 0
fi

if [ "$BACKUP_ENABLED" = "true" ] && [ -f "$TARGET_PATH" ]; then
  ts="$(date +%Y%m%d%H%M%S)"
  cp "$TARGET_PATH" "${TARGET_PATH}.bak-${ts}"
  log "Backup criado: ${TARGET_PATH}.bak-${ts}"
fi

cp "$SOURCE_PATH" "$TARGET_PATH"
log "Template sincronizado (${MODE_NORMALIZED}): ${TARGET_PATH}"

exit 0
