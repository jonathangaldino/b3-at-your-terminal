# Melhoria 07: Relatórios e Visualização

**Prioridade:** P2 (Média)
**Complexidade:** Média-Alta
**Impacto:** Médio-Alto

---

## 📋 Visão Geral

Melhorar visualização de dados através de gráficos no terminal, exportação de relatórios em PDF profissionais, dashboards HTML interativos e histórico de evolução da carteira.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Dados difíceis de visualizar**
   - Tabelas longas difíceis de interpretar
   - Tendências invisíveis em texto puro
   - Impossível ter visão panorâmica

2. **Falta de relatórios profissionais**
   - Não tem como mostrar para contador
   - Difícil compartilhar resultados
   - Sem documentação visual

3. **Impossível ver evolução**
   - Não sabe se está melhorando
   - Sem histórico visual
   - Difícil avaliar decisões passadas

### Benefícios mensuráveis:

- 📊 **Compreensão rápida:** Gráficos > tabelas
- 📄 **Profissionalismo:** PDFs para contadores/parceiros
- 📈 **Insights visuais:** Tendências óbvias
- 🎯 **Motivação:** Ver progresso visualmente

---

## 🏗️ Arquitetura Proposta

### Componentes principais:

```
internal/reports/
├── pdf/
│   ├── generator.go     # Gerador de PDF
│   ├── monthly.go       # Relatório mensal
│   ├── annual.go        # Relatório anual
│   └── charts.go        # Gráficos para PDF
├── html/
│   ├── dashboard.go     # Dashboard HTML
│   ├── templates.go     # Templates HTML
│   └── assets.go        # CSS/JS embutidos
├── terminal/
│   ├── charts.go        # Gráficos ASCII
│   └── sparklines.go    # Mini-gráficos
└── history.go           # Snapshots históricos

cmd/b3cli/
├── report.go            # Comando principal
├── report_monthly.go
├── report_annual.go
└── dashboard.go
```

---

## 💡 Implementação Técnica

### 1. Gráficos no Terminal (ASCII)

**Biblioteca:** `github.com/guptarohit/asciigraph`

```go
import "github.com/guptarohit/asciigraph"

// PlotEvolution plota evolução do patrimônio
func PlotEvolution(snapshots []PortfolioSnapshot) string {
    data := make([]float64, len(snapshots))

    for i, snapshot := range snapshots {
        data[i] = snapshot.TotalValue.InexactFloat64()
    }

    graph := asciigraph.Plot(data,
        asciigraph.Height(10),
        asciigraph.Width(60),
        asciigraph.Caption("Evolução do Patrimônio (R$)"),
    )

    return graph
}

// Example output:
//  150,000 ┤                                                  ╭─
//  140,000 ┤                                            ╭─────╯
//  130,000 ┤                                      ╭─────╯
//  120,000 ┤                                ╭─────╯
//  110,000 ┤                          ╭─────╯
//  100,000 ┤                    ╭─────╯
//   90,000 ┤              ╭─────╯
//   80,000 ┤        ╭─────╯
//   70,000 ┤  ╭─────╯
//   60,000 ┼──╯
//           Evolução do Patrimônio (R$)
```

**Gráfico de pizza (ASCII):**

```go
// PlotAllocationPie plota alocação em formato "pizza" ASCII
func PlotAllocationPie(allocation *AllocationBreakdown) string {
    var output strings.Builder

    output.WriteString("📊 Alocação da Carteira\n\n")

    for name, item := range allocation.ByType {
        // Criar barra proporcional
        barLength := int(item.Percentage.InexactFloat64() / 2)  // 50% = 25 chars
        bar := strings.Repeat("█", barLength)
        spaces := strings.Repeat("░", 50-barLength)

        output.WriteString(fmt.Sprintf(
            "%-20s %5.1f%% %s%s R$ %s\n",
            name,
            item.Percentage.InexactFloat64(),
            bar,
            spaces,
            item.TotalInvested.StringFixed(2),
        ))
    }

    return output.String()
}

// Output:
// 📊 Alocação da Carteira
//
// Ações               65.2% ████████████████████████████████░░░░░░░░░░░░░░░░░░ R$ 81,850.00
// FIIs                28.5% ██████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ R$ 35,750.00
// BDRs                 6.3% ███░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ R$  7,900.00
```

