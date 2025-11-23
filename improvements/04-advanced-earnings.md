# Melhoria 04: Proventos Avançados (Advanced Earnings)

**Prioridade:** P1 (Alta)
**Complexidade:** Média
**Impacto:** Alto

---

## 📋 Visão Geral

Expandir as funcionalidades de proventos além do tracking básico: calendários, Dividend Yield detalhado, IR retido, reinvestimento, e análises preditivas.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Falta de previsibilidade**
   - "Quando vou receber o próximo provento?"
   - "Quanto vou receber de dividendos este mês?"
   - Impossível planejar renda passiva

2. **Análise superficial de DY**
   - DY é calculado apenas no total
   - Não sabe qual ativo tem melhor yield
   - Difícil comparar performance de FIIs

3. **Controle de IR deficiente**
   - FIIs têm 15% retido na fonte
   - Usuário não sabe quanto foi retido
   - Complicado para declaração de IRPF

4. **Gestão de reinvestimento**
   - Perdeu a conta de quais proventos foram reinvestidos
   - Difícil calcular efeito de DRIP (Dividend Reinvestment Plan)

### Benefícios mensuráveis:

- 📅 **Previsibilidade:** Saber exatamente quando e quanto receberá
- 💰 **Otimização:** Identificar ativos com melhor DY
- 📊 **Análise profunda:** Entender consistência de pagamentos
- 🎯 **Meta de renda:** Acompanhar progresso para independência financeira

---

## 🏗️ Arquitetura Proposta

### Componentes principais:

```
internal/earnings/
├── calendar.go          # Calendário de proventos
├── yield.go             # Cálculo de DY detalhado
├── tax.go               # IR retido
├── reinvestment.go      # Tracking de reinvestimento
├── analysis.go          # Análises e previsões
├── consistency.go       # Scoring de consistência
└── projections.go       # Projeções futuras

cmd/b3cli/
├── earnings_calendar.go
├── earnings_yield.go
├── earnings_tax.go
└── earnings_analysis.go
```

---

## 💡 Implementação Técnica

### 1. Calendário de Proventos

**Estruturas:**

```go
type EarningsCalendar struct {
    Upcoming    []UpcomingEarning    // Próximos proventos
    Received    []EarningEvent       // Recebidos este mês
    Projected   []ProjectedEarning   // Projeções baseadas em histórico
}

type UpcomingEarning struct {
    Ticker          string
    Type            string              // Dividendo, JCP, Rendimento
    PaymentDate     time.Time           // Data de pagamento
    ExDate          time.Time           // Data-ex (opcional)
    AmountPerShare  decimal.Decimal     // R$ por ação/cota
    TotalAmount     decimal.Decimal     // Total estimado
    Source          string              // "historical", "announced", "manual"
}

type ProjectedEarning struct {
    Ticker          string
    Month           time.Time
    AverageAmount   decimal.Decimal     // Média histórica
    Confidence      int                 // 0-100%
    BasedOnPayments int                 // Número de pagamentos usados
}
```

**Implementação:**

```go
// GenerateCalendar gera calendário de proventos
func (w *Wallet) GenerateCalendar() *EarningsCalendar {
    calendar := &EarningsCalendar{
        Upcoming:  make([]UpcomingEarning, 0),
        Received:  make([]EarningEvent, 0),
        Projected: make([]ProjectedEarning, 0),
    }

    now := time.Now()
    thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

    // Proventos recebidos este mês
    for _, asset := range w.GetActiveAssets() {
        for _, earning := range asset.Earnings {
            if earning.Date.Year() == now.Year() && earning.Date.Month() == now.Month() {
                calendar.Received = append(calendar.Received, EarningEvent{
                    Ticker:      earning.Ticker,
                    Type:        earning.Type,
                    Date:        earning.Date,
                    Amount:      earning.TotalAmount,
                })
            }
        }
    }

    // Projetar próximos meses baseado em histórico
    for _, asset := range w.GetActiveAssets() {
        // Analisar padrão de pagamentos
        pattern := w.analyzePaymentPattern(asset)

        // Gerar projeções para próximos 3-6 meses
        projections := generateProjections(asset, pattern, 6)
        calendar.Projected = append(calendar.Projected, projections...)
    }

    return calendar
}

// analyzePaymentPattern analisa padrão de pagamentos de um ativo
func (w *Wallet) analyzePaymentPattern(asset *Asset) *PaymentPattern {
    // Agrupar por mês
    monthlyPayments := make(map[int][]decimal.Decimal)

    for _, earning := range asset.Earnings {
        month := int(earning.Date.Month())
        monthlyPayments[month] = append(monthlyPayments[month], earning.TotalAmount)
    }

    // Calcular médias mensais
    pattern := &PaymentPattern{
        Ticker:           asset.ID,
        MonthlyAverages:  make(map[int]decimal.Decimal),
        Frequency:        determineFrequency(monthlyPayments),
        IsConsistent:     isConsistent(monthlyPayments),
    }

    for month, payments := range monthlyPayments {
        sum := decimal.Zero
        for _, p := range payments {
            sum = sum.Add(p)
        }
        avg := sum.Div(decimal.NewFromInt(int64(len(payments))))
        pattern.MonthlyAverages[month] = avg
    }

    return pattern
}

type PaymentPattern struct {
    Ticker          string
    MonthlyAverages map[int]decimal.Decimal  // Média por mês (1-12)
    Frequency       string                    // "monthly", "bimonthly", "quarterly", "irregular"
    IsConsistent    bool                      // Pagamentos consistentes?
}
```

