package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/john/b3-project/internal/config"
	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
)

var assetsBuyCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy assets",
	Long:  "Interactive interface to buy assets and add them to your wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		walletPath, err := config.GetCurrentWallet()
		if err != nil {
			return fmt.Errorf("no wallet is currently open. Use 'b3cli wallet open <path>' first")
		}

		w, err := wallet.Load(walletPath)
		if err != nil {
			return fmt.Errorf("failed to load wallet: %w", err)
		}

		p := tea.NewProgram(newBuyModel(w, walletPath))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}

		return nil
	},
}

var assetsSellCmd = &cobra.Command{
	Use:   "sell",
	Short: "Sell assets",
	Long:  "Interactive interface to sell assets from your wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		walletPath, err := config.GetCurrentWallet()
		if err != nil {
			return fmt.Errorf("no wallet is currently open. Use 'b3cli wallet open <path>' first")
		}

		w, err := wallet.Load(walletPath)
		if err != nil {
			return fmt.Errorf("failed to load wallet: %w", err)
		}

		p := tea.NewProgram(newSellModel(w, walletPath))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}

		return nil
	},
}

type transactionType int

const (
	buyTransaction transactionType = iota
	sellTransaction
)

type transactModel struct {
	wallet       *wallet.Wallet
	walletPath   string
	txType       transactionType
	currentField int
	inputs       []textinput.Model
	err          error
	confirmed    bool
	showSummary  bool
	cancelled    bool
}

func newBuyModel(w *wallet.Wallet, walletPath string) transactModel {
	inputs := make([]textinput.Model, 4)

	// Ticker
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "e.g., BBAS3"
	inputs[0].Focus()
	inputs[0].CharLimit = 10
	inputs[0].Width = 30

	// Date
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "YYYY-MM-DD (empty for today)"
	inputs[1].CharLimit = 10
	inputs[1].Width = 30

	// Quantity
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "e.g., 100"
	inputs[2].CharLimit = 10
	inputs[2].Width = 30

	// Unit Price
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "e.g., 25.50"
	inputs[3].CharLimit = 15
	inputs[3].Width = 30

	return transactModel{
		wallet:     w,
		walletPath: walletPath,
		txType:     buyTransaction,
		inputs:     inputs,
	}
}

func newSellModel(w *wallet.Wallet, walletPath string) transactModel {
	inputs := make([]textinput.Model, 4)

	// Ticker
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "e.g., BBAS3"
	inputs[0].Focus()
	inputs[0].CharLimit = 10
	inputs[0].Width = 30

	// Date
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "YYYY-MM-DD (empty for today)"
	inputs[1].CharLimit = 10
	inputs[1].Width = 30

	// Quantity
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "e.g., 100"
	inputs[2].CharLimit = 10
	inputs[2].Width = 30

	// Unit Price
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "e.g., 25.50"
	inputs[3].CharLimit = 15
	inputs[3].Width = 30

	return transactModel{
		wallet:     w,
		walletPath: walletPath,
		txType:     sellTransaction,
		inputs:     inputs,
	}
}

