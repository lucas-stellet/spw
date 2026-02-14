# Conceito Geral: Migração para SQLite Local

## Motivação

O SPW armazena todos os artefatos de specs como arquivos (Markdown, JSON) em árvores de diretórios complexas sob `.spec-workflow/specs/<spec>/`. Isso causa três problemas fundamentais:

1. **Poluição no repositório** -- dezenas de arquivos transientes (`status.json`, `_handoff.md`, `_wave-summary.json`, `_latest.json`, `_iteration-state.json`) poluem o working tree e diffs do git. Mesmo com `.gitattributes` colapsando esses arquivos em PRs, eles ainda existem no disco e no index.

2. **Impossibilidade de busca semântica** -- não existe forma de buscar "qual spec implementou autenticação JWT?" ou "quais foram os problemas recorrentes nos checkpoints?". Os dados estão espalhados em centenas de arquivos JSON e Markdown sem indexação.

3. **Queries complexas requerem I/O excessivo** -- operações como `ScanWaves`, `ResolveCheckpoint` e `HandleGuardStop` precisam varrer diretórios recursivamente, ler múltiplos JSONs e resolver fallbacks. Uma query SQL simples substituiria dezenas de `os.ReadDir` + `os.ReadFile`.

## Visão

### Arquitetura Alvo

```
.spec-workflow/
├── .spw-index.db              <-- global: FTS5 + sqlite-vec cross-spec
├── spw-config.toml            <-- mantém (user-editable)
└── specs/<spec-name>/
    ├── requirements.md        <-- mantém (dashboard MCP)
    ├── design.md              <-- mantém (dashboard MCP)
    ├── tasks.md               <-- mantém (dashboard MCP)
    └── spec.db                <-- tudo mais aqui
```

**Dois bancos de dados:**

- **`spec.db`** (um por spec) -- contém todos os dados de runtime: runs, subagents, waves, tasks (cache), impl logs, approvals, handoffs. Portável, sem cross-spec locks durante execução.
- **`.spw-index.db`** (global) -- índice cross-spec com FTS5 para busca full-text e sqlite-vec (opt-in) para busca vetorial com embeddings.

### O que mantém como arquivo

| Arquivo | Motivo |
|---------|--------|
| `requirements.md`, `design.md`, `tasks.md` | Dashboards consumidos pelo MCP e humanos |
| `spw-config.toml` | Configuração user-editable em TOML |
| `STATUS-SUMMARY.md` | Output human-readable (não source of truth) |
| User templates | Editáveis pelo usuário |
| Workflows (`.claude/workflows/spw/*.md`) | Claude Code lê como slash commands |

## Decisões Técnicas

| Decisão | Escolha | Motivo |
|---------|---------|--------|
| SQLite driver | `modernc.org/sqlite` (pure Go) | Sem CGO, cross-compile trivial, FTS5 incluso |
| Vector search | FTS5 primário, `sqlite-vec` opt-in | FTS5 funciona sem deps extras; vec requer extensão nativa |
| Embeddings | Ollama (`nomic-embed-text`, 768 dims) | Prático para devs, fail-open se indisponível |
| Visualização | `charmbracelet/glamour` (terminal) + VS Code | Glamour é a lib por trás do `glow` |
| DB por spec | Sim, `spec.db` por spec | Portável, sem cross-spec locks |
| Índice global | `.spw-index.db` na raiz | Para busca cross-spec |
| YAML frontmatter | `gopkg.in/yaml.v3` | Frontmatter padronizado no `spw finalizar` |

## Diagrama da Arquitetura

```
                         +-----------------------+
                         |    Claude Code CLI     |
                         |  (slash commands /spw) |
                         +----------+------------+
                                    |
                         +----------v------------+
                         |   SPW Go CLI (hooks)   |
                         |   cli/cmd/spw/main.go  |
                         +----------+------------+
                                    |
                    +---------------+---------------+
                    |                               |
           +--------v--------+            +---------v---------+
           |   tools package  |            |   hook package     |
           | dispatch_init    |            | statusline         |
           | dispatch_setup   |            | guard_stop         |
           | dispatch_handoff |            | guard_prompt       |
           | wave_update      |            | session_start      |
           | task_mark        |            +---+---------------+
           | impl_log         |                |
           +--------+--------+                |
                    |                          |
           +--------v--------------------------v--------+
           |              store package                  |
           |  SpecStore (spec.db)  |  IndexStore (.spw)  |
           +---------+------------+----------+----------+
                     |                       |
              +------v------+         +------v------+
              |   spec.db   |         | .spw-index  |
              |  per spec   |         |    .db      |
              +-------------+         +-------------+
              | runs        |         | specs       |
              | subagents   |         | documents   |
              | waves       |         | docs_fts    |
              | tasks       |         | docs_vec    |
              | impl_logs   |         +-------------+
              | handoffs    |
              | approvals   |
              | artifacts   |
              | spec_meta   |
              +-------------+
```

## Padrão Harvest

O padrão central da migração é o **harvest** (colheita). Durante execução, os subagents continuam escrevendo arquivos normalmente -- os workflows Markdown não mudam. No ponto de handoff, os arquivos são colhidos para o banco de dados.

### Fluxo de Execução com Harvest

```
1. DispatchInit              2. DispatchSetup           3. Subagent executa
   |                            |                          |
   v                            v                          v
   CREATE run-NNN dir           CREATE <agent>/            WRITE brief.md
   INSERT runs table            WRITE brief.md             WRITE report.md
                                INSERT subagents           WRITE status.json

4. DispatchHandoff (harvest point)
   |
   v
   READ */status.json      <-- lê do filesystem
   WRITE _handoff.md       <-- escreve no filesystem
   HARVEST to DB           <-- colhe tudo para spec.db
     UPDATE subagents SET report, status, summary
     UPDATE runs SET handoff_md, all_pass
```

### Por que harvest e não write-through?

- **Compatibilidade total**: workflows existentes referenciam caminhos no filesystem (`<runDir>/<subagent>/status.json`). Mudar isso quebraria todos os commands e workflows.
- **Transacional**: o harvest acontece num ponto bem definido (handoff), onde todos os arquivos já foram escritos. Não há risco de estado parcial.
- **Gradual**: novos reads podem consultar o DB primeiro e cair para filesystem se vazio. Não há big-bang.

## Fases da Migração

| Fase | Nome | Objetivo |
|------|------|----------|
| 1 | Store + Dual-Write | Criar pacote `store`, schema SQLite, integrar dual-write nos pontos de escrita |
| 2 | DB-First Reads | Trocar reads para DB-first com file fallback |
| 3 | Cleanup + Vector | FTS5, sqlite-vec opt-in, embeddings, cleanup de arquivos transientes |
| 4 | Novos Comandos | `spw finalizar`, `spw view`, `spw search`, `spw summary` |

A ordem de implementação prioriza valor imediato:

```
Fase 1 (store)  ──>  Fase 4.1 (finalizar)  ──>  Fase 4.4 (summary)
                 ──>  Fase 4.2 (view)       ──>  Fase 2 (db-first reads)
                                             ──>  Fase 3 (FTS5 + vec)
                                             ──>  Fase 4.3 (search)
```

## Dependências Novas

```
modernc.org/sqlite                          # Pure Go SQLite (sem CGO)
github.com/charmbracelet/glamour            # Markdown rendering terminal
gopkg.in/yaml.v3                            # YAML frontmatter
github.com/asg017/sqlite-vec-go-bindings    # sqlite-vec (opt-in, Fase 3)
```
