# Roadmap - B3 At Your Terminal

> Plano de desenvolvimento de novas funcionalidades baseadas no ecossistema da B3 e nas necessidades dos investidores brasileiros.

---

## 🎯 Legenda de Prioridade

- **P0** - Crítico: Funcionalidades essenciais para a maioria dos usuários
- **P1** - Alta: Funcionalidades muito úteis que agregam valor significativo
- **P2** - Média: Melhorias incrementais e features de nicho
- **P3** - Baixa: Nice-to-have, funcionalidades experimentais

---

## 📊 1. Declaração de Imposto de Renda (IRPF)

**Prioridade: P0**

A funcionalidade mais solicitada por investidores brasileiros. Cálculo automático de impostos e geração de relatórios para declaração anual.

### Features

- [ ] **Cálculo de ganho de capital** em vendas de ativos
  - Identificar vendas acima de R$ 20.000/mês (ações) - obrigatório IR
  - Identificar vendas de FII (sempre tributado)
  - Calcular prejuízos acumulados para compensação
  - Distinção entre day trade (20%) e swing trade (15%)

- [ ] **Geração de DARF** (documento de arrecadação)
  - Calcular valor de IR devido por mês
  - Gerar código de barras para pagamento
  - Alertas de vencimento (último dia útil do mês seguinte)

- [ ] **Relatório anual para IRPF**
  - Gerar dados para ficha "Bens e Direitos"
  - Calcular lucros/prejuízos por ano fiscal
  - Exportar para formatos compatíveis (txt, csv, pdf)
  - Separar por tipo de ativo (ações, FII, BDR)

- [ ] **Isenções fiscais**
  - Detectar vendas isentas (< R$ 20k/mês em ações)
  - Rastrear vendas de FII isentas de IR (rendimentos)
  - Alertas de otimização fiscal

### Comandos sugeridos
```bash
b3cli tax calculate 2024           # Calcular impostos do ano
b3cli tax darf 2024-11              # Gerar DARF de novembro/2024
b3cli tax irpf 2024                 # Relatório anual IRPF
b3cli tax capital-gains             # Visão geral de ganhos/prejuízos
```

---

## 📈 2. Análise de Carteira (Portfolio Analytics)

**Prioridade: P0**

Ferramentas para análise profunda da carteira e tomada de decisão.

### Features

- [ ] **Alocação de ativos**
  - Distribuição por tipo (ações, FIIs, BDRs, ETFs)
  - Distribuição por segmento/setor
  - Visualização de concentração
  - Alertas de sobre-exposição

- [ ] **Métricas de performance**
  - ROI total da carteira
  - ROI por ativo
  - Dividend Yield médio (DY)
  - DY projetado anual
  - Comparação com benchmarks (IBOV, IFIX)

- [ ] **Diversificação**
  - Índice de concentração (HHI - Herfindahl-Hirschman)
  - Número efetivo de ativos
  - Sugestões de rebalanceamento
  - Análise de correlação entre ativos

- [ ] **Valor de mercado vs custo**
  - Integração com cotações atuais
  - Lucro/prejuízo não realizado
  - Variação percentual por ativo
  - Total investido vs valor atual

### Comandos sugeridos
```bash
b3cli portfolio summary             # Resumo geral da carteira
b3cli portfolio allocation          # Análise de alocação
b3cli portfolio performance         # Métricas de performance
b3cli portfolio diversification     # Análise de diversificação
b3cli portfolio rebalance           # Sugestões de rebalanceamento
```

---

## 🏢 3. Eventos Corporativos (Corporate Actions)

**Prioridade: P1**

Suporte completo a eventos corporativos que afetam a quantidade e preço médio dos ativos.

### Features

- [ ] **Desdobramento (Stock Split)**
  - Ajustar quantidade e preço médio proporcionalmente
  - Ex: split 1:2 dobra quantidade, divide preço médio por 2
  - Histórico de desdobramentos por ativo

- [ ] **Grupamento (Reverse Split)**
  - Reduzir quantidade e aumentar preço médio proporcionalmente
  - Ex: grupamento 10:1 divide quantidade por 10, multiplica preço por 10

- [ ] **Bonificação**
  - Adicionar ações/cotas gratuitas
  - Ajustar preço médio (diluição)
  - Rastrear origem das bonificações

- [ ] **Direitos de Subscrição** (melhorar feature existente)
  - Exercício de direitos (compra via subscrição)
  - Venda de direitos de subscrição
  - Rastreamento de sobras

- [ ] **Fusões e Aquisições**
  - Conversão de ativos (empresa A vira empresa B)
  - Proporção de conversão
  - Ajuste de preço médio

