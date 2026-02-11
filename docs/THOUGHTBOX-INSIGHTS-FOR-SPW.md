# Thoughtbox -> SPW: O que aproveitar em funcionalidade, design e filosofia

## Objetivo deste documento

Este documento descreve, de forma detalhada, o que o projeto Thoughtbox oferece hoje e como essas ideias podem ser adaptadas para o SPW sem quebrar os principios atuais do kit (thin orchestrator, subagent-first, aprovacoes MCP, handoff por arquivo e checkpoints por wave).

A intencao nao e copiar o Thoughtbox como produto, e sim extrair os padroes que aumentam:

- visibilidade operacional do fluxo
- auditabilidade de decisao
- confiabilidade de execucao
- qualidade do output dos agentes
- melhoria continua baseada em evidencia

---

## Resumo executivo

O SPW ja e forte em orquestracao, guardrails e estrutura de artefatos por fase. O Thoughtbox complementa isso com tres forcas principais que o SPW ainda pode explorar melhor:

1. Observabilidade de processo em tempo real
2. Contratos comportamentais para validar que agentes estao realmente raciocinando
3. Disciplina explicita de qualidade anti-slop e melhoria continua com historico de execucoes

Se o SPW incorporar esses tres pilares, ele deixa de ser apenas um "motor de fluxo" e passa a ser tambem um "sistema observavel e evolutivo" para engenharia assistida por agentes.

---

## O que o Thoughtbox traz (e por que isso importa)

## 1) Ledger auditavel de raciocinio e colaboracao

No Thoughtbox, o foco nao e so gerar resultado final; o caminho de decisao e modelado como dado de primeira classe.

O que existe la:

- trilha persistente de pensamento e decisoes
- relacionamento explicito entre problema, proposta, revisao e consenso
- historico navegavel por sessao

Por que isso importa para o SPW:

- hoje o SPW guarda briefs/reports/status por subagente, mas nao tem uma camada unica de "timeline executiva" do processo
- quando algo falha, o debug tende a ser artesanal (abrir varios arquivos e reconstruir historia)
- com um ledger de eventos e decisoes, incidentes de fluxo ficam mais rapidos de entender e corrigir

## 2) Observabilidade em tempo real (nao so artefato final)

Thoughtbox tem uma visao de observabilidade onde o processo e monitorado enquanto acontece (eventos, estado, metricas, UI dedicada).

O que existe la:

- stream de eventos em JSONL
- emissao non-blocking (falha de observabilidade nao derruba execucao)
- camada de visualizacao operacional

Por que isso importa para o SPW:

- o SPW hoje e "forte em estrutura" e "fraco em telemetria de runtime"
- nao existe uma visao consolidada de progresso por comando/fase/wave em tempo real
- para times, uma visao operacional compartilhada reduz custo de coordenacao

## 3) Progressive disclosure de capacidades

Thoughtbox aplica acesso por estagio, liberando capacidades conforme contexto e maturidade da sessao.

O que existe la:

- gates de estagio explicitos
- restricao operacional por fase da sessao
- transicoes de estagio controladas e auditaveis

Por que isso importa para o SPW:

- SPW ja tem gates fortes (aproval, checkpoint, exec policy), mas pode evoluir para "policy por estagio de comando"
- isso reduz uso incorreto de comandos e aumenta seguranca operacional

## 4) Contratos comportamentais para agentes

Thoughtbox usa verificacoes para evitar falso positivo de "funciona" quando o agente so retornou algo plausivel, mas sem raciocinio real.

O que existe la:

- testes de variancia (entradas diferentes -> saidas diferentes)
- acoplamento ao conteudo de entrada
- verificacao de existencia de trilha de raciocinio
- camadas de julgamento semantico

Por que isso importa para o SPW:

- SPW valida muito bem formato e fluxo, mas pode validar melhor "substancia"
- melhora qualidade de PRD/design/tasks/qa-check em cenarios ambiguos
- reduz output generico que passa em checklist mas falha na pratica

## 5) Anti-slop explicito e mensuravel

