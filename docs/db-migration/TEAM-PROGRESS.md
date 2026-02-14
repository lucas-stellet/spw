# SQLite Migration — Team Progress

## Status Geral

| Fase | Status | Agente | Início | Fim |
|------|--------|--------|--------|-----|
| Análise da Codebase | pending | analyzer | - | - |
| Planejamento de Tarefas | pending | planner | - | - |
| Implementação | pending | impl-* | - | - |
| Validação | pending | validator-* | - | - |

---

## Fase 1: Análise da Codebase

**Agente:** `analyzer`
**Objetivo:** Mapear todos os arquivos existentes que serão criados ou modificados, documentar interfaces, tipos, e padrões usados no código atual.

### Checklist de Análise

- [ ] Mapear `cli/internal/tools/dispatch_handoff.go` — pontos de integração dual-write
- [ ] Mapear `cli/internal/tools/dispatch_init.go` — pontos de integração dual-write
- [ ] Mapear `cli/internal/tasks/mark.go` — ponto de sync DB
- [ ] Mapear `cli/internal/specdir/` — constantes e paths existentes
- [ ] Mapear `cli/internal/spec/stage.go` — ClassifyStage para DB-first
- [ ] Mapear `cli/internal/wave/` — scanner, checkpoint, summary, resume
- [ ] Mapear `cli/internal/tools/runs.go` — RunsLatestUnfinished
- [ ] Mapear `cli/internal/tools/dispatch_status.go` — DispatchReadStatus
- [ ] Mapear `cli/internal/hook/statusline.go` — detectActiveSpec
- [ ] Mapear `cli/internal/hook/guard_stop.go` — checkRunCompleteness
- [ ] Mapear `cli/internal/cli/root.go` — registro de comandos
- [ ] Mapear `cli/internal/config/config.go` — estrutura de config para [store]
- [ ] Documentar tipos existentes em `cli/internal/tasks/types.go`
- [ ] Documentar padrões de erro e retorno usados no projeto

### Output
Resultado será salvo em `docs/db-migration/CODEBASE-ANALYSIS.md`

---

## Fase 2: Planejamento de Tarefas

**Agente:** `planner`
**Objetivo:** Com base na análise, criar tarefas granulares e independentes para os implementadores.

### Output
Tarefas serão criadas no TaskList do time e documentadas abaixo.

---

## Fase 3: Implementação

Tarefas independentes serão atribuídas a agentes implementadores em paralelo.

| Task ID | Descrição | Agente | Status |
|---------|-----------|--------|--------|
| (será preenchido pelo planner) | | | |

---

## Fase 4: Validação

Cada implementação será validada por um agente validador.

| Task Validada | Agente Validador | Resultado | Notas |
|---------------|------------------|-----------|-------|
| (será preenchido após implementação) | | | |
