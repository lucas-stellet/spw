---
name: spw:design-research
description: Pesquisa técnica para gerar design.md com base em requirements aprovados
argument-hint: "<spec-name> [--focus <tema>] [--web-depth low|medium|high]"
---

<objective>
Gerar insumos de arquitetura e implementação para o design da spec.
Saída: `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md`.
</objective>

<preconditions>
- `requirements.md` da spec existe e está aprovado.
- Ler também steering docs se existirem (`product.md`, `tech.md`, `structure.md`).
</preconditions>

<workflow>
1. Ler:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/steering/*.md` (se existir)
2. Fazer análise de codebase:
   - padrões existentes
   - componentes/utilitários reaproveitáveis
   - pontos de integração e riscos
3. Fazer pesquisa externa (web) para bibliotecas/padrões relevantes ao problema.
4. Para projetos Elixir/Phoenix, incluir checagem explícita de:
   - boundaries de contextos (Ecto)
   - padrões de LiveView/Phoenix
   - necessidade real de processos (OTP)
5. Escrever `DESIGN-RESEARCH.md` com:
   - recomendações principais
   - alternativas e trade-offs
   - referências/padrões que serão adotados
   - riscos técnicos e mitigação
</workflow>

<acceptance_criteria>
- [ ] Todo requisito funcional relevante tem ao menos uma recomendação técnica associada.
- [ ] Há seção de reuso de código existente.
- [ ] Há seção de riscos e decisões recomendadas.
</acceptance_criteria>
