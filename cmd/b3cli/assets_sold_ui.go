package main

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/wallet"
)

type assetsSoldModel struct {
	soldAssets map[string]*wallet.Asset
	tickers    []string
}

var (
	soldTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	soldTickerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Bold(true)

	soldValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	soldLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	soldStatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Italic(true)

	soldHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginTop(1)

	soldHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

func initialAssetsSoldModel(soldAssets map[string]*wallet.Asset) assetsSoldModel {
	// Get tickers and sort
	tickers := make([]string, 0, len(soldAssets))
	for ticker := range soldAssets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	return assetsSoldModel{
		soldAssets: soldAssets,
		tickers:    tickers,
	}
}

func (m assetsSoldModel) Init() tea.Cmd {
	return nil
}

func (m assetsSoldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m assetsSoldModel) View() string {
	var b strings.Builder

	// Título
	b.WriteString(soldTitleStyle.Render("🔴 Ativos Vendidos Completamente"))
	b.WriteString("\n")
	b.WriteString(soldLabelStyle.Render(fmt.Sprintf("Total: %d", len(m.tickers))))
	b.WriteString("\n\n")

	// Exibir cada ativo vendido
	for _, ticker := range m.tickers {
		asset := m.soldAssets[ticker]

		b.WriteString(soldTickerStyle.Render(fmt.Sprintf("%-10s", ticker)))
		b.WriteString(" ")
		b.WriteString(soldStatusStyle.Render("Vendido"))
		b.WriteString("\n")

		b.WriteString(soldLabelStyle.Render("  Investido: "))
		b.WriteString(soldValueStyle.Render(fmt.Sprintf("R$ %s", asset.TotalInvestedValue.StringFixed(2))))
		b.WriteString(soldLabelStyle.Render(" • PM: "))
		b.WriteString(soldValueStyle.Render(fmt.Sprintf("R$ %s", asset.AveragePrice.StringFixed(4))))
		b.WriteString("\n\n")
	}

	// Hint sobre histórico
	b.WriteString(soldHintStyle.Render("ℹ  Estes ativos foram vendidos completamente mas seu histórico"))
	b.WriteString("\n")
	b.WriteString(soldHintStyle.Render("   de transações ainda está disponível em transactions.yaml"))
	b.WriteString("\n")

	b.WriteString(soldHelpStyle.Render("\nq/esc: sair"))

	return docStyle.Render(b.String())
}
