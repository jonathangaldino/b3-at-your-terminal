package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/john/b3-project/internal/wallet"
	"github.com/john/b3-project/internal/wallet/events"
)

type splitViewMode int

const (
	splitViewSelectAsset splitViewMode = iota
	splitViewInputs
	splitViewConfirm
	splitViewResult
)

type splitModel struct {
	walletPath    string
	wallet        *wallet.Wallet
	mode          splitViewMode
	assetList     list.Model
	selectedAsset string
	inputs        []textinput.Model
	currentInput  int
	result        *events.SplitResult
	err           error
	cancelled     bool
}

// Asset item for the list
type splitAssetItem struct {
	ticker string
	qty    int
	price  string
}

func (i splitAssetItem) FilterValue() string { return i.ticker }
func (i splitAssetItem) Title() string {
	return fmt.Sprintf("%s - %d shares", i.ticker, i.qty)
}
func (i splitAssetItem) Description() string {
	return fmt.Sprintf("Average Price: R$ %s", i.price)
}

func newSplitModel(walletPath string) splitModel {
	// Load wallet
	w, err := wallet.Load(walletPath)
	if err != nil {
		return splitModel{
			walletPath: walletPath,
			err:        fmt.Errorf("failed to load wallet: %w", err),
		}
	}

	// Build asset list (only active assets)
	items := []list.Item{}
	tickers := make([]string, 0)
	for ticker, asset := range w.Assets {
		if asset.Quantity > 0 {
			tickers = append(tickers, ticker)
		}
	}
	sort.Strings(tickers)

	for _, ticker := range tickers {
		asset := w.Assets[ticker]
		items = append(items, splitAssetItem{
			ticker: ticker,
			qty:    asset.Quantity,
			price:  asset.AveragePrice.StringFixed(2),
		})
	}

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Asset for Split"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	// Create text inputs
	inputs := make([]textinput.Model, 2)

	// Ratio input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "e.g., 1:2"
	inputs[0].CharLimit = 10
	inputs[0].Width = 30
	inputs[0].Focus()

	// Date input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "YYYY-MM-DD"
	inputs[1].CharLimit = 10
	inputs[1].Width = 30

	return splitModel{
		walletPath: walletPath,
		wallet:     w,
		mode:       splitViewSelectAsset,
		assetList:  l,
		inputs:     inputs,
	}
}

func (m splitModel) Init() tea.Cmd {
	return nil
}

func (m splitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.assetList.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case splitViewSelectAsset:
			return m.updateSelectAsset(msg)
		case splitViewInputs:
			return m.updateInputs(msg)
		case splitViewConfirm:
			return m.updateConfirm(msg)
		case splitViewResult:
			return m.updateResult(msg)
		}
	}

	// Update list if in select mode
	if m.mode == splitViewSelectAsset {
		var cmd tea.Cmd
		m.assetList, cmd = m.assetList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m splitModel) updateSelectAsset(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.cancelled = true
		return m, tea.Quit

	case "enter":
		// Get selected asset
		if item, ok := m.assetList.SelectedItem().(splitAssetItem); ok {
			m.selectedAsset = item.ticker
			m.mode = splitViewInputs
			m.inputs[0].Focus()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.assetList, cmd = m.assetList.Update(msg)
	return m, cmd
}

func (m splitModel) updateInputs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		// Go back to asset selection
		m.mode = splitViewSelectAsset
		m.currentInput = 0
		m.inputs[0].Focus()
		m.inputs[1].Blur()
		m.err = nil
		return m, nil

	case "enter", "tab", "down":
		if m.currentInput < len(m.inputs)-1 {
			// Move to next input
			m.currentInput++
			m.inputs[m.currentInput].Focus()
			for i := 0; i < len(m.inputs); i++ {
				if i != m.currentInput {
					m.inputs[i].Blur()
				}
			}
			return m, nil
		} else {
			// All inputs filled, show preview
			m.mode = splitViewConfirm
			return m, nil
		}

	case "up", "shift+tab":
		if m.currentInput > 0 {
			m.currentInput--
			m.inputs[m.currentInput].Focus()
			for i := 0; i < len(m.inputs); i++ {
				if i != m.currentInput {
					m.inputs[i].Blur()
				}
			}
		}
		return m, nil
	}

	// Update current input
	var cmd tea.Cmd
	m.inputs[m.currentInput], cmd = m.inputs[m.currentInput].Update(msg)
	return m, cmd
}

func (m splitModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "esc", "n":
		// Go back to inputs
		m.mode = splitViewInputs
		m.err = nil
		return m, nil

	case "enter", "y":
		// Apply split
		err := m.applySplit()
		if err != nil {
			m.err = err
			return m, nil
		}
		m.mode = splitViewResult
		return m, nil
	}

	return m, nil
}

