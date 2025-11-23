# Melhoria 08: Metas e Planejamento Financeiro

**Prioridade:** P2 (Média)
**Complexidade:** Média
**Impacto:** Alto

---

## 📋 Visão Geral

Ferramentas para definir metas financeiras, rastrear aportes, sugerir rebalanceamento e simular cenários futuros.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Investimento sem direção**
   - Não sabe se está no caminho certo
   - Falta de objetivos claros
   - Impossível medir progresso

2. **Decisões de alocação aleatórias**
   - "Onde devo investir próximo aporte?"
   - Carteira desbalanceada sem perceber
   - Compra de ativos por impulso

3. **Sem visão de longo prazo**
   - Não sabe quanto tempo até independência financeira
   - Impossível planejar aposentadoria
   - Falta de perspectiva

### Benefícios:

- 🎯 **Clareza:** Saber exatamente onde quer chegar
- 📊 **Acompanhamento:** Medir progresso objetivamente
- 🧭 **Direcionamento:** Decisões baseadas em estratégia
- 💡 **Motivação:** Ver que está próximo da meta

---

## 🏗️ Implementação

### Estruturas principais:

```go
type Goal struct {
    ID              string
    Name            string
    Type            GoalType         // "wealth", "income", "fire"
    TargetAmount    decimal.Decimal  // Meta de patrimônio ou renda
    TargetDate      time.Time        // Quando quer atingir
    CurrentProgress decimal.Decimal  // Progresso atual (%)
    MonthlyTarget   decimal.Decimal  // Quanto precisa aportar/mês
}

type GoalType string

const (
    GoalWealth  GoalType = "wealth"  // Meta de patrimônio
    GoalIncome  GoalType = "income"  // Meta de renda passiva
    GoalFIRE    GoalType = "fire"    // Independência financeira
)

type Contribution struct {
    ID          string
    Date        time.Time
    Amount      decimal.Decimal
    Type        string  // "regular", "extra"
    Description string
}

type RebalanceStrategy struct {
    TargetAllocations map[string]decimal.Decimal  // Ticker → % desejado
    CurrentState      map[string]decimal.Decimal  // Ticker → % atual
    Suggestions       []RebalanceSuggestion
}

type RebalanceSuggestion struct {
    Action      string          // "buy", "sell", "hold"
    Ticker      string
    Quantity    int
    Amount      decimal.Decimal
    Reason      string
}
```

### Funcionalidades-chave:

```go
// CreateGoal cria uma meta financeira
func (w *Wallet) CreateGoal(name string, goalType GoalType, target decimal.Decimal, targetDate time.Time) (*Goal, error)

// TrackProgress calcula progresso em direção à meta
func (w *Wallet) TrackProgress(goalID string) decimal.Decimal

// AddContribution registra aporte
func (w *Wallet) AddContribution(amount decimal.Decimal, date time.Time) (*Contribution, error)

// CalculateMonthlyTarget calcula quanto precisa aportar/mês
func (w *Wallet) CalculateMonthlyTarget(goal *Goal) decimal.Decimal

// SimulateFuture simula evolução da carteira
func (w *Wallet) SimulateFuture(years int, monthlyContribution decimal.Decimal, avgReturn decimal.Decimal) *Projection

// SuggestRebalance sugere rebalanceamento
func (w *Wallet) SuggestRebalance(strategy *RebalanceStrategy) []RebalanceSuggestion
```

---

## 🎨 Comandos CLI

```bash
# Criar meta
b3cli goals set --target=1000000 --years=10
b3cli goals set --type=income --target=5000/month

# Ver progresso
b3cli goals track
b3cli goals list

# Registrar aporte
b3cli contributions add 5000 --date=2024-11-23
b3cli contributions history

# Rebalanceamento
b3cli rebalance --target-file=strategy.yaml
b3cli rebalance simulate

# Simulações
b3cli simulate --years=10 --monthly=5000 --return=10%
b3cli simulate fire --expenses=8000
```

