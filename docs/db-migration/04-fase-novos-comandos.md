# Fase 4: Novos Comandos

## Visão Geral

Quatro novos subcomandos do CLI `spw`:

| Comando | Objetivo | Dependência |
|---------|----------|-------------|
| `spw finalizar <spec>` | Marcar spec como completa, gerar resumo, indexar | Fase 1 |
| `spw view <spec> [tipo]` | Visualizar artefatos do DB no terminal ou VS Code | Fase 1 |
| `spw search "query"` | Busca full-text (e vetorial) cross-spec | Fase 3 |
| `spw summary <spec>` | Gerar resumo de progresso on-demand | Fase 1 |

Todos são registrados em `cli/internal/cli/root.go`:

```go
cmd.AddCommand(newFinalizarCmd())
cmd.AddCommand(newViewCmd())
cmd.AddCommand(newSearchCmd())
cmd.AddCommand(newSummaryCmd())
```

---

## `spw finalizar <spec-name>`

### Arquivo: `cli/internal/cli/finalizar_cmd.go`

### Assinatura

```bash
spw finalizar <spec-name> [flags]

Flags:
  --export       Exportar COMPLETION-SUMMARY.md para disco
  --force        Pular verificação de post-mortem
  --keep-files   Não limpar arquivos transientes do disco
```

### Algoritmo

```
1. specdir.Resolve(cwd, specName)          -- validar que spec existe
2. tasks.ParseFile(tasksPath)              -- parsear tasks.md
3. Validar: todas tasks com status [x]     -- se não, erro
4. Se !--force: verificar post-mortem/report.md existe
5. store.Open(specDir)                     -- abrir/criar spec.db
6. store.Migrate()                         -- garantir schema atualizado
7. store.HarvestAll()                      -- colher TODOS artefatos do filesystem
8. summary.GenerateCompletion(s, specDir)  -- gerar completion summary
9. store.SaveCompletionSummary(fm, body)   -- salvar no DB
10. store.SetMeta("status", "completed")   -- marcar spec como completa
    store.SetMeta("stage", "complete")
    store.SetMeta("completed_at", now)
11. index.OpenIndex(workspaceRoot)         -- abrir indice global
12. index.UpsertSpec(name, "complete", dbPath)
13. index.IndexDocuments(...)              -- indexar todos os documentos
14. Se embeddings_enabled:
      embed.GenerateEmbedding(content)     -- gerar embeddings
      index.StoreEmbedding(docID, vec)     -- armazenar no vec0
15. Se --export:
      Escrever COMPLETION-SUMMARY.md na raiz da spec
16. Se cleanup_after_finalize && !--keep-files:
      cleanupTransientFiles(specDir)       -- limpar transientes
17. Imprimir estatísticas finais
```

### Output Exemplo

```
Finalizing spec: user-authentication

  Tasks: 8/8 completed
  Waves: 3 (all passed)
  Artifacts harvested: 47
  Documents indexed: 23
  Embeddings generated: 23

  COMPLETION-SUMMARY.md exported to:
    .spec-workflow/specs/user-authentication/COMPLETION-SUMMARY.md

  Transient files cleaned: 156 files removed

Done. Spec "user-authentication" marked as completed.
```

---

## YAML Frontmatter -- Especificação Completa

### Schema

```yaml
---
spec: string                    # nome da spec (obrigatório)
status: "completed"             # sempre "completed" no finalizar
completed_at: datetime          # ISO 8601: 2026-02-14T10:30:00Z
duration_days: int              # dias entre criação e conclusão

tasks_count: int                # total de tasks
waves_count: int                # total de waves executadas
checkpoint_passes: int          # total de checkpoints que passaram
checkpoint_failures: int        # total de checkpoints que falharam

files_changed:                  # lista de arquivos modificados
  - string                      # caminhos relativos ao workspace root

technologies:                   # linguagens/frameworks detectados
  - string

tags:                           # tags semânticas inferidas
  - string

summary: string                 # resumo em uma frase (multi-line com >)
---
```

