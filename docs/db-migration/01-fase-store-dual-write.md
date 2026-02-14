# Fase 1: Store Package + Schema + Dual-Write

## Objetivo

Criar o pacote `cli/internal/store/` com schema SQLite completo, operações CRUD, sistema de migrations, e integrar dual-write nos pontos de escrita existentes. Nesta fase, o filesystem continua sendo a fonte de verdade -- o DB é populado em paralelo.

---

## Schema SQL Completo

O schema é embeddado no binário via `//go:embed schema.sql` e versionado com migrations incrementais.

### Tabela `schema_version`

```sql
CREATE TABLE IF NOT EXISTS schema_version (
    version   INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Tabela `spec_meta`

Singleton key/value com metadados da spec.

```sql
CREATE TABLE spec_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Chaves esperadas: name, stage, status, created_at, completed_at
-- status: "active" | "completed"
-- stage: "requirements" | "design" | "planning" | "execution" | "qa" | "post-mortem" | "complete"
```

### Tabela `runs`

Cada execução de um comando (dispatch_init) gera um run.

```sql
CREATE TABLE runs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id      TEXT NOT NULL,               -- "run-001", "run-002"
    command     TEXT NOT NULL,               -- "prd", "design-research", "exec", etc.
    phase       TEXT NOT NULL,               -- "prd", "design", "planning", "execution", "qa", "post-mortem"
    category    TEXT,                        -- "pipeline", "audit", "wave-execution"
    subcategory TEXT,                        -- "research", "synthesis", "artifact", "code", "implementation", "validation"
    wave_num    INTEGER,                     -- NULL se não é wave command
    comms_path  TEXT NOT NULL,               -- caminho relativo dentro do spec dir
    status      TEXT NOT NULL DEFAULT 'created',  -- "created", "running", "pass", "fail", "blocked"
    handoff_md  TEXT,                        -- conteúdo do _handoff.md
    all_pass    INTEGER,                     -- 1 se todos subagents passaram, 0 se não
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),

    UNIQUE(run_id, command, phase)
);

CREATE INDEX idx_runs_command ON runs(command);
CREATE INDEX idx_runs_wave ON runs(wave_num);
CREATE INDEX idx_runs_status ON runs(status);
```

### Tabela `subagents`

Cada subagent dentro de um run.

```sql
CREATE TABLE subagents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id      INTEGER NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,               -- nome do subagent
    model       TEXT,                        -- alias do modelo ("haiku", "opus", "sonnet")
    brief_md    TEXT,                        -- conteúdo do brief.md
    report_md   TEXT,                        -- conteúdo do report.md
    status      TEXT,                        -- "pass", "fail", "blocked", NULL
    summary     TEXT,                        -- summary do status.json
    status_json TEXT,                        -- raw JSON do status.json
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),

    UNIQUE(run_id, name)
);