### 2. Relatórios PDF

**Biblioteca:** `github.com/go-pdf/fpdf`

```go
import "github.com/go-pdf/fpdf"

type PDFReport struct {
    pdf    *fpdf.Fpdf
    wallet *Wallet
}

func NewPDFReport(wallet *Wallet) *PDFReport {
    pdf := fpdf.New("P", "mm", "A4", "")
    pdf.SetFont("Arial", "", 12)

    return &PDFReport{
        pdf:    pdf,
        wallet: wallet,
    }
}

// GenerateMonthlyReport gera relatório mensal em PDF
func (r *PDFReport) GenerateMonthlyReport(year int, month int) error {
    r.pdf.AddPage()

    // Cabeçalho
    r.addHeader(fmt.Sprintf("Relatório Mensal - %s/%d", time.Month(month), year))

    // Resumo
    r.addSection("Resumo do Mês")
    r.addMonthSummary(year, month)

    // Transações
    r.addSection("Transações do Mês")
    r.addTransactionsTable(year, month)

    // Proventos
    r.addSection("Proventos Recebidos")
    r.addEarningsTable(year, month)

    // Alocação
    r.addSection("Alocação da Carteira")
    r.addAllocationChart()

    // Performance
    r.addSection("Performance")
    r.addPerformanceTable()

    return r.pdf.OutputFileAndClose(fmt.Sprintf("relatorio-%d-%02d.pdf", year, month))
}

func (r *PDFReport) addHeader(title string) {
    r.pdf.SetFont("Arial", "B", 20)
    r.pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
    r.pdf.Ln(5)

    // Data de geração
    r.pdf.SetFont("Arial", "I", 10)
    r.pdf.CellFormat(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")), "", 1, "C", false, 0, "")
    r.pdf.Ln(10)
}

func (r *PDFReport) addSection(title string) {
    r.pdf.SetFont("Arial", "B", 14)
    r.pdf.CellFormat(0, 8, title, "B", 1, "L", false, 0, "")
    r.pdf.Ln(3)
    r.pdf.SetFont("Arial", "", 10)
}

func (r *PDFReport) addMonthSummary(year, month int) {
    // Calcular métricas do mês
    summary := r.wallet.GetMonthlySummary(year, month)

    data := [][]string{
        {"Total Investido", fmt.Sprintf("R$ %s", summary.TotalInvested.StringFixed(2))},
        {"Valor de Mercado", fmt.Sprintf("R$ %s", summary.MarketValue.StringFixed(2))},
        {"Lucro/Prejuízo", fmt.Sprintf("R$ %s (%.2f%%)", summary.UnrealizedPL.StringFixed(2), summary.UnrealizedPLPct.InexactFloat64())},
        {"Proventos do Mês", fmt.Sprintf("R$ %s", summary.EarningsMonth.StringFixed(2))},
        {"Transações", fmt.Sprintf("%d compras, %d vendas", summary.PurchaseCount, summary.SaleCount)},
    }

    for _, row := range data {
        r.pdf.CellFormat(80, 6, row[0], "1", 0, "L", false, 0, "")
        r.pdf.CellFormat(0, 6, row[1], "1", 1, "R", false, 0, "")
    }

    r.pdf.Ln(5)
}

func (r *PDFReport) addTransactionsTable(year, month int) {
    transactions := r.wallet.GetTransactionsByMonth(year, month)

    // Cabeçalho da tabela
    r.pdf.SetFont("Arial", "B", 9)
    r.pdf.CellFormat(25, 6, "Data", "1", 0, "C", false, 0, "")
    r.pdf.CellFormat(20, 6, "Tipo", "1", 0, "C", false, 0, "")
    r.pdf.CellFormat(20, 6, "Ticker", "1", 0, "C", false, 0, "")
    r.pdf.CellFormat(20, 6, "Qtd", "1", 0, "C", false, 0, "")
    r.pdf.CellFormat(25, 6, "Preço", "1", 0, "C", false, 0, "")
    r.pdf.CellFormat(30, 6, "Valor", "1", 1, "C", false, 0, "")

    // Dados
    r.pdf.SetFont("Arial", "", 9)
    for _, tx := range transactions {
        r.pdf.CellFormat(25, 6, tx.Date.Format("02/01/2006"), "1", 0, "L", false, 0, "")
        r.pdf.CellFormat(20, 6, tx.Type, "1", 0, "C", false, 0, "")
        r.pdf.CellFormat(20, 6, tx.Ticker, "1", 0, "C", false, 0, "")
        r.pdf.CellFormat(20, 6, tx.Quantity.String(), "1", 0, "R", false, 0, "")
        r.pdf.CellFormat(25, 6, fmt.Sprintf("R$ %s", tx.Price.StringFixed(2)), "1", 0, "R", false, 0, "")
        r.pdf.CellFormat(30, 6, fmt.Sprintf("R$ %s", tx.Amount.StringFixed(2)), "1", 1, "R", false, 0, "")
    }

    r.pdf.Ln(5)
}
```

