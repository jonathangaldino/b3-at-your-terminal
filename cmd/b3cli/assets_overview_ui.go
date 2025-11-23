package main

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/wallet"
)

type assetsOverviewModel struct {
	wallet       *wallet.Wallet
	activeAssets map[string]*wallet.Asset
	soldAssets   map[string]*wallet.Asset
	groups       []groupInfo
}

type groupInfo struct {
	key     wallet.GroupKey
	assets  []*wallet.Asset
	tickers []string
}

var (
	assetsTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

	assetsGroupStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("141")).
				MarginTop(1)

	assetsTickerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("117")).
				Bold(true)

	assetsValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))

	assetsQuantityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226"))

	assetsPMStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))

	assetsHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginTop(1)

	assetsLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	assetsHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

func initialAssetsOverviewModel(w *wallet.Wallet) assetsOverviewModel {
	activeAssets := w.GetActiveAssets()
	soldAssets := w.GetSoldAssets()
	groups := w.GroupActiveAssetsByTypeAndSegment()

	// Sort groups
	sortedGroups := make([]groupInfo, 0, len(groups))
	for key, assets := range groups {
		tickers := make([]string, 0, len(assets))
		for ticker, asset := range w.Assets {
			for _, a := range assets {
				if a == asset {
					tickers = append(tickers, ticker)
					break
				}
			}
		}
		sort.Strings(tickers)

		sortedGroups = append(sortedGroups, groupInfo{
			key:     key,
			assets:  assets,
			tickers: tickers,
		})
	}

	sort.Slice(sortedGroups, func(i, j int) bool {
		if sortedGroups[i].key.Type != sortedGroups[j].key.Type {
			return sortedGroups[i].key.Type < sortedGroups[j].key.Type
		}
		return sortedGroups[i].key.Segment < sortedGroups[j].key.Segment
	})

	return assetsOverviewModel{
		wallet:       w,
		activeAssets: activeAssets,
		soldAssets:   soldAssets,
		groups:       sortedGroups,
	}
}

func (m assetsOverviewModel) Init() tea.Cmd {
	return nil
}

func (m assetsOverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m assetsOverviewModel) View() string {
	var b strings.Builder

	// Título
	b.WriteString(assetsTitleStyle.Render("📊 Resumo de Ativos"))
	b.WriteString("\n")
	b.WriteString(assetsLabelStyle.Render(fmt.Sprintf("Ativos em carteira: %d", len(m.activeAssets))))
	b.WriteString("\n")

	// Exibir cada grupo
	for _, group := range m.groups {
		subType := group.key.Type
		if subType == "" {
			subType = "(sem classificação)"
		}
		segment := group.key.Segment
		if segment == "" {
			segment = "(sem segmento)"
		}

		b.WriteString("\n")
		b.WriteString(assetsGroupStyle.Render(fmt.Sprintf("📁 %s / %s", subType, segment)))
		b.WriteString("\n")

		for _, ticker := range group.tickers {
			asset := m.wallet.Assets[ticker]

			b.WriteString("  ")
			b.WriteString(assetsTickerStyle.Render(fmt.Sprintf("%-8s", ticker)))
			b.WriteString(" ")
			b.WriteString(assetsQuantityStyle.Render(fmt.Sprintf("%3d", asset.Quantity)))
			b.WriteString(assetsLabelStyle.Render(" ativos"))
			b.WriteString(" • ")
			b.WriteString(assetsLabelStyle.Render("Investido: "))
			b.WriteString(assetsValueStyle.Render(fmt.Sprintf("R$ %10s", asset.TotalInvestedValue.StringFixed(2))))
			b.WriteString(" • ")
			b.WriteString(assetsLabelStyle.Render("PM: "))
			b.WriteString(assetsPMStyle.Render(fmt.Sprintf("R$ %8s", asset.AveragePrice.StringFixed(4))))
			b.WriteString("\n")
		}
	}

	// Hint sobre ativos vendidos
	if len(m.soldAssets) > 0 {
		b.WriteString("\n")
		b.WriteString(assetsHintStyle.Render(fmt.Sprintf("ℹ  Você possui %d ativo(s) vendido(s) completamente.", len(m.soldAssets))))
		b.WriteString("\n")
		b.WriteString(assetsHintStyle.Render("   Use 'b3cli assets sold' para visualizá-los."))
	}

	b.WriteString("\n")
	b.WriteString(assetsHelpStyle.Render("\nq/esc: sair"))

	return docStyle.Render(b.String())
}