CREATE INDEX idx_subagents_run ON subagents(run_id);
CREATE INDEX idx_subagents_status ON subagents(status);
```

### Tabela `waves`

Estado de cada wave de execução.

```sql
CREATE TABLE waves (
    wave_num        INTEGER PRIMARY KEY,
    status          TEXT NOT NULL DEFAULT 'pending', -- "pending", "executing", "checkpoint", "passed", "failed"
    task_ids        TEXT,                            -- JSON array: [1, 2, 3]
    exec_runs       INTEGER NOT NULL DEFAULT 0,     -- contagem de runs de execução
    check_runs      INTEGER NOT NULL DEFAULT 0,     -- contagem de runs de checkpoint
    summary_status  TEXT,                            -- status do _wave-summary.json
    summary_text    TEXT,                            -- texto do summary
    summary_source  TEXT,                            -- "execution" ou "checkpoint"
    latest_json     TEXT,                            -- raw JSON do _latest.json
    stale           INTEGER NOT NULL DEFAULT 0,      -- 1 se summary está desatualizado
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Tabela `tasks`

Cache do estado parseado de `tasks.md`. A fonte de verdade continua sendo o arquivo.

```sql
CREATE TABLE tasks (
    id          INTEGER PRIMARY KEY,         -- task ID numérico
    title       TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending', -- "pending", "done", "in_progress"
    wave        INTEGER,                     -- wave atribuída
    depends_on  TEXT,                        -- JSON array: [1, 2]
    files       TEXT,                        -- lista de arquivos (texto livre)
    tdd         INTEGER NOT NULL DEFAULT 0,  -- 1 se task tem TDD
    is_deferred INTEGER NOT NULL DEFAULT 0,  -- 1 se task foi adiada
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_wave ON tasks(wave);
```

### Tabela `impl_logs`

Implementation logs por task.

```sql
CREATE TABLE impl_logs (
    task_id      INTEGER PRIMARY KEY,
    content      TEXT NOT NULL,
    content_hash TEXT NOT NULL,              -- SHA256 do conteúdo
    wave_num     INTEGER,
    title        TEXT,
    files        TEXT,                       -- lista de arquivos modificados
    path         TEXT NOT NULL,              -- caminho original do arquivo
    created_at   TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Tabela `handoffs`

Registros de handoff separados (caso necessário para histórico).

```sql
CREATE TABLE handoffs (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id   INTEGER NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    content  TEXT NOT NULL,                  -- conteúdo do _handoff.md
    all_pass INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),

    UNIQUE(run_id)
);
```

### Tabela `approvals`

Registros de aprovação importados dos JSONs.

```sql
CREATE TABLE approvals (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_type    TEXT NOT NULL,               -- "requirements", "design", "tasks"
    approval_id TEXT NOT NULL,               -- ID da aprovação
    file_path   TEXT,                        -- caminho do arquivo aprovado
    raw_json    TEXT NOT NULL,               -- JSON original completo
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),

    UNIQUE(doc_type, approval_id)
);

CREATE INDEX idx_approvals_doc ON approvals(doc_type);
```

### Tabela `artifacts`

Artefatos genéricos (briefs, reports, summaries, etc.) com histórico.

```sql
CREATE TABLE artifacts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    phase         TEXT NOT NULL,              -- "prd", "design", "planning", "execution", "qa", "post-mortem"
    rel_path      TEXT NOT NULL,              -- caminho relativo dentro da spec
    artifact_type TEXT NOT NULL,              -- "brief", "report", "status", "summary", "log", "plan"
    content       TEXT,                       -- conteúdo do arquivo
    content_hash  TEXT,                       -- SHA256
    metadata      TEXT,                       -- JSON com metadados extras
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now')),

    UNIQUE(phase, rel_path)
);

