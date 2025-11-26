package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/wallet"
	"github.com/shopspring/decimal"
)

type reportType int

const (
	reportTypeSelect reportType = iota
	reportTypeAnnual
	reportTypeMonthly
	reportTypeYearSelect
)

type reportsModel struct {
	wallet       *wallet.Wallet
	currentView  reportType
	cursor       int
	selectedYear int
	years        []int
	err          error
}

var (
	reportTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	reportSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	reportNormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	reportHelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	reportValueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	reportYearStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
)

func initialReportsModel(w *wallet.Wallet) reportsModel {
	return reportsModel{
		wallet:      w,
		currentView: reportTypeSelect,
		cursor:      0,
	}
}

func (m reportsModel) Init() tea.Cmd {
	return nil
}

func (m reportsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			// Voltar para seleção anterior
			if m.currentView == reportTypeYearSelect {
				m.currentView = reportTypeSelect
				m.cursor = 1 // Voltar para "Mensal"
			} else if m.currentView == reportTypeAnnual || m.currentView == reportTypeMonthly {
				m.currentView = reportTypeSelect
				m.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.currentView == reportTypeSelect || m.currentView == reportTypeYearSelect {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.currentView == reportTypeSelect {
				if m.cursor < 1 {
					m.cursor++
				}
			} else if m.currentView == reportTypeYearSelect {
				if m.cursor < len(m.years)-1 {
					m.cursor++
				}
			}

		case "enter":
			return m.handleEnter()
		}
	}

	return m, nil
}

func (m reportsModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentView {
	case reportTypeSelect:
		if m.cursor == 0 {
			// Anual
			m.currentView = reportTypeAnnual
		} else {
			// Mensal - verificar se precisa selecionar ano
			years := m.getAvailableYears()
			if len(years) == 0 {
				m.err = fmt.Errorf("nenhum provento encontrado")
				return m, nil
			} else if len(years) == 1 {
				m.selectedYear = years[0]
				m.currentView = reportTypeMonthly
			} else {
				m.years = years
				m.cursor = len(years) - 1 // Default: ano mais recente
				m.currentView = reportTypeYearSelect
			}
		}

	case reportTypeYearSelect:
		m.selectedYear = m.years[m.cursor]
		m.currentView = reportTypeMonthly
	}

	return m, nil
}