### 3. Dashboard HTML

**Geração de dashboard estático:**

```go
type HTMLDashboard struct {
    wallet *Wallet
}

const dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Dashboard - B3 Wallet</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 20px;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .card {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #eee;
        }
        .positive { color: #10b981; }
        .negative { color: #ef4444; }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #eee;
        }
        th {
            background: #f9fafb;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Dashboard de Investimentos</h1>
            <p>Atualizado em: {{.LastUpdate}}</p>
        </div>

        <div class="grid">
            <div class="card">
                <h2>💰 Resumo Geral</h2>
                <div class="metric">
                    <span>Total Investido</span>
                    <strong>R$ {{.TotalInvested}}</strong>
                </div>
                <div class="metric">
                    <span>Valor de Mercado</span>
                    <strong>R$ {{.MarketValue}}</strong>
                </div>
                <div class="metric">
                    <span>Lucro/Prejuízo</span>
                    <strong class="{{.PLClass}}">R$ {{.UnrealizedPL}} ({{.UnrealizedPLPct}}%)</strong>
                </div>
                <div class="metric">
                    <span>Proventos Totais</span>
                    <strong>R$ {{.TotalEarnings}}</strong>
                </div>
            </div>

            <div class="card">
                <h2>📈 Performance</h2>
                <canvas id="performanceChart"></canvas>
            </div>

            <div class="card">
                <h2>🥧 Alocação</h2>
                <canvas id="allocationChart"></canvas>
            </div>
        </div>

        <div class="card">
            <h2>📋 Ativos em Carteira</h2>
            <table>
                <thead>
                    <tr>
                        <th>Ticker</th>
                        <th>Quantidade</th>
                        <th>Preço Médio</th>
                        <th>Preço Atual</th>
                        <th>Investido</th>
                        <th>Valor Atual</th>
                        <th>Lucro/Prejuízo</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Assets}}
                    <tr>
                        <td><strong>{{.Ticker}}</strong></td>
                        <td>{{.Quantity}}</td>
                        <td>R$ {{.AvgPrice}}</td>
                        <td>R$ {{.CurrentPrice}}</td>
                        <td>R$ {{.Invested}}</td>
                        <td>R$ {{.CurrentValue}}</td>
                        <td class="{{.PLClass}}">R$ {{.PL}} ({{.PLPct}}%)</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>

    <script>
        // Performance Chart
        const perfCtx = document.getElementById('performanceChart').getContext('2d');
        new Chart(perfCtx, {
            type: 'line',
            data: {
                labels: {{.MonthLabels}},
                datasets: [{
                    label: 'Patrimônio',
                    data: {{.MonthValues}},
                    borderColor: '#667eea',
                    tension: 0.4
                }]
            }
        });

        // Allocation Chart
        const allocCtx = document.getElementById('allocationChart').getContext('2d');
        new Chart(allocCtx, {
            type: 'doughnut',
            data: {
                labels: {{.AllocationLabels}},
                datasets: [{
                    data: {{.AllocationValues}},
                    backgroundColor: ['#667eea', '#764ba2', '#f093fb', '#4facfe']
                }]
            }
        });
    </script>
</body>
</html>
`