### 2. Dividend Yield Avançado

**Estruturas:**

```go
type YieldAnalysis struct {
    Ticker          string
    DY12M           decimal.Decimal  // DY últimos 12 meses
    DYProjected     decimal.Decimal  // DY projetado anual
    DYAvg3Y         decimal.Decimal  // DY médio 3 anos
    DYGrowth        decimal.Decimal  // Crescimento YoY
    MonthlyBreakdown map[string]decimal.Decimal
    History         []YieldHistoryPoint
}

type YieldHistoryPoint struct {
    Date    time.Time
    DY      decimal.Decimal
    Amount  decimal.Decimal
}
```

**Implementação:**

```go
// CalculateYieldAnalysis calcula DY detalhado de um ativo
func (w *Wallet) CalculateYieldAnalysis(ticker string) (*YieldAnalysis, error) {
    asset, exists := w.Assets[ticker]
    if !exists {
        return nil, fmt.Errorf("ativo %s não encontrado", ticker)
    }

    analysis := &YieldAnalysis{
        Ticker:           ticker,
        MonthlyBreakdown: make(map[string]decimal.Decimal),
        History:          make([]YieldHistoryPoint, 0),
    }

    // DY últimos 12 meses
    oneYearAgo := time.Now().AddDate(-1, 0, 0)
    earningsLast12M := decimal.Zero

    for _, earning := range asset.Earnings {
        if earning.Date.After(oneYearAgo) {
            earningsLast12M = earningsLast12M.Add(earning.TotalAmount)

            // Breakdown mensal
            monthKey := earning.Date.Format("2006-01")
            analysis.MonthlyBreakdown[monthKey] = analysis.MonthlyBreakdown[monthKey].Add(earning.TotalAmount)
        }
    }

    if !asset.TotalInvestedValue.IsZero() {
        analysis.DY12M = earningsLast12M.Div(asset.TotalInvestedValue).Mul(decimal.NewFromInt(100))
    }

    // DY médio 3 anos
    threeYearsAgo := time.Now().AddDate(-3, 0, 0)
    earnings3Y := decimal.Zero
    years := 0

    for year := 0; year < 3; year++ {
        yearStart := time.Now().AddDate(-year-1, 0, 0)
        yearEnd := time.Now().AddDate(-year, 0, 0)
        yearEarnings := decimal.Zero

        for _, earning := range asset.Earnings {
            if earning.Date.After(yearStart) && earning.Date.Before(yearEnd) {
                yearEarnings = yearEarnings.Add(earning.TotalAmount)
            }
        }

        if yearEarnings.GreaterThan(decimal.Zero) {
            earnings3Y = earnings3Y.Add(yearEarnings)
            years++
        }
    }

    if years > 0 && !asset.TotalInvestedValue.IsZero() {
        avgEarnings := earnings3Y.Div(decimal.NewFromInt(int64(years)))
        analysis.DYAvg3Y = avgEarnings.Div(asset.TotalInvestedValue).Mul(decimal.NewFromInt(100))
    }

    // Projetar DY anual baseado em padrão
    pattern := w.analyzePaymentPattern(asset)
    yearlyProjection := decimal.Zero
    for _, avg := range pattern.MonthlyAverages {
        yearlyProjection = yearlyProjection.Add(avg)
    }

    if !asset.TotalInvestedValue.IsZero() {
        analysis.DYProjected = yearlyProjection.Div(asset.TotalInvestedValue).Mul(decimal.NewFromInt(100))
    }

    return analysis, nil
}

// RankByYield ordena ativos por DY
func (w *Wallet) RankByYield() []YieldRanking {
    rankings := make([]YieldRanking, 0)

    for _, asset := range w.GetActiveAssets() {
        dy12m := w.CalculateDY12M(asset)

        rankings = append(rankings, YieldRanking{
            Ticker:          asset.ID,
            DY12M:           dy12m,
            TotalEarnings:   asset.TotalEarnings,
            TotalInvested:   asset.TotalInvestedValue,
        })
    }

    // Ordenar por DY (maior primeiro)
    sort.Slice(rankings, func(i, j int) bool {
        return rankings[i].DY12M.GreaterThan(rankings[j].DY12M)
    })

    return rankings
}

type YieldRanking struct {
    Ticker        string
    DY12M         decimal.Decimal
    TotalEarnings decimal.Decimal
    TotalInvested decimal.Decimal
}
```

