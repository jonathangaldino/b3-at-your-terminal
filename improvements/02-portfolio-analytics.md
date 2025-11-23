# Melhoria 02: Análise de Carteira (Portfolio Analytics)

**Prioridade:** P0 (Crítica)
**Complexidade:** Média
**Impacto:** Muito Alto

---

## 📋 Visão Geral

Ferramentas avançadas para análise profunda da carteira, métricas de performance, alocação de ativos e sugestões de otimização.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Falta de visibilidade sobre a carteira**
   - Usuário não sabe se está bem diversificado
   - Difícil ver o "quadro geral" dos investimentos
   - Impossível comparar performance entre ativos

2. **Decisões de investimento sem dados**
   - "Devo comprar mais desse ativo?"
   - "Estou muito exposto a um setor?"
   - "Minha carteira está performando bem?"

3. **Ausência de benchmarks**
   - Não sabe se está batendo o IBOV/IFIX
   - Sem métricas objetivas de sucesso
   - Difícil justificar estratégia

### Benefícios mensuráveis:

- 📊 **Decisões mais informadas:** Reduz "achismos" em 80%+
- 🎯 **Melhor alocação:** Identifica concentrações de risco
- 📈 **Performance otimizada:** Comparação objetiva com mercado
- 💡 **Insights acionáveis:** Sugestões práticas de melhoria

---

## 🏗️ Arquitetura Proposta

### Componentes principais:

```
internal/analytics/
├── allocation.go        # Alocação de ativos
├── performance.go       # Métricas de ROI, rentabilidade
├── diversification.go   # Análise de diversificação
├── benchmarks.go        # Comparação com índices
├── metrics.go           # Cálculos de métricas
└── scoring.go           # Sistema de pontuação

cmd/b3cli/
├── portfolio.go         # Comando principal
├── portfolio_summary.go # Resumo geral
├── portfolio_allocation.go
├── portfolio_performance.go
└── portfolio_insights.go
```

---

## 💡 Implementação Técnica

### 1. Alocação de Ativos

**Estruturas de dados:**

```go
type AllocationBreakdown struct {
    ByType        map[string]AllocationItem  // ações, FIIs, BDRs
    BySubType     map[string]AllocationItem  // blue chips, small caps
    BySegment     map[string]AllocationItem  // tecnologia, bancos
    BySector      map[string]AllocationItem  // financeiro, energia
    TopHoldings   []Holding                  // Top 10 posições
    Concentration ConcentrationMetrics
}

type AllocationItem struct {
    Name            string
    TotalInvested   decimal.Decimal
    Percentage      decimal.Decimal  // % do total da carteira
    Count           int              // Número de ativos
    Assets          []string         // Lista de tickers
}

type Holding struct {
    Ticker          string
    TotalInvested   decimal.Decimal
    Percentage      decimal.Decimal
    Quantity        int
    AveragePrice    decimal.Decimal
}

type ConcentrationMetrics struct {
    HHI             decimal.Decimal  // Herfindahl-Hirschman Index
    EffectiveNumber int              // Número efetivo de ativos
    Top3Percentage  decimal.Decimal  // % dos 3 maiores
    Top5Percentage  decimal.Decimal  // % dos 5 maiores
    Top10Percentage decimal.Decimal  // % dos 10 maiores
}
```

**Implementação:**

