package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/shopspring/decimal"
)

// View states for earnings add TUI
type addEarningView int

const (
	viewAssetSelect addEarningView = iota
	viewEarningForm
	viewConfirmation
	viewSuccess
)

// Styles for the TUI
var (
	addTitleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	addLabelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	addValueStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	addSelectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	addHelpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	addErrorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	addSuccessStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	addHighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	addDocStyle       = lipgloss.NewStyle().Margin(1, 2)
)

// earningAssetItem implements list.Item for Bubble Tea list
type earningAssetItem struct {
	ticker   string
	quantity int
}

func (i earningAssetItem) FilterValue() string { return i.ticker }
func (i earningAssetItem) Title() string       { return i.ticker }
func (i earningAssetItem) Description() string {
	return fmt.Sprintf("%d shares", i.quantity)
}

// addEarningModel represents the state of the add earning TUI
type addEarningModel struct {
	wallet       *wallet.Wallet
	walletPath   string
	currentView  addEarningView

	// Asset selection
	assetList      list.Model
	selectedTicker string

	// Form inputs
	quantityInput textinput.Model
	totalInput    textinput.Model
	dateInput     textinput.Model
	earningTypes  []string
	selectedType  int
	focusedField  int

	// State
	err   error
	saved bool
}

// newAddEarningModel creates a new model with active assets
func newAddEarningModel(w *wallet.Wallet, walletPath string) addEarningModel {
	// Get active assets and create list items
	activeAssets := w.GetActiveAssets()
	items := []list.Item{}
	tickers := make([]string, 0, len(activeAssets))

	for ticker := range activeAssets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	for _, ticker := range tickers {
		asset := activeAssets[ticker]
		items = append(items, earningAssetItem{
			ticker:   ticker,
			quantity: asset.Quantity,
		})
	}

	// Configure list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Asset"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	// Configure input fields
	quantityInput := textinput.New()
	quantityInput.Placeholder = "e.g., 100 or 100.5"
	quantityInput.CharLimit = 15
	quantityInput.Width = 30

	totalInput := textinput.New()
	totalInput.Placeholder = "e.g., 150.75"
	totalInput.CharLimit = 15
	totalInput.Width = 30

	dateInput := textinput.New()
	dateInput.Placeholder = "YYYY-MM-DD (empty for today)"
	dateInput.CharLimit = 10
	dateInput.Width = 30

	earningTypes := []string{
		"Dividendo",
		"Juros Sobre Capital Próprio",
		"Rendimento",
	}

	return addEarningModel{
		wallet:        w,
		walletPath:    walletPath,
		currentView:   viewAssetSelect,
		assetList:     l,
		quantityInput: quantityInput,
		totalInput:    totalInput,
		dateInput:     dateInput,
		earningTypes:  earningTypes,
		selectedType:  0,
	}
}

func (m addEarningModel) Init() tea.Cmd {
	return nil
}

func (m addEarningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.currentView == viewAssetSelect {
			h, v := addDocStyle.GetFrameSize()
			m.assetList.SetSize(msg.Width-h, msg.Height-v)
		}
		return m, nil

	case tea.KeyMsg:
		switch m.currentView {
		case viewAssetSelect:
			return m.updateAssetSelectView(msg)
		case viewEarningForm:
			return m.updateEarningFormView(msg)
		case viewConfirmation:
			return m.updateConfirmationView(msg)
		case viewSuccess:
			return m, tea.Quit
		}
	}

	// Update list in asset selection view
	if m.currentView == viewAssetSelect {
		var cmd tea.Cmd
		m.assetList, cmd = m.assetList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m addEarningModel) updateAssetSelectView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "enter":
		selected := m.assetList.SelectedItem()
		if selected != nil {
			item := selected.(earningAssetItem)
			m.selectedTicker = item.ticker
			m.currentView = viewEarningForm
			m.focusedField = 0
			m.quantityInput.Focus()
			m.totalInput.Blur()
			m.dateInput.Blur()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.assetList, cmd = m.assetList.Update(msg)
	return m, cmd
}

func (m addEarningModel) updateEarningFormView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc":
		m.currentView = viewAssetSelect
		m.err = nil
		return m, nil

	case "enter":
		// Proceed to confirmation
		m.currentView = viewConfirmation
		return m, nil

	case "tab", "shift+tab", "up", "down":
		// Navigate between fields
		s := msg.String()

		if s == "up" || s == "shift+tab" {
			m.focusedField--
		} else {
			m.focusedField++
		}

		// Wrap around (4 fields: quantity, total, type, date)
		if m.focusedField > 3 {
			m.focusedField = 0
		} else if m.focusedField < 0 {
			m.focusedField = 3
		}

		// Update focus
		m.quantityInput.Blur()
		m.totalInput.Blur()
		m.dateInput.Blur()

		switch m.focusedField {
		case 0:
			m.quantityInput.Focus()
		case 1:
			m.totalInput.Focus()
		case 3:
			m.dateInput.Focus()
		}

		return m, nil

	case "left", "right":
		// Cycle earning type (only when type field is focused)
		if m.focusedField == 2 {
			if msg.String() == "right" {
				m.selectedType = (m.selectedType + 1) % len(m.earningTypes)
			} else {
				m.selectedType--
				if m.selectedType < 0 {
					m.selectedType = len(m.earningTypes) - 1
				}
			}
		}
		return m, nil
	}

	// Handle character input for text fields
	var cmd tea.Cmd
	switch m.focusedField {
	case 0:
		m.quantityInput, cmd = m.quantityInput.Update(msg)
	case 1:
		m.totalInput, cmd = m.totalInput.Update(msg)
	case 3:
		m.dateInput, cmd = m.dateInput.Update(msg)
	}

	return m, cmd
}

