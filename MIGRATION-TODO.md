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

- [x] Fase 7: Consolidar configuracao runtime em `.spec-workflow/spw-config.toml`
- [x] Atualizar `copy-ready/install.sh` para instalar o TOML apenas em `.spec-workflow/spw-config.toml`
- [x] Atualizar comandos/workflows/hooks para ler `.spec-workflow/spw-config.toml` como caminho canônico
- [x] Manter compatibilidade legado com fallback de leitura para `.spw/spw-config.toml`
- [x] Atualizar docs canonicos (`README.md`, `AGENTS.md`, `docs/SPW-WORKFLOW.md`, `hooks/README.md`, `copy-ready/README.md`)
- [x] Remover duplicacao de config no kit (`copy-ready/.spw/spw-config.toml`)

## Validacao final

- [x] Rodar checklist minimo de validacao do AGENTS.md
- [x] Verificar espelho completo (`commands`, `workflows`, `templates`, `hooks`, `config`)
- [x] Confirmar que nao houve regressao de contratos (`spw:tasks-plan`, `spw:exec`, `spw:checkpoint`, `spw:plan`, `spw:design-draft`, `spw:status`, `spw:qa`)