```go
// CalculateAllocation calcula a alocação da carteira
func (w *Wallet) CalculateAllocation() *AllocationBreakdown {
    breakdown := &AllocationBreakdown{
        ByType:    make(map[string]AllocationItem),
        BySubType: make(map[string]AllocationItem),
        BySegment: make(map[string]AllocationItem),
    }

    // Total da carteira
    totalInvested := decimal.Zero
    for _, asset := range w.GetActiveAssets() {
        totalInvested = totalInvested.Add(asset.TotalInvestedValue)
    }

    // Quebrar por tipo, subtype, segmento
    for _, asset := range w.GetActiveAssets() {
        percentage := asset.TotalInvestedValue.Div(totalInvested).Mul(decimal.NewFromInt(100))

        // Por tipo
        addToAllocation(breakdown.ByType, asset.Type, asset, percentage)

        // Por subtipo
        if asset.SubType != "" {
            addToAllocation(breakdown.BySubType, asset.SubType, asset, percentage)
        }

        // Por segmento
        if asset.Segment != "" {
            addToAllocation(breakdown.BySegment, asset.Segment, asset, percentage)
        }
    }

    // Calcular top holdings
    breakdown.TopHoldings = w.getTopHoldings(10, totalInvested)

    // Calcular métricas de concentração
    breakdown.Concentration = calculateConcentration(breakdown.TopHoldings, totalInvested)

    return breakdown
}

// calculateConcentration calcula o índice HHI e outras métricas
func calculateConcentration(holdings []Holding, total decimal.Decimal) ConcentrationMetrics {
    // HHI = Σ(peso²) × 10000
    hhi := decimal.Zero
    for _, h := range holdings {
        weight := h.Percentage.Div(decimal.NewFromInt(100))
        hhi = hhi.Add(weight.Mul(weight))
    }
    hhi = hhi.Mul(decimal.NewFromInt(10000))

    // Número efetivo = 1 / Σ(peso²)
    effectiveN := decimal.NewFromInt(1).Div(hhi.Div(decimal.NewFromInt(10000)))

    // Top N percentages
    top3 := sumPercentages(holdings[:min(3, len(holdings))])
    top5 := sumPercentages(holdings[:min(5, len(holdings))])
    top10 := sumPercentages(holdings[:min(10, len(holdings))])

    return ConcentrationMetrics{
        HHI:             hhi.Round(2),
        EffectiveNumber: int(effectiveN.IntPart()),
        Top3Percentage:  top3,
        Top5Percentage:  top5,
        Top10Percentage: top10,
    }
}
```

### 2. Métricas de Performance

**Estruturas:**

```go
type PerformanceMetrics struct {
    TotalInvested     decimal.Decimal
    CurrentValue      decimal.Decimal  // Se integrado com cotações
    UnrealizedPL      decimal.Decimal  // Profit/Loss não realizado
    UnrealizedPLPct   decimal.Decimal  // % de lucro/prejuízo
    RealizedGains     decimal.Decimal  // Ganhos realizados (vendas)
    RealizedLosses    decimal.Decimal  // Perdas realizadas
    TotalEarnings     decimal.Decimal  // Proventos recebidos
    TotalReturn       decimal.Decimal  // Retorno total (ganhos + proventos)
    ROI               decimal.Decimal  // Return on Investment (%)
    AverageDY         decimal.Decimal  // Dividend Yield médio
    CAGR              decimal.Decimal  // Compound Annual Growth Rate
    ByAsset           []AssetPerformance
}

type AssetPerformance struct {
    Ticker            string
    TotalInvested     decimal.Decimal
    CurrentValue      decimal.Decimal
    UnrealizedPL      decimal.Decimal
    UnrealizedPLPct   decimal.Decimal
    Earnings          decimal.Decimal
    TotalReturn       decimal.Decimal
    DY12M             decimal.Decimal  // Dividend Yield 12 meses
    ROI               decimal.Decimal
}
```

**Implementação:**

