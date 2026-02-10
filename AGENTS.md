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
- `workflows/spw/*.md` <-> `copy-ready/.claude/workflows/spw/*.md`
- `workflows/spw/overlays/teams/*.md` <-> `copy-ready/.claude/workflows/spw/overlays/teams/*.md`
- `workflows/spw/overlays/noop.md` <-> `copy-ready/.claude/workflows/spw/overlays/noop.md`
- `workflows/spw/overlays/active/*.md` <-> `copy-ready/.claude/workflows/spw/overlays/active/*.md` (symlinks)
- `templates/user-templates/**` <-> `copy-ready/.spec-workflow/user-templates/**`
- `config/spw-config.toml` <-> `copy-ready/.spec-workflow/spw-config.toml`
- `hooks/*.js|*.sh` <-> `copy-ready/.claude/hooks/*`
- `hooks/claude-hooks.snippet.json` alinhado com `copy-ready/.claude/settings.json.example`

## Regras operacionais obrigatórias

1. Respeitar paths canônicos SPW: usar `.spec-workflow/specs/<spec-name>/` (nunca `.specs/`).
2. Runtime config canônico: `.spec-workflow/spw-config.toml` (com fallback legado para `.spw/spw-config.toml`).
3. Manter localidade de artefatos: pesquisa/planejamento ficam dentro da spec ativa; apoio em `.spec-workflow/specs/<spec-name>/research/`.
4. Aprovação é MCP-only: checar status via MCP; não substituir por aprovação manual em chat.
5. Preservar contrato dos comandos (`spw:prd`, `spw:plan`, `spw:tasks-plan`, `spw:exec`, `spw:checkpoint`, `spw:status`, `spw:post-mortem`, `spw:qa`, `spw:qa-check`, `spw:qa-exec`) e atualizar docs se comportamento mudar.
6. Padrão thin-orchestrator obrigatório: `commands/` são wrappers finos (máx. 60 linhas) e a lógica detalhada fica em `workflows/`.
7. Em `spw:tasks-plan`, manter semântica + precedência:
   - `--mode initial`: gera apenas wave executável inicial
   - `--mode next-wave`: adiciona apenas próxima wave executável
   - sem `--mode`, usar `[planning].tasks_generation_strategy`:
     - `rolling-wave`: gera uma wave executável por ciclo
     - `all-at-once`: gera todas as waves executáveis em uma execução
   - `--max-wave-size` sobrescreve `[planning].max_wave_size`; sem argumento, usar config
8. Em `spw:exec`, execução é via subagentes por tarefa (inclusive waves sequenciais de 1 tarefa); orquestrador não implementa código direto.
9. Se `execution.require_user_approval_between_waves=true`, não avançar wave sem autorização explícita do usuário.
10. Se `execution.commit_per_task="auto"` ou `"manual"`, exigir commit atômico por tarefa; se `"manual"`, parar com comandos git explícitos; se `"none"`, pular enforcement de commit por tarefa. Respeitar gate de worktree limpo quando habilitado.
11. `spw update` deve atualizar primeiro o próprio binário (`spw`) e, em seguida, limpar cache local do kit antes de atualizar, para evitar templates/comandos stale.
12. Em comandos longos com subagentes (`spw:prd`, `spw:design-research`, `spw:tasks-plan`, `spw:tasks-check`, `spw:checkpoint`, `spw:post-mortem`, `spw:qa`, `spw:qa-check`, `spw:qa-exec`), se existir run incompleto, é obrigatório AskUserQuestion (`continue-unfinished` ou `delete-and-restart`); o agente não pode escolher reiniciar sozinho.
13. Compatibilidade com dashboard (`spec-workflow-mcp`) em `tasks.md` é obrigatória:
   - checkbox apenas em linhas de tarefa (`- [ ]`, `- [-]`, `- [x]` com ID numérico)
   - IDs de tarefa devem ser únicos no arquivo (sem duplicatas)
   - task row usa `-` (nunca `*` para linha de tarefa)
   - nunca usar checkbox aninhado em DoD/metadados
   - metadados devem ser bullets normais (`- ...`), nunca checkbox
   - `Files` deve ser parseável em uma linha (`- Files: a, b`)
   - usar metadados com underscore: `_Requirements: ..._`, `_Leverage: ..._` (quando houver), `_Prompt: ..._` (fechando com `_`)
   - `_Prompt` deve incluir `Role|Task|Restrictions|Success`
