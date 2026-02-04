# SPW Copy-Ready Kit

Pacote pronto para copiar em qualquer projeto com `spec-workflow-mcp`.

## O que vem no kit

- `.claude/commands/spw/*.md` (comandos de planning/exec/checkpoint)
- `.claude/hooks/session-start-sync-tasks-template.sh` (hook de sync)
- `.claude/settings.json.example` (snippet de configuracao de hook)
- `.spec-workflow/spw-config.toml` (config central com comentarios extensivos)
- `.spec-workflow/user-templates/*.md` (templates custom)
- `.spec-workflow/user-templates/prd-template.md` (template PRD)
- `.spec-workflow/user-templates/variants/tasks-template.tdd-*.md` (variantes ON/OFF)

## Como instalar no projeto alvo

Na raiz do projeto alvo:

```bash
cp -R /CAMINHO/PARA/spw/copy-ready/. .
```

Depois:

1. Mescle `.claude/settings.json.example` no seu `.claude/settings.json`.
2. Ajuste `.spec-workflow/spw-config.toml` (principalmente `execution.tdd_default`).
3. Inicie uma nova sessao para o hook sincronizar o `tasks-template.md`.

## Compatibilidade com spec-workflow

Este kit usa apenas:
- `.spec-workflow/user-templates/` para sobrescrever templates custom
- `.spec-workflow/spw-config.toml` para configuracao do fluxo

Nao altera templates default de `.spec-workflow/templates/`.

## Comandos disponiveis

- `/spw:prd` (do zero: gera requirements em PRD)
- `/spw:plan` (com requirements existente: gera design/tasks; valida aprovacao via MCP)
- `/spw:design-research`
- `/spw:design-draft`
- `/spw:tasks-plan`
- `/spw:tasks-check`
- `/spw:exec`
- `/spw:checkpoint`
