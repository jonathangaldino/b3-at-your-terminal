# Melhoria 01: Declaração de Imposto de Renda (IRPF)

**Prioridade:** P0 (Crítica)
**Complexidade:** Alta
**Impacto:** Muito Alto

---

## 📋 Visão Geral

Sistema completo para cálculo de impostos, geração de DARF e preparação de dados para a declaração anual do Imposto de Renda Pessoa Física (IRPF).

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Eliminação de erros manuais**
   - Cálculo automático de ganho de capital
   - Redução de risco de multas da Receita Federal
   - Dados precisos e auditáveis

2. **Economia de tempo**
   - Processo manual pode levar horas/dias
   - Automação reduz para minutos
   - Evita necessidade de contratar contador (para casos simples)

3. **Economia financeira**
   - Identifica automaticamente vendas isentas
   - Otimiza compensação de prejuízos
   - Pode economizar centenas a milhares de reais em IR

4. **Tranquilidade e conformidade**
   - Certeza de estar em dia com a Receita Federal
   - Relatórios prontos para envio
   - Histórico completo para eventuais fiscalizações

### Benefícios mensuráveis:

- ⏱️ **Economia de tempo:** 10-20 horas/ano → 30 minutos/ano
- 💰 **Economia potencial:** R$ 500 - R$ 2.000/ano (evitando erros e otimizando)
- 📊 **Redução de erros:** 90%+ (comparado a cálculo manual)
- 😌 **Redução de estresse:** Imensurável

---

## 🏗️ Arquitetura Proposta

### Componentes principais:

```
internal/tax/
├── calculator.go         # Cálculo de ganho de capital
├── darf.go              # Geração de DARF
├── irpf.go              # Relatórios IRPF
├── exemptions.go        # Regras de isenção
├── losses.go            # Gestão de prejuízos
└── validators.go        # Validação de dados fiscais

cmd/b3cli/
├── tax.go               # Comando principal
├── tax_calculate.go     # UI para cálculo
├── tax_darf.go          # UI para DARF
└── tax_irpf.go          # UI para IRPF
```

---

## 💡 Implementação Técnica

### 1. Cálculo de Ganho de Capital

**Conceito:**
- Ganho = Preço de Venda - Custo de Aquisição
- Custo = Preço Médio × Quantidade Vendida
- IR devido = Ganho × Alíquota

**Estrutura de dados:**

```go
type CapitalGain struct {
    Month          time.Time
    Ticker         string
    SaleDate       time.Time
    Quantity       int
    SalePrice      decimal.Decimal
    AverageCost    decimal.Decimal
    TotalSale      decimal.Decimal  // Preço × Quantidade
    TotalCost      decimal.Decimal  // Custo × Quantidade
    Gain           decimal.Decimal  // TotalSale - TotalCost
    TaxRate        decimal.Decimal  // 15% swing ou 20% day trade
    TaxDue         decimal.Decimal  // Gain × TaxRate
    IsExempt       bool             // Vendas < R$ 20k (ações)
    IsDayTrade     bool
}

type MonthlyTaxReport struct {
    Month          time.Time
    TotalSales     decimal.Decimal
    TotalGains     decimal.Decimal
    TotalLosses    decimal.Decimal
    NetGain        decimal.Decimal  // Gains - Losses
    AccumulatedLoss decimal.Decimal // Prejuízos anteriores
    TaxableAmount  decimal.Decimal  // NetGain - AccumulatedLoss
    TaxDue         decimal.Decimal
    IsExempt       bool
    Transactions   []CapitalGain
}
```

**Regras de negócio:**