### Struct Go

```go
// Package: cli/internal/summary/frontmatter.go

type CompletionFrontmatter struct {
    Spec               string    `yaml:"spec"`
    Status             string    `yaml:"status"`
    CompletedAt        time.Time `yaml:"completed_at"`
    DurationDays       int       `yaml:"duration_days"`
    TasksCount         int       `yaml:"tasks_count"`
    WavesCount         int       `yaml:"waves_count"`
    CheckpointPasses   int       `yaml:"checkpoint_passes"`
    CheckpointFailures int       `yaml:"checkpoint_failures"`
    FilesChanged       []string  `yaml:"files_changed"`
    Technologies       []string  `yaml:"technologies"`
    Tags               []string  `yaml:"tags"`
    Summary            string    `yaml:"summary"`
}

type ProgressFrontmatter struct {
    Spec          string    `yaml:"spec"`
    Status        string    `yaml:"status"`      // "in_progress"
    Stage         string    `yaml:"stage"`        // stage atual
    AsOf          time.Time `yaml:"as_of"`
    TasksDone     int       `yaml:"tasks_done"`
    TasksTotal    int       `yaml:"tasks_total"`
    TasksPending  int       `yaml:"tasks_pending"`
    CurrentWave   int       `yaml:"current_wave"`
    WavesTotal    int       `yaml:"waves_total"`
    FilesChanged  []string  `yaml:"files_changed,omitempty"`
    Technologies  []string  `yaml:"technologies,omitempty"`
}
```

### Exemplo Completo de COMPLETION-SUMMARY.md

```markdown
---
spec: user-authentication
status: completed
completed_at: 2026-02-14T10:30:00Z
duration_days: 5
tasks_count: 8
waves_count: 3
checkpoint_passes: 3
checkpoint_failures: 1
files_changed:
  - src/auth/login.go
  - src/auth/middleware.go
  - src/auth/session.go
  - src/auth/jwt.go
  - src/auth/login_test.go
  - migrations/004_add_sessions.sql
technologies:
  - Go
  - SQL
tags:
  - authentication
  - security
  - jwt
  - middleware
summary: >
  Implemented user authentication with JWT tokens including login endpoint,
  auth middleware, session management, and comprehensive test coverage.
---

# Completion Summary: user-authentication

## Tasks Completed

| # | Task | Wave | Files |
|---|------|------|-------|
| 1 | Implement JWT token generation | 1 | src/auth/jwt.go |
| 2 | Create login endpoint | 1 | src/auth/login.go |
| 3 | Add auth middleware | 2 | src/auth/middleware.go |
| 4 | Implement session storage | 2 | src/auth/session.go, migrations/004_add_sessions.sql |
| 5 | Add token refresh | 2 | src/auth/jwt.go |
| 6 | Write unit tests | 3 | src/auth/login_test.go |
| 7 | Integration testing | 3 | src/auth/login_test.go |
| 8 | Documentation | 3 | docs/auth.md |

## Wave History

| Wave | Status | Tasks | Execution Runs | Checkpoint |
|------|--------|-------|----------------|------------|
| 1 | passed | 1, 2 | 1 | passed |
| 2 | passed | 3, 4, 5 | 2 | passed (2nd attempt) |
| 3 | passed | 6, 7, 8 | 1 | passed |

## Key Decisions

- JWT with RS256 for token signing (recommended by design-research)
- Session storage in PostgreSQL with 24h TTL
- Middleware pattern with context propagation
```

---

## Regras de Inferência

### `InferTechnologies(files []string) []string`

**Arquivo:** `cli/internal/summary/infer.go`

Mapeamento de extensão para tecnologia:

