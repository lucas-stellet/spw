# Oráculo

![Version](https://img.shields.io/badge/version-2.0-blue)
![License: MIT](https://img.shields.io/badge/license-MIT-green)

> *"Conhece-te a ti mesmo."* — Inscrição no Templo de Apolo em Delfos

## Índice

- [O que é o Oráculo?](#o-que-é-o-oráculo)
- [Por que Oráculo?](#por-que-oráculo)
- [Início Rápido](#início-rápido)
- [Onde começar](#onde-começar)
- [Instalação](#instalação)
- [Armazenamento Local](#armazenamento-local)
- [Comandos de Entrada](#comandos-de-entrada)
- [Arquitetura do Orquestrador Enxuto](#arquitetura-do-orquestrador-enxuto)
- [Compatibilidade com Dashboard](#compatibilidade-com-dashboard-spec-workflow-mcp)
- [Mermaid para Design de Arquitetura](#mermaid-para-design-de-arquitetura)
- [Validação QA (3 Fases)](#validação-qa-3-fases)
- [Glossário](#glossário)

## O que é o Oráculo?

Nos templos da Grécia Antiga, o Oráculo de Delfos era consultado antes de qualquer empreitada importante — guerras, fundações de cidades, decisões políticas. Ninguém agia sem antes buscar a sabedoria das Pítias, que canalizavam o conhecimento de Apolo em profecias estruturadas e interpretáveis.

**Oráculo** traz esse mesmo princípio para o desenvolvimento de software com Claude Code. Assim como os antigos não partiam para a batalha sem consultar Delfos, Oráculo impede que um agente de IA ataque uma feature inteira de uma vez. Em vez disso, o trabalho é decomposto em fases ritualizadas — cada uma com seus portões de qualidade e pontos de aprovação — como os estágios de uma consulta oracular:

| Fase | Analogia | O que acontece |
|------|----------|----------------|
| **Profecia** (PRD) | A pergunta ao Oráculo | Requisitos são extraídos e estruturados |
| **Interpretação** (Design) | A Pítia traduz a visão | Pesquisa, arquitetura e decisões técnicas |
| **Tábuas** (Planning) | As tábuas de pedra com a resposta | Tarefas executáveis são geradas em ondas |
| **Execução** (Exec) | Os generais implementam a profecia | Implementação com checkpoints automáticos |
| **Julgamento** (QA) | O tribunal valida o cumprimento | Testes planejados, verificados e executados |

Cada fase despacha **agentes especializados** com roteamento de modelo: Haiku faz o reconhecimento leve (como batedores), Opus conduz o raciocínio complexo (como os sábios do templo), e Sonnet executa a implementação (como os artesãos). Os agentes se comunicam por artefatos no sistema de arquivos — não por chat — tornando cada handoff reproduzível e auditável.

Você orquestra tudo por slash commands no Claude Code (ex: `/oraculo:prd`, `/oraculo:exec`) enquanto o `spec-workflow-mcp` serve como fonte de verdade para artefatos e aprovações.

## Por que Oráculo?

O Oráculo de Delfos não era simplesmente um adivinho. Era um **sistema**:

- **Estrutura ritualística** — Perguntas tinham formato. Respostas seguiam protocolo. Nada era improvisado. Oráculo impõe a mesma disciplina: cada fase tem seu formato, seus artefatos, seus portões.

- **Intermediários especializados** — A Pítia profetizava, os sacerdotes interpretavam, os escribas registravam. Oráculo replica isso com subagentes: cada um tem seu papel, seu modelo, seu escopo.

- **Sabedoria acumulada** — O templo mantinha registros de consultas passadas. Os post-mortems do Oráculo indexam lições aprendidas que alimentam decisões futuras.

- **Nunca agir sem consultar** — O maior pecado na Grécia Antiga era a *hybris*: agir com arrogância, sem pedir orientação. Oráculo garante que nenhum agente implemente código sem antes passar pelos portões de planejamento.

## Início Rápido

Após a [instalação](#instalação), execute dentro de uma sessão Claude Code:

1. `/oraculo:prd minha-feature` — Gera o documento de requisitos a partir da descrição
2. `/oraculo:plan minha-feature` — Cria design e decompõe em tarefas executáveis
3. `/oraculo:exec minha-feature` — Implementa em ondas com checkpoints automáticos
4. `/oraculo:qa minha-feature` — Constrói e executa plano de validação QA

Cada comando cuida do despacho de subagentes, handoff de arquivos e portões de qualidade. Entre passos, artefatos ficam em `.spec-workflow/specs/minha-feature/` e aprovações fluem pelo `spec-workflow-mcp`.

## Onde começar

- Este arquivo é a fonte principal de uso e operação.
- Regras operacionais para agentes e contribuidores estão em `AGENTS.md`.
- `docs/ORACULO-WORKFLOW.md`, `hooks/README.md` e `copy-ready/README.md` são ponteiros leves para este README.

## Instalação

### 1. Instalar a CLI

O script de bootstrap baixa o binário Go compilado do último Release no GitHub e instala em `~/.local/bin/oraculo`. Requer `curl` e `tar`.

```bash
curl -fsSL https://raw.githubusercontent.com/lucas-stellet/oraculo/main/scripts/bootstrap.sh | bash
```

**A partir de um clone local (build do source):**

```bash
cd cli && go build -o ~/.local/bin/oraculo ./cmd/oraculo/
```

Execute `oraculo` sem argumentos para ver os comandos disponíveis. Use `oraculo update` para auto-atualizar.

### 2. Instalar no projeto

Na raiz do projeto:

```bash
oraculo install
```

Copia comandos, workflows, hooks, config e skills para o projeto. Para instalação manual: `cp -R /path/to/oraculo/copy-ready/. .`

### 3. Checklist pós-instalação

Obrigatório:
1. Merge de `.claude/settings.json.example` no seu `.claude/settings.json` (se necessário).
2. Revise `.spec-workflow/oraculo.toml`, especialmente `[planning].tasks_generation_strategy` e `[planning].max_wave_size`.
3. Inicie uma nova sessão Claude Code para o hook SessionStart sincronizar o template de tasks.

Opcional:
- Habilite enforcement de skills por fase: `skills.design.enforce_required` e `skills.implementation.enforce_required` em `oraculo.toml`.
- Habilite a statusline do Oráculo (veja `.claude/settings.json.example`).
- Habilite hooks de enforcement com `hooks.enforcement_mode = "warn"` ou `"block"` em `oraculo.toml`.
- Para validação QA com browser e exploração de URLs no planejamento, adicione Playwright MCP:
  ```
  claude mcp add playwright -- npx @playwright/mcp@latest --headless --isolated
  ```

Skills padrão são copiados para `.claude/skills/` durante a instalação. O skill `test-driven-development` está no catálogo padrão; `qa-validation-planning` está disponível para fases QA. Em fases de implementação (`oraculo:exec`, `oraculo:checkpoint`), TDD é obrigatório apenas quando `[execution].tdd_default = true`.

Variantes de template TDD: `user-templates/variants/` contém `tasks-template.tdd-on.md` e `tasks-template.tdd-off.md`. A chave `[templates].tasks_template_mode` controla a seleção (`auto`, `on`, `off`). O hook SessionStart sincroniza a variante ativa ao início de cada sessão.

> **Path legado:** o Oráculo também verifica `.spw/spw-config.toml` como fallback se `.spec-workflow/oraculo.toml` não for encontrado.

### Instalação Global

Para quem trabalha em múltiplos projetos, o Oráculo suporta instalação em duas camadas que evita duplicar ~72 arquivos por projeto:

| Modo | Comando | O que instala | Onde |
|------|---------|---------------|------|
| **Global** | `oraculo install --global` | Comandos, workflows, hooks, skills | `~/.claude/` |
| **Init de Projeto** | `oraculo init` | Config, templates, snippets, .gitattributes | `.spec-workflow/`, `CLAUDE.md`, `AGENTS.md` |
| **Completo (padrão)** | `oraculo install` | Tudo (comportamento inalterado) | `.claude/` + `.spec-workflow/` |

**Setup:**

```bash
# Uma vez: instalar globalmente
oraculo install --global

# Por projeto: inicializar config específica
cd meu-projeto
oraculo init
```

**Como funciona:** Claude Code resolve paths `@.claude/` com prioridade local, fallback global. Se o projeto tem install local (`oraculo install`), ele prevalece sobre o global.

**Limitações:**
- Workflows globais usam config padrão (sem guidelines do projeto). Projetos que precisam de guidelines customizadas devem usar `oraculo install`.
- Overlays de Agent Teams ficam como noop globalmente. Projetos usando Agent Teams precisam de `oraculo install` para ativação local.

### Comandos da CLI

| Comando | Descrição |
|---------|-----------|
| `oraculo install` | Instala Oráculo no projeto atual (instalação local completa) |
| `oraculo install --global` | Instala comandos, workflows, hooks e skills em `~/.claude/` |
| `oraculo init` | Inicializa config, templates e snippets específicos do projeto |
| `oraculo update` | Auto-atualiza o binário via GitHub Releases |
| `oraculo doctor` | Verifica saúde da instalação (versão, config, hooks, comandos, workflows, skills) |
| `oraculo status` | Resumo rápido do kit e skills |
| `oraculo skills` | Status de skills instalados/disponíveis/ausentes |
| `oraculo skills install [--elixir]` | Instala skills gerais (ou Elixir com a flag) |

<details>
<summary>Referência rápida de config (todas as seções)</summary>

| Seção | Chave(s) | Descrição |
|-------|----------|-----------|
| `[statusline]` | `cache_ttl_seconds`, `base_branches`, `sticky_spec`, `show_token_cost` | Comportamento do hook StatusLine |
| `[templates]` | `sync_tasks_template_on_session_start`, `tasks_template_mode` | Seleção de variante do template de tasks |
| `[safety]` | `backup_before_overwrite` | Backup antes de sobrescrever arquivos spec |
| `[verification]` | `inline_audit_max_iterations` | Máx. retentativas de audit inline |
| `[qa]` | `max_scenarios_per_wave` | Dimensionamento de waves QA |
| `[hooks]` | `verbose`, `recent_run_window_minutes`, `guard_prompt_require_spec`, `guard_paths`, `guard_wave_layout`, `guard_stop_handoff` | Toggles por guard de hook |
| `[execution]` | `require_clean_worktree_for_wave_pass`, `manual_tasks_require_human_handoff`, `tdd_default` | Portões de execução |
| `[planning]` | `tasks_generation_strategy`, `max_wave_size` | Estratégia de planejamento em ondas |
| `[post_mortem_memory]` | `enabled`, `max_entries_for_design` | Indexação de lições post-mortem |
| `[agent_teams]` | `enabled`, `exclude_phases`, `require_delegate_mode` | Toggle de Agent Teams |

Veja `.spec-workflow/oraculo.toml` para documentação completa de cada chave.

</details>

| `oraculo finalizar <spec>` | Marca spec como completo, gera sumário com frontmatter YAML |
| `oraculo view <spec> [type]` | Visualiza artefatos no terminal ou VS Code |
| `oraculo search <query>` | Busca full-text (FTS5) em specs indexadas |
| `oraculo summary <spec>` | Gera resumo de progresso sob demanda |

#### Ferramentas de workflow (usadas por subagentes)

| Comando | Descrição |
|---------|-----------|
| `oraculo tools verify-task <spec> --task-id N [--check-commit]` | Verifica existência de artefatos da task |
| `oraculo tools impl-log register <spec> --task-id N --wave NN --title T --files F --changes C` | Cria log de implementação para task concluída |
| `oraculo tools impl-log check <spec> --task-ids 1,2,3` | Verifica se logs de implementação existem |
| `oraculo tools task-mark <spec> --task-id N --status done` | Atualiza checkbox da task em tasks.md |
| `oraculo tools wave-status <spec>` | Resolução completa de estado da wave |
| `oraculo tools wave-update <spec> --wave NN --status pass --tasks 3,4,7` | Escreve resumo da wave e JSON de estado |
| `oraculo tools dispatch-init-audit --run-dir R --type T` | Cria diretório de audit dentro de um run |
| `oraculo tools audit-iteration start --run-dir R --type T [--max N]` | Inicializa tracking de iteração de audit |
| `oraculo tools audit-iteration check --run-dir R --type T` | Verifica se outra retentativa é permitida |
| `oraculo tools audit-iteration advance --run-dir R --type T --result R` | Avança contador de iteração |

### Agent Teams (opcional)

Agent Teams é desabilitado por padrão. Para habilitar, defina `[agent_teams].enabled = true` em `oraculo.toml`. O instalador lê essa configuração e alterna symlinks em `.claude/workflows/oraculo/overlays/active/`.

Configuração adicional (automática pelo instalador):
- `env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1"` em `.claude/settings.json`
- `teammateMode = "in-process"` (altere manualmente para `"tmux"` se desejar)
- Symlinks de overlay: `cd .claude/workflows/oraculo/overlays/active && ln -sf ../teams/<cmd>.md <cmd>.md`

Quando habilitado, Oráculo cria um time para qualquer fase não listada em `[agent_teams].exclude_phases`. `oraculo:exec` exige delegate mode quando `[agent_teams].require_delegate_mode = true`.

## Armazenamento Local

Oráculo armazena dados estruturados em bancos SQLite (driver Go puro, sem CGO, modo WAL):

- **`spec.db`** — Banco por spec em `.spec-workflow/specs/<spec-name>/spec.db`. O sistema de dispatch colhe automaticamente artefatos de subagentes (briefs, reports, status) no DB durante handoff (dual-write). Três arquivos gerenciados pelo MCP permanecem em disco como fonte de verdade: `requirements.md`, `design.md`, `tasks.md`.
- **`.oraculo-index.db`** — Índice global em `.spec-workflow/.oraculo-index.db` com FTS5 full-text search. Alimenta `oraculo search <query>`.

`oraculo finalizar <spec>` marca um spec como completo e gera um bloco de sumário YAML frontmatter, tornando o spec pesquisável no índice global.

## Comandos de Entrada

As fases seguem a jornada oracular:

- `oraculo:prd` → fluxo zero-to-PRD de requisitos *(a pergunta ao Oráculo)*
- `oraculo:plan` → meta-orquestrador de design/tasks: encadeia pesquisa → draft → plano → checagem *(a interpretação)*
- `oraculo:tasks-plan` → geração de tasks (`rolling-wave` ou `all-at-once`) *(as tábuas com a resposta)*
- `oraculo:exec` → execução em batch com checkpoints *(os generais implementam)*
- `oraculo:checkpoint` → portão de qualidade (PASS/BLOCKED) *(o templo valida)*
- `oraculo:status` → resumo de onde o workflow parou + próximos comandos *(consulta ao templo)*
- `oraculo:post-mortem` → analisa commits pós-spec e registra lições *(as crônicas)*
- `oraculo:qa` → constrói plano de validação QA *(o tribunal)*
- `oraculo:qa-check` → valida seletores e rastreabilidade *(verificação dos testemunhos)*
- `oraculo:qa-exec` → executa plano validado *(o veredito)*

## Arquitetura do Orquestrador Enxuto

Oráculo usa orquestradores enxutos com um sistema de padrões de despacho:
- wrappers de comando em `.claude/commands/oraculo/*.md`
- workflows detalhados em `.claude/workflows/oraculo/*.md`
- políticas de despacho compartilhadas em `.claude/workflows/oraculo/shared/dispatch-{pipeline,audit,wave}.md`
- políticas transversais em `.claude/workflows/oraculo/shared/*.md`

### Categorias de Despacho

Cada workflow declara uma seção `<dispatch_pattern>` como **fonte única de verdade** para metadados de despacho (`category`, `phase`, `comms_path`, `artifacts`). A CLI parseia essa seção dos workflows embutidos na inicialização.

| Categoria | Política | Comandos |
|-----------|----------|----------|
| **Pipeline** | `dispatch-pipeline.md` | `prd`, `design-research`, `design-draft`, `tasks-plan`, `qa`, `post-mortem` |
| **Audit** | `dispatch-audit.md` | `tasks-check`, `qa-check`, `checkpoint` |
| **Wave Execution** | `dispatch-wave.md` | `exec`, `qa-exec` |

Guardrails de checkpoint (comandos audit):
- Orquestradores são observadores read-only — NUNCA criam/modificam/deletam artefatos fora de comms para resolver um auditor BLOCKED (anti-self-heal).
- Se QUALQUER auditor retorna `blocked`, o veredito final DEVE ser BLOCKED (consistência de handoff).
- Briefs nunca afirmam fatos sobre o código — instruem auditores a verificar.
- `oraculo:exec` deve parar e instruir o usuário a rodar `oraculo:checkpoint` em sessão separada (isolamento de sessão).

As 5 regras core do thin-dispatch:
1. Orquestrador lê apenas `status.json` após dispatch (nunca `report.md` em caso de pass).
2. Briefs contêm paths de filesystem para reports anteriores (nunca conteúdo).
3. Sintetizadores/agregadores leem direto do disco.
4. Estrutura de run segue layout da categoria.
5. Resume pula subagentes completos, sempre reexecuta o estágio final.

Lógica específica por comando é injetada via `<extensions>` em pontos nomeados (`pre_pipeline`, `pre_dispatch`, `post_dispatch`, `post_pipeline`, `inter_wave`, `per_task`).

### Agent Teams

Agent Teams usa base + overlay via symlinks:
- workflow base: `.claude/workflows/oraculo/<command>.md`
- overlay ativo: `.claude/workflows/oraculo/overlays/active/<command>.md` (symlink)
- teams off: symlink → `../noop.md`
- teams on: symlink → `../teams/<command>.md`

Wrappers permanecem intencionalmente enxutos e delegam 100% da lógica aos workflows.

Guardrail de contexto de execução (`oraculo:exec`):
- Antes de leituras amplas, despacha `execution-state-scout` (modelo de implementação, default `sonnet`).
- Scout retorna apenas estado compacto: status do checkpoint, task `[-]` em progresso, próximas tasks executáveis, e ação necessária (`resume|wait-user-authorization|manual-handoff|done|blocked`).
- Orquestrador então lê apenas arquivos com escopo de task (evita `requirements.md`/`design.md` completos, exceto para blockers).

Defaults de planejamento em `.spec-workflow/oraculo.toml`:

```toml
[planning]
tasks_generation_strategy = "rolling-wave" # ou "all-at-once"
max_wave_size = 3
```

- `rolling-wave`: cada ciclo de planejamento cria uma wave executável.
  - Loop típico: `tabuas` → `exec` → `checkpoint` → `tabuas` (próxima wave)...
- `all-at-once`: um passo de planejamento cria todas as waves.
- Args explícitos da CLI sobrescrevem config (`--mode`, `--max-wave-size`).

Memória post-mortem em `.spec-workflow/oraculo.toml`:

```toml
[post_mortem_memory]
enabled = true
max_entries_for_design = 5
```

- `oraculo:pos-mortem` escreve reports em `.spec-workflow/post-mortems/<spec-name>/`.
- Índice compartilhado: `.spec-workflow/post-mortems/INDEX.md` (usado por design/planning quando habilitado).
- Fases de design/planning carregam lições indexadas com priorização por recência/tags.

Tratamento de runs inacabados para comandos longos:
- Antes de criar um novo run-id, inspeciona a pasta de runs da fase.
- Se existe run inacabado, pede decisão explícita do usuário:
  - `continue-unfinished`
  - `delete-and-restart`
- Nunca escolhe automaticamente.
- Se decisão indisponível, para com `WAITING_FOR_USER_DECISION`.

Reconciliação de aprovações para comandos com gate MCP:
- Primeiro lê estado de aprovação dos campos `spec-status`.
- Se status ausente/desconhecido/inconsistente, resolve ID de aprovação e confirma via MCP `approvals status`.
- `STATUS-SUMMARY.md` é output-only, nunca fonte de verdade.

Comunicação file-first entre subagentes é armazenada em diretórios `_comms/` organizados por fase:
- prd: `.spec-workflow/specs/<spec-name>/prd/_comms/run-NNN/`
- design: `.spec-workflow/specs/<spec-name>/design/_comms/{design-research,design-draft}/run-NNN/`
- planning: `.spec-workflow/specs/<spec-name>/planning/_comms/{tasks-plan,tasks-check}/run-NNN/`
- execution: `.spec-workflow/specs/<spec-name>/execution/waves/wave-NN/{execution,checkpoint}/run-NNN/`
- qa: `.spec-workflow/specs/<spec-name>/qa/_comms/{qa,qa-check}/run-NNN/`
- qa-exec: `.spec-workflow/specs/<spec-name>/qa/_comms/qa-exec/waves/wave-NN/run-NNN/`
- post-mortem: `.spec-workflow/specs/<spec-name>/post-mortem/_comms/run-NNN/`

Formato de `<run-id>`: `run-NNN` (sequencial com zero-padding, ex: `run-001`).

YAML frontmatter (metadados opcionais) é incluído nos templates de spec sob a chave `oraculo` para classificação de documentos por subagentes.

## Compatibilidade com Dashboard (`spec-workflow-mcp`)

Para manter `tasks.md` compatível com renderização + parsing + validação de aprovação do Dashboard:

- Checkbox markers apenas em linhas reais de task:
  - `- [ ] <id>. <description>`
  - `- [-] <id>. <description>`
  - `- [x] <id>. <description>`
- Use `-` como marcador (nunca `*`).
- Nunca use checkboxes aninhados em blocos de metadados.
- IDs numéricos no início (`1`, `1.1`, `2.3`, ...), únicos no arquivo inteiro.
- Metadados como bullets regulares (`- ...`), nunca checkbox.
- `Files` parseável em linha única:
  - `- Files: path/to/file.ext, test/path/to/file_test.ext`
- Campos de metadados com underscore:
  - `_Requirements: ..._`
  - `_Leverage: ..._`
  - `_Prompt: ..._`
- `_Prompt` estruturado como:
  - `Role: ... | Task: ... | Restrictions: ... | Success: ...`

## Mermaid para Design de Arquitetura

Oráculo inclui o skill `mermaid-architecture` para fases de design:
- arquivo do skill: `skills/mermaid-architecture/SKILL.md`
- config padrão: listado em `[skills.design].optional`

Exemplos de arquitetura cobertos:
- fronteiras de módulos/camadas (`flowchart`)
- visão de containers/sistema (`flowchart`)
- fluxo de request com paths de sucesso/erro (`sequenceDiagram`)
- pipeline event-driven (`flowchart`)
- ciclo de vida de workflow (`stateDiagram-v2`)

Em `oraculo:design-draft`, `design.md` deve incluir pelo menos um diagrama Mermaid válido na seção `## Architecture`.

## Validação QA (3 Fases)

O QA segue uma cadeia de planejar → verificar → executar, como um tribunal grego:

```
oraculo:julgamento (plano) → oraculo:julgamento-check (validar) → oraculo:julgamento-exec (executar)
```

### `oraculo:julgamento` (planejamento)
- Pergunta ao usuário o que validar quando foco não é explícito
- Seleciona `Playwright MCP`, `Bruno CLI`, ou `hybrid` por risco/escopo
- Produz `QA-TEST-PLAN.md` com seletores/endpoints concretos por cenário
- Usa ferramentas de automação browser do Playwright MCP pré-configurado

### `oraculo:julgamento-check` (validação)
- Valida plano de teste contra código real (a ÚNICA fase que lê arquivos de implementação)
- Verifica existência de seletores/endpoints via `qa-selector-verifier`
- Checa rastreabilidade e viabilidade de dados
- Produz `QA-CHECK.md` com mapa verificado (test-id → seletor → file:line)
- Decisão PASS/BLOCKED gateia `oraculo:julgamento-exec`

### `oraculo:julgamento-exec` (execução)
- Executa plano validado usando apenas seletores verificados do `QA-CHECK.md`
- **Nunca lê arquivos fonte de implementação** — drift de seletores é logado como defeito
- Suporta `--scope smoke|regression|full` e `--rerun-failed true|false`
- Produz `QA-EXECUTION-REPORT.md` e `QA-DEFECT-REPORT.md` com decisão GO/NO-GO

Enforcement de hooks:
- `warn` → apenas diagnósticos
- `block` → nega ações violadoras
- Detalhes: `AGENTS.md` + `.spec-workflow/oraculo.toml`

## Glossário

- **Agent Teams**: Modo opcional onde Oráculo instancia múltiplos agentes Claude Code para trabalhar em paralelo numa fase. Controlado por `[agent_teams].enabled` em `oraculo.toml`.

- **Checkpoint** *(Portão)*: Gate de qualidade executado após cada wave via `oraculo:checkpoint`. Produz report PASS/BLOCKED que determina se a próxima wave pode prosseguir.

- **Dispatch Pattern** *(Padrão de Despacho)*: Estratégia de orquestração de um comando. Uma de três categorias: Pipeline (estágios sequenciais com sintetizador), Audit (revisores paralelos com agregador), ou Wave Execution (ciclos iterativos com checkpoints). Declarado via `<dispatch_pattern>` em cada workflow.

- **File-First Communication** *(Comunicação por Artefato)*: Subagentes se comunicam exclusivamente via artefatos no filesystem (`brief.md`, `report.md`, `status.json`) — nunca por chat. Armazenados em diretórios `_comms/`.

- **Overlay**: Mecanismo baseado em symlinks que alterna comportamento entre modo solo (symlink para `noop.md`) e Agent Teams (symlink para `teams/<cmd>.md`).

- **Rolling Wave** *(Onda Progressiva)*: Estratégia onde tasks são geradas uma wave por vez, permitindo que waves futuras incorporem lições da execução anterior. Config: `[planning].tasks_generation_strategy = "rolling-wave"`.

- **Scout** *(Batedor)*: Subagente leve despachado antes de uma wave para coletar estado de execução sem ler specs completos. Retorna estado compacto de resume para o orquestrador.

- **Synthesizer** *(Sintetizador)*: Subagente final num Pipeline que lê todos os reports anteriores do disco e produz o artefato consolidado.

- **Thin Dispatch** *(Despacho Enxuto)*: Princípio arquitetural central: orquestradores leem apenas `status.json` após cada subagente, passam paths entre estágios, e delegam lógica detalhada aos workflows.

- **Wave** *(Onda)*: Batch de tasks executadas juntas em `oraculo:exec`. Cada wave é seguida por checkpoint. Tamanho controlado por `[planning].max_wave_size`.

- **spec.db**: Banco SQLite por spec que armazena artefatos colhidos de subagentes. Criado automaticamente via dual-write durante dispatch handoff.

- **Harvest** *(Colheita)*: Padrão onde dispatch-handoff coleta arquivos de subagentes (`brief.md`, `report.md`, `status.json`) para `spec.db` após cada subagente completar.

---

<p align="center">
  <em>"Μηδὲν ἄγαν"</em> — Nada em excesso.<br>
  <small>Segunda inscrição no Templo de Apolo em Delfos.</small>
</p>
