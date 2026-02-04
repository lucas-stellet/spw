---
name: spw:tasks-plan
description: Cria tasks.md orientado a waves, paralelismo e TDD por tarefa
argument-hint: "<spec-name> [--max-wave-size 3] [--allow-no-test-exception true|false]"
---

<objective>
Gerar `.spec-workflow/specs/<spec-name>/tasks.md` com execução previsível e paralela.
</objective>

<rules>
- Cada tarefa deve ser autocontida.
- Cada tarefa deve ter teste e comando de verificação.
- Exceção de teste só com justificativa explícita.
- Planejar dependências para maximizar paralelismo por wave.
</rules>

<workflow>
1. Ler:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/design.md`
   - `.spec-workflow/user-templates/tasks-template.md` (preferencial)
   - fallback: `.spec-workflow/templates/tasks-template.md`
2. Construir DAG de dependências entre tarefas.
3. Atribuir `Wave: N` respeitando `max-wave-size`.
4. Para cada tarefa, preencher:
   - `Depends On`
   - `Files`
   - `Test Plan`
   - `Verification Command`
   - `Requirements`
   - `No-Test Justification` (quando necessário)
5. Salvar em `.spec-workflow/specs/<spec-name>/tasks.md`.
6. Solicitar aprovação.
</workflow>

<acceptance_criteria>
- [ ] Todas as tarefas possuem `Requirements`.
- [ ] Todas as tarefas possuem teste ou exceção documentada.
- [ ] Waves respeitam o limite configurado.
</acceptance_criteria>