CREATE INDEX idx_artifacts_phase ON artifacts(phase);
CREATE INDEX idx_artifacts_type ON artifacts(artifact_type);
```

### Tabela `artifact_history`

Log append-only de mudanças em artefatos.

```sql
CREATE TABLE artifact_history (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    artifact_id  INTEGER NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    content      TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    changed_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_artifact_history_artifact ON artifact_history(artifact_id);
```

### Tabela `audit_iterations`

Estado de iterações de auditoria inline.

```sql
CREATE TABLE audit_iterations (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    run_dir          TEXT NOT NULL,           -- caminho do run dir
    audit_type       TEXT NOT NULL,           -- "audit" ou "checkpoint"
    current_iteration INTEGER NOT NULL DEFAULT 0,
    max_iterations   INTEGER NOT NULL DEFAULT 3,
    history          TEXT,                    -- JSON array com histórico
    created_at       TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at       TEXT NOT NULL DEFAULT (datetime('now')),

    UNIQUE(run_dir, audit_type)
);
```

### Tabela `completion_summary`

Singleton com o resumo de finalização.

```sql
CREATE TABLE completion_summary (
    id         INTEGER PRIMARY KEY CHECK (id = 1), -- singleton
    frontmatter TEXT NOT NULL,               -- YAML frontmatter
    body        TEXT NOT NULL,               -- Markdown body
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
```

---

## Package `store` -- API Design

### `store.go` -- Constructor e Lifecycle

```go
package store

import "database/sql"

// SpecStore gerencia o banco de dados de uma spec individual.
type SpecStore struct {
    db      *sql.DB
    specDir string  // caminho absoluto do diretório da spec
    name    string  // nome da spec
}

// Open abre (ou cria) o spec.db no diretório da spec.
// Configura WAL mode e busy_timeout para concorrência.
func Open(specDir string) (*SpecStore, error)

// Close fecha a conexão com o banco.
func (s *SpecStore) Close() error

// Name retorna o nome da spec.
func (s *SpecStore) Name() string

// DB retorna a conexão para queries diretas (uso avançado).
func (s *SpecStore) DB() *sql.DB
```

**Pragmas aplicados no Open:**

```sql
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;
PRAGMA foreign_keys = ON;
PRAGMA synchronous = NORMAL;
```

### `migrate.go` -- Sistema de Migrations

```go
// Migrate aplica todas as migrations pendentes.
func (s *SpecStore) Migrate() error

// CurrentVersion retorna a versão atual do schema.
func (s *SpecStore) CurrentVersion() (int, error)

// NeedsMigration verifica se há migrations pendentes.
func (s *SpecStore) NeedsMigration() (bool, error)
```

O sistema usa `schema_version` para tracking. Migrations são funções Go registradas em ordem:

```go
var migrations = []Migration{
    {Version: 1, Up: migrateV1},  // schema inicial completo
}

type Migration struct {
    Version int
    Up      func(tx *sql.Tx) error
}
```

### `harvest.go` -- Funções de Colheita

```go
// HarvestRunDir colhe todos os arquivos de um run dir para o banco.
// Lê status.json, brief.md, report.md de cada subagent e _handoff.md do run.
func (s *SpecStore) HarvestRunDir(runDir, command string, waveNum *int) error

// HarvestArtifact colhe um único artefato do filesystem para o banco.
func (s *SpecStore) HarvestArtifact(phase, relPath, absPath string) error

// HarvestImplLog colhe um implementation log para o banco.
func (s *SpecStore) HarvestImplLog(taskID int, absPath string) error

// HarvestApprovals importa todos os approval JSONs de uma spec.
func (s *SpecStore) HarvestApprovals(approvalsDir string) error

// HarvestAll colhe todos os artefatos remanescentes do filesystem.
// Usado pelo `spw migrate` e `spw finalizar`.
func (s *SpecStore) HarvestAll() error
```

### `types.go` -- Structs

```go
type Run struct {
    ID          int
    RunID       string
    Command     string
    Phase       string
    Category    string
    Subcategory string
    WaveNum     *int
    CommsPath   string
    Status      string
    HandoffMD   *string
    AllPass     *bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Subagent struct {
    ID         int
    RunID      int
    Name       string
    Model      string
    BriefMD    *string
    ReportMD   *string
    Status     *string
    Summary    *string
    StatusJSON *string
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type WaveState struct {
    WaveNum       int
    Status        string
    TaskIDs       []int
    ExecRuns      int
    CheckRuns     int
    SummaryStatus *string
    SummaryText   *string
    SummarySource *string
    LatestJSON    *string
    Stale         bool
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type Task struct {
    ID         int
    Title      string
    Status     string
    Wave       *int
    DependsOn  []int
    Files      string
    TDD        bool
    IsDeferred bool
    UpdatedAt  time.Time
}

type ImplLog struct {
    TaskID      int
    Content     string
    ContentHash string
    WaveNum     *int
    Title       string
    Files       string
    Path        string
    CreatedAt   time.Time
}
```

### `index.go` -- IndexStore (DB Global)

```go
// IndexStore gerencia o banco de dados global de índice cross-spec.
type IndexStore struct {
    db *sql.DB
}

// OpenIndex abre (ou cria) o .spw-index.db no workspace root.
func OpenIndex(workspaceRoot string) (*IndexStore, error)

// Close fecha a conexão.
func (idx *IndexStore) Close() error

// UpsertSpec registra ou atualiza uma spec no índice.
func (idx *IndexStore) UpsertSpec(name, stage, dbPath string) error

// IndexDocument indexa um documento para busca FTS5.
func (idx *IndexStore) IndexDocument(spec, docType, phase, title, content string) error

// Search executa busca FTS5 e retorna resultados ranqueados.
func (idx *IndexStore) Search(query string, specFilter string, limit int) ([]SearchResult, error)

type SearchResult struct {
    Spec    string
    Type    string
    Phase   string
    Title   string
    Snippet string
    Rank    float64
}
```

**Schema do IndexStore:**

```sql
CREATE TABLE specs (
    name      TEXT PRIMARY KEY,
    stage     TEXT NOT NULL,
    status    TEXT NOT NULL DEFAULT 'active',
    db_path   TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE documents (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    spec    TEXT NOT NULL REFERENCES specs(name),
    type    TEXT NOT NULL,       -- "brief", "report", "checkpoint", "impl-log", "completion"
    phase   TEXT NOT NULL,
    title   TEXT NOT NULL,
    snippet TEXT,                -- primeiros 200 chars
    content TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE VIRTUAL TABLE documents_fts USING fts5(
    title, content, spec, type,
    content=documents, content_rowid=id
);

-- Triggers para manter FTS5 sincronizado
CREATE TRIGGER documents_ai AFTER INSERT ON documents BEGIN
    INSERT INTO documents_fts(rowid, title, content, spec, type)
    VALUES (new.id, new.title, new.content, new.spec, new.type);
END;

CREATE TRIGGER documents_ad AFTER DELETE ON documents BEGIN
    INSERT INTO documents_fts(documents_fts, rowid, title, content, spec, type)
    VALUES ('delete', old.id, old.title, old.content, old.spec, old.type);
END;

CREATE TRIGGER documents_au AFTER UPDATE ON documents BEGIN
    INSERT INTO documents_fts(documents_fts, rowid, title, content, spec, type)
    VALUES ('delete', old.id, old.title, old.content, old.spec, old.type);
    INSERT INTO documents_fts(rowid, title, content, spec, type)
    VALUES (new.id, new.title, new.content, new.spec, new.type);
END;
```

---

## Pontos de Integração Dual-Write

Cada ponto de integração segue o padrão: operação no filesystem completa com sucesso, depois escreve no DB. Se o DB falhar, loga warning mas não bloqueia a operação.

### 1. Run Creation -- `dispatch_init.go`

**Arquivo:** `cli/internal/tools/dispatch_init.go` (linhas 88-93)

**Contexto:** Após `os.MkdirAll(runDir, 0o755)` criar o diretório do run.

```go
// Após os.MkdirAll(runDir, 0o755) na linha 93:
if s, err := store.Open(specDir); err == nil {
    defer s.Close()
    s.CreateRun(command, runID, meta.Phase, waveNumPtr, commsPath)
    // Erro ignorado -- filesystem é fonte de verdade
}
```

**Dados inseridos:** `{run_id, command, phase, category, subcategory, wave_num, comms_path}`

### 2. Subagent Creation -- `dispatch_setup.go`

**Arquivo:** `cli/internal/tools/dispatch_setup.go` (linhas 28-91)

**Contexto:** Após `os.WriteFile(briefFullPath, briefContent, 0o644)` escrever o brief.

```go
// Após os.WriteFile(briefFullPath, ...) na seção de criação:
if s, err := store.Open(specDir); err == nil {
    defer s.Close()
    s.CreateSubagent(runDBID, subagentName, modelAlias, string(briefContent))
}
```

**Dados inseridos:** `{run_id (FK), name, model, brief_md}`

### 3. Handoff Generation -- `dispatch_handoff.go`

**Arquivo:** `cli/internal/tools/dispatch_handoff.go` (linhas 83-86)

**Contexto:** Após `os.WriteFile(handoffPath, handoffContent, 0o644)` e leitura de todos os `status.json`.

```go
// Após os.WriteFile(handoffPath, ...) na linha 86:
if s, err := store.Open(specDir); err == nil {
    defer s.Close()
    s.HarvestRunDir(runDir, command, waveNumPtr)
    // Colhe: handoff_md, all_pass, subagent statuses/reports
}
```

**Dados atualizados:**
- `runs`: SET `handoff_md`, `all_pass`, `status`
- `subagents`: SET `report_md`, `status`, `summary`, `status_json` (por subagent)

### 4. Wave Summary Update -- `wave_update.go`

**Arquivo:** `cli/internal/tools/wave_update.go` (linhas 47-71)

**Contexto:** Após escrever `_wave-summary.json` e `_latest.json`.

```go
// Após writeJSONFile(_wave-summary.json) e writeJSONFile(_latest.json):
if s, err := store.Open(specDir); err == nil {
    defer s.Close()
    s.UpsertWave(waveNum, status, taskIDs, summaryStatus, summaryText, source, latestJSON)
}
```

**Dados upserted:** `{wave_num, status, task_ids, summary_status, summary_text, summary_source, latest_json}`

### 5. Task Mark -- `task_mark.go`

**Arquivo:** `cli/internal/tools/task_mark.go` (linhas 62-63)

**Contexto:** Após `os.WriteFile(tasksPath, updatedContent, 0o644)` marcar o checkbox no arquivo.

```go
// Após os.WriteFile(tasksPath, ...) na linha 63:
if s, err := store.Open(specDir); err == nil {
    defer s.Close()
    s.UpdateTaskStatus(taskID, newStatus)
}
```

**Dados atualizados:** `tasks.status` para o task_id específico

### 6. Implementation Log Register -- `impl_log.go`

**Arquivo:** `cli/internal/tools/impl_log.go` (linha 44)

**Contexto:** Após `os.WriteFile(logPath, content, 0o644)` escrever o log.

```go
// Após os.WriteFile(logPath, ...) na linha 44:
if s, err := store.Open(specDir); err == nil {
    defer s.Close()
    s.HarvestImplLog(taskID, logPath)
}
```

**Dados inseridos:** `{task_id, content, content_hash, wave_num, title, files, path}`

### 7. Audit Iteration State -- `audit_iteration.go`

**Arquivo:** `cli/internal/tools/audit_iteration.go` (linhas 86-88 start, 239-241 advance)

**Contexto:** Após escrever `_iteration-state.json`.

```go
// Após os.WriteFile(_iteration-state.json, ...):
if s, err := store.Open(specDir); err == nil {
    defer s.Close()
    s.UpsertAuditIteration(runDir, auditType, currentIter, maxIter, history)
}
```

### 8. Statusline Cache -- `cache.go`

**Arquivo:** `cli/internal/hook/cache.go` (linha 50)

**Contexto:** Após escrever `.spw-cache/statusline.json`. Este é P2 e pode ser migrado para DB-only futuramente.

---

## Sistema de Migration

### Design

O sistema de migrations é simples e embeddado no binário:

```go
//go:embed schema.sql
var schemaSQL string

var migrations = []Migration{
    {
        Version: 1,
        Up: func(tx *sql.Tx) error {
            _, err := tx.Exec(schemaSQL)
            return err
        },
    },
    // Futuras migrations aqui
}
```

### Execução

```go
func (s *SpecStore) Migrate() error {
    current, _ := s.CurrentVersion()
    for _, m := range migrations {
        if m.Version > current {
            tx, _ := s.db.Begin()
            if err := m.Up(tx); err != nil {
                tx.Rollback()
                return err
            }
            tx.Exec("INSERT INTO schema_version (version) VALUES (?)", m.Version)
            tx.Commit()
        }
    }
    return nil
}
```

### Auto-migrate

O `Open()` chama `Migrate()` automaticamente. Se o banco não existe, cria e aplica todas as migrations. Se existe, aplica apenas as pendentes.

---

## Comando `spw migrate <spec>`

Novo subcomando para migrar specs existentes (que já possuem artefatos no filesystem) para o banco de dados.

### Uso

```bash
spw migrate my-feature          # migra uma spec
spw migrate --all               # migra todas as specs
spw migrate my-feature --dry-run  # mostra o que seria migrado
```

### Algoritmo

1. Resolver spec dir via `specdir.Resolve(cwd, specName)`
2. Abrir (ou criar) `spec.db` via `store.Open(specDir)`
3. Rodar `Migrate()` para garantir schema atualizado
4. Chamar `HarvestAll()` que:
   - Varre `_comms/` de cada fase, colhendo runs e subagents
   - Colhe `_wave-summary.json` e `_latest.json` de cada wave
   - Parseia `tasks.md` e popula tabela `tasks`
   - Colhe implementation logs de `_implementation-logs/`
   - Importa approval JSONs de `.spec-workflow/approvals/<spec>/`
   - Define `spec_meta` (name, stage via ClassifyStage, status)
5. Reportar estatísticas: runs colhidos, subagents, waves, tasks, logs

### Implementação

**Arquivo:** `cli/internal/cli/migrate_cmd.go`

```go
cmd := &cobra.Command{
    Use:   "migrate [spec-name]",
    Short: "Migrate existing spec artifacts to SQLite database",
    Args:  cobra.MaximumNArgs(1),
}
cmd.Flags().Bool("all", false, "Migrate all specs")
cmd.Flags().Bool("dry-run", false, "Show what would be migrated without writing")
```

---

## Arquivos a Criar

| Arquivo | Descrição |
|---------|-----------|
| `cli/internal/store/store.go` | SpecStore constructor, Open/Close, pragmas |
| `cli/internal/store/migrate.go` | Schema versioning e migrations |
| `cli/internal/store/schema.sql` | Schema completo (embedded) |
| `cli/internal/store/harvest.go` | Funções de colheita filesystem -> DB |
| `cli/internal/store/types.go` | Structs: Run, Subagent, WaveState, Task, etc. |
| `cli/internal/store/index.go` | IndexStore para DB global |
| `cli/internal/store/queries.go` | Métodos CRUD: CreateRun, CreateSubagent, UpsertWave, etc. |
| `cli/internal/cli/migrate_cmd.go` | Comando `spw migrate` |

## Arquivos a Modificar

| Arquivo | Mudança |
|---------|---------|
| `cli/go.mod` | Adicionar `modernc.org/sqlite`, `gopkg.in/yaml.v3` |
| `cli/internal/specdir/paths.go` | Adicionar `SpecDB`, `CompletionSummaryMD` |
| `cli/internal/tools/dispatch_init.go` | Dual-write após criar run dir |
| `cli/internal/tools/dispatch_setup.go` | Dual-write após criar subagent |
| `cli/internal/tools/dispatch_handoff.go` | Harvest após gerar handoff |
| `cli/internal/tools/wave_update.go` | Dual-write após atualizar wave |
| `cli/internal/tools/task_mark.go` | Sync após marcar task |
| `cli/internal/tools/impl_log.go` | Harvest após registrar log |
| `cli/internal/tools/audit_iteration.go` | Sync após atualizar iteração |
| `cli/internal/cli/root.go` | Registrar comando migrate |