| Extensão | Tecnologia |
|----------|-----------|
| `.go` | Go |
| `.ts`, `.tsx` | TypeScript |
| `.js`, `.jsx` | JavaScript |
| `.py` | Python |
| `.rs` | Rust |
| `.java` | Java |
| `.kt` | Kotlin |
| `.rb` | Ruby |
| `.sql` | SQL |
| `.sh`, `.bash` | Shell |
| `.css`, `.scss` | CSS |
| `.html` | HTML |
| `.yaml`, `.yml` | YAML |
| `.toml` | TOML |
| `.json` | JSON |
| `.proto` | Protocol Buffers |
| `.graphql`, `.gql` | GraphQL |
| `.dockerfile`, `Dockerfile` | Docker |
| `.tf` | Terraform |

**Regras adicionais:**
- Se contém `_test.go` ou `.test.ts` -> adicionar "Testing" (sem duplicar)
- Se contém `migrations/` no path -> adicionar "Database Migrations"
- Deduplicar e ordenar alfabeticamente

### `InferTags(tasks []Task) []string`

Mapeamento de keywords nos títulos de tasks para tags:

| Keyword Pattern | Tag |
|----------------|-----|
| `auth`, `login`, `jwt`, `oauth`, `session` | `authentication` |
| `database`, `db`, `sql`, `migration`, `schema` | `database` |
| `api`, `endpoint`, `route`, `handler` | `api` |
| `test`, `spec`, `coverage` | `testing` |
| `ui`, `frontend`, `component`, `page` | `frontend` |
| `deploy`, `ci`, `cd`, `pipeline` | `devops` |
| `security`, `encrypt`, `hash`, `permission` | `security` |
| `config`, `setting`, `env` | `configuration` |
| `refactor`, `cleanup`, `rename` | `refactoring` |
| `fix`, `bug`, `patch`, `hotfix` | `bugfix` |
| `doc`, `readme`, `guide` | `documentation` |
| `cache`, `redis`, `memcache` | `caching` |
| `queue`, `worker`, `async`, `job` | `async` |
| `middleware` | `middleware` |
| `webhook`, `event`, `notification` | `events` |

**Regras:**
- Case-insensitive matching
- Deduplicar
- Ordenar alfabeticamente
- Limitar a 10 tags

### `CollectFilesChanged(tasks []Task) []string`

Coleta o campo `files` de cada task, parsing conforme formato do `tasks.md`:

```go
func CollectFilesChanged(tasks []Task) []string {
    seen := make(map[string]bool)
    var files []string
    for _, t := range tasks {
        if t.Status != "done" {
            continue
        }
        // Parse files field: pode ser comma-separated ou multi-line
        for _, f := range parseFilesList(t.Files) {
            f = strings.TrimSpace(f)
            if f != "" && !seen[f] {
                seen[f] = true
                files = append(files, f)
            }
        }
    }
    sort.Strings(files)
    return files
}
```

---

## `spw view <spec> [artifact-type]`

### Arquivo: `cli/internal/cli/view_cmd.go`

### Assinatura

```bash
spw view <spec-name> [artifact-type] [flags]

Artifact types:
  overview              Visão geral da spec (default)
  report                Report de subagent
  brief                 Brief de subagent
  checkpoint            Report de checkpoint
  implementation-log    Log de implementação de task
  wave-summary          Resumo de wave
  completion-summary    Resumo de finalização

Flags:
  --wave int      Número da wave (para checkpoint, wave-summary)
  --run int       Número do run
  --task string   ID da task (para implementation-log)
  --vscode        Abrir no VS Code
  --raw           Output raw Markdown (sem rendering)
```

### Exemplos

```bash
# Visão geral da spec
spw view my-feature

# Checkpoint da wave 2
spw view my-feature checkpoint --wave 2

# Implementation log da task 3
spw view my-feature implementation-log --task 3

# Report do último run de design-research
spw view my-feature report --run 1

# Abrir no VS Code
spw view my-feature completion-summary --vscode

# Raw markdown (para piping)
spw view my-feature overview --raw | less
```

### Rendering com Glamour