func (m addEarningModel) updateConfirmationView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "n", "N":
		// Go back to form
		m.currentView = viewEarningForm
		m.err = nil
		return m, nil

	case "enter", "y", "Y":
		if m.err != nil {
			// Error shown, go back to form
			m.currentView = viewEarningForm
			m.err = nil
			return m, nil
		}

		// Save earning
		if err := m.saveEarning(); err != nil {
			m.err = err
			return m, nil
		}

		m.saved = true
		m.currentView = viewSuccess
		return m, tea.Quit
	}

	return m, nil
}

func (m addEarningModel) View() string {
	switch m.currentView {
	case viewAssetSelect:
		return m.viewAssetSelect()
	case viewEarningForm:
		return m.viewEarningForm()
	case viewConfirmation:
		return m.viewConfirmation()
	case viewSuccess:
		return m.viewSuccess()
	}
	return ""
}

func (m addEarningModel) viewAssetSelect() string {
	var b strings.Builder

	b.WriteString(addTitleStyle.Render("Add Earning - Select Asset"))
	b.WriteString("\n\n")
	b.WriteString(m.assetList.View())
	b.WriteString("\n")
	b.WriteString(addHelpStyle.Render("enter: select • /: search • q: quit"))

	return addDocStyle.Render(b.String())
}

