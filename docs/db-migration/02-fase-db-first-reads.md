# Fase 2: DB-First Reads com File Fallback

## Objetivo

Trocar todos os reads do filesystem para consultar o banco de dados primeiro, caindo para leitura de arquivo apenas quando o DB está vazio ou ausente. Isso permite que specs migradas (via `spw migrate`) se beneficiem imediatamente de queries mais rápidas, enquanto specs não-migradas continuam funcionando normalmente.

---

## Padrão de Implementação

Todas as funções seguem o mesmo padrão DB-first + fallback:

```go
func ReadSomething(specDir string, args...) (Result, error) {
    // 1. Tentar DB
    if s, err := store.Open(specDir); err == nil {
        defer s.Close()
        if result, err := s.QuerySomething(args...); err == nil {
            return result, nil
        }
        // DB vazio ou query sem resultado -- cair para arquivo
    }

    // 2. Fallback para filesystem
    return readSomethingFromFile(specDir, args...)
}
```

**Regras do padrão:**

1. Se `store.Open()` falha (spec.db não existe), vai direto para fallback -- sem erro
2. Se a query retorna `sql.ErrNoRows`, vai para fallback -- o dado ainda não foi migrado
3. Se a query retorna erro inesperado, loga warning e vai para fallback
4. O fallback é a implementação original (renomeada com sufixo `FromFile`)

---

## Lista Completa de Funções a Modificar

### Package `specdir` (`cli/internal/specdir/specdir.go`)

#### `ReadStatusJSON(path string) (StatusDoc, error)`

**Uso atual:** Lê `status.json` de um subagent dir.

**DB-first query:**
```go
func (s *SpecStore) GetSubagentStatusByPath(relPath string) (StatusDoc, error) {
    // Extrai o run dir e subagent name do path
    // SELECT status, summary FROM subagents
    // JOIN runs ON subagents.run_id = runs.id
    // WHERE subagents.name = ? AND runs.comms_path LIKE ?
}
```

**Chamadores:** `ResolveCheckpoint`, `GenerateSummary`, `DispatchHandoff`, `checkRunCompleteness`

#### `ReadLatestJSON(path string) (LatestDoc, error)`

**Uso atual:** Lê `_latest.json` de um wave dir.

**DB-first query:**
```go
func (s *SpecStore) GetWaveLatest(waveNum int) (LatestDoc, error) {
    // SELECT latest_json FROM waves WHERE wave_num = ?
    // Deserializa o JSON armazenado
}
```

**Chamadores:** `ResolveCheckpoint`, `waveStatusResult`, `resolveCheckpointStatus`

#### `LatestRunDir(dir string) (string, int, error)`

**Uso atual:** Varre diretório por `run-NNN` e retorna o maior.

**DB-first query:**
```go
func (s *SpecStore) GetLatestRunID(command, phase string, waveNum *int) (string, int, error) {
    // SELECT run_id FROM runs
    // WHERE command = ? AND phase = ? AND (wave_num = ? OR wave_num IS NULL)
    // ORDER BY id DESC LIMIT 1
}
```

**Chamadores:** `ResolveCheckpoint`, `GenerateSummary`, `inspectRunDir`

#### `ListWaveDirs(specDir string) ([]WaveDir, error)`

**Uso atual:** Varre `execution/waves/` por diretórios `wave-NN`.

**DB-first query:**
```go
func (s *SpecStore) ListWaves() ([]WaveDir, error) {
    // SELECT wave_num FROM waves ORDER BY wave_num
    // Retorna como WaveDir structs
}
```

**Chamadores:** `ScanWaves`, `WaveResolveCurrent`, `waveStatusResult`

---

### Package `spec` (`cli/internal/spec/`)

#### `ClassifyStage(specDir string) (string, error)` -- `stage.go`

**Uso atual:** Verifica existência de ~10 arquivos para determinar o stage.