```go
// CalculateMonthlyTax calcula o imposto devido em um mês
func (w *Wallet) CalculateMonthlyTax(year int, month int) (*MonthlyTaxReport, error) {
    // 1. Obter todas as vendas do mês
    sales := w.GetSalesByMonth(year, month)

    // 2. Para cada venda, calcular ganho/prejuízo
    var gains []CapitalGain
    totalSales := decimal.Zero

    for _, sale := range sales {
        // Obter preço médio do ativo na data da venda
        avgPrice := w.GetAveragePriceAtDate(sale.Ticker, sale.Date)

        gain := CapitalGain{
            Month:       time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC),
            Ticker:      sale.Ticker,
            SaleDate:    sale.Date,
            Quantity:    sale.Quantity.IntPart(),
            SalePrice:   sale.Price,
            AverageCost: avgPrice,
            TotalSale:   sale.Amount,
            TotalCost:   avgPrice.Mul(sale.Quantity),
            IsDayTrade:  w.IsDayTrade(sale),
        }

        gain.Gain = gain.TotalSale.Sub(gain.TotalCost)
        gain.TaxRate = getTaxRate(gain.IsDayTrade)

        totalSales = totalSales.Add(gain.TotalSale)
        gains = append(gains, gain)
    }

    // 3. Verificar isenção (R$ 20.000 para ações)
    isExempt := isExemptFromTax(sales, totalSales)

    // 4. Calcular total de ganhos e perdas
    totalGains, totalLosses := calculateGainsAndLosses(gains)

    // 5. Aplicar prejuízos acumulados
    accumulatedLoss := w.GetAccumulatedLoss(year, month)
    taxableAmount := totalGains.Sub(totalLosses).Sub(accumulatedLoss)

    if taxableAmount.LessThan(decimal.Zero) {
        taxableAmount = decimal.Zero
    }

    // 6. Calcular imposto devido
    taxDue := decimal.Zero
    if !isExempt && taxableAmount.GreaterThan(decimal.Zero) {
        // Média ponderada das alíquotas
        taxDue = calculateWeightedTax(gains, taxableAmount)
    }

    return &MonthlyTaxReport{
        Month:           time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC),
        TotalSales:      totalSales,
        TotalGains:      totalGains,
        TotalLosses:     totalLosses,
        NetGain:         totalGains.Sub(totalLosses),
        AccumulatedLoss: accumulatedLoss,
        TaxableAmount:   taxableAmount,
        TaxDue:          taxDue,
        IsExempt:        isExempt,
        Transactions:    gains,
    }, nil
}

// Regras de isenção
func isExemptFromTax(sales []Transaction, totalSales decimal.Decimal) bool {
    // Verificar se todas são ações (não FII)
    allStocks := true
    for _, sale := range sales {
        asset := w.GetAsset(sale.Ticker)
        if asset.SubType == "fundos imobiliários" {
            allStocks = false
            break
        }
    }

    // Ações: isento se vendas < R$ 20.000
    if allStocks {
        threshold := decimal.NewFromInt(20000)
        return totalSales.LessThan(threshold)
    }

    // FII: nunca isento
    return false
}
```

### 2. Geração de DARF

**Estrutura:**

```go
type DARF struct {
    ReferenceMonth time.Time       // Mês de referência (venda)
    DueDate        time.Time       // Último dia útil do mês seguinte
    TaxCode        string          // 6015 (swing) ou 8523 (day trade)
    TaxableAmount  decimal.Decimal // Valor base de cálculo
    TaxDue         decimal.Decimal // Valor do imposto
    Barcode        string          // Código de barras para pagamento
    PaymentStatus  string          // "pending", "paid", "overdue"
}

// GenerateDARF gera um DARF para um mês específico
func GenerateDARF(report *MonthlyTaxReport) (*DARF, error) {
    if report.TaxDue.LessThanOrEqual(decimal.Zero) {
        return nil, fmt.Errorf("sem imposto devido neste mês")
    }

    // Data de vencimento: último dia útil do mês seguinte
    dueDate := getLastBusinessDay(report.Month.AddDate(0, 1, 0))

    // Código do imposto
    taxCode := "6015" // Swing trade (comum)
    if hasOnlyDayTrades(report) {
        taxCode = "8523" // Day trade
    }

    // Gerar código de barras (simplificado)
    barcode := generateBarcode(taxCode, report.TaxDue, dueDate)

    return &DARF{
        ReferenceMonth: report.Month,
        DueDate:        dueDate,
        TaxCode:        taxCode,
        TaxableAmount:  report.TaxableAmount,
        TaxDue:         report.TaxDue.Round(2),
        Barcode:        barcode,
        PaymentStatus:  "pending",
    }, nil
}
```

