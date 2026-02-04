---
name: spw:plan
description: Planejamento tecnico a partir de requirements existentes, com gate de aprovacao MCP
argument-hint: "<spec-name> [--max-wave-size 3]"
---

<objective>
Executar o fluxo completo de planejamento tecnico para uma spec que JA tem `requirements.md` existente.
Antes de seguir, este comando valida e, se preciso, solicita aprovacao via MCP.
Este comando NAO cria PRD nem faz descoberta inicial de produto.
</objective>

<when_to_use>
- Use quando a spec ja tem contexto de produto definido e `requirements.md` existente.
- Entrada esperada: `.spec-workflow/specs/<spec-name>/requirements.md`.
</when_to_use>

<preconditions>
- `requirements.md` existe para `<spec-name>`.
- Se nao existir, parar com BLOCKED e orientar: `rode /spw:prd <spec-name>`.
- Nao assumir aprovacao por existencia de arquivo; validar aprovacao via MCP.
</preconditions>

<pipeline>
0. Validar existencia de `.spec-workflow/specs/<spec-name>/requirements.md`.
0.1 Validar status via MCP `spec-status`:
    - checar `documents.requirements.approved`
0.2 Se nao aprovado:
    - solicitar aprovacao via MCP com `request-approval` para `docType: "requirements"`
    - informar usuario para revisar no dashboard/UI
    - acompanhar com `get-approval-status`
    - somente continuar quando status = `approved`
    - se status = `rejected` ou `changes-requested`, parar com BLOCKED
0.3 Se ja aprovado:
    - continuar pipeline normalmente
1. `spw:design-research <spec-name>`
2. `spw:design-draft <spec-name>`
3. `spw:tasks-plan <spec-name> --max-wave-size <N>`
4. `spw:tasks-check <spec-name>`
</pipeline>

<rules>
- Se `tasks-check` retornar BLOCKED, corrigir `tasks.md` e reexecutar check.
- Nao iniciar execucao de codigo sem design e tasks aprovados.
- Nao tentar "adivinhar requisitos" neste comando; requisitos vem do PRD/requirements com aprovacao MCP.
- Gate obrigatorio: requirements sem aprovacao MCP bloqueia o `spw:plan`.
</rules>