Thoughtbox trata qualidade textual e evidencial como regra operacional, nao so boa pratica.

O que existe la:

- rejeicao de afirmacoes numericas sem fonte
- exigencia de evidencias clicaveis e completas
- diferenciacao clara entre "estimado" e "medido"

Por que isso importa para o SPW:

- artefatos de especificacao podem parecer robustos, mas conter claims sem base
- anti-slop no pipeline evita ruido cedo e melhora confianca no processo

## 6) Historico append-only para melhoria continua

Thoughtbox registra ciclos em historico estruturado para aprender com iteracoes anteriores.

O que existe la:

- logs historicos por run
- custo, resultado e artefatos por ciclo
- consultas simples para tendencia e eficiencia

Por que isso importa para o SPW:

- SPW tem post-mortem por spec, mas nao tem scorecard acumulado de runtime do sistema
- sem serie historica, fica dificil otimizar estrategia de waves e subagentes

---

## O que ja existe no SPW e nao deve ser perdido

Qualquer adicao deve preservar os pilares atuais:

- thin dispatch (orchestrator le apenas status.json no fluxo feliz)
- file-first handoff
- estrutura por fase (`prd/`, `design/`, `planning/`, `execution/`, `qa/`, `post-mortem/`)
- aprovacao MCP como fonte unica de verdade
- sem avance automatico entre waves quando policy exige aprovacao do usuario

Em outras palavras: a evolucao deve ser incremental e compativel, nao uma troca de arquitetura.

---

## Propostas concretas para trazer ao SPW

## Proposta A - Event Ledger do workflow (prioridade P0)

### O que e

Criar um log append-only por spec para eventos de runtime do SPW, por exemplo:

`/.spec-workflow/specs/<spec-name>/execution/events.jsonl`

### Evento minimo sugerido

- `run_started`
- `run_resumed`
- `subagent_dispatched`
- `subagent_pass`
- `subagent_blocked`
- `wave_closed`
- `checkpoint_pass`
- `checkpoint_blocked`
- `approval_wait`
- `user_continue_unfinished`
- `user_delete_and_restart`
- `command_completed`

### Beneficios

- auditoria operacional clara
- debug mais rapido de bloqueios
- base para metricas de melhoria continua

### Riscos

- excesso de volume de log
- acoplamento acidental entre eventos e logica de decisao

### Mitigacao

- schema estavel e simples
- emissao non-blocking
- logs como observabilidade, nunca como source of truth de aprovacao

---

## Proposta B - SPW Observatory (prioridade P1)

### O que e

Uma visao operacional de progresso, consumindo apenas artefatos ja existentes + event ledger.

### Escopo inicial (MVP)

- status por fase/comando
- progresso por wave
- ultimo subagente executado
- bloqueios abertos
- proxima acao recomendada

### Formato possivel

- primeiro TUI/CLI (`spw observability`)
- depois UI web leve se necessario

### Beneficios

- reduz tempo de orientacao para usuarios e lideres
- melhora colaboracao em Agent Teams
- facilita revisao de execucao sem abrir dezenas de arquivos

---

## Proposta C - Behavioral Contracts para comandos SPW (prioridade P1)

### O que e

Adicionar uma camada de testes de comportamento do proprio fluxo do SPW, alem dos testes estruturais.

### Exemplos de contrato

- `exec` nunca avanca wave sem aprovacao quando `require_user_approval_between_waves=true`
- orchestrator nao le `report.md` em caminho feliz
- `qa-exec` nao le codigo-fonte
- `tasks-plan` respeita precedencia `--mode` > config

### Entrega recomendada

- `docs/BEHAVIORAL-CONTRACTS.md` com matriz de contratos
- suite automatizada no CLI para validar contratos criticos

### Beneficios

- previne regressao silenciosa de guardrails
- aumenta confianca em mudancas futuras no kit

---

## Proposta D - Anti-slop enforcement nos artefatos (prioridade P1)

### O que e

Aplicar regras objetivas de qualidade em PRD/design/tasks/qa-check/post-mortem.

### Regras iniciais