**DB-first query:**
```go
func (s *SpecStore) GetStage() (string, error) {
    // SELECT value FROM spec_meta WHERE key = 'stage'
}
```

**Nota:** Se `spec_meta.status == "completed"`, retorna `"complete"` diretamente sem verificar arquivos.

**Chamadores:** statusline hook, `spw status`

#### `CheckApproval(cwd, specName, docType string) (ApprovalResult, error)` -- `approval.go`

**Uso atual:** Varre `approvals/<spec>/approval_*.json`.

**DB-first query:**
```go
func (s *SpecStore) GetApproval(docType string) (*Approval, error) {
    // SELECT approval_id, raw_json FROM approvals WHERE doc_type = ?
    // ORDER BY created_at DESC LIMIT 1
}
```

**Chamadores:** comandos de workflow que checam aprovações

#### `List(cwd string) ([]string, error)` -- `list.go`

**Uso atual:** `os.ReadDir` em `.spec-workflow/specs/`.

**DB-first (via IndexStore):**
```go
func (idx *IndexStore) ListSpecs() ([]SpecEntry, error) {
    // SELECT name, stage, status FROM specs ORDER BY name
}
```

**Nota:** Faz fallback para `os.ReadDir` pois specs podem existir no disco sem estarem indexadas.

---

### Package `wave` (`cli/internal/wave/`)

#### `ScanWaves(specDir string) ([]WaveInfo, error)` -- `scanner.go`

**Uso atual:** Varre diretórios de waves, conta runs, resolve checkpoints.

**DB-first query:**
```go
func (s *SpecStore) ScanWaves() ([]WaveInfo, error) {
    // SELECT w.wave_num, w.status, w.exec_runs, w.check_runs,
    //        w.summary_status, w.summary_text, w.stale
    // FROM waves w
    // ORDER BY w.wave_num
    //
    // Para cada wave, busca subagent status do último run:
    // SELECT s.name, s.status, s.summary
    // FROM subagents s JOIN runs r ON s.run_id = r.id
    // WHERE r.wave_num = ? AND r.command = 'exec'
    // ORDER BY r.id DESC
}
```

**Chamadores:** `ComputeResume`, `waveStatusResult`, statusline

#### `ResolveCheckpoint(specDir string, waveNum int) (CheckpointResult, error)` -- `checkpoint.go`

**Uso atual:** Multi-step resolution com fallbacks: `_latest.json` -> `_wave-summary.json` -> `run-NNN/release-gate-decider/status.json`.

**DB-first query:**
```go
func (s *SpecStore) ResolveCheckpoint(waveNum int) (CheckpointResult, error) {
    // 1. Checar waves.latest_json (equivalente a _latest.json)
    // 2. Se não tem, buscar último run de checkpoint:
    //    SELECT * FROM runs
    //    WHERE wave_num = ? AND command = 'checkpoint'
    //    ORDER BY id DESC LIMIT 1
    // 3. Buscar subagents do run (especialmente release-gate-decider):
    //    SELECT status, summary FROM subagents
    //    WHERE run_id = ? AND name = 'release-gate-decider'
}
```

**Nota:** A lógica de prioridade (latest > wave-summary > ground truth) deve ser replicada exatamente na query.

#### `GenerateSummary(specDir string, waveNum int) (Summary, error)` -- `summary.go`

**Uso atual:** Agrega status de todos os subagents de execution e checkpoint runs.

**DB-first query:**
```go
func (s *SpecStore) GenerateSummary(waveNum int) (Summary, error) {
    // SELECT r.command, s.name, s.status, s.summary
    // FROM subagents s JOIN runs r ON s.run_id = r.id
    // WHERE r.wave_num = ?
    // ORDER BY r.command, r.id, s.name
}
```

#### `ComputeResume(specDir string) (ResumeState, error)` -- `resume.go`

**Uso atual:** Delega para `ScanWaves`. Se `ScanWaves` já usa DB-first, este funciona automaticamente.

