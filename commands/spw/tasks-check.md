---
name: spw:tasks-check
description: Valida qualidade do tasks.md (rastreabilidade, dependências, testes)
argument-hint: "<spec-name>"
---

<objective>
Validar se o `tasks.md` está pronto para execução com subagentes.
</objective>

<checks>
1. Rastreabilidade:
   - toda tarefa referencia `Requirements`
   - todo requirement tem pelo menos uma tarefa
2. Dependências:
   - sem ciclos
   - ordem de wave compatível com `Depends On`
3. Paralelismo:
   - tarefas da mesma wave não conflitam em arquivos críticos
4. Testes:
   - toda tarefa tem `Test Plan` + `Verification Command`
   - exceções têm justificativa explícita
5. Definição de pronto:
   - critérios objetivos por tarefa
</checks>

<output>
Gerar `.spec-workflow/specs/<spec-name>/TASKS-CHECK.md` contendo:
- PASS/BLOCKED
- achados por severidade
- ajustes recomendados no tasks.md
</output>