- claim numerico exige fonte ou marcador de estimativa
- evidencias devem ser URL completa quando aplicavel
- separacao explicita entre dado medido e inferencia
- linguagem de confianca do tipo: `evidencia`, `hipotese`, `estimativa`

### Beneficios

- melhora qualidade editorial e tecnica
- reduz retrabalho por interpretacao ambigua

---

## Proposta E - Scorecard historico de execucao (prioridade P2)

### O que e

Um historico append-only por run para medir eficiencia e qualidade do fluxo ao longo do tempo.

### Metricas sugeridas

- tempo por fase e por wave
- taxa de bloqueio por comando
- percentual de rerun por causa de incompletude
- taxa de drift de seletor em QA
- custo estimado por comando/modelo (quando aplicavel)

### Beneficios

- orienta tuning de `max_wave_size`
- orienta tuning de modelos por tipo de subagente
- transforma post-mortem em melhoria sistematica

---

## Proposta F - Catalogo de operacoes e contrato de comando (prioridade P2)

### O que e

Documento e schema declarando para cada comando:

- pre-condicoes
- entradas
- saidas esperadas
- gates obrigatorios
- invariantes que nunca podem quebrar

### Beneficios

- onboarding mais rapido
- menos ambiguidades em evolucao de comando
- melhor base para testes gerados automaticamente

---

## Diretriz de design para adocao no SPW

Para manter compatibilidade com a arquitetura atual:

1. Tudo novo deve ser observabilidade-first e side-effect-safe.
2. Nenhum novo modulo deve substituir aprovacao MCP como fonte de verdade.
3. O orchestrator continua thin: status-only no caminho feliz.
4. Visualizacao e scorecards devem ler arquivos, nao reescrever estado do fluxo.
5. Novos gates devem ser graduais (warn -> block), igual ao modelo de hooks.

---

## Roadmap recomendado

## Fase 1 (1-2 semanas) - Base de observabilidade

- Implementar `events.jsonl` por spec
- Adicionar utilitario para leitura de eventos
- Definir schema de evento versao 1

Resultado esperado:

- timeline operacional auditavel
- base pronta para UI/metricas

## Fase 2 (2-3 semanas) - Qualidade e contratos

- Introduzir anti-slop linter em comandos de sintese
- Criar matriz de behavioral contracts do SPW
- Cobrir contratos mais criticos em testes

Resultado esperado:

- maior robustez de output e de guardrails

## Fase 3 (2-4 semanas) - Observatory e scorecard

- Entregar visao consolidada de fase/wave/bloqueios
- Gerar scorecard historico por spec
- Publicar guia de operacao para times

Resultado esperado:

- operacao previsivel, mensuravel e otimizavel

---

## Criterios de sucesso (objetivos)

- reducao no tempo medio para diagnosticar bloqueio de run
- reducao de reruns por erro de fluxo
- aumento de completude de evidencia em artefatos
- reducao de saidas genericas em design/tasks/qa-check
- melhoria de throughput por wave sem perda de qualidade

---

## Riscos de implementacao e como evitar

## Risco 1: Complexidade excessiva

Sintoma:

- SPW vira um produto de observabilidade em vez de kit de workflow

Prevencao:

- priorizar MVPs pequenos
- cada feature nova precisa mostrar ganho em tempo/qualidade

## Risco 2: Telemetria invadir logica de execucao

Sintoma:

- decisoes dependerem do ledger de evento

Prevencao:

- evento so observa
- decisao continua em artefatos canonicos + MCP

## Risco 3: Sobrecarga de manutencao

Sintoma:

- custo de manter docs e regras cresce muito

Prevencao:

- schema e contratos curtos
- automacao de validacao no CI

---

## Conclusao

O Thoughtbox nao precisa ser replicado no SPW para gerar valor.

A melhor estrategia e absorver seus padroes mais fortes:

- observabilidade operacional
- contratos comportamentais
- anti-slop e melhoria continua baseada em historico

Com isso, o SPW mantem sua identidade (thin orchestrator + gates MCP + file-first) e ganha uma camada que hoje falta: visibilidade e evolucao sistematica do proprio processo.