### 3. Relatório Anual IRPF

**Funcionalidades:**

```go
type IRPFReport struct {
    Year            int
    TotalInvested   decimal.Decimal  // Posição em 31/12
    Purchases       []AssetPosition  // Compras do ano
    Sales           []AssetPosition  // Vendas do ano
    CapitalGains    decimal.Decimal  // Total de ganhos
    CapitalLosses   decimal.Decimal  // Total de perdas
    TaxPaid         decimal.Decimal  // Total de IR pago (DARFs)
    MonthlyReports  []MonthlyTaxReport
    AssetPositions  []AssetPosition  // Posição final por ativo
}

type AssetPosition struct {
    Ticker          string
    Quantity        int
    AveragePrice    decimal.Decimal
    TotalInvested   decimal.Decimal
    SubType         string          // "ações" ou "fundos imobiliários"
}

// GenerateAnnualIRPFReport gera relatório completo do ano
func (w *Wallet) GenerateAnnualIRPFReport(year int) (*IRPFReport, error) {
    report := &IRPFReport{Year: year}

    // 1. Calcular impostos de cada mês
    for month := 1; month <= 12; month++ {
        monthReport, err := w.CalculateMonthlyTax(year, month)
        if err != nil {
            continue
        }

        if monthReport.TaxDue.GreaterThan(decimal.Zero) {
            report.MonthlyReports = append(report.MonthlyReports, *monthReport)
            report.TaxPaid = report.TaxPaid.Add(monthReport.TaxDue)
        }

        report.CapitalGains = report.CapitalGains.Add(monthReport.TotalGains)
        report.CapitalLosses = report.CapitalLosses.Add(monthReport.TotalLosses)
    }

    // 2. Posição final dos ativos (31/12)
    snapshot := w.GetSnapshotAtDate(time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC))
    report.AssetPositions = snapshot.Positions

    for _, pos := range snapshot.Positions {
        report.TotalInvested = report.TotalInvested.Add(pos.TotalInvested)
    }

    // 3. Resumo de compras e vendas do ano
    report.Purchases = w.GetPurchasesByYear(year)
    report.Sales = w.GetSalesByYear(year)

    return report, nil
}
```

---

## 🎨 Interface do Usuário (CLI)

### Comandos propostos:

```bash
# Calcular imposto de um mês
b3cli tax calculate 2024-11
# Saída:
# 📊 Cálculo de Imposto - Novembro 2024
#
# Vendas totais: R$ 45.320,00
# Ganho bruto: R$ 3.450,00
# Prejuízos: R$ 0,00
# Prejuízos acumulados: R$ 1.200,00
#
# Base de cálculo: R$ 2.250,00
# Alíquota: 15% (swing trade)
# Imposto devido: R$ 337,50
#
# Status: Não isento (vendas > R$ 20.000)
# Vencimento DARF: 20/12/2024

# Gerar DARF
b3cli tax darf 2024-11
# Saída:
# 📄 DARF - Novembro 2024
#
# Código: 6015 (Ações - Swing Trade)
# Valor: R$ 337,50
# Vencimento: 20/12/2024
#
# Código de barras:
# 60152024120000033750...
#
# Instruções:
# 1. Copie o código de barras acima
# 2. Acesse o site do seu banco
# 3. Pague até 20/12/2024
#
# Salvo em: ./minha-carteira/tax/darf-2024-11.pdf

# Relatório anual
b3cli tax irpf 2024
# Abre TUI interativo com:
# - Resumo anual
# - DARFs pagos/pendentes
# - Dados para ficha "Bens e Direitos"
# - Ganhos e perdas detalhados
# - Opção de exportar PDF/CSV

# Ganhos de capital detalhados
b3cli tax capital-gains --year=2024
# Lista todas as vendas com lucro/prejuízo

# Verificar isenções
b3cli tax exemptions 2024
# Mostra quais meses tiveram isenção e por quê
```

