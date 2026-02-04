---
name: spw:design-draft
description: Cria ou atualiza design.md usando requirements + DESIGN-RESEARCH
argument-hint: "<spec-name>"
---

<objective>
Gerar `.spec-workflow/specs/<spec-name>/design.md` com rastreabilidade forte para requirements.
</objective>

<workflow>
1. Ler:
   - `.spec-workflow/specs/<spec-name>/requirements.md`
   - `.spec-workflow/specs/<spec-name>/DESIGN-RESEARCH.md` (se existir)
   - `.spec-workflow/user-templates/design-template.md` (preferencial)
   - fallback: `.spec-workflow/templates/design-template.md`
2. Preencher design com foco em:
   - mapeamento `REQ-ID -> decisão técnica`
   - arquitetura e boundaries
   - reuso de código existente
   - estratégia de teste (unit, integration, e2e)
3. Salvar em `.spec-workflow/specs/<spec-name>/design.md`.
4. Solicitar aprovação (workflow normal do spec-workflow).
</workflow>

<acceptance_criteria>
- [ ] Existe matriz de rastreabilidade de requisitos.
- [ ] Decisões técnicas estão justificadas.
- [ ] Estratégia de testes está explícita.
</acceptance_criteria>
