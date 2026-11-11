package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/wallet"
)

type distributionView int

const (
	viewSubType distributionView = iota
	viewSegment
)

type distributionModel struct {
	subTypeDistribution []wallet.DistributionItem
	segmentDistribution []wallet.DistributionItem
	currentView         distributionView
	width               int
	height              int
}

func initialDistributionModel(subType, segment []wallet.DistributionItem) distributionModel {
	return distributionModel{
		subTypeDistribution: subType,
		segmentDistribution: segment,
		currentView:         viewSubType,
	}
}

func (m distributionModel) Init() tea.Cmd {
	return nil
}

func (m distributionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "right", "l"))):
			// Switch to next view
			if m.currentView == viewSubType {
				m.currentView = viewSegment
			} else {
				m.currentView = viewSubType
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "left", "h"))):
			// Switch to previous view
			if m.currentView == viewSegment {
				m.currentView = viewSubType
			} else {
				m.currentView = viewSegment
			}
			return m, nil
		}
	}

	return m, nil
}

func (m distributionModel) View() string {
	var content strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginBottom(1)

	content.WriteString(titleStyle.Render("Portfolio Distribution"))
	content.WriteString("\n\n")

	// Tabs
	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 2)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 2)

	subTypeTab := "By SubType"
	segmentTab := "By Segment"

	if m.currentView == viewSubType {
		content.WriteString(activeTabStyle.Render(subTypeTab))
		content.WriteString(inactiveTabStyle.Render(segmentTab))
	} else {
		content.WriteString(inactiveTabStyle.Render(subTypeTab))
		content.WriteString(activeTabStyle.Render(segmentTab))
	}
	content.WriteString("\n\n")

	// Distribution content
	if m.currentView == viewSubType {
		content.WriteString(m.renderDistribution(m.subTypeDistribution, "SubType"))
	} else {
		content.WriteString(m.renderDistribution(m.segmentDistribution, "Segment"))
	}

	// Footer
	content.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(helpStyle.Render("Tab/←/→: switch view • q/Esc: quit"))

	return content.String()
}

func (m distributionModel) renderDistribution(items []wallet.DistributionItem, label string) string {
	if len(items) == 0 {
		noDataStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		return noDataStyle.Render(fmt.Sprintf("No %s data available. Use 'b3cli assets manage' to set %s for your assets.", label, label))
	}

	var content strings.Builder

	// Check if all items are "(not set)"
	allNotSet := true
	for _, item := range items {
		if item.Category != "" {
			allNotSet = false
			break
		}
	}

	if allNotSet && len(items) == 1 {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

		content.WriteString(warningStyle.Render("⚠ No assets have " + label + " configured"))
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render(fmt.Sprintf("To see your portfolio distribution, you need to configure the %s field for your assets.", label)))
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("Use the following command to configure your assets:"))
		content.WriteString("\n")

		cmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		content.WriteString("  " + cmdStyle.Render("b3cli assets manage"))
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("This will open an interactive interface where you can set:"))
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("  • SubType: e.g., 'stocks', 'real estate funds'"))
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("  • Segment: e.g., 'banking', 'technology', 'energy'"))

		return content.String()
	}

	// Table header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Underline(true)

	content.WriteString(headerStyle.Render(fmt.Sprintf("%-40s %12s %18s", label, "Percentage", "Amount (R$)")))
	content.WriteString("\n\n")

	// Calculate column widths
	maxWidth := 40
	if m.width > 0 && m.width < 80 {
		maxWidth = m.width - 35 // Reserve space for percentage and amount
		if maxWidth < 20 {
			maxWidth = 20
		}
	}

	// Table rows
	categoryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	percentageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	amountStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	for _, item := range items {
		category := item.Category
		if category == "" {
			category = "(not set)"
		}

		// Truncate category if too long
		if len(category) > maxWidth {
			category = category[:maxWidth-3] + "..."
		}

		// Format percentage
		percentage := fmt.Sprintf("%.2f%%", item.Percentage.InexactFloat64())

		// Format amount
		amount := fmt.Sprintf("R$ %s", item.Amount.StringFixed(2))

		// Render row
		content.WriteString(
			categoryStyle.Render(fmt.Sprintf("%-40s", category)) + " " +
				percentageStyle.Render(fmt.Sprintf("%11s", percentage)) + " " +
				amountStyle.Render(fmt.Sprintf("%18s", amount)),
		)
		content.WriteString("\n")

		// Render bar chart
		barWidth := 50
		filledWidth := int(item.Percentage.InexactFloat64() / 100.0 * float64(barWidth))
		if filledWidth > barWidth {
			filledWidth = barWidth
		}

		barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
		emptyBarStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)
		content.WriteString(barStyle.Render(bar[:filledWidth]))
		content.WriteString(emptyBarStyle.Render(bar[filledWidth:]))
		content.WriteString("\n\n")
	}

	return content.String()
}
