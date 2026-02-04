#!/usr/bin/env bash
set -euo pipefail

# Instala o kit copy-ready na raiz do projeto corrente.
# Uso:
#   ./install.sh
#
# Comportamento:
# - Copia arquivos do kit para o projeto atual
# - Nao sobrescreve .claude/settings.json (apenas avisa para merge manual)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_ROOT="$(pwd)"

echo "[spw-kit] Instalando no projeto: ${TARGET_ROOT}"

# Copia tudo, exceto o install e settings example (tratados separadamente)
rsync -a \
  --exclude 'install.sh' \
  --exclude '.claude/settings.json.example' \
  "${SCRIPT_DIR}/" "${TARGET_ROOT}/"

if [ ! -f "${TARGET_ROOT}/.claude/settings.json" ]; then
  mkdir -p "${TARGET_ROOT}/.claude"
  cp "${SCRIPT_DIR}/.claude/settings.json.example" "${TARGET_ROOT}/.claude/settings.json"
  echo "[spw-kit] Criado .claude/settings.json com hook de SessionStart."
else
  echo "[spw-kit] .claude/settings.json ja existe."
  echo "[spw-kit] Mescle manualmente o bloco de hook de ${SCRIPT_DIR}/.claude/settings.json.example"
fi

chmod +x "${TARGET_ROOT}/.claude/hooks/session-start-sync-tasks-template.sh" || true

echo "[spw-kit] Instalacao concluida."
echo "[spw-kit] Proximo passo: ajustar .spec-workflow/spw-config.toml"