14. Em `design.md`, incluir ao menos um diagrama Mermaid válido em `## Architecture` (fluxo principal), preferindo a skill `mermaid-architecture` para padronização.
   - usar bloco fenced com marcador de linguagem `mermaid` em minúsculo
15. UX do CLI: `spw` deve mostrar help por padrão; instalação explícita via `spw install`.
16. Em gates de aprovação (`spw:prd`, `spw:status`, `spw:plan`, `spw:design-draft`, `spw:tasks-plan`), quando `spec-status` vier incompleto/ambíguo, reconciliar via MCP `approvals status` (resolvendo `approvalId` por `spec-status` e, se necessário, por `.spec-workflow/approvals/<spec-name>/`); nunca decidir por `overallStatus`/fases apenas e nunca usar `STATUS-SUMMARY.md` como fonte de verdade.
17. Em `spw:post-mortem`, salvar relatórios em `.spec-workflow/post-mortems/<spec-name>/` com front matter YAML (`spec`, `topic`, `tags`, `range_from`, `range_to`) e atualizar `.spec-workflow/post-mortems/INDEX.md`.
18. Com `[post_mortem_memory].enabled=true`, comandos de design/planning (`spw:prd`, `spw:design-research`, `spw:design-draft`, `spw:tasks-plan`, `spw:tasks-check`) devem consultar o índice de post-mortems e aplicar no máximo `[post_mortem_memory].max_entries_for_design` entradas relevantes.
19. Catálogo padrão de skills: não incluir `requesting-code-review`; manter alinhamento entre `copy-ready/install.sh`, `config/spw-config.toml` e `copy-ready/.spec-workflow/spw-config.toml`.
20. `test-driven-development` pertence ao catálogo comum; em `spw:exec`/`spw:checkpoint`, só vira obrigatório quando `[execution].tdd_default=true`.
21. Em `spw:exec` (normal e teams), antes de leitura ampla o orquestrador deve despachar `execution-state-scout` (modelo implementation/sonnet por padrão) para consolidar checkpoint, tarefa `[-]` em progresso, próxima(s) executável(eis) e ação de retomada; o principal deve consumir apenas o resumo compacto e então ler contexto por tarefa.
22. Em `spw:qa`, quando o foco não for informado, perguntar explicitamente ao usuário o alvo de validação e escolher `playwright|bruno|hybrid` com justificativa de risco/escopo. O plano deve incluir seletores/endpoints concretos por cenário (CSS, `data-testid`, rotas, métodos HTTP).
23. Em validações com Playwright no `spw:qa`/`spw:qa-exec`, utilizar tools do servidor Playwright MCP pré-configurado; nunca invocar npx ou scripts Node diretamente para automação de browser.
24. Cobertura de Agent Teams para comandos subagent-first usa symlinks em `workflows/spw/overlays/active/` (apontando para `../noop.md` quando desabilitado ou `../teams/<cmd>.md` quando habilitado); por padrão todas as fases são elegíveis (`[agent_teams].exclude_phases = []`); fases podem ser excluídas adicionando-as a `exclude_phases`.
25. Em `spw:qa-check`, validar seletores/endpoints do plano contra código fonte real (único comando QA que lê arquivos de implementação); produzir mapa verificado em `QA-CHECK.md`.
26. Em `spw:qa-exec`, nunca ler arquivos fonte de implementação; usar apenas seletores verificados de `QA-CHECK.md`. Se seletor falhar em runtime, registrar como defeito "selector drift" e recomendar `spw:qa-check`.

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
- `bash -n scripts/validate-thin-orchestrator.sh`
- `scripts/validate-thin-orchestrator.sh`
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
