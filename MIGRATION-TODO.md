# Thin-Orchestrator Migration TODO

Status: completed

## Big-Bang Checklist

- [x] Fase 1: Fundacao de arquitetura
- [x] Fase 2: Migrar comandos base (`commands/spw/*.md`) para wrappers finos
- [x] Fase 3: Unificar `spw-teams` com base + overlay
- [x] Fase 4: Sincronizacao de espelhos + instalador
- [x] Fase 5: Documentacao e governanca
- [x] Fase 6: Hardening e release

## Nova etapa solicitada (config TOML)

- [x] Fase 7: Mover configuracao runtime de `.spec-workflow/spw-config.toml` para `.spw/spw-config.toml`
- [x] Atualizar `copy-ready/install.sh` para instalar o TOML em `.spw/spw-config.toml`
- [x] Atualizar comandos/workflows para ler `.spw/spw-config.toml`
- [x] Definir compatibilidade temporaria: fallback de leitura para `.spec-workflow/spw-config.toml` durante a transicao
- [x] Atualizar docs canonicos (`README.md`, `AGENTS.md`, `docs/SPW-WORKFLOW.md`, `hooks/README.md`, `copy-ready/README.md`)
- [x] Adicionar validacao automatica para garantir que o path novo seja o canonico

## Validacao final

- [x] Rodar checklist minimo de validacao do AGENTS.md
- [x] Verificar espelho completo (`commands`, `workflows`, `templates`, `hooks`, `config`)
- [x] Confirmar que nao houve regressao de contratos (`spw:tasks-plan`, `spw:exec`, `spw:checkpoint`, `spw:plan`, `spw:design-draft`, `spw:status`, `spw:qa`)