- [ ] **Incorporação e Cisão**
  - Transferência de posições entre tickers
  - Histórico completo de transformações

### Comandos sugeridos
```bash
b3cli events split ITSA4 1:2 2024-05-01         # Registrar desdobramento
b3cli events merge ITSA4 2:1 2024-05-01          # Registrar grupamento
b3cli events bonus PETR4 100 2024-05-01          # Registrar bonificação
b3cli events acquisition LAME3 LAME4 1:1         # Registrar conversão
b3cli events history ITSA4                       # Ver histórico de eventos
```

---

## 💰 4. Proventos Avançados (Advanced Earnings)

**Prioridade: P1**

Expandir funcionalidades de proventos além do tracking básico.

### Features

- [ ] **Dividend Yield por ativo**
  - DY nos últimos 12 meses (DY12M)
  - DY projetado baseado em histórico
  - Comparação de DY entre ativos

- [ ] **Calendário de proventos**
  - Data-com e data-ex
  - Data de pagamento
  - Alertas de proventos a receber
  - Projeções baseadas em histórico

- [ ] **Imposto retido na fonte**
  - Rastrear 15% de IR retido em FIIs
  - IR sobre JCP (15%)
  - Separar valor bruto vs líquido

- [ ] **Reinvestimento de proventos**
  - Marcar proventos que foram reinvestidos
  - Vincular provento → nova compra
  - Calcular DRIP (Dividend Reinvestment)

- [ ] **Análise de proventos**
  - Proventos recebidos por ano/mês/trimestre
  - Crescimento de proventos YoY
  - Consistência de pagamentos (scoring)
  - Payout ratio estimado

### Comandos sugeridos
```bash
b3cli earnings calendar                     # Calendário de proventos
b3cli earnings yield MXRF11                 # DY de um ativo específico
b3cli earnings analysis                     # Análise detalhada
b3cli earnings reinvest <earning-id>        # Marcar reinvestimento
b3cli earnings tax-report 2024              # Relatório de IR retido
```

---

## 🔄 5. Importação e Exportação Avançada

**Prioridade: P1**

Facilitar integração com outras ferramentas e brokers.

### Features

- [ ] **Integração CEI (Canal Eletrônico do Investidor)**
  - Login e scraping automatizado (ou API se disponível)
  - Importar todas as movimentações diretamente da B3
  - Sincronização periódica
  - Evitar importação manual de Excel

- [ ] **Suporte multi-corretoras**
  - Parser para formatos de diferentes corretoras:
    - Clear (XP)
    - Rico
    - Inter
    - BTG
    - Nubank
    - Avenue (para BDRs)
  - Detecção automática de formato

- [ ] **Exportação para contabilidade**
  - Formato CNAB para contadores
  - Planilhas padronizadas para IRPF
  - JSON/CSV para integração com outros sistemas

- [ ] **Backup e restore**
  - Exportar wallet completa (transações + metadados)
  - Importar de backup
  - Versionamento de backups
  - Compressão automática

### Comandos sugeridos
```bash
b3cli import cei --user=CPF --password=SENHA       # Importar do CEI
b3cli import broker clear transacoes.xlsx           # Importar de corretora
b3cli export irpf 2024 --format=pdf                 # Exportar para IRPF
b3cli backup create ./backup/2024-11-23.zip         # Criar backup
b3cli backup restore ./backup/2024-11-23.zip        # Restaurar backup
```

---

## 📊 6. Cotações e Dados de Mercado

**Prioridade: P2**

Integração com fontes de dados de mercado para informações em tempo real.

### Features

- [ ] **Integração com APIs de cotações**
  - Yahoo Finance (gratuito)
  - Alpha Vantage (gratuito com limite)
  - Brapi (API brasileira)
  - B3 oficial (se disponível)

- [ ] **Cotações em tempo real**
  - Preço atual de cada ativo
  - Variação diária (%, R$)
  - Atualização sob demanda ou automática

- [ ] **Valor de mercado da carteira**
  - Calcular valor total atual (quantity × preço atual)
  - Lucro/prejuízo não realizado
  - Variação total da carteira (%, R$)

- [ ] **Alertas de preço**
  - Notificar quando ativo atingir preço alvo
  - Alertas de queda acentuada
  - Stop loss sugerido

- [ ] **Dados fundamentalistas básicos**
  - P/VP, P/L, ROE
  - Dividend Yield atual
  - Informações básicas da empresa

### Comandos sugeridos
```bash
b3cli market update                         # Atualizar cotações
b3cli market prices                         # Ver preços atuais
b3cli market portfolio-value                # Valor de mercado da carteira
b3cli market alerts set ITSA4 --price=12.50 # Criar alerta de preço
b3cli market fundamentals PETR4             # Dados fundamentalistas
```