### 3. IR Retido na Fonte

**Implementação:**

```go
type TaxWithheld struct {
    Ticker          string
    GrossAmount     decimal.Decimal  // Valor bruto
    WithheldAmount  decimal.Decimal  // IR retido (15%)
    NetAmount       decimal.Decimal  // Valor líquido
    TaxRate         decimal.Decimal  // 15% para FII
    CanOffset       bool             // Pode compensar no IR?
}

// CalculateTaxWithheld calcula IR retido nos proventos
func (w *Wallet) CalculateTaxWithheld(year int) []TaxWithheld {
    results := make([]TaxWithheld, 0)

    for _, asset := range w.Assets {
        gross := decimal.Zero
        net := decimal.Zero

        // FIIs têm 15% retido em rendimentos
        isFII := asset.SubType == "fundos imobiliários"

        for _, earning := range asset.Earnings {
            if earning.Date.Year() != year {
                continue
            }

            // Apenas rendimentos de FII têm IR retido
            if isFII && earning.Type == "Rendimento" {
                gross = gross.Add(earning.TotalAmount)
                // Valor líquido (já recebido)
                net = net.Add(earning.TotalAmount)
            }
        }

        if gross.GreaterThan(decimal.Zero) {
            // Reverter cálculo para obter bruto
            // Net = Gross × 0.85
            // Gross = Net / 0.85
            actualGross := net.Div(decimal.NewFromFloat(0.85))
            withheld := actualGross.Sub(net)

            results = append(results, TaxWithheld{
                Ticker:         asset.ID,
                GrossAmount:    actualGross.Round(2),
                WithheldAmount: withheld.Round(2),
                NetAmount:      net.Round(2),
                TaxRate:        decimal.NewFromInt(15),
                CanOffset:      true,  // Pode compensar na declaração
            })
        }
    }

    return results
}
```

### 4. Reinvestimento de Proventos

**Estruturas:**

```go
type Reinvestment struct {
    ID              string
    EarningID       string              // Provento que foi reinvestido
    EarningDate     time.Time
    EarningAmount   decimal.Decimal
    PurchaseDate    time.Time
    PurchaseTicker  string              // Pode ser diferente
    PurchaseQty     decimal.Decimal
    PurchasePrice   decimal.Decimal
    TransactionHash string
}

// LinkReinvestment vincula provento a uma compra
func (w *Wallet) LinkReinvestment(earningTicker string, earningDate time.Time, purchaseHash string) (*Reinvestment, error) {
    // Encontrar o earning
    var earning *parser.Earning
    asset := w.Assets[earningTicker]

    for i := range asset.Earnings {
        if asset.Earnings[i].Date.Equal(earningDate) {
            earning = &asset.Earnings[i]
            break
        }
    }

    if earning == nil {
        return nil, fmt.Errorf("provento não encontrado")
    }

    // Encontrar a transação de compra
    var purchase *parser.Transaction
    for _, a := range w.Assets {
        for i := range a.Negotiations {
            if a.Negotiations[i].Hash == purchaseHash {
                purchase = &a.Negotiations[i]
                break
            }
        }
    }

    if purchase == nil {
        return nil, fmt.Errorf("transação não encontrada")
    }

    // Criar vínculo
    reinv := &Reinvestment{
        ID:              generateUUID(),
        EarningID:       earning.Hash,
        EarningDate:     earning.Date,
        EarningAmount:   earning.TotalAmount,
        PurchaseDate:    purchase.Date,
        PurchaseTicker:  purchase.Ticker,
        PurchaseQty:     purchase.Quantity,
        PurchasePrice:   purchase.Price,
        TransactionHash: purchase.Hash,
    }

    w.Reinvestments = append(w.Reinvestments, reinv)

    return reinv, nil
}

// CalculateDRIPEffect calcula efeito do reinvestimento automático
func (w *Wallet) CalculateDRIPEffect() *DRIPAnalysis {
    analysis := &DRIPAnalysis{
        TotalReinvested:     decimal.Zero,
        SharesAcquired:      decimal.Zero,
        CompoundingEffect:   decimal.Zero,
    }

    for _, reinv := range w.Reinvestments {
        analysis.TotalReinvested = analysis.TotalReinvested.Add(reinv.EarningAmount)
        analysis.SharesAcquired = analysis.SharesAcquired.Add(reinv.PurchaseQty)

        // Proventos gerados pelas ações reinvestidas
        // (simplificado - seria recursivo na realidade)
        asset := w.Assets[reinv.PurchaseTicker]
        futureEarnings := estimateFutureEarnings(asset, reinv.PurchaseQty)
        analysis.CompoundingEffect = analysis.CompoundingEffect.Add(futureEarnings)
    }

    return analysis
}

type DRIPAnalysis struct {
    TotalReinvested   decimal.Decimal
    SharesAcquired    decimal.Decimal
    CompoundingEffect decimal.Decimal
}
```

