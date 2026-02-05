# AGENTS.md

## Objetivo do projeto

SPW é um kit de comandos/templates para `spec-workflow-mcp`, com execução subagent-first, gates explícitos de aprovação e checkpoints por wave.

## Fontes canônicas (ordem de leitura)

1. `README.md` (entrada rápida e comandos principais)
2. `docs/SPW-WORKFLOW.md` (contrato completo de lifecycle)
3. `hooks/README.md` (hooks + guardrails `warn|block`)
4. `config/spw-config.toml` (defaults operacionais)
5. `copy-ready/README.md` (distribuição para projetos-alvo)

## Mapa de arquivos que devem ficar em espelho

- `commands/spw/*.md` <-> `copy-ready/.claude/commands/spw/*.md`
- `commands/spw-teams/*.md` <-> `copy-ready/.claude/commands/spw-teams/*.md`
- `templates/user-templates/**` <-> `copy-ready/.spec-workflow/user-templates/**`
- `config/spw-config.toml` <-> `copy-ready/.spec-workflow/spw-config.toml`
- `hooks/*.js|*.sh` <-> `copy-ready/.claude/hooks/*`
- `hooks/claude-hooks.snippet.json` alinhado com `copy-ready/.claude/settings.json.example`

## Regras operacionais obrigatórias

1. Respeitar paths canônicos SPW: usar `.spec-workflow/specs/<spec-name>/` (nunca `.specs/`).
2. Manter localidade de artefatos: pesquisa/planejamento ficam dentro da spec ativa; apoio em `.spec-workflow/specs/<spec-name>/research/`.
3. Aprovação é MCP-only: checar status via MCP; não substituir por aprovação manual em chat.
4. Preservar contrato dos comandos (`spw:prd`, `spw:plan`, `spw:tasks-plan`, `spw:exec`, `spw:checkpoint`, `spw:status`) e atualizar docs se comportamento mudar.
5. Em `spw:tasks-plan`, manter semântica de modo:
   - `initial`: gera apenas wave executável inicial
   - `next-wave`: adiciona apenas próxima wave executável
6. Em `spw:exec`, execução é via subagentes por tarefa (inclusive waves sequenciais de 1 tarefa); orquestrador não implementa código direto.
7. Se `execution.require_user_approval_between_waves=true`, não avançar wave sem autorização explícita do usuário.
8. Se `execution.commit_per_task=true`, exigir commit atômico por tarefa; respeitar gate de worktree limpo quando habilitado.

## File-first comms (não quebrar)

Para comandos que exigem handoff por arquivos, garantir presença de:

- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

Ausência desses arquivos deve resultar em `BLOCKED`.

## Checklist mínimo de validação

- `bash -n bin/spw`
- `bash -n scripts/install-spw-bin.sh`
- `bash -n hooks/session-start-sync-tasks-template.sh`
- `bash -n copy-ready/install.sh`
- `node hooks/spw-statusline.js <<< '{"workspace":{"current_dir":"'"$(pwd)"'"}}'`
- `node hooks/spw-guard-user-prompt.js <<< '{"prompt":"/spw:plan"}'`
- `node hooks/spw-guard-paths.js <<< '{"cwd":"'"$(pwd)"'","tool_input":{"file_path":"docs/DESIGN-RESEARCH.md"}}'`
- `node hooks/spw-guard-stop.js <<< '{}'`

## Sincronização de documentação

Se mudar comportamento, defaults ou guardrails, atualizar no mesmo patch:

- `README.md`
- `docs/SPW-WORKFLOW.md`
- `hooks/README.md`
- `copy-ready/README.md`