---

## 📱 7. Relatórios e Visualização

**Prioridade: P2**

Melhorar a visualização de dados com gráficos e relatórios profissionais.

### Features

- [ ] **Gráficos no terminal**
  - Evolução do patrimônio ao longo do tempo
  - Pizza de alocação por setor
  - Barras de performance por ativo
  - Biblioteca: termui, asciigraph

- [ ] **Exportação de relatórios PDF**
  - Relatório mensal/anual completo
  - Gráficos profissionais
  - Sumário executivo
  - Biblioteca: gofpdf, go-chart

- [ ] **Dashboard HTML**
  - Página web estática gerada localmente
  - Gráficos interativos (Chart.js)
  - Tabelas ordenáveis
  - Sem necessidade de servidor

- [ ] **Histórico de evolução**
  - Snapshots mensais automáticos
  - Linha do tempo da carteira
  - Comparação entre períodos
  - Taxa de crescimento (CAGR)

### Comandos sugeridos
```bash
b3cli report monthly 2024-11                # Relatório mensal
b3cli report annual 2024 --pdf              # Relatório anual em PDF
b3cli report dashboard --output=./dash      # Gerar dashboard HTML
b3cli report evolution --from=2023-01       # Evolução histórica
```

---

## 🎯 8. Metas e Planejamento Financeiro

**Prioridade: P2**

Ferramentas para ajudar no planejamento de longo prazo.

### Features

- [ ] **Definição de metas**
  - Meta de patrimônio
  - Meta de renda passiva mensal
  - Prazo para atingir metas
  - Acompanhamento de progresso

- [ ] **Aportes e contribuições**
  - Registrar aportes mensais
  - Histórico de contribuições
  - Calcular taxa de poupança
  - Projetar evolução com aportes regulares

- [ ] **Rebalanceamento inteligente**
  - Definir alocação alvo (% por setor/ativo)
  - Comparar alocação atual vs alvo
  - Sugerir compras/vendas para rebalancear
  - Considerar custos de transação

- [ ] **Simulações e projeções**
  - Projetar patrimônio futuro (Monte Carlo)
  - Simular diferentes cenários de aportes
  - Calcular independência financeira (FIRE)
  - Considerar inflação e reinvestimento

### Comandos sugeridos
```bash
b3cli goals set --target=1000000 --years=10     # Definir meta
b3cli goals track                               # Ver progresso
b3cli contributions add 5000 2024-11-23         # Registrar aporte
b3cli rebalance --target-allocation=config.yaml # Sugestão de rebalanceamento
b3cli simulate --monthly-contribution=5000      # Simular evolução
```

---

## 🔧 9. Transações Avançadas

**Prioridade: P2**

Suporte a tipos de transações mais complexas.

### Features

- [ ] **Day Trade**
  - Identificar automaticamente day trades
  - Separar de swing trades
  - Cálculo de IR específico (20%)
  - Tracking de prejuízos em day trade

- [ ] **Opções (Calls e Puts)**
  - Compra/venda de opções
  - Prêmios recebidos/pagos
  - Exercício de opções
  - Vencimento de opções (expiração)
  - Greeks básicos (Delta, Gamma, Theta)

- [ ] **BDRs (Brazilian Depositary Receipts)**
  - Suporte completo a BDRs
  - Conversão de moeda (USD → BRL)
  - Tracking de dividendos em USD
  - Imposto específico (15% sobre ganho de capital)

- [ ] **ETFs**
  - Identificar automaticamente ETFs
  - Composição do ETF (se possível)
  - Tracking específico de custos (taxa de administração)

- [ ] **Renda Fixa**
  - Suporte a Tesouro Direto
  - CDB, LCI, LCA
  - Debêntures
  - Cálculo de rendimentos

### Comandos sugeridos
```bash
b3cli trade daytrade                        # Ver day trades
b3cli options buy call PETR4 --strike=40    # Registrar compra de call
b3cli bdr overview                          # Ver BDRs em carteira
b3cli fixed-income add CDB --amount=10000   # Adicionar renda fixa
```

---

## 🏦 10. Multi-Corretora e Consolidação

**Prioridade: P3**

Gerenciar investimentos em múltiplas corretoras.

### Features

- [ ] **Tracking de corretoras**
  - Identificar corretora de cada transação
  - Posições separadas por corretora
  - Consolidação total

- [ ] **Custos e taxas**
  - Taxa de corretagem por operação
  - Taxa de custódia
  - Emolumentos B3
  - ISS sobre corretagem
  - Comparação de custos entre corretoras

