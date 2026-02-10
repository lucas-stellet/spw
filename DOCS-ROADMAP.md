# Documentation Roadmap

Roadmap para preparar a documentacao do SPW para publicacao no GitHub.
Baseado na auditoria completa realizada em 2026-02-10.

---

## Phase 1 — Correcoes e limpeza

Corrigir problemas concretos encontrados na auditoria.

### Paths desatualizados

- [x] `README.md` linhas ~285-300: substituir `_agent-comms/qa/` por paths fase-based (`qa/_comms/`)
- [x] `CLAUDE.md` linha ~95: atualizar referencia generica a `_agent-comms/`
- [x] `copy-ready/README.md`: sincronizar mirror apos correcao do README
- [x] `workflows/spw/status.md` linha ~75: atualizar referencia legada
- [x] `copy-ready/.claude/workflows/spw/status.md`: sincronizar mirror

### Exemplo com arquivo inexistente

- [x] `CLAUDE.md` smoke test: trocar `docs/DESIGN-RESEARCH.md` por path valido
- [x] `AGENTS.md` smoke test: mesma correcao

### Arquivos historicos

- [x] Deletar `MIGRATION-TODO.md` (migracaoo ja concluida, info no git history)
- [x] Deletar `docs/REFACTOR-PLAN.md` (plano ja executado, info no git history)

### Validacao Phase 1

- [x] Rodar `scripts/validate-thin-orchestrator.sh` — mirrors em sync
- [x] Grep por `_agent-comms` em todos os `.md` — zero resultados (exceto contexto de migracao no status.md se mantido)

---

## Phase 2 — README para publicacao

Tornar o README acessivel para quem nunca viu o projeto.

### Estrutura

- [x] Adicionar Table of Contents no topo do README
- [x] Adicionar secao "What is SPW?" com explicacao em 3-4 frases para leigos
- [x] Adicionar secao "Quick Start" com exemplo de fluxo real (prd → plan → exec)
- [x] Revisar secao de instalacao — garantir que funciona para usuario externo
- [x] Adicionar badges (version, license) no topo

### Glossario

- [x] Adicionar secao "Glossary" no README ou doc separado com termos-chave:
  - wave, thin-dispatch, synthesizer, scout, checkpoint, file-first, dispatch-pattern, rolling-wave, agent teams, overlay

### Sincronizacao

- [x] Sincronizar `copy-ready/README.md` com README atualizado

---

## Phase 3 — Arquivos open-source padrao

Arquivos esperados em todo repositorio publico.

- [ ] Criar `LICENSE` (escolher: MIT, Apache 2.0, ou outra)
- [ ] Criar `CONTRIBUTING.md` basico (como contribuir, pre-requisitos, validacao)
- [ ] Criar `CHANGELOG.md` com historico de versoes (pelo menos v2.0 atual)

---

## Phase 4 — Internacionalizacao

Remover barreiras de idioma para contribuidores externos.

- [ ] Traduzir `AGENTS.md` para ingles (ou criar versao bilingue)
- [ ] Revisar todos os docs para consistencia de idioma (ingles como padrao publico)
- [ ] Avaliar se comentarios no `spw-config.toml` precisam de revisao

---

## Phase 5 — Documentacao avancada (pos-publicacao)

Melhorias de qualidade para adocao mais ampla.

- [ ] Adicionar FAQ (rolling-wave vs all-at-once, TDD defaults, quando usar teams)
- [ ] Adicionar diagrama visual do fluxo de comandos (Mermaid no README)
- [ ] Criar `docs/HOOKS.md` completo (expandir o ponteiro atual em `hooks/README.md`)
- [ ] Adicionar exemplos de output para cada comando SPW
- [ ] Avaliar criacao de site de documentacao (GitHub Pages / mdbook)

---

## Notas

- Sempre rodar `scripts/validate-thin-orchestrator.sh` apos qualquer alteracao em arquivos com mirror
- Manter `docs/SPW-WORKFLOW.md` e `hooks/README.md` como ponteiros leves (nao duplicar conteudo)
- `CLAUDE.md` e `AGENTS.md` servem audiencias diferentes do README — nao fundir
