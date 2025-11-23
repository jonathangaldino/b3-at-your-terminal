package main

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/wallet"
)

type parseResultsModel struct {
	wallet *wallet.Wallet
}

var (
	parseResultsTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

	parseResultsHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("141")).
				MarginTop(1)

	parseResultsTickerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("117")).
				Bold(true)

	parseResultsValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))

	parseResultsLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	parseResultsTypeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226"))

	parseResultsTableHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("141")).
					Bold(true)

	parseResultsHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)
)

func initialParseResultsModel(w *wallet.Wallet) parseResultsModel {
	return parseResultsModel{
		wallet: w,
	}
}

func (m parseResultsModel) Init() tea.Cmd {
	return nil
}

func (m parseResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m parseResultsModel) View() string {
	var b strings.Builder

	// Título
	b.WriteString(parseResultsTitleStyle.Render("📊 Resumo do Processamento"))
	b.WriteString("\n\n")

	// Resumo geral
	b.WriteString(parseResultsHeaderStyle.Render("=== RESUMO ==="))
	b.WriteString("\n")
	b.WriteString(parseResultsLabelStyle.Render(fmt.Sprintf("Total de transações únicas: ")))
	b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%d", len(m.wallet.Transactions))))
	b.WriteString("\n")

	totalEarnings := 0
	for _, asset := range m.wallet.Assets {
		totalEarnings += len(asset.Earnings)
	}
	b.WriteString(parseResultsLabelStyle.Render(fmt.Sprintf("Total de proventos: ")))
	b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%d", totalEarnings)))
	b.WriteString("\n")

	b.WriteString(parseResultsLabelStyle.Render(fmt.Sprintf("Total de ativos diferentes: ")))
	b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%d", len(m.wallet.Assets))))
	b.WriteString("\n")

	// Seção de ativos
	b.WriteString("\n")
	b.WriteString(parseResultsHeaderStyle.Render("=== ATIVOS ==="))
	b.WriteString("\n")

	// Ordenar tickers
	tickers := make([]string, 0, len(m.wallet.Assets))
	for ticker := range m.wallet.Assets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	for _, ticker := range tickers {
		asset := m.wallet.Assets[ticker]
		b.WriteString("\n")
		b.WriteString(parseResultsTickerStyle.Render(fmt.Sprintf("[%s]", ticker)))
		b.WriteString(" - ")
		b.WriteString(parseResultsTypeStyle.Render(asset.Type))
		b.WriteString("\n")

		b.WriteString(parseResultsLabelStyle.Render("  Negociações: "))
		b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%d", len(asset.Negotiations))))
		b.WriteString("\n")

		b.WriteString(parseResultsLabelStyle.Render("  Preço Médio: "))
		b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("R$ %s", asset.AveragePrice.StringFixed(4))))
		b.WriteString("\n")

		b.WriteString(parseResultsLabelStyle.Render("  Valor Total Investido: "))
		b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("R$ %s", asset.TotalInvestedValue.StringFixed(2))))
		b.WriteString("\n")

		b.WriteString(parseResultsLabelStyle.Render("  Quantidade em carteira: "))
		b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%d", asset.Quantity)))
		b.WriteString("\n")

		if len(asset.Earnings) > 0 {
			b.WriteString(parseResultsLabelStyle.Render("  Proventos recebidos: "))
			b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%d (Total: R$ %s)", len(asset.Earnings), asset.TotalEarnings.StringFixed(2))))
			b.WriteString("\n")
		}
	}

	// Seção de transações (apenas as últimas 10)
	b.WriteString("\n")
	b.WriteString(parseResultsHeaderStyle.Render("=== TRANSAÇÕES (Últimas 10) ==="))
	b.WriteString("\n")

	// Cabeçalho da tabela
	b.WriteString(parseResultsTableHeaderStyle.Render(fmt.Sprintf("%-20s | %-10s | %-6s | %-8s | %6s | %8s | %12s",
		"Hash", "Data", "Tipo", "Ticker", "Qtd", "Preço", "Valor")))
	b.WriteString("\n")
	b.WriteString(parseResultsLabelStyle.Render(strings.Repeat("─", 90)))
	b.WriteString("\n")

	// Mostrar últimas 10 transações
	start := 0
	if len(m.wallet.Transactions) > 10 {
		start = len(m.wallet.Transactions) - 10
	}

	for i := start; i < len(m.wallet.Transactions); i++ {
		t := m.wallet.Transactions[i]
		hashShort := t.Hash[:16] + "..."
		if len(t.Hash) < 16 {
			hashShort = t.Hash
		}

		b.WriteString(parseResultsLabelStyle.Render(fmt.Sprintf("%-20s | ", hashShort)))
		b.WriteString(parseResultsLabelStyle.Render(fmt.Sprintf("%-10s | ", t.Date.Format("02/01/2006"))))
		b.WriteString(parseResultsTypeStyle.Render(fmt.Sprintf("%-6s", t.Type)))
		b.WriteString(parseResultsLabelStyle.Render(" | "))
		b.WriteString(parseResultsTickerStyle.Render(fmt.Sprintf("%-8s", t.Ticker)))
		b.WriteString(parseResultsLabelStyle.Render(" | "))
		b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%6s", t.Quantity.StringFixed(0))))
		b.WriteString(parseResultsLabelStyle.Render(" | "))
		b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%8s", t.Price.StringFixed(4))))
		b.WriteString(parseResultsLabelStyle.Render(" | "))
		b.WriteString(parseResultsValueStyle.Render(fmt.Sprintf("%12s", t.Amount.StringFixed(2))))
		b.WriteString("\n")
	}

	if len(m.wallet.Transactions) > 10 {
		b.WriteString("\n")
		b.WriteString(parseResultsLabelStyle.Render(fmt.Sprintf("... e mais %d transações", len(m.wallet.Transactions)-10)))
		b.WriteString("\n")
	}

	b.WriteString(parseResultsHelpStyle.Render("\nq/esc: sair"))

	return docStyle.Render(b.String())
}
