---
name: spw:prd
description: Planejamento do zero em formato PRD (descoberta guiada) para gerar requirements.md
argument-hint: "<spec-name> [--source <url-ou-arquivo.md>]"
---

<objective>
Gerar ou atualizar `.spec-workflow/specs/<spec-name>/requirements.md` no formato PRD.

Este comando combina:
- GSD: escopo v1/v2/out-of-scope, REQ-ID, criterios testaveis e rastreabilidade.
- superpowers: perguntas uma por vez, recomendacoes com trade-off, validacao incremental por secao.
</objective>

<when_to_use>
- Use quando a spec ainda NAO tem requirements aprovados (planejamento do zero).
- Use tambem quando precisar rediscutir requirements com base em novas fontes de produto.
</when_to_use>

<out_of_scope>
- Este comando nao gera `design.md`.
- Este comando nao gera `tasks.md`.
- Proximo passo apos aprovacao do PRD: `spw:plan <spec-name>`.
</out_of_scope>

<inputs>
- `spec-name` (obrigatorio)
- `--source` (opcional): URL (GitHub/Linear/ClickUp/etc.) ou arquivo markdown.
</inputs>

<source_handling>
Se `--source` for informado e parecer URL (`http://` ou `https://`) ou markdown (`.md`), executar gate de leitura:

1. Perguntar com AskUserQuestion:
   - header: "Fonte"
   - question: "Detectei uma fonte externa. Quer usar um MCP especifico para ler essa fonte?"
   - options:
     - "Sim, escolher MCP (Recomendado)" — Escolher explicitamente o conector
     - "Auto" — Tentar MCP compativel, com fallback para leitura direta
     - "Nao" — Ler sem MCP

2. Se escolher "Sim, escolher MCP", perguntar:
   - header: "MCP"
   - question: "Qual MCP deseja usar para esta fonte?"
   - options:
     - "GitHub" — Issues/PRs/repositorios
     - "Linear" — Issues/projetos do Linear
     - "ClickUp" — Tasks/listas do ClickUp
     - "Web/Browser" — Leitura via web fetch
     - "Arquivo local markdown" — Leitura direta de arquivo local

3. Se MCP escolhido nao estiver disponivel no ambiente, avisar claramente e perguntar fallback:
   - "Ler sem MCP"
   - "Escolher outro MCP"
</source_handling>

<workflow>
1. Ler contexto existente:
   - `.spec-workflow/specs/<spec-name>/requirements.md` (se existir)
   - `.spec-workflow/specs/<spec-name>/design.md` (se existir)
   - `.spec-workflow/steering/*.md` (se existir)
2. Se houver `--source`, processar fonte com o gate de MCP acima.
3. Conduzir descoberta com perguntas uma por vez (nao despejar formulario):
   - problema, publico e contexto de uso
   - objetivo principal e sucesso esperado
   - escopo v1, v2 e fora de escopo
   - restricoes e dependencias
   - riscos e questoes abertas
4. Sempre que houver ambiguidade, propor 2-3 opcoes com recomendacao explicita.
5. Escrever PRD em blocos (200-300 palavras por secao) e validar secao a secao.
6. Preencher template PRD com prioridade de busca:
   - `.spec-workflow/user-templates/prd-template.md` (preferencial)
   - fallback: `.spec-workflow/templates/prd-template.md`
   - fallback final: estrutura embutida deste comando
7. Salvar artefatos:
   - Canonico: `.spec-workflow/specs/<spec-name>/requirements.md`
   - Espelho para produto: `.spec-workflow/specs/<spec-name>/PRD.md`
8. Confirmar readiness para design:
   - requisitos funcionais com REQ-ID e criterios de aceitacao
   - NFRs e metricas
   - secoes de v2 e out-of-scope
</workflow>

<acceptance_criteria>
- [ ] Documento final esta em formato PRD, mas compativel com o fluxo de requirements do spec-workflow.
- [ ] Cada requisito funcional tem REQ-ID, prioridade e criterio de aceitacao verificavel.
- [ ] Existe separacao explicita entre v1, v2 e out-of-scope.
- [ ] Se `--source` foi usado, houve pergunta explicita sobre uso de MCP.
- [ ] PRD aprovado pelo usuario antes de avancar para design/tasks.
</acceptance_criteria>