func (m splitModel) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m splitModel) applySplit() error {
	// Parse ratio
	ratioStr := m.inputs[0].Value()
	ratio, err := events.ParseSplitRatio(ratioStr)
	if err != nil {
		return fmt.Errorf("invalid ratio: %w", err)
	}

	// Parse date
	dateStr := m.inputs[1].Value()
	eventDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
	}

	// Apply split
	result, err := events.ApplySplit(m.wallet, m.selectedAsset, ratio, eventDate)
	if err != nil {
		return fmt.Errorf("failed to apply split: %w", err)
	}

	m.result = result

	// Save wallet
	if err := m.wallet.Save(m.walletPath); err != nil {
		return fmt.Errorf("failed to save wallet: %w", err)
	}

	return nil
}

func (m splitModel) View() string {
	if m.err != nil && m.mode != splitViewConfirm && m.mode != splitViewResult {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress Enter to continue or Esc to cancel", m.err))
	}

	switch m.mode {
	case splitViewSelectAsset:
		return m.viewSelectAsset()
	case splitViewInputs:
		return m.viewInputs()
	case splitViewConfirm:
		return m.viewConfirm()
	case splitViewResult:
		return m.viewResult()
	}

	return ""
}

func (m splitModel) viewSelectAsset() string {
	return docStyle.Render(m.assetList.View())
}

func (m splitModel) viewInputs() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Apply Split (Stock Split)"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Asset: %s\n\n", selectedItemStyle.Render(m.selectedAsset)))

	b.WriteString("Split Ratio (e.g., 1:2, 1:3):\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("Event Date (YYYY-MM-DD):\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Tab/Enter: Next field • Shift+Tab/Up: Previous field • Esc: Back • Ctrl+C: Quit"))

	return docStyle.Render(b.String())
}

func (m splitModel) viewConfirm() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Preview: Split"))
	b.WriteString("\n\n")

	// Parse inputs for preview
	ratioStr := m.inputs[0].Value()
	ratio, err := events.ParseSplitRatio(ratioStr)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Invalid ratio: %v\n\nPress Esc to go back", err))
	}

	dateStr := m.inputs[1].Value()
	_, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Invalid date: %v\n\nPress Esc to go back", err))
	}

	asset := m.wallet.Assets[m.selectedAsset]

	b.WriteString(fmt.Sprintf("Asset: %s\n", selectedItemStyle.Render(m.selectedAsset)))
	b.WriteString(fmt.Sprintf("Ratio: %s\n", events.FormatSplitRatio(ratio)))
	b.WriteString(fmt.Sprintf("Event Date: %s\n\n", dateStr))

	// Calculate preview values
	multiplier := float64(ratio.To) / float64(ratio.From)
	newQty := float64(asset.Quantity) * multiplier
	newPrice, _ := asset.AveragePrice.Float64()
	newPrice /= multiplier

	b.WriteString("BEFORE:\n")
	b.WriteString(fmt.Sprintf("  Quantity: %d shares\n", asset.Quantity))
	b.WriteString(fmt.Sprintf("  Avg Price: R$ %s\n\n", asset.AveragePrice.StringFixed(2)))

	b.WriteString("AFTER:\n")
	b.WriteString(fmt.Sprintf("  Quantity: ~%.0f shares\n", newQty))
	b.WriteString(fmt.Sprintf("  Avg Price: ~R$ %.2f\n\n", newPrice))

	// Count transactions that will be affected
	txBefore := 0
	for _, tx := range asset.Negotiations {
		eventDate, _ := time.Parse("2006-01-02", dateStr)
		if tx.Date.Before(eventDate) {
			txBefore++
		}
	}

	b.WriteString(fmt.Sprintf("Transactions to adjust: %d\n\n", txBefore))

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v\n\n", m.err)))
	}

	b.WriteString(helpStyle.Render("Y/Enter: Apply • N/Esc: Cancel • Ctrl+C: Quit"))

	return docStyle.Render(b.String())
}

func (m splitModel) viewResult() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("✓ Split Applied Successfully"))
	b.WriteString("\n\n")

	if m.result != nil {
		b.WriteString(fmt.Sprintf("Asset: %s\n", selectedItemStyle.Render(m.result.Ticker)))
		b.WriteString(fmt.Sprintf("Ratio: %s\n", events.FormatSplitRatio(m.result.Ratio)))
		b.WriteString(fmt.Sprintf("Event Date: %s\n\n", m.result.EventDate.Format("2006-01-02")))

		b.WriteString("CHANGES:\n")
		b.WriteString(fmt.Sprintf("  Quantity: %d → %d shares\n", m.result.QuantityBefore, m.result.QuantityAfter))
		b.WriteString(fmt.Sprintf("  Avg Price: R$ %s → R$ %s\n", m.result.PriceBefore.StringFixed(2), m.result.PriceAfter.StringFixed(2)))
		b.WriteString(fmt.Sprintf("  Transactions adjusted: %d\n\n", m.result.TransactionsAdjusted))

		b.WriteString("✓ Wallet saved successfully\n\n")
	}

	b.WriteString(helpStyle.Render("Press Enter to exit"))

	return docStyle.Render(b.String())
}