// GenerateDashboard gera dashboard HTML
func (d *HTMLDashboard) GenerateDashboard(outputDir string) error {
    // Preparar dados
    data := d.prepareTemplateData()

    // Executar template
    tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
    if err != nil {
        return err
    }

    // Criar arquivo
    outputPath := filepath.Join(outputDir, "dashboard.html")
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    return tmpl.Execute(file, data)
}
```

### 4. Histórico de Evolução

**Snapshots periódicos:**

```go
type PortfolioSnapshot struct {
    Date            time.Time
    TotalInvested   decimal.Decimal
    MarketValue     decimal.Decimal
    UnrealizedPL    decimal.Decimal
    TotalEarnings   decimal.Decimal
    AssetCount      int
    TopHoldings     []string
}

// CreateSnapshot cria snapshot da carteira atual
func (w *Wallet) CreateSnapshot() *PortfolioSnapshot {
    snapshot := &PortfolioSnapshot{
        Date:          time.Now(),
        TotalInvested: decimal.Zero,
        MarketValue:   decimal.Zero,
        AssetCount:    len(w.GetActiveAssets()),
    }

    for _, asset := range w.GetActiveAssets() {
        snapshot.TotalInvested = snapshot.TotalInvested.Add(asset.TotalInvestedValue)
        snapshot.TotalEarnings = snapshot.TotalEarnings.Add(asset.TotalEarnings)

        if price, ok := w.GetCurrentPrice(asset.ID); ok {
            value := price.Mul(decimal.NewFromInt(int64(asset.Quantity)))
            snapshot.MarketValue = snapshot.MarketValue.Add(value)
        }
    }

    snapshot.UnrealizedPL = snapshot.MarketValue.Sub(snapshot.TotalInvested)

    w.Snapshots = append(w.Snapshots, snapshot)

    return snapshot
}

// GetEvolutionData retorna dados de evolução
func (w *Wallet) GetEvolutionData(months int) []PortfolioSnapshot {
    cutoff := time.Now().AddDate(0, -months, 0)

    result := make([]PortfolioSnapshot, 0)
    for _, snapshot := range w.Snapshots {
        if snapshot.Date.After(cutoff) {
            result = append(result, snapshot)
        }
    }

    return result
}
```

---

## 🎨 Interface do Usuário (CLI)

### Comandos propostos:

```bash
# Relatórios PDF
b3cli report monthly 2024-11 --output=relatorio-nov.pdf
b3cli report annual 2024 --output=relatorio-2024.pdf

# Dashboard HTML
b3cli dashboard generate --output=./dashboard/

# Gráficos no terminal
b3cli report evolution --months=12
b3cli report allocation --chart

# Snapshots
b3cli snapshot create
b3cli snapshot list
```

---

## 📊 Métricas de Sucesso

- ✅ PDFs profissionais gerados em < 5 segundos
- ✅ Dashboards acessíveis offline
- ✅ Gráficos ASCII legíveis
- ✅ Snapshots automáticos mensais

---

**Estimativa de implementação:** 2-3 semanas
**ROI para usuários:** Médio-Alto (profissionalismo + insights visuais)