```go
// CalculatePerformance calcula métricas de performance da carteira
func (w *Wallet) CalculatePerformance() *PerformanceMetrics {
    metrics := &PerformanceMetrics{
        ByAsset: make([]AssetPerformance, 0),
    }

    totalEarnings := decimal.Zero
    totalInvested := decimal.Zero

    for _, asset := range w.GetActiveAssets() {
        totalInvested = totalInvested.Add(asset.TotalInvestedValue)
        totalEarnings = totalEarnings.Add(asset.TotalEarnings)

        // Performance por ativo
        assetPerf := AssetPerformance{
            Ticker:        asset.ID,
            TotalInvested: asset.TotalInvestedValue,
            Earnings:      asset.TotalEarnings,
        }

        // Se temos cotação atual, calcular P/L não realizado
        if currentPrice, ok := w.GetCurrentPrice(asset.ID); ok {
            currentValue := currentPrice.Mul(decimal.NewFromInt(int64(asset.Quantity)))
            assetPerf.CurrentValue = currentValue
            assetPerf.UnrealizedPL = currentValue.Sub(asset.TotalInvestedValue)
            assetPerf.UnrealizedPLPct = assetPerf.UnrealizedPL.
                Div(asset.TotalInvestedValue).
                Mul(decimal.NewFromInt(100))
        }

        // Calcular DY12M
        assetPerf.DY12M = w.CalculateDY12M(asset)

        // ROI = (Valor Atual - Investido + Proventos) / Investido × 100
        if !assetPerf.CurrentValue.IsZero() {
            assetPerf.TotalReturn = assetPerf.UnrealizedPL.Add(assetPerf.Earnings)
            assetPerf.ROI = assetPerf.TotalReturn.
                Div(assetPerf.TotalInvested).
                Mul(decimal.NewFromInt(100))
        }

        metrics.ByAsset = append(metrics.ByAsset, assetPerf)
    }

    metrics.TotalInvested = totalInvested
    metrics.TotalEarnings = totalEarnings

    // Calcular métricas gerais
    metrics.AverageDY = w.CalculatePortfolioDY()

    // Se temos cotações, calcular totais
    if w.HasPriceData() {
        for _, ap := range metrics.ByAsset {
            metrics.CurrentValue = metrics.CurrentValue.Add(ap.CurrentValue)
            metrics.UnrealizedPL = metrics.UnrealizedPL.Add(ap.UnrealizedPL)
        }

        metrics.UnrealizedPLPct = metrics.UnrealizedPL.
            Div(metrics.TotalInvested).
            Mul(decimal.NewFromInt(100))

        metrics.TotalReturn = metrics.UnrealizedPL.Add(metrics.TotalEarnings)
        metrics.ROI = metrics.TotalReturn.
            Div(metrics.TotalInvested).
            Mul(decimal.NewFromInt(100))
    }

    return metrics
}

// CalculateDY12M calcula o Dividend Yield dos últimos 12 meses
func (w *Wallet) CalculateDY12M(asset *Asset) decimal.Decimal {
    // Obter proventos dos últimos 12 meses
    oneYearAgo := time.Now().AddDate(-1, 0, 0)
    earningsLast12M := decimal.Zero

    for _, earning := range asset.Earnings {
        if earning.Date.After(oneYearAgo) {
            earningsLast12M = earningsLast12M.Add(earning.TotalAmount)
        }
    }

    // DY = (Proventos / Investimento) × 100
    if asset.TotalInvestedValue.IsZero() {
        return decimal.Zero
    }

    return earningsLast12M.Div(asset.TotalInvestedValue).Mul(decimal.NewFromInt(100))
}
```

### 3. Análise de Diversificação

**Implementação:**