---

### Package `tools` (`cli/internal/tools/`)

#### `RunsLatestUnfinished(cwd, phaseDir string, raw bool)` -- `runs.go`

**Uso atual:** Varre `run-NNN/` dirs, verifica existência de `_handoff.md`, `brief.md`, `report.md`, `status.json` por subagent.

**DB-first query:**
```go
func (s *SpecStore) GetLatestUnfinishedRun(command, phase string) (*UnfinishedRun, error) {
    // SELECT r.id, r.run_id, r.command, r.status
    // FROM runs r
    // WHERE r.command = ? AND r.phase = ? AND r.status NOT IN ('pass', 'fail')
    // ORDER BY r.id DESC LIMIT 1
    //
    // Se encontrou, buscar subagents:
    // SELECT name, brief_md IS NOT NULL as has_brief,
    //        report_md IS NOT NULL as has_report,
    //        status
    // FROM subagents WHERE run_id = ?
}
```

**Chamadores:** Workflows que verificam runs incompletos antes de iniciar novo

#### `DispatchReadStatus(cwd, subagentName, runDir string, raw bool)` -- `dispatch_status.go`

**Uso atual:** `os.ReadFile` em `<runDir>/<subagent>/status.json`.

**DB-first query:**
```go
func (s *SpecStore) GetSubagentStatus(runID int, name string) (StatusDoc, error) {
    // SELECT status, summary, status_json FROM subagents
    // WHERE run_id = ? AND name = ?
}
```

#### `waveStatusResult(cwd, specName string)` -- `wave_status.go`

**Uso atual:** Combina `ListWaveDirs`, `ReadLatestJSON`, `_wave-summary.json` e `tasks.md` parsing.

**DB-first:** Se `ScanWaves`, `ReadLatestJSON` e task parsing já usam DB-first, esta função se beneficia automaticamente pela cadeia de chamadas.

#### `WaveResolveCurrent(cwd, specName string, raw bool)` -- `wave_resolve.go`

**Uso atual:** `os.Stat` e `os.ReadDir` em `execution/waves/`.

**DB-first query:**
```go
func (s *SpecStore) GetCurrentWaveNum() (int, error) {
    // SELECT MAX(wave_num) FROM waves
}
```

---

### Package `hook` (`cli/internal/hook/`)

#### `detectActiveSpec(dir string)` -- `statusline.go`

**Uso atual:** Varre diretórios de specs, verifica mtime, usa cache.

**DB-first (via IndexStore):**
```go
func (idx *IndexStore) GetActiveSpec() (string, error) {
    // SELECT name FROM specs
    // WHERE status = 'active'
    // ORDER BY updated_at DESC LIMIT 1
}
```

**Nota:** Mantém fallback para `detectSpecFromGit` e `detectSpecByMtime` para specs não-indexadas.

#### `checkRunCompleteness(runDir string)` -- `guard_stop.go`

**Uso atual:** `os.Stat` em `_handoff.md`, `os.ReadDir` para subagent dirs, verifica `brief.md`, `report.md`, `status.json` por subagent.

**DB-first query:**
```go
func (s *SpecStore) CheckRecentRunCompleteness(since time.Time) ([]IncompleteRun, error) {
    // SELECT r.run_id, r.command, r.handoff_md IS NOT NULL as has_handoff,
    //        COUNT(s.id) as total_subagents,
    //        COUNT(s.status) as completed_subagents
    // FROM runs r
    // LEFT JOIN subagents s ON s.run_id = r.id
    // WHERE r.created_at > ?
    // GROUP BY r.id
    // HAVING has_handoff = 0 OR completed_subagents < total_subagents
}
```

---

## Estratégia de Testing

### Testes de Fallback

Cada função modificada precisa de testes verificando os três cenários:

```go
func TestReadStatusJSON_DBFirst(t *testing.T) {
    // Setup: criar spec.db com dados, sem arquivo no disco
    // Assert: lê do DB corretamente
}

func TestReadStatusJSON_Fallback(t *testing.T) {
    // Setup: spec.db não existe, arquivo status.json no disco
    // Assert: lê do arquivo corretamente
}

func TestReadStatusJSON_DBEmpty(t *testing.T) {
    // Setup: spec.db existe mas sem o registro, arquivo no disco
    // Assert: fallback para arquivo
}
```

### Testes de Consistência

Verificar que DB e filesystem retornam resultados idênticos:

```go
func TestScanWaves_Consistency(t *testing.T) {
    // Setup: criar spec com waves no filesystem + migrar para DB
    // Assert: ScanWaves via DB == ScanWaves via filesystem
}
```

### Testes com `:memory:` DB

Usar banco em memória para testes rápidos:

```go
func setupTestStore(t *testing.T) *SpecStore {
    s := &SpecStore{}
    db, _ := sql.Open("sqlite", ":memory:")
    s.db = db
    s.Migrate()
    return s
}
```

### Matriz de Cobertura

| Função | DB-first | Fallback | Consistência |
|--------|----------|----------|--------------|
| `ReadStatusJSON` | x | x | x |
| `ReadLatestJSON` | x | x | x |
| `LatestRunDir` | x | x | x |
| `ListWaveDirs` | x | x | x |
| `ClassifyStage` | x | x | x |
| `CheckApproval` | x | x | x |
| `ScanWaves` | x | x | x |
| `ResolveCheckpoint` | x | x | x |
| `GenerateSummary` | x | x | - |
| `RunsLatestUnfinished` | x | x | x |
| `DispatchReadStatus` | x | x | x |
| `detectActiveSpec` | x | x | - |
| `checkRunCompleteness` | x | x | x |

---

## Ordem de Migração Recomendada

A migração dos reads deve seguir a ordem de dependência entre funções:

### Camada 1 -- Reads Primitivos (sem dependências)

Estas funções fazem I/O direto e não chamam outras funções da lista:

1. `ReadStatusJSON` -- usado por muitas outras funções
2. `ReadLatestJSON` -- usado por checkpoint e wave status
3. `LatestRunDir` -- usado por checkpoint e summary
4. `ListWaveDirs` -- usado por ScanWaves

### Camada 2 -- Reads Compostos (dependem da Camada 1)

Estas funções chamam reads primitivos que já foram migrados:

5. `ResolveCheckpoint` -- usa ReadLatestJSON, ReadStatusJSON, LatestRunDir
6. `ScanWaves` -- usa ListWaveDirs, LatestRunDir, ResolveCheckpoint
7. `GenerateSummary` -- usa ReadStatusJSON, LatestRunDir

### Camada 3 -- Reads de Alto Nível (dependem da Camada 2)

8. `ComputeResume` -- delega para ScanWaves (automático)
9. `RunsLatestUnfinished` -- read independente mas complexo
10. `ClassifyStage` -- pode usar spec_meta diretamente
11. `CheckApproval` -- read independente

### Camada 4 -- Hooks e Tools

12. `detectActiveSpec` -- usa IndexStore
13. `checkRunCompleteness` -- query direta no DB
14. `DispatchReadStatus` -- query simples
15. `waveStatusResult` -- se beneficia da cadeia

---

## Riscos e Mitigações

| Risco | Mitigação |
|-------|-----------|
| DB e filesystem dessincronizados | Fallback garante resultado correto; `spw migrate` resincroniza |
| Performance do Open() em cada read | Connection pool ou cache no processo; spec.db é pequeno |
| Checkpoint priority logic diverge | Testes de consistência obrigatórios |
| spec.db corrompido | Fallback para filesystem; `spw migrate` reconstrói |
| Concorrência (múltiplos hooks) | WAL mode + busy_timeout no SQLite |