```go
import "github.com/charmbracelet/glamour"

func renderMarkdown(content, filename string, vscode, raw bool) error {
    if raw {
        fmt.Print(content)
        return nil
    }

    if vscode {
        cmd := exec.Command("code", "--stdin", "--filename", filename)
        cmd.Stdin = strings.NewReader(content)
        return cmd.Run()
    }

    // Glamour rendering no terminal
    r, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(100),
    )
    out, _ := r.Render(content)
    fmt.Print(out)
    return nil
}
```

### Overview Default

Quando nenhum artifact-type é especificado, mostra um overview composto:

```
Spec: user-authentication
Stage: execution (wave 2 of 3)
Status: active

Tasks: 5/8 done, 3 in wave 3
Files changed: 6

Waves:
  Wave 1: passed (2 tasks, 1 exec run, 1 checkpoint)
  Wave 2: passed (3 tasks, 2 exec runs, 1 checkpoint)
  Wave 3: executing (3 tasks, 0 exec runs)

Recent activity:
  2026-02-14 10:30 - Wave 2 checkpoint passed
  2026-02-14 09:15 - Wave 2 execution run 2
  2026-02-13 16:00 - Wave 2 execution run 1 (checkpoint failed)
```

---

## `spw search "query"`

### Arquivo: `cli/internal/cli/search_cmd.go`

### Assinatura

```bash
spw search <query> [flags]

Flags:
  --spec string    Filtrar por spec
  --limit int      Máximo de resultados (default: 5)
  --type string    Filtrar por tipo de documento
  --vector         Forçar busca vetorial (requer Ollama)
```

### Algoritmo

```
1. OpenIndex(workspaceRoot)
2. Se --vector ou (embeddings_enabled && vec table existe):
     embedding := GenerateEmbedding(query)
     Se embedding != nil:
       results := HybridSearch(query, embedding, opts)
     Senão:
       results := Search(query, opts)  -- fallback FTS5
3. Senão:
     results := Search(query, opts)    -- FTS5 only
4. Render resultados formatados
```

### Output Exemplo

```
Search: "authentication implementation"
Found 3 results:

  1. [completion] user-authentication
     Implemented user authentication with JWT tokens including login endpoint,
     auth middleware, session management...
     Score: 0.92

  2. [impl-log] user-authentication / task-2
     Created login endpoint with password hashing using bcrypt,
     input validation, and rate limiting...
     Score: 0.78

  3. [report] api-refactor / design-research / run-001 / security-analyst
     Analyzed authentication patterns including OAuth2, JWT, and session-based
     approaches for the API refactoring...
     Score: 0.65
```

---

## `spw summary <spec>`

### Arquivo: `cli/internal/cli/summary_cmd.go`

### Assinatura

```bash
spw summary <spec-name> [flags]

Flags:
  --export    Exportar PROGRESS-SUMMARY.md para disco
  --vscode    Abrir no VS Code
  --raw       Raw Markdown
```

### Diferença entre summary e finalizar

| Aspecto | `spw summary` | `spw finalizar` |
|---------|---------------|-----------------|
| Requer todas tasks completas | Nao | Sim |
| Funciona em qualquer stage | Sim | Apenas post-execution |
| Altera estado | Nao (read-only) | Sim (marca completed) |
| Indexa no global DB | Nao | Sim |
| Gera embeddings | Nao | Sim (se habilitado) |
| Cleanup de arquivos | Nao | Sim (se habilitado) |
| Frontmatter type | `ProgressFrontmatter` | `CompletionFrontmatter` |

### Algoritmo

```
1. specdir.Resolve(cwd, specName)
2. tasks.ParseFile(tasksPath)
3. Se spec.db existe:
     store.Open(specDir)
     waves := store.ScanWaves()
4. Senão:
     waves := wave.ScanWaves(specDir)
5. fm := summary.GenerateProgress(tasks, waves, specDir)
6. body := summary.RenderProgressBody(tasks, waves)
7. content := formatFrontmatter(fm) + body
8. Se --export: escrever PROGRESS-SUMMARY.md
9. Se --vscode: abrir no VS Code
10. Senão: render com glamour
```

