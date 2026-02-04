---
name: spw:exec
description: Executa tasks.md em batches com checkpoints obrigatórios
argument-hint: "<spec-name> [--batch-size 3] [--strict true|false]"
---

<objective>
Executar tarefas em lotes, com pausa obrigatória para checkpoint de qualidade.
</objective>

<workflow>
1. Ler `tasks.md` e selecionar tarefas pendentes por wave.
2. Executar até `batch-size` tarefas por lote (preferir paralelismo seguro).
3. Para cada tarefa:
   - marcar `[-]`
   - executar com subagente
   - validar spec compliance + code quality
   - registrar log de implementação
   - marcar `[x]`
4. Ao fim do lote, executar `spw:checkpoint <spec-name>`.
5. Só continuar quando checkpoint passar.
</workflow>

<strict_mode>
Com `--strict true` (padrão):
- bloqueia continuidade se checkpoint falhar.
- bloqueia continuidade se houver tarefa sem rastreabilidade de requirement.
</strict_mode>
