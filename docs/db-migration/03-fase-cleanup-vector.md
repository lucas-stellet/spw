# Fase 3: Cleanup + Vector Search

## Objetivo

Parar de manter arquivos transientes no disco após finalização, adicionar busca full-text via FTS5 no índice global, e opcionalmente habilitar busca vetorial com sqlite-vec + Ollama embeddings.

---

## FTS5 -- Full-Text Search

### Setup no IndexStore

O FTS5 já está definido no schema do IndexStore (Fase 1). Nesta fase, implementamos os triggers de sincronização e a API de busca.

### Triggers de Sincronização

```sql
-- Inserção: indexa automaticamente novos documentos
CREATE TRIGGER documents_ai AFTER INSERT ON documents BEGIN
    INSERT INTO documents_fts(rowid, title, content, spec, type)
    VALUES (new.id, new.title, new.content, new.spec, new.type);
END;

-- Deleção: remove do índice
CREATE TRIGGER documents_ad AFTER DELETE ON documents BEGIN
    INSERT INTO documents_fts(documents_fts, rowid, title, content, spec, type)
    VALUES ('delete', old.id, old.title, old.content, old.spec, old.type);
END;

-- Atualização: remove e reinsere
CREATE TRIGGER documents_au AFTER UPDATE ON documents BEGIN
    INSERT INTO documents_fts(documents_fts, rowid, title, content, spec, type)
    VALUES ('delete', old.id, old.title, old.content, old.spec, old.type);
    INSERT INTO documents_fts(rowid, title, content, spec, type)
    VALUES (new.id, new.title, new.content, new.spec, new.type);
END;
```

### API de Busca FTS5

```go
// Search executa busca FTS5 com ranking BM25.
func (idx *IndexStore) Search(query string, opts SearchOpts) ([]SearchResult, error) {
    q := `
        SELECT d.spec, d.type, d.phase, d.title,
               snippet(documents_fts, 1, '<mark>', '</mark>', '...', 32) as snippet,
               rank
        FROM documents_fts
        JOIN documents d ON d.id = documents_fts.rowid
        WHERE documents_fts MATCH ?
    `
    args := []interface{}{query}

    if opts.Spec != "" {
        q += " AND d.spec = ?"
        args = append(args, opts.Spec)
    }

    q += " ORDER BY rank LIMIT ?"
    args = append(args, opts.Limit)

    // Execute e retorna resultados
}

type SearchOpts struct {
    Spec  string
    Limit int  // default: 5
}

type SearchResult struct {
    Spec    string
    Type    string   // "brief", "report", "checkpoint", "impl-log", "completion"
    Phase   string
    Title   string
    Snippet string   // com marcação de highlight
    Rank    float64  // BM25 score
}
```

### Documentos Indexados

O `spw finalizar` indexa os seguintes documentos no `.spw-index.db`:

| Tipo | Fonte | Titulo |
|------|-------|--------|
| `completion` | completion_summary body | `<spec>: Completion Summary` |
| `report` | subagents.report_md | `<spec>/<command>/run-NNN/<agent>` |
| `checkpoint` | checkpoint reports | `<spec>/wave-NN/checkpoint` |
| `impl-log` | impl_logs.content | `<spec>/task-NN` |
| `brief` | subagents.brief_md | `<spec>/<command>/run-NNN/<agent>` |

---

## sqlite-vec -- Busca Vetorial (Opt-in)

### Dependência

```
github.com/asg017/sqlite-vec-go-bindings
```

**Nota:** sqlite-vec requer extensão nativa. A feature é opt-in -- se a extensão não carrega, o sistema continua funcionando apenas com FTS5.

### Schema Condicional

```go
func (idx *IndexStore) initVecTable() error {
    // Tenta carregar extensão sqlite-vec
    _, err := idx.db.Exec("SELECT vec_version()")
    if err != nil {
        // sqlite-vec não disponível -- skip silently
        return nil
    }

    _, err = idx.db.Exec(`
        CREATE VIRTUAL TABLE IF NOT EXISTS documents_vec
        USING vec0(embedding float[768])
    `)
    return err
}
```

A tabela `documents_vec` usa `rowid` correspondente ao `documents.id`, permitindo JOIN direto.

### Busca Híbrida (FTS5 + Vector)

```go
// HybridSearch combina FTS5 text match com KNN vector similarity.
func (idx *IndexStore) HybridSearch(query string, embedding []float32, opts SearchOpts) ([]SearchResult, error) {
    // 1. FTS5 match -> top N*2 candidatos com BM25 score
    // 2. Vector KNN nos candidatos -> distance score
    // 3. Combinar scores: hybrid = alpha * bm25_norm + (1-alpha) * (1 - vec_distance)
    // 4. Re-rank e retornar top N
}
```

**Alpha default:** 0.7 (favorece texto sobre semântica). Configurável via `[store]`.

---

## Ollama Embedding Pipeline

### Arquivo: `cli/internal/store/embed.go`

```go
// GenerateEmbedding gera embedding via Ollama API.
// Retorna nil se Ollama não está disponível (fail-open).
func GenerateEmbedding(text string, cfg EmbedConfig) ([]float32, error) {
    client := &http.Client{Timeout: 10 * time.Second}

    body := map[string]interface{}{
        "model":  cfg.Model,
        "prompt": text,
    }
    jsonBody, _ := json.Marshal(body)

    resp, err := client.Post(cfg.URL+"/api/embeddings", "application/json", bytes.NewReader(jsonBody))
    if err != nil {
        return nil, nil  // fail-open: Ollama indisponível
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, nil  // fail-open
    }

    var result struct {
        Embedding []float32 `json:"embedding"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Embedding, nil
}