---

## 🎨 Interface do Usuário (CLI)

### Comandos propostos:

```bash
# Calendário de proventos
b3cli earnings calendar
b3cli earnings calendar --month=2024-12

# Dividend Yield por ativo
b3cli earnings yield MXRF11
b3cli earnings yield --all --sort=dy

# Ranking de DY
b3cli earnings ranking

# IR retido
b3cli earnings tax 2024
b3cli earnings tax 2024 --export=pdf

# Vincular reinvestimento
b3cli earnings reinvest link

# Análise de proventos
b3cli earnings analysis MXRF11
b3cli earnings analysis --all
```

### TUI - Calendário:

```
╔══════════════════════════════════════════════════════════════════════╗
║              📅 CALENDÁRIO DE PROVENTOS - DEZEMBRO 2024              ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  💰 RECEBIDOS ESTE MÊS                                               ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  05/12  MXRF11   Rendimento      R$   125.50                        ║
║  10/12  HGLG11   Rendimento      R$    98.20                        ║
║  15/12  ITSA4    Dividendo       R$    45.00                        ║
║  20/12  PETR4    JCP             R$   180.00                        ║
║                                                                      ║
║  Total recebido: R$ 448.70                                           ║
║                                                                      ║
║  📊 PROJETADOS (baseado em histórico)                                ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  Jan/25  ~R$ 520.00  (Confiança: 85%)                                ║
║  Fev/25  ~R$ 480.00  (Confiança: 80%)                                ║
║  Mar/25  ~R$ 550.00  (Confiança: 75%)                                ║
║                                                                      ║
║  [↑↓] Navegar  [Enter] Detalhes  [q] Sair                           ║
╚══════════════════════════════════════════════════════════════════════╝
```

### TUI - Ranking DY:

```
╔══════════════════════════════════════════════════════════════════════╗
║              🏆 RANKING POR DIVIDEND YIELD (12M)                     ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  #   Ticker    DY 12M    Proventos    Investido    Tipo             ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  1   MXRF11    12.5%     R$ 1,250     R$ 10,000    FII              ║
║  2   HGLG11    11.8%     R$   590     R$  5,000    FII              ║
║  3   ITSA4      8.2%     R$   328     R$  4,000    Ação             ║
║  4   PETR4      6.5%     R$   650     R$ 10,000    Ação             ║
║  5   BBDC4      5.1%     R$   255     R$  5,000    Ação             ║
║  6   VALE3      3.8%     R$   304     R$  8,000    Ação             ║
║                                                                      ║
║  Média da carteira: 8.2% a.a.                                        ║
║                                                                      ║
║  [↑↓] Navegar  [Enter] Análise  [q] Sair                            ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 🔄 Fluxo de Trabalho do Usuário

### Mensal:
1. `b3cli earnings calendar`
2. Ver o que receberá no mês
3. Planejar reinvestimentos

### Análise de ativos:
1. `b3cli earnings yield --all`
2. Identificar ativos com melhor DY
3. Decidir onde alocar próximo aporte

### Anual (IRPF):
1. `b3cli earnings tax 2024`
2. Saber quanto foi retido em FIIs
3. Informar na declaração

---

## 📊 Métricas de Sucesso

- ✅ Projeções com >80% de acurácia
- ✅ Usuários conseguem planejar renda passiva
- ✅ Redução de 100% em erros de IR retido
- ✅ Aumento de reinvestimento consciente

---

**Estimativa de implementação:** 1-2 semanas
**ROI para usuários:** Alto (otimização de renda passiva)