### TUI para relatório anual:

```
╔══════════════════════════════════════════════════════════════════════╗
║              📊 RELATÓRIO IRPF 2024                                  ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  RESUMO DO ANO                                                       ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  Total investido (31/12): R$ 125.430,00                              ║
║  Ganhos de capital:       R$  12.450,00                              ║
║  Prejuízos:               R$   2.100,00                              ║
║  Imposto pago (DARFs):    R$   1.552,50                              ║
║                                                                      ║
║  BENS E DIREITOS                                                     ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  ITSA4  - 500 cotas  - R$ 5.450,00 - Código 31 (Ações)              ║
║  PETR4  - 200 cotas  - R$ 6.800,00 - Código 31 (Ações)              ║
║  MXRF11 - 300 cotas  - R$ 3.150,00 - Código 73 (FII)                ║
║  ...                                                                 ║
║                                                                      ║
║  AÇÕES                                                               ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  [Tab] DARFs  [Enter] Exportar  [↑↓] Navegar  [q] Sair              ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 🔄 Fluxo de Trabalho do Usuário

### Mensal (5 minutos):
1. Vendeu ações no mês?
2. `b3cli tax calculate 2024-11`
3. Tem imposto? `b3cli tax darf 2024-11`
4. Pagar DARF no banco

### Anual (15 minutos):
1. `b3cli tax irpf 2024`
2. Revisar relatório
3. Exportar PDF
4. Preencher IRPF com dados prontos

---

## 🧪 Casos de Teste

```go
func TestTaxCalculation(t *testing.T) {
    // Caso 1: Venda isenta (< R$ 20k)
    // Caso 2: Venda com lucro
    // Caso 3: Venda com prejuízo
    // Caso 4: Compensação de prejuízos
    // Caso 5: FII (nunca isento)
    // Caso 6: Day trade (alíquota 20%)
    // Caso 7: Múltiplas vendas no mês
}
```

---

## 📊 Métricas de Sucesso

- ✅ 100% dos cálculos conferem com planilhas manuais
- ✅ DARFs gerados são aceitos pelo sistema da Receita
- ✅ Relatórios IRPF aprovados por contadores
- ✅ Redução de 95% no tempo de preparação do IR
- ✅ Zero erros em auditorias fiscais

---

## 🚧 Riscos e Mitigações

### Riscos:
1. **Mudança na legislação** → Manter código modular e configurável
2. **Erros de cálculo** → Testes extensivos, validação com contadores
3. **Interpretação incorreta** → Disclaimers claros, sugerir consulta profissional

### Disclaimers necessários:
```
⚠️ IMPORTANTE: Esta ferramenta é auxiliar.
Consulte um contador para casos complexos.
O usuário é responsável pela exatidão da declaração.
```

---

## 📚 Referências

- [IN RFB 1585/2015](http://normas.receita.fazenda.gov.br/sijut2consulta/link.action?idAto=66619) - Ganho de capital
- [Perguntão IRPF 2024](https://www.gov.br/receitafederal/pt-br/assuntos/meu-imposto-de-renda/perguntas-e-respostas)
- [Como declarar ações](https://www.gov.br/receitafederal/pt-br/assuntos/meu-imposto-de-renda/preenchimento/rendimentos-de-aplicacoes-financeiras-e-ganho-de-capital)

---

## 🎓 Próximos Passos

1. Implementar estruturas de dados básicas
2. Desenvolver algoritmo de cálculo de ganho de capital
3. Criar gerador de DARF (inicialmente em texto)
4. Implementar TUI para visualização
5. Adicionar exportação PDF
6. Testes extensivos com casos reais
7. Validação com contadores
8. Beta test com usuários reais

---

**Estimativa de implementação:** 3-4 semanas
**ROI estimado para usuários:** Altíssimo (economia de tempo e dinheiro)