### Output Exemplo (renderizado)

```
---
spec: user-authentication
status: in_progress
stage: execution
as_of: 2026-02-14T10:30:00Z
tasks_done: 5
tasks_total: 8
tasks_pending: 3
current_wave: 3
waves_total: 3
files_changed:
  - src/auth/login.go
  - src/auth/middleware.go
technologies:
  - Go
  - SQL
---

# Progress Summary: user-authentication

## Task Status

  Done: 5/8 (62.5%)
  Pending: 3 (wave 3)

| # | Task | Status | Wave |
|---|------|--------|------|
| 1 | Implement JWT token generation | done | 1 |
| 2 | Create login endpoint | done | 1 |
| 3 | Add auth middleware | done | 2 |
| 4 | Implement session storage | done | 2 |
| 5 | Add token refresh | done | 2 |
| 6 | Write unit tests | pending | 3 |
| 7 | Integration testing | pending | 3 |
| 8 | Documentation | pending | 3 |

## Wave Status

| Wave | Status | Tasks | Checkpoint |
|------|--------|-------|------------|
| 1 | passed | 1, 2 | passed |
| 2 | passed | 3, 4, 5 | passed |
| 3 | executing | 6, 7, 8 | -- |
```

---

## Package `summary` (`cli/internal/summary/`)

### Arquivos

| Arquivo | Conteúdo |
|---------|----------|
| `summary.go` | `GenerateCompletion()`, `GenerateProgress()`, `RenderCompletionBody()`, `RenderProgressBody()` |
| `frontmatter.go` | Structs `CompletionFrontmatter`, `ProgressFrontmatter`, `FormatFrontmatter()` |
| `infer.go` | `InferTechnologies()`, `InferTags()`, `CollectFilesChanged()` |

### API Principal

```go
// GenerateCompletion gera o frontmatter e body de um completion summary.
func GenerateCompletion(s *store.SpecStore, specDir string, tasks []tasks.Task) (CompletionFrontmatter, string, error)

// GenerateProgress gera o frontmatter e body de um progress summary.
func GenerateProgress(specDir string, taskList []tasks.Task, waves []wave.WaveInfo) (ProgressFrontmatter, string, error)

// FormatFrontmatter serializa frontmatter para YAML com delimitadores ---.
func FormatFrontmatter(fm interface{}) (string, error) {
    data, _ := yaml.Marshal(fm)
    return fmt.Sprintf("---\n%s---\n\n", string(data)), nil
}

// RenderCompletionBody gera o corpo Markdown do completion summary.
func RenderCompletionBody(tasks []tasks.Task, waves []wave.WaveInfo) string

// RenderProgressBody gera o corpo Markdown do progress summary.
func RenderProgressBody(tasks []tasks.Task, waves []wave.WaveInfo) string
```

---

## Registrar Comandos

### `cli/internal/cli/root.go`

Adicionar os quatro novos comandos ao root:

```go
func newRootCmd() *cobra.Command {
    cmd := &cobra.Command{...}

    // Comandos existentes
    cmd.AddCommand(newHookCmd())
    // ...

    // Novos comandos (migração DB)
    cmd.AddCommand(newFinalizarCmd())
    cmd.AddCommand(newViewCmd())
    cmd.AddCommand(newSearchCmd())
    cmd.AddCommand(newSummaryCmd())
    cmd.AddCommand(newMigrateCmd())  // da Fase 1

    return cmd
}
```

---

## Dependências

| Pacote | Uso | Fase |
|--------|-----|------|
| `modernc.org/sqlite` | SQLite pure Go | 1 |
| `gopkg.in/yaml.v3` | YAML frontmatter | 4 (finalizar) |
| `github.com/charmbracelet/glamour` | Rendering Markdown no terminal | 4 (view) |
| `github.com/asg017/sqlite-vec-go-bindings` | sqlite-vec (opt-in) | 3 |