- [ ] **Transferência entre corretoras**
  - Registrar transferência de custódia
  - Manter histórico completo
  - Não considerar como venda/compra

### Comandos sugeridos
```bash
b3cli brokers list                              # Listar corretoras
b3cli brokers positions --broker=clear          # Posições em uma corretora
b3cli brokers fees --year=2024                  # Custos por corretora
b3cli brokers transfer ITSA4 100 clear→rico     # Transferir custódia
```

---

## 🔐 11. Segurança e Privacy

**Prioridade: P3**

Funcionalidades para proteger dados sensíveis.

### Features

- [ ] **Criptografia de wallets**
  - Criptografar wallet.yaml com senha
  - Descriptografar ao abrir
  - AES-256 encryption

- [ ] **Anonimização de dados**
  - Modo "demo" com dados fictícios
  - Ocultar valores em screenshots
  - Exportar sem CPF/dados pessoais

- [ ] **Auditoria de operações**
  - Log de todas as operações
  - Histórico de modificações
  - Rollback de alterações indevidas

### Comandos sugeridos
```bash
b3cli wallet encrypt --password=SENHA       # Criptografar wallet
b3cli wallet decrypt --password=SENHA       # Descriptografar
b3cli privacy anonymize                     # Gerar dados anônimos
```

---

## 🌐 12. Funcionalidades Web/Mobile (Futuro)

**Prioridade: P3**

Expansão para além da CLI.

### Features

- [ ] **API REST local**
  - Servidor HTTP local
  - Endpoints para todas as operações
  - Autenticação JWT
  - Documentação OpenAPI/Swagger

- [ ] **Interface Web**
  - Frontend React/Vue
  - Mesmas funcionalidades da CLI
  - Responsivo (mobile-friendly)
  - Gráficos interativos

- [ ] **App Mobile** (longo prazo)
  - Flutter ou React Native
  - Sincronização com desktop
  - Notificações push
  - Modo offline

---

## 🚀 Roadmap de Implementação Sugerido

### Fase 1: Essencial (Q1 2025)
- ✅ IRPF básico (cálculo de ganho de capital)
- ✅ Eventos corporativos (split, grupamento, bonificação)
- ✅ Portfolio analytics básico (alocação, ROI)

### Fase 2: Crescimento (Q2 2025)
- ✅ Integração CEI
- ✅ DARF e relatórios fiscais
- ✅ Proventos avançados (DY, calendário)
- ✅ Importação multi-corretoras

### Fase 3: Profissional (Q3 2025)
- ✅ Cotações de mercado
- ✅ Day trade e opções
- ✅ Relatórios em PDF
- ✅ Metas e rebalanceamento

### Fase 4: Premium (Q4 2025)
- ✅ BDRs e renda fixa
- ✅ Dashboard HTML
- ✅ API REST local
- ✅ Multi-corretora avançado

---

## 📝 Como Contribuir

Quer ajudar a implementar alguma dessas funcionalidades?

1. Escolha uma feature do roadmap
2. Abra uma issue no GitHub discutindo a implementação
3. Faça um fork e crie uma branch (`feature/nome-da-feature`)
4. Implemente seguindo as regras do projeto (ver CLAUDE.md)
5. Abra um Pull Request

**Priorize features marcadas como P0 e P1 para maior impacto!**

---

## 🎓 Referências

### Regulamentação
- [Instrução CVM 600](http://conteudo.cvm.gov.br/legislacao/instrucoes/inst600.html) - Mercado de Valores Mobiliários
- [Receita Federal - IRPF](https://www.gov.br/receitafederal/pt-br/assuntos/meu-imposto-de-renda)
- [B3 - Regulamentos](https://www.b3.com.br/pt_br/regulacao/)

### APIs e Integrações
- [CEI B3](https://cei.b3.com.br/)
- [Brapi - API Brasileira](https://brapi.dev/)
- [Yahoo Finance API](https://finance.yahoo.com/)
- [Alpha Vantage](https://www.alphavantage.co/)

### Conceitos Financeiros
- [Como declarar ações no IR](https://www.gov.br/receitafederal/pt-br/assuntos/meu-imposto-de-renda/preenchimento/rendimentos-de-aplicacoes-financeiras-e-ganho-de-capital)
- [Imposto sobre Day Trade](https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/tributos/irpf-imposto-de-renda-pessoa-fisica)
- [Como funcionam FIIs](https://www.gov.br/investidor/pt-br/investir/tipos-de-investimentos/fundos-de-investimento-imobiliario)

---

**Última atualização:** 23 de Novembro de 2024

**Status:** 🚧 Documento vivo - será atualizado conforme o projeto evolui