type EmbedConfig struct {
    Model string  // default: "nomic-embed-text"
    URL   string  // default: "http://localhost:11434"
}
```

### Modelo: `nomic-embed-text`

- **Dimensão:** 768
- **Contexto:** 8192 tokens
- **Velocidade:** ~50ms por embedding em hardware médio
- **Instalação:** `ollama pull nomic-embed-text`

### Fluxo de Indexação com Embeddings

```
spw finalizar <spec>
    |
    v
HarvestAll() --> IndexDocument() para cada documento
    |
    v
GenerateEmbedding(content)
    |
    +-- Ollama disponível --> INSERT INTO documents_vec
    |
    +-- Ollama indisponível --> skip (apenas FTS5)
```

---

## Cleanup Strategy

### Princípio

Arquivos transientes NÃO são limpos durante execução ativa. Os workflows Markdown referenciam caminhos no filesystem, e limpar durante execução quebraria o fluxo. A limpeza acontece apenas em dois momentos:

1. **`spw finalizar`** -- limpa após marcar spec como completa
2. **Manual** -- `spw cleanup <spec>` (futuro)

### O que é limpo

| Diretório/Arquivo | Condição |
|-------------------|----------|
| `<spec>/_comms/` (todos os run dirs) | Sempre (já colhidos no DB) |
| `execution/waves/wave-NN/_wave-summary.json` | Sempre (no DB) |
| `execution/waves/wave-NN/_latest.json` | Sempre (no DB) |
| `execution/waves/wave-NN/execution/` | Sempre (runs no DB) |
| `execution/waves/wave-NN/checkpoint/` | Sempre (runs no DB) |
| `execution/_implementation-logs/` | Sempre (no DB) |
| `.spec-workflow/approvals/<spec>/` | Sempre (no DB) |

### O que NÃO é limpo

| Arquivo | Motivo |
|---------|--------|
| `requirements.md` | Dashboard MCP |
| `design.md` | Dashboard MCP |
| `tasks.md` | Dashboard MCP |
| `STATUS-SUMMARY.md` | Human-readable output |
| `COMPLETION-SUMMARY.md` | Gerado pelo finalizar (se `--export`) |

### Flag `--keep-files`

```bash
spw finalizar my-feature              # limpa transientes
spw finalizar my-feature --keep-files  # preserva tudo no disco
```

### Implementação

```go
func cleanupTransientFiles(specDir string) error {
    // Definir padrões a limpar
    patterns := []string{
        "prd/_comms",
        "design/_comms",
        "planning/_comms",
        "execution/_comms",
        "execution/waves/*/execution",
        "execution/waves/*/checkpoint",
        "execution/waves/*/_wave-summary.json",
        "execution/waves/*/_latest.json",
        "execution/_implementation-logs",
        "qa/_comms",
        "qa/qa-artifacts",
        "post-mortem/_comms",
    }

    for _, pattern := range patterns {
        matches, _ := filepath.Glob(filepath.Join(specDir, pattern))
        for _, m := range matches {
            os.RemoveAll(m)
        }
    }

    // Limpar approvals separadamente (fora do spec dir)
    approvalsDir := filepath.Join(specDir, "../../approvals", filepath.Base(specDir))
    os.RemoveAll(approvalsDir)

    return nil
}
```

---

## Configuração `[store]`

Nova seção no `spw-config.toml`:

```toml
[store]
# Habilitar geração de embeddings via Ollama
embeddings_enabled = true

# Modelo Ollama para embeddings
embeddings_model = "nomic-embed-text"

# URL do servidor Ollama
embeddings_url = "http://localhost:11434"

# Limpar arquivos transientes após `spw finalizar`
cleanup_after_finalize = true

# Alpha para busca híbrida (0.0 = só vetor, 1.0 = só texto)
hybrid_search_alpha = 0.7
```

### Atualização em `config.go`

```go
type StoreConfig struct {
    EmbeddingsEnabled    bool    `toml:"embeddings_enabled"`
    EmbeddingsModel      string  `toml:"embeddings_model"`
    EmbeddingsURL        string  `toml:"embeddings_url"`
    CleanupAfterFinalize bool    `toml:"cleanup_after_finalize"`
    HybridSearchAlpha    float64 `toml:"hybrid_search_alpha"`
}

// Defaults
func defaultStoreConfig() StoreConfig {
    return StoreConfig{
        EmbeddingsEnabled:    true,
        EmbeddingsModel:      "nomic-embed-text",
        EmbeddingsURL:        "http://localhost:11434",
        CleanupAfterFinalize: true,
        HybridSearchAlpha:    0.7,
    }
}
```

---

## Tarefas

| # | Tarefa | Dependência |
|---|--------|-------------|
| 1 | Implementar FTS5 triggers no IndexStore schema | Fase 1 completa |
| 2 | Implementar `Search()` no IndexStore com BM25 ranking | 1 |
| 3 | Implementar `embed.go` com Ollama pipeline | -- |
| 4 | Implementar `initVecTable()` condicional no IndexStore | -- |
| 5 | Implementar `HybridSearch()` com combinação FTS5 + vec | 2, 3, 4 |
| 6 | Implementar `cleanupTransientFiles()` | -- |
| 7 | Integrar cleanup no `spw finalizar` | 6 |
| 8 | Adicionar seção `[store]` ao config | -- |
| 9 | Atualizar `config.go` com `StoreConfig` struct | 8 |
| 10 | Testes: FTS5 search, embedding fail-open, cleanup | 1-9 |