```go
type DiversificationAnalysis struct {
    Score             int                // 0-100
    NumberOfAssets    int
    NumberOfSectors   int
    NumberOfSegments  int
    ConcentrationRisk string             // "baixo", "médio", "alto"
    Recommendations   []string
    Issues            []DiversificationIssue
}

type DiversificationIssue struct {
    Severity    string  // "critical", "warning", "info"
    Type        string  // "concentration", "sector-exposure", etc
    Description string
    Suggestion  string
}

// AnalyzeDiversification analisa a diversificação da carteira
func (w *Wallet) AnalyzeDiversification() *DiversificationAnalysis {
    analysis := &DiversificationAnalysis{
        Recommendations: make([]string, 0),
        Issues:          make([]DiversificationIssue, 0),
    }

    allocation := w.CalculateAllocation()

    // Contar ativos, setores, segmentos
    analysis.NumberOfAssets = len(w.GetActiveAssets())
    analysis.NumberOfSectors = len(allocation.BySegment)

    // Avaliar concentração
    hhi := allocation.Concentration.HHI.IntPart()

    if hhi < 1500 {
        analysis.ConcentrationRisk = "baixo"
    } else if hhi < 2500 {
        analysis.ConcentrationRisk = "médio"
        analysis.Issues = append(analysis.Issues, DiversificationIssue{
            Severity:    "warning",
            Type:        "concentration",
            Description: "Carteira moderadamente concentrada",
            Suggestion:  "Considere aumentar o número de ativos",
        })
    } else {
        analysis.ConcentrationRisk = "alto"
        analysis.Issues = append(analysis.Issues, DiversificationIssue{
            Severity:    "critical",
            Type:        "concentration",
            Description: "Carteira muito concentrada - risco elevado",
            Suggestion:  "Diversifique! Objetivo: HHI < 2000",
        })
    }

    // Verificar se há ativos com > 20% da carteira
    for _, holding := range allocation.TopHoldings {
        if holding.Percentage.GreaterThan(decimal.NewFromInt(20)) {
            analysis.Issues = append(analysis.Issues, DiversificationIssue{
                Severity:    "warning",
                Type:        "single-asset-exposure",
                Description: fmt.Sprintf("%s representa %.1f%% da carteira", holding.Ticker, holding.Percentage),
                Suggestion:  "Considere reduzir exposição a um único ativo",
            })
        }
    }

    // Verificar setores com > 30% da carteira
    for sector, alloc := range allocation.BySegment {
        if alloc.Percentage.GreaterThan(decimal.NewFromInt(30)) {
            analysis.Issues = append(analysis.Issues, DiversificationIssue{
                Severity:    "info",
                Type:        "sector-exposure",
                Description: fmt.Sprintf("Setor '%s' representa %.1f%%", sector, alloc.Percentage),
                Suggestion:  "Considere diversificar em outros setores",
            })
        }
    }

    // Calcular score (0-100)
    analysis.Score = calculateDiversificationScore(analysis, allocation)

    return analysis
}

func calculateDiversificationScore(analysis *DiversificationAnalysis, allocation *AllocationBreakdown) int {
    score := 100

    // Penalizar por HHI alto
    hhi := allocation.Concentration.HHI.IntPart()
    if hhi > 2500 {
        score -= 30
    } else if hhi > 2000 {
        score -= 20
    } else if hhi > 1500 {
        score -= 10
    }

    // Penalizar por poucos ativos
    if analysis.NumberOfAssets < 5 {
        score -= 20
    } else if analysis.NumberOfAssets < 10 {
        score -= 10
    }

    // Penalizar por poucos setores
    if analysis.NumberOfSectors < 3 {
        score -= 15
    } else if analysis.NumberOfSectors < 5 {
        score -= 5
    }

    // Bonificar diversificação boa
    if analysis.NumberOfAssets >= 15 && analysis.NumberOfSectors >= 5 && hhi < 1500 {
        score += 10
    }

    // Limitar entre 0-100
    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }

    return score
}
```

---

## 🎨 Interface do Usuário (CLI)

### Comandos propostos:

```bash
# Resumo geral
b3cli portfolio summary

# Alocação de ativos
b3cli portfolio allocation
b3cli portfolio allocation --by=segment
b3cli portfolio allocation --by=type

# Performance
b3cli portfolio performance
b3cli portfolio performance --sort=roi

# Diversificação
b3cli portfolio diversification

# Insights e recomendações
b3cli portfolio insights
```

### TUI - Portfolio Summary:

```
╔══════════════════════════════════════════════════════════════════════╗
║              📊 RESUMO DA CARTEIRA                                   ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  💰 VALORES                                                          ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  Total investido:      R$ 125.430,00                                 ║
║  Valor atual:          R$ 142.850,00                                 ║
║  Lucro não realizado:  R$  17.420,00  (+13.89%)  ↑                  ║
║  Proventos recebidos:  R$   8.250,00                                 ║
║  Retorno total:        R$  25.670,00  (+20.46%)                      ║
║                                                                      ║
║  📊 COMPOSIÇÃO                                                       ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  Ações:         65.2%  ████████████████░░░░░░░░                     ║
║  FIIs:          28.5%  ███████░░░░░░░░░░░░░░░░░                     ║
║  BDRs:           6.3%  ██░░░░░░░░░░░░░░░░░░░░░░                     ║
║                                                                      ║
║  🎯 DIVERSIFICAÇÃO                                                   ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  Score:            78/100  ⭐⭐⭐⭐☆                                  ║
║  Ativos:           18                                                ║
║  Setores:          8                                                 ║
║  HHI:              1,245  (baixa concentração)                       ║
║  DY médio:         8.2% a.a.                                         ║
║                                                                      ║
║  [Tab] Detalhes  [Enter] Análise  [q] Sair                          ║
╚══════════════════════════════════════════════════════════════════════╝
```

### TUI - Alocação por Segmento:

```
╔══════════════════════════════════════════════════════════════════════╗
║              📈 ALOCAÇÃO POR SEGMENTO                                ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Bancos                 28.5%  ███████████░░░░░  R$ 35.750           ║
║  Energia                18.2%  ███████░░░░░░░░░  R$ 22.850           ║
║  Varejo                 15.3%  ██████░░░░░░░░░░  R$ 19.200           ║
║  Tecnologia             12.8%  █████░░░░░░░░░░░  R$ 16.050           ║
║  Logística (FII)        10.5%  ████░░░░░░░░░░░░  R$ 13.200           ║
║  Papel (FII)             8.7%  ███░░░░░░░░░░░░░  R$ 10.900           ║
║  Utilities               6.0%  ██░░░░░░░░░░░░░░  R$  7.480           ║
║                                                                      ║
║  🎯 Análise de concentração:                                         ║
║     ✓ Boa diversificação entre setores                              ║
║     ⚠ Setor bancário acima de 25% (considere reduzir)               ║
║                                                                      ║
║  [↑↓] Navegar  [Enter] Detalhes  [Esc] Voltar                       ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 🔄 Fluxo de Trabalho do Usuário

### Semanal/Mensal:
1. `b3cli portfolio summary` - Ver snapshot geral
2. Avaliar se está no caminho das metas
3. Identificar oportunidades

### Antes de investir:
1. `b3cli portfolio allocation --by=segment`
2. Ver onde está concentrado
3. Decidir onde alocar próximo aporte

### Trimestral:
1. `b3cli portfolio performance`
2. Ver ativos com melhor/pior performance
3. `b3cli portfolio diversification`
4. Avaliar necessidade de rebalanceamento

---

## 📊 Métricas de Sucesso

- ✅ Usuário consegue responder "Como está minha carteira?" em < 30 segundos
- ✅ 90%+ dos usuários acham os insights úteis
- ✅ Redução de 50% em decisões de investimento "no escuro"
- ✅ Aumento da diversificação média dos usuários

---

## 🚀 Expansões Futuras

1. **Benchmarking**
   - Comparar com IBOV, IFIX
   - Beta da carteira
   - Sharpe ratio

2. **Simulações**
   - "E se eu comprar X ações de PETR4?"
   - Impacto na alocação

3. **Alertas**
   - "Você está 30% em bancos - risco!"
   - "Top 3 ativos = 60% da carteira"

---

**Estimativa de implementação:** 2 semanas
**ROI para usuários:** Alto (decisões mais inteligentes = melhor performance)
