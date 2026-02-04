# SPW Hook: SessionStart Template Sync

Este hook sincroniza automaticamente o template ativo de tasks com base em `.spec-workflow/spw-config.toml`.

## Arquivos

- Script: `spw/hooks/session-start-sync-tasks-template.sh`
- Config: `.spec-workflow/spw-config.toml`
- Variantes esperadas:
  - `.spec-workflow/user-templates/variants/tasks-template.tdd-on.md`
  - `.spec-workflow/user-templates/variants/tasks-template.tdd-off.md`
- Alvo ativo:
  - `.spec-workflow/user-templates/tasks-template.md`

## Instalação no projeto

1. Copie o TOML de exemplo:

```bash
mkdir -p .spec-workflow
cp spw/config/spw-config.toml .spec-workflow/spw-config.toml
```

2. Copie as variantes de template:

```bash
mkdir -p .spec-workflow/user-templates/variants
cp spw/templates/user-templates/variants/tasks-template.tdd-on.md .spec-workflow/user-templates/variants/
cp spw/templates/user-templates/variants/tasks-template.tdd-off.md .spec-workflow/user-templates/variants/
```

3. Registre o hook de SessionStart no seu `.claude/settings.json` (ou configuração equivalente), apontando para:

```text
<repo>/spw/hooks/session-start-sync-tasks-template.sh
```

Exemplo de estrutura JSON no arquivo `spw/hooks/claude-hooks.snippet.json`.

## Teste manual rápido

```bash
./spw/hooks/session-start-sync-tasks-template.sh
```

Saída esperada:
- informa se sincronizou template
- informa se já estava sincronizado
- informa se faltou config/template (sem quebrar sessão, por padrão)