---

## 📊 TUI - Acompanhamento de Meta:

```
╔══════════════════════════════════════════════════════════════════════╗
║              🎯 METAS FINANCEIRAS                                    ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Meta: Independência Financeira                                      ║
║  Tipo: Patrimônio de R$ 1.000.000                                    ║
║  Prazo: Dezembro 2034 (10 anos)                                      ║
║                                                                      ║
║  Progresso                                                           ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  ████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  14.2%                ║
║                                                                      ║
║  Atual:          R$ 142,850                                          ║
║  Faltam:         R$ 857,150                                          ║
║  Tempo restante: 120 meses                                           ║
║                                                                      ║
║  📊 Para atingir a meta:                                             ║
║  Aporte mensal necessário: R$ 4,250                                  ║
║  (assumindo 8% a.a. de retorno)                                      ║
║                                                                      ║
║  Aportes nos últimos 12 meses:                                       ║
║  Total: R$ 48,000 (R$ 4,000/mês em média)                            ║
║                                                                      ║
║  ⚠ Você está R$ 250/mês abaixo da meta!                              ║
║                                                                      ║
║  [E] Editar  [S] Simular  [q] Sair                                  ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 💡 Exemplo de Simulação:

```go
// SimulateFuture simula evolução com aportes mensais
func (w *Wallet) SimulateFuture(years int, monthlyContribution decimal.Decimal, avgReturn decimal.Decimal) *Projection {
    projection := &Projection{
        StartingValue: w.CalculateMarketValue(),
        Years:         years,
        MonthlyPoints: make([]ProjectionPoint, years*12),
    }

    currentValue := projection.StartingValue
    monthlyReturn := avgReturn.Div(decimal.NewFromInt(12)).Div(decimal.NewFromInt(100))

    for month := 0; month < years*12; month++ {
        // Adicionar aporte
        currentValue = currentValue.Add(monthlyContribution)

        // Aplicar rentabilidade
        returns := currentValue.Mul(monthlyReturn)
        currentValue = currentValue.Add(returns)

        projection.MonthlyPoints[month] = ProjectionPoint{
            Month: month,
            Value: currentValue,
        }
    }

    projection.FinalValue = currentValue
    projection.TotalContributed = monthlyContribution.Mul(decimal.NewFromInt(int64(years * 12)))
    projection.TotalReturns = currentValue.Sub(projection.StartingValue).Sub(projection.TotalContributed)

    return projection
}

type Projection struct {
    StartingValue    decimal.Decimal
    FinalValue       decimal.Decimal
    TotalContributed decimal.Decimal
    TotalReturns     decimal.Decimal
    Years            int
    MonthlyPoints    []ProjectionPoint
}

type ProjectionPoint struct {
    Month int
    Value decimal.Decimal
}
```

---

## 🚀 Casos de Uso Reais:

**1. Independência Financeira (FIRE)**
```bash
b3cli goals set fire --monthly-expenses=8000
# Calcula quanto precisa de patrimônio (regra dos 4%)
# Meta: R$ 2.400.000 (8000 × 12 ÷ 0.04)
```

**2. Aposentadoria**
```bash
b3cli goals set wealth --target=5000000 --years=30
b3cli simulate --years=30 --monthly=3000
```

**3. Compra de imóvel**
```bash
b3cli goals set wealth --target=500000 --years=5 --name="Casa própria"
```

---

## 📊 Métricas de Sucesso

- ✅ Usuários com metas claras investem 40% mais
- ✅ Acompanhamento aumenta consistência de aportes
- ✅ Rebalanceamento melhora performance em 15-20%
- ✅ Simulações reduzem ansiedade e melhoram decisões

---

**Estimativa de implementação:** 1-2 semanas
**ROI para usuários:** Muito Alto (direcionamento estratégico)
