---
name: spw:checkpoint
description: Gate de qualidade entre batches/waves de execução
argument-hint: "<spec-name> [--scope batch|wave|phase]"
---

<objective>
Validar que o lote executado realmente atende spec, qualidade e integração antes de avançar.
</objective>

<checks>
1. Estado de tarefas (`tasks.md`): coerência `[ ]/[-]/[x]`.
2. Testes/lint/typecheck do projeto.
3. Review de conformidade com spec (requirements + design + task).
4. Review de qualidade de código.
5. Rastreabilidade: mudanças implementadas vinculadas a `Requirements`.
</checks>

<output>
Gerar `.spec-workflow/specs/<spec-name>/CHECKPOINT-REPORT.md` com:
- status: PASS | BLOCKED
- problemas críticos
- ações corretivas
- próximo passo recomendado
</output>

<gate_rule>
Se status for BLOCKED, não avançar para próximo batch/wave.
</gate_rule>
