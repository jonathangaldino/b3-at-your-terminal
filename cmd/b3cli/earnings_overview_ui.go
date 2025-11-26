package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/wallet"
	"github.com/shopspring/decimal"
)

type overviewModel struct {
	wallet     *wallet.Wallet
	categories map[string]*EarningsByType
	types      []string
	total      decimal.Decimal
	totalCount int
}

type EarningsByType struct {
	Count       int
	TotalAmount decimal.Decimal
	Assets      map[string]decimal.Decimal
}

var (
	overviewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1).
				MarginTop(1)

	overviewHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("141")).
				MarginTop(1)

	overviewValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true)

	overviewLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	overviewPercentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226"))

	overviewAssetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("117"))

	overviewSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
)

func initialOverviewModel(w *wallet.Wallet) overviewModel {
	categories := map[string]*EarningsByType{
		"Rendimento":                  {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
		"Dividendo":                   {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
		"Juros Sobre Capital Próprio": {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
		"Resgate":                     {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
	}

	totalGeneral := decimal.Zero
	totalCount := 0

	// Agrupar earnings por tipo
	for ticker, asset := range w.Assets {
		for _, earning := range asset.Earnings {
			if cat, exists := categories[earning.Type]; exists {
				cat.Count++
				cat.TotalAmount = cat.TotalAmount.Add(earning.TotalAmount)

				if current, ok := cat.Assets[ticker]; ok {
					cat.Assets[ticker] = current.Add(earning.TotalAmount)
				} else {
					cat.Assets[ticker] = earning.TotalAmount
				}

				totalGeneral = totalGeneral.Add(earning.TotalAmount)
				totalCount++
			}
		}
	}

	types := []string{"Rendimento", "Dividendo", "Juros Sobre Capital Próprio", "Resgate"}

	return overviewModel{
		wallet:     w,
		categories: categories,
		types:      types,
		total:      totalGeneral,
		totalCount: totalCount,
	}
}

func (m overviewModel) Init() tea.Cmd {
	return nil
}

func (m overviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m overviewModel) View() string {
	var b strings.Builder

	// Título
	b.WriteString(overviewTitleStyle.Render("💰 Resumo Geral de Proventos"))
	b.WriteString("\n\n")

	// Resumo geral
	b.WriteString(overviewLabelStyle.Render("Total de pagamentos recebidos: "))
	b.WriteString(overviewValueStyle.Render(fmt.Sprintf("%d", m.totalCount)))
	b.WriteString("\n")

	b.WriteString(overviewLabelStyle.Render("Valor total recebido: "))
	b.WriteString(overviewValueStyle.Render(fmt.Sprintf("R$ %s", m.total.StringFixed(2))))
	b.WriteString("\n")

	typeLabels := map[string]string{
		"Rendimento":                  "📊 RENDIMENTOS",
		"Dividendo":                   "💵 DIVIDENDOS",
		"Juros Sobre Capital Próprio": "🏦 JUROS SOBRE CAPITAL PRÓPRIO (JCP)",
		"Resgate":                     "🔄 RESGATES",
	}

	// Exibir por categoria
	for _, earningType := range m.types {
		cat := m.categories[earningType]
		if cat.Count == 0 {
			continue
		}

		b.WriteString("\n")
		b.WriteString(overviewSeparatorStyle.Render(strings.Repeat("─", 60)))
		b.WriteString("\n\n")
		b.WriteString(overviewHeaderStyle.Render(typeLabels[earningType]))
		b.WriteString("\n\n")

		// Quantidade e valor
		b.WriteString(overviewLabelStyle.Render("  Quantidade de pagamentos: "))
		b.WriteString(overviewValueStyle.Render(fmt.Sprintf("%d", cat.Count)))
		b.WriteString("\n")

		b.WriteString(overviewLabelStyle.Render("  Valor total: "))
		b.WriteString(overviewValueStyle.Render(fmt.Sprintf("R$ %s", cat.TotalAmount.StringFixed(2))))
		b.WriteString("\n")

		// Percentual
		if !m.total.IsZero() {
			percentage := cat.TotalAmount.Div(m.total).Mul(decimal.NewFromInt(100))
			b.WriteString(overviewLabelStyle.Render("  Percentual do total: "))
			b.WriteString(overviewPercentStyle.Render(fmt.Sprintf("%.2f%%", percentage.InexactFloat64())))
			b.WriteString("\n")
		}

		// Listar ativos
		b.WriteString("\n")
		b.WriteString(overviewLabelStyle.Render("  Ativos que pagaram:"))
		b.WriteString("\n")

		// Criar slice para ordenar
		type AssetEarning struct {
			Ticker string
			Amount decimal.Decimal
		}
		var assetList []AssetEarning
		for ticker, amount := range cat.Assets {
			assetList = append(assetList, AssetEarning{ticker, amount})
		}

		// Ordenar por valor (maior primeiro)
		for i := 0; i < len(assetList); i++ {
			for j := i + 1; j < len(assetList); j++ {
				if assetList[j].Amount.GreaterThan(assetList[i].Amount) {
					assetList[i], assetList[j] = assetList[j], assetList[i]
				}
			}
		}

		// Exibir lista de ativos
		for _, item := range assetList {
			b.WriteString("    ")
			b.WriteString(overviewAssetStyle.Render(fmt.Sprintf("%-8s", item.Ticker)))
			b.WriteString("  ")
			b.WriteString(overviewValueStyle.Render(fmt.Sprintf("R$ %10s", item.Amount.StringFixed(2))))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(overviewSeparatorStyle.Render(strings.Repeat("─", 60)))
	b.WriteString("\n")
	b.WriteString(reportHelpStyle.Render("\nq/esc: sair"))

	return docStyle.Render(b.String())
}
