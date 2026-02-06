# AGENTS.md

## Objetivo do projeto

SPW é um kit de comandos/templates para `spec-workflow-mcp`, com execução subagent-first, gates explícitos de aprovação e checkpoints por wave.

## Fontes canônicas (ordem de leitura)

1. `README.md` (fonte principal para instalação/uso/workflow)
2. `AGENTS.md` (regras operacionais para agentes e contribuição)
3. `config/spw-config.toml` (defaults operacionais)

Observação:
- `docs/SPW-WORKFLOW.md`, `hooks/README.md` e `copy-ready/README.md` devem permanecer enxutos e apontar para `README.md`.

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
4. Preservar contrato dos comandos (`spw:prd`, `spw:plan`, `spw:tasks-plan`, `spw:exec`, `spw:checkpoint`, `spw:status`, `spw:post-mortem`) e atualizar docs se comportamento mudar.
5. Em `spw:tasks-plan`, manter semântica + precedência:
   - `--mode initial`: gera apenas wave executável inicial
   - `--mode next-wave`: adiciona apenas próxima wave executável
   - sem `--mode`, usar `[planning].tasks_generation_strategy`:
     - `rolling-wave`: gera uma wave executável por ciclo
     - `all-at-once`: gera todas as waves executáveis em uma execução
   - `--max-wave-size` sobrescreve `[planning].max_wave_size`; sem argumento, usar config
6. Em `spw:exec`, execução é via subagentes por tarefa (inclusive waves sequenciais de 1 tarefa); orquestrador não implementa código direto.
7. Se `execution.require_user_approval_between_waves=true`, não avançar wave sem autorização explícita do usuário.
8. Se `execution.commit_per_task=true`, exigir commit atômico por tarefa; respeitar gate de worktree limpo quando habilitado.
9. `spw update` deve atualizar primeiro o próprio binário (`spw`) e, em seguida, limpar cache local do kit antes de atualizar, para evitar templates/comandos stale.
10. Em comandos longos com subagentes (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`, `spw:post-mortem`), se existir run incompleto, é obrigatório AskUserQuestion (`continue-unfinished` ou `delete-and-restart`); o agente não pode escolher reiniciar sozinho.
11. Compatibilidade com dashboard (`spec-workflow-mcp`) em `tasks.md` é obrigatória:
   - checkbox apenas em linhas de tarefa (`- [ ]`, `- [-]`, `- [x]` com ID numérico)
   - IDs de tarefa devem ser únicos no arquivo (sem duplicatas)
   - task row usa `-` (nunca `*` para linha de tarefa)
   - nunca usar checkbox aninhado em DoD/metadados
   - metadados devem ser bullets normais (`- ...`), nunca checkbox
   - `Files` deve ser parseável em uma linha (`- Files: a, b`)
   - usar metadados com underscore: `_Requirements: ..._`, `_Leverage: ..._` (quando houver), `_Prompt: ..._` (fechando com `_`)
   - `_Prompt` deve incluir `Role|Task|Restrictions|Success`
12. Em `design.md`, incluir ao menos um diagrama Mermaid válido em `## Architecture` (fluxo principal), preferindo a skill `mermaid-architecture` para padronização.
   - usar bloco fenced com marcador de linguagem `mermaid` em minúsculo
13. UX do CLI: `spw` deve mostrar help por padrão; instalação explícita via `spw install`.
14. Em `spw:post-mortem`, salvar relatórios em `.spec-workflow/post-mortems/<spec-name>/` com front matter YAML (`spec`, `topic`, `tags`, `range_from`, `range_to`) e atualizar `.spec-workflow/post-mortems/INDEX.md`.
15. Com `[post_mortem_memory].enabled=true`, comandos de design/planning (`spw:prd`, `spw:design-research`, `spw:design-draft`, `spw:tasks-plan`, `spw:tasks-check`) devem consultar o índice de post-mortems e aplicar no máximo `[post_mortem_memory].max_entries_for_design` entradas relevantes.

## File-first comms (não quebrar)

Para comandos que exigem handoff por arquivos, garantir presença de:

- `<subagent>/brief.md`
- `<subagent>/report.md`
- `<subagent>/status.json`
- `<run-dir>/_handoff.md`

Ausência desses arquivos deve resultar em `BLOCKED`.

## Checklist mínimo de validação

- `bash -n bin/spw`
- `bash -n scripts/bootstrap.sh`
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
- `AGENTS.md`
- `docs/SPW-WORKFLOW.md` (ponteiro)
- `hooks/README.md` (ponteiro)
- `copy-ready/README.md` (ponteiro)