func (m transactModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m transactModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.showSummary {
				m.showSummary = false
				m.err = nil
				return m, nil
			}
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			if m.showSummary {
				if m.err != nil {
					// Error shown, go back to edit
					m.showSummary = false
					m.err = nil
					return m, nil
				}
				// Confirm transaction
				if err := m.saveTransaction(); err != nil {
					m.err = err
					return m, nil
				}
				m.confirmed = true
				return m, tea.Quit
			}

			// Move to next field or show summary
			if m.currentField < len(m.inputs)-1 {
				m.currentField++
				m.inputs[m.currentField].Focus()
				for i := 0; i < len(m.inputs); i++ {
					if i != m.currentField {
						m.inputs[i].Blur()
					}
				}
				return m, nil
			}

			// All fields filled, show summary
			m.showSummary = true
			return m, nil

		case "tab", "shift+tab":
			if m.showSummary {
				return m, nil
			}

			if msg.String() == "tab" {
				m.currentField = (m.currentField + 1) % len(m.inputs)
			} else {
				m.currentField--
				if m.currentField < 0 {
					m.currentField = len(m.inputs) - 1
				}
			}

			for i := 0; i < len(m.inputs); i++ {
				if i == m.currentField {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, nil

		case "n":
			if m.showSummary {
				m.showSummary = false
				m.err = nil
				return m, nil
			}
		}
	}

	// Handle character input
	if !m.showSummary {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	return m, nil
}

func (m *transactModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m transactModel) View() string {
	if m.cancelled {
		return "Operation cancelled.\n"
	}

	if m.confirmed {
		return fmt.Sprintf("✓ Transaction saved successfully!\n")
	}

	if m.showSummary {
		return m.renderSummary()
	}

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	if m.txType == buyTransaction {
		b.WriteString(titleStyle.Render("Buy Assets"))
	} else {
		b.WriteString(titleStyle.Render("Sell Assets"))
	}
	b.WriteString("\n\n")

	labels := []string{"Ticker:", "Date:", "Quantity:", "Unit Price:"}
	for i, label := range labels {
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to continue, Esc to cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m transactModel) renderSummary() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	if m.err != nil {
		b.WriteString(errorStyle.Render("Error"))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(labelStyle.Render("Press Enter to go back, Esc to cancel"))
		b.WriteString("\n")
		return b.String()
	}

	// Parse inputs
	ticker := parser.NormalizeTicker(m.inputs[0].Value())
	dateStr := strings.TrimSpace(m.inputs[1].Value())
	quantityStr := strings.TrimSpace(m.inputs[2].Value())
	priceStr := strings.TrimSpace(m.inputs[3].Value())

	var date time.Time
	var err error
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			m.err = fmt.Errorf("invalid date format. Use YYYY-MM-DD")
			return m.renderSummary()
		}
	}

	quantity, err := decimal.NewFromString(quantityStr)
	if err != nil || quantity.LessThanOrEqual(decimal.Zero) {
		m.err = fmt.Errorf("invalid quantity. Must be a positive number")
		return m.renderSummary()
	}

	price, err := decimal.NewFromString(priceStr)
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		m.err = fmt.Errorf("invalid price. Must be a positive number")
		return m.renderSummary()
	}

	amount := quantity.Mul(price)

	// Validate for sell using wallet method
	if m.txType == sellTransaction {
		if err := m.wallet.CanSell(ticker, quantity); err != nil {
			m.err = err
			return m.renderSummary()
		}
	}

	b.WriteString(titleStyle.Render("Transaction Summary"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Ticker:       "))
	b.WriteString(valueStyle.Render(ticker))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Date:         "))
	b.WriteString(valueStyle.Render(date.Format("2006-01-02")))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Quantity:     "))
	b.WriteString(valueStyle.Render(quantity.StringFixed(4)))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Unit Price:   "))
	b.WriteString(valueStyle.Render("R$ " + price.StringFixed(2)))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Total Amount: "))
	b.WriteString(valueStyle.Render("R$ " + amount.StringFixed(2)))
	b.WriteString("\n\n")

	// Show comparison with average price for buy operations
	if m.txType == buyTransaction {
		asset, exists := m.wallet.Assets[ticker]
		if exists && asset.AveragePrice.GreaterThan(decimal.Zero) {
			b.WriteString(labelStyle.Render("Current Average Price: "))
			b.WriteString(valueStyle.Render("R$ " + asset.AveragePrice.StringFixed(2)))
			b.WriteString("\n")

			diff := price.Sub(asset.AveragePrice)
			diffPercent := diff.Div(asset.AveragePrice).Mul(decimal.NewFromInt(100))

			if price.GreaterThan(asset.AveragePrice) {
				b.WriteString(warningStyle.Render(fmt.Sprintf("⚠ Buying ABOVE average price (+%.2f%%, +R$ %s)",
					diffPercent.InexactFloat64(), diff.StringFixed(2))))
			} else if price.LessThan(asset.AveragePrice) {
				b.WriteString(successStyle.Render(fmt.Sprintf("✓ Buying BELOW average price (%.2f%%, R$ %s)",
					diffPercent.InexactFloat64(), diff.StringFixed(2))))
			} else {
				b.WriteString(valueStyle.Render("= Buying at EXACT average price"))
			}
			b.WriteString("\n\n")
		}
	}

	if m.txType == sellTransaction {
		asset, _ := m.wallet.Assets[ticker]
		b.WriteString(labelStyle.Render(fmt.Sprintf("Remaining after sale: %d shares",
			asset.Quantity-int(quantity.IntPart()))))
		b.WriteString("\n\n")
	}

	b.WriteString(successStyle.Render("Proceed with this transaction?"))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Press Enter to confirm, N to edit, Esc to cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m *transactModel) saveTransaction() error {
	// Parse inputs
	ticker := parser.NormalizeTicker(m.inputs[0].Value())
	dateStr := strings.TrimSpace(m.inputs[1].Value())
	quantityStr := strings.TrimSpace(m.inputs[2].Value())
	priceStr := strings.TrimSpace(m.inputs[3].Value())

	var date time.Time
	var err error
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format")
		}
	}

	quantity, err := decimal.NewFromString(quantityStr)
	if err != nil {
		return fmt.Errorf("invalid quantity")
	}

	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		return fmt.Errorf("invalid price")
	}

	amount := quantity.Mul(price)

	// Create transaction
	txType := "Compra"
	if m.txType == sellTransaction {
		txType = "Venda"
	}

	transaction := parser.Transaction{
		Date:        date,
		Type:        txType,
		Institution: "Manual Entry",
		Ticker:      ticker,
		Quantity:    quantity,
		Price:       price,
		Amount:      amount,
	}

	// Use wallet method to add transaction (handles validation, dedup, recalc)
	if err := m.wallet.AddTransaction(transaction); err != nil {
		return err
	}

	// Save wallet
	if err := m.wallet.Save(m.walletPath); err != nil {
		return fmt.Errorf("failed to save wallet: %w", err)
	}

	return nil
}

func init() {
	assetsCmd.AddCommand(assetsBuyCmd)
	assetsCmd.AddCommand(assetsSellCmd)
}