func (m reportsModel) getAvailableYears() []int {
	yearsSet := make(map[int]bool)

	for _, asset := range m.wallet.Assets {
		for _, earning := range asset.Earnings {
			yearsSet[earning.Date.Year()] = true
		}
	}

	var years []int
	for year := range yearsSet {
		years = append(years, year)
	}

	// Ordenar anos
	for i := 0; i < len(years); i++ {
		for j := i + 1; j < len(years); j++ {
			if years[j] < years[i] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}

	return years
}

func (m reportsModel) View() string {
	switch m.currentView {
	case reportTypeSelect:
		return m.viewSelectType()
	case reportTypeAnnual:
		return m.viewAnnualReport()
	case reportTypeMonthly:
		return m.viewMonthlyReport()
	case reportTypeYearSelect:
		return m.viewYearSelect()
	}
	return ""
}

func (m reportsModel) viewSelectType() string {
	var b strings.Builder

	b.WriteString(reportTitleStyle.Render("📊 Relatórios de Proventos"))
	b.WriteString("\n\n")
	b.WriteString("Selecione o tipo de relatório:\n\n")

	// Opção Anual
	cursor := " "
	if m.cursor == 0 {
		cursor = "►"
		b.WriteString(reportSelectedStyle.Render(fmt.Sprintf("%s Anual (resumo por ano)", cursor)))
	} else {
		b.WriteString(reportNormalStyle.Render(fmt.Sprintf("%s Anual (resumo por ano)", cursor)))
	}
	b.WriteString("\n")

	// Opção Mensal
	cursor = " "
	if m.cursor == 1 {
		cursor = "►"
		b.WriteString(reportSelectedStyle.Render(fmt.Sprintf("%s Mensal (resumo por mês)", cursor)))
	} else {
		b.WriteString(reportNormalStyle.Render(fmt.Sprintf("%s Mensal (resumo por mês)", cursor)))
	}
	b.WriteString("\n")

	b.WriteString(reportHelpStyle.Render("\n↑/↓: navegar • enter: selecionar • q: sair"))

	return docStyle.Render(b.String())
}

func (m reportsModel) viewYearSelect() string {
	var b strings.Builder

	b.WriteString(reportTitleStyle.Render("📅 Selecione o Ano"))
	b.WriteString("\n\n")
	b.WriteString("Anos disponíveis:\n\n")

	for i, year := range m.years {
		cursor := " "
		yearStr := fmt.Sprintf("%d", year)
		if i == m.cursor {
			cursor = "►"
			b.WriteString(reportSelectedStyle.Render(fmt.Sprintf("%s %s", cursor, yearStr)))
		} else {
			b.WriteString(reportNormalStyle.Render(fmt.Sprintf("%s %s", cursor, yearStr)))
		}
		b.WriteString("\n")
	}

	b.WriteString(reportHelpStyle.Render("\n↑/↓: navegar • enter: selecionar • esc: voltar • q: sair"))

	return docStyle.Render(b.String())
}

func (m reportsModel) viewAnnualReport() string {
	var b strings.Builder

	// Agrupar proventos por ano
	yearlyData := make(map[int]decimal.Decimal)
	var years []int

	for _, asset := range m.wallet.Assets {
		for _, earning := range asset.Earnings {
			year := earning.Date.Year()
			if _, exists := yearlyData[year]; !exists {
				years = append(years, year)
				yearlyData[year] = decimal.Zero
			}
			yearlyData[year] = yearlyData[year].Add(earning.TotalAmount)
		}
	}

	// Ordenar anos
	for i := 0; i < len(years); i++ {
		for j := i + 1; j < len(years); j++ {
			if years[j] < years[i] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}

	b.WriteString(reportTitleStyle.Render("📈 Relatório Anual de Proventos"))
	b.WriteString("\n\n")

	totalGeral := decimal.Zero
	for _, year := range years {
		amount := yearlyData[year]
		totalGeral = totalGeral.Add(amount)

		yearLine := fmt.Sprintf("%d:", year)
		valueLine := fmt.Sprintf("R$ %s", amount.StringFixed(2))

		b.WriteString(reportYearStyle.Render(fmt.Sprintf("%-6s", yearLine)))
		b.WriteString(" ")
		b.WriteString(reportValueStyle.Render(valueLine))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 40))
	b.WriteString("\n")
	b.WriteString(reportSelectedStyle.Render(fmt.Sprintf("Total geral: R$ %s", totalGeral.StringFixed(2))))
	b.WriteString("\n")

	// Calcular média anual
	if len(years) > 0 {
		media := totalGeral.Div(decimal.NewFromInt(int64(len(years))))
		b.WriteString(reportNormalStyle.Render(fmt.Sprintf("Média anual: R$ %s", media.StringFixed(2))))
	}

	b.WriteString(reportHelpStyle.Render("\n\nesc: voltar • q: sair"))

	return docStyle.Render(b.String())
}

func (m reportsModel) viewMonthlyReport() string {
	var b strings.Builder

	monthlyData := make(map[int]decimal.Decimal)
	monthNames := []string{
		"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
		"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
	}

	for _, asset := range m.wallet.Assets {
		for _, earning := range asset.Earnings {
			if earning.Date.Year() == m.selectedYear {
				month := int(earning.Date.Month())
				if _, exists := monthlyData[month]; !exists {
					monthlyData[month] = decimal.Zero
				}
				monthlyData[month] = monthlyData[month].Add(earning.TotalAmount)
			}
		}
	}

	b.WriteString(reportTitleStyle.Render(fmt.Sprintf("📅 Relatório Mensal de Proventos - %d", m.selectedYear)))
	b.WriteString("\n\n")

	totalAnual := decimal.Zero
	monthsWithPayments := 0

	for month := 1; month <= 12; month++ {
		if amount, exists := monthlyData[month]; exists && !amount.IsZero() {
			totalAnual = totalAnual.Add(amount)
			monthsWithPayments++

			monthLine := fmt.Sprintf("%-12s:", monthNames[month-1])
			valueLine := fmt.Sprintf("R$ %s", amount.StringFixed(2))

			b.WriteString(reportNormalStyle.Render(monthLine))
			b.WriteString(" ")
			b.WriteString(reportValueStyle.Render(valueLine))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 40))
	b.WriteString("\n")
	b.WriteString(reportSelectedStyle.Render(fmt.Sprintf("Total do ano: R$ %s", totalAnual.StringFixed(2))))
	b.WriteString("\n")

	if monthsWithPayments > 0 {
		media := totalAnual.Div(decimal.NewFromInt(int64(monthsWithPayments)))
		b.WriteString(reportNormalStyle.Render(fmt.Sprintf("Média mensal (meses com pagamento): R$ %s", media.StringFixed(2))))
	}

	b.WriteString(reportHelpStyle.Render("\n\nesc: voltar • q: sair"))

	return docStyle.Render(b.String())
}