func (m addEarningModel) viewEarningForm() string {
	var b strings.Builder

	b.WriteString(addTitleStyle.Render(fmt.Sprintf("Add Earning - %s", m.selectedTicker)))
	b.WriteString("\n\n")

	// Quantity field
	if m.focusedField == 0 {
		b.WriteString(addSelectedStyle.Render("► Quantity:"))
	} else {
		b.WriteString(addLabelStyle.Render("  Quantity:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.quantityInput.View())
	b.WriteString("\n\n")

	// Total Value field
	if m.focusedField == 1 {
		b.WriteString(addSelectedStyle.Render("► Total Value:"))
	} else {
		b.WriteString(addLabelStyle.Render("  Total Value:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.totalInput.View())
	b.WriteString("\n\n")

	// Type field (cycling)
	if m.focusedField == 2 {
		b.WriteString(addSelectedStyle.Render("► Type:"))
	} else {
		b.WriteString(addLabelStyle.Render("  Type:"))
	}
	b.WriteString("\n  ")
	typeDisplay := fmt.Sprintf("◄ %s ►", m.earningTypes[m.selectedType])
	if m.focusedField == 2 {
		b.WriteString(addSelectedStyle.Render(typeDisplay))
		b.WriteString(addHelpStyle.Render(" (←/→ to change)"))
	} else {
		b.WriteString(addValueStyle.Render(m.earningTypes[m.selectedType]))
	}
	b.WriteString("\n\n")

	// Date field
	if m.focusedField == 3 {
		b.WriteString(addSelectedStyle.Render("► Date:"))
	} else {
		b.WriteString(addLabelStyle.Render("  Date:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.dateInput.View())
	b.WriteString("\n\n")

	b.WriteString(addHelpStyle.Render("tab/↑/↓: navigate • enter: next • esc: back • q: quit"))

	return addDocStyle.Render(b.String())
}

func (m addEarningModel) viewConfirmation() string {
	var b strings.Builder

	// Validate and calculate
	quantity, err := decimal.NewFromString(strings.TrimSpace(m.quantityInput.Value()))
	if err != nil || quantity.LessThanOrEqual(decimal.Zero) {
		m.err = fmt.Errorf("invalid quantity. Must be a positive number")
	}

	totalAmount, err := decimal.NewFromString(strings.TrimSpace(m.totalInput.Value()))
	if err != nil || totalAmount.LessThanOrEqual(decimal.Zero) {
		if m.err == nil {
			m.err = fmt.Errorf("invalid total value. Must be a positive number")
		}
	}

	var date time.Time
	dateStr := strings.TrimSpace(m.dateInput.Value())
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			if m.err == nil {
				m.err = fmt.Errorf("invalid date format. Use YYYY-MM-DD")
			}
		}
	}

	// Calculate unit price
	var unitPrice decimal.Decimal
	if m.err == nil && !quantity.IsZero() {
		unitPrice = totalAmount.Div(quantity).Round(4)
	}

	// Show error if exists
	if m.err != nil {
		b.WriteString(addErrorStyle.Render("Error"))
		b.WriteString("\n\n")
		b.WriteString(addErrorStyle.Render(m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(addLabelStyle.Render("Press Enter to go back, Esc to cancel"))
		b.WriteString("\n")
		return addDocStyle.Render(b.String())
	}

	// Show confirmation summary
	b.WriteString(addTitleStyle.Render("Confirm Earning"))
	b.WriteString("\n\n")

	b.WriteString(addLabelStyle.Render("Ticker:       "))
	b.WriteString(addValueStyle.Render(m.selectedTicker))
	b.WriteString("\n")

	b.WriteString(addLabelStyle.Render("Quantity:     "))
	b.WriteString(addValueStyle.Render(quantity.StringFixed(4)))
	b.WriteString("\n")

	b.WriteString(addLabelStyle.Render("Total Value:  "))
	b.WriteString(addValueStyle.Render("R$ " + totalAmount.StringFixed(2)))
	b.WriteString("\n")

	b.WriteString(addLabelStyle.Render("Unit Price:   "))
	b.WriteString(addHighlightStyle.Render("R$ " + unitPrice.StringFixed(4)))
	b.WriteString(addHelpStyle.Render(" (calculated)"))
	b.WriteString("\n")

	b.WriteString(addLabelStyle.Render("Type:         "))
	b.WriteString(addValueStyle.Render(m.earningTypes[m.selectedType]))
	b.WriteString("\n")

	b.WriteString(addLabelStyle.Render("Date:         "))
	b.WriteString(addValueStyle.Render(date.Format("2006-01-02")))
	b.WriteString("\n\n")

	b.WriteString(addSuccessStyle.Render("✓ Proceed with this earning?"))
	b.WriteString("\n")
	b.WriteString(addHelpStyle.Render("enter/Y: confirm • N/esc: back • q: quit"))
	b.WriteString("\n")

	return addDocStyle.Render(b.String())
}

func (m addEarningModel) viewSuccess() string {
	var b strings.Builder

	b.WriteString(addSuccessStyle.Render("✓ Earning added successfully!"))
	b.WriteString("\n\n")

	b.WriteString(addValueStyle.Render(fmt.Sprintf("%s - %s", m.selectedTicker, m.earningTypes[m.selectedType])))
	b.WriteString("\n")

	totalAmount, _ := decimal.NewFromString(strings.TrimSpace(m.totalInput.Value()))
	dateStr := strings.TrimSpace(m.dateInput.Value())
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		date, _ = time.Parse("2006-01-02", dateStr)
	}

	b.WriteString(addValueStyle.Render(fmt.Sprintf("R$ %s on %s", totalAmount.StringFixed(2), date.Format("2006-01-02"))))
	b.WriteString("\n")

	return addDocStyle.Render(b.String())
}

// calculateUnitPrice calculates unit price from total and quantity
func (m addEarningModel) calculateUnitPrice() (decimal.Decimal, error) {
	total, err := decimal.NewFromString(strings.TrimSpace(m.totalInput.Value()))
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid total value")
	}

	quantity, err := decimal.NewFromString(strings.TrimSpace(m.quantityInput.Value()))
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid quantity")
	}

	if quantity.IsZero() {
		return decimal.Zero, fmt.Errorf("quantity cannot be zero")
	}

	return total.Div(quantity).Round(4), nil
}

// createEarning creates an Earning struct from the form inputs
func (m addEarningModel) createEarning() (parser.Earning, error) {
	// Parse inputs
	quantity, err := decimal.NewFromString(strings.TrimSpace(m.quantityInput.Value()))
	if err != nil {
		return parser.Earning{}, fmt.Errorf("invalid quantity")
	}

	totalAmount, err := decimal.NewFromString(strings.TrimSpace(m.totalInput.Value()))
	if err != nil {
		return parser.Earning{}, fmt.Errorf("invalid total value")
	}

	unitPrice := totalAmount.Div(quantity).Round(4)

	// Parse date
	var date time.Time
	dateStr := strings.TrimSpace(m.dateInput.Value())
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return parser.Earning{}, fmt.Errorf("invalid date format: %w", err)
		}
	}

	return parser.Earning{
		Date:        date,
		Type:        m.earningTypes[m.selectedType],
		Ticker:      m.selectedTicker,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalAmount: totalAmount,
		// Hash will be calculated by wallet.AddEarning
	}, nil
}

// saveEarning saves the earning to the wallet
func (m *addEarningModel) saveEarning() error {
	earning, err := m.createEarning()
	if err != nil {
		return err
	}

	// Use wallet method (validates, deduplicates, recalculates)
	if err := m.wallet.AddEarning(earning); err != nil {
		return err
	}

	// Save wallet to disk
	if err := m.wallet.Save(m.walletPath); err != nil {
		return fmt.Errorf("failed to save wallet: %w", err)
	}

	return nil
}
