package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/john/b3-project/internal/config"
	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
)

// Estilos
var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type viewMode int

const (
	viewList viewMode = iota
	viewEdit
	viewConfirmMerge
	viewConfirmCreate
)

// assetItem implementa list.Item para bubbles/list
type assetItem struct {
	ticker string
	qty    int
	pm     string
}

func (i assetItem) FilterValue() string { return i.ticker }
func (i assetItem) Title() string {
	return fmt.Sprintf("%s (%d ativos)", i.ticker, i.qty)
}
func (i assetItem) Description() string {
	return fmt.Sprintf("PM: R$ %s", i.pm)
}

// model representa o estado da aplicação
type model struct {
	mode           viewMode
	list           list.Model
	wallet         *wallet.Wallet
	walletPath     string
	selectedAsset  *wallet.Asset
	typeInput      textinput.Model
	subTypeInput   textinput.Model
	segmentInput   textinput.Model
	focusIndex     int
	err            error
	saved          bool
	mergeResult    *wallet.MergeResult
	normalizedName string // Ticker sem F para confirmar criação
}

func initialModel(w *wallet.Wallet, walletPath string) model {
	// Criar lista de assets
	items := []list.Item{}
	tickers := make([]string, 0, len(w.Assets))
	for ticker := range w.Assets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	for _, ticker := range tickers {
		asset := w.Assets[ticker]
		// Incluir apenas ativos com quantity >= 1
		if asset.Quantity >= 1 {
			items = append(items, assetItem{
				ticker: ticker,
				qty:    asset.Quantity,
				pm:     asset.AveragePrice.StringFixed(4),
			})
		}
	}

	// Configurar lista
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Selecione um ativo para gerenciar"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	// Configurar inputs
	ti := textinput.New()
	ti.Placeholder = "Ex: renda variável"
	ti.CharLimit = 50

	sti := textinput.New()
	sti.Placeholder = "Ex: ações, fundos imobiliários"
	sti.CharLimit = 50

	si := textinput.New()
	si.Placeholder = "Ex: tecnologia, energia"
	si.CharLimit = 50

	return model{
		mode:         viewList,
		list:         l,
		wallet:       w,
		walletPath:   walletPath,
		typeInput:    ti,
		subTypeInput: sti,
		segmentInput: si,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case viewList:
			return m.updateListView(msg)
		case viewEdit:
			return m.updateEditView(msg)
		case viewConfirmMerge:
			return m.updateConfirmMergeView(msg)
		case viewConfirmCreate:
			return m.updateConfirmCreateView(msg)
		}
	}

	var cmd tea.Cmd
	if m.mode == viewList {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "enter":
		selected := m.list.SelectedItem()
		if selected != nil {
			item := selected.(assetItem)
			m.selectedAsset = m.wallet.Assets[item.ticker]
			m.mode = viewEdit
			m.focusIndex = 0

			// Preencher inputs com valores atuais
			m.typeInput.SetValue(m.selectedAsset.Type)
			m.typeInput.Focus()
			m.subTypeInput.SetValue(m.selectedAsset.SubType)
			m.subTypeInput.Blur()
			m.segmentInput.SetValue(m.selectedAsset.Segment)
			m.segmentInput.Blur()
			m.saved = false
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateEditView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.mode = viewList
		m.saved = false
		m.err = nil
		return m, nil

	case "f", "F":
		// Corrigir ativo fracionário (apenas se termina com F)
		if wallet.IsFractionalTicker(m.selectedAsset.ID) {
			m.mode = viewConfirmMerge
			m.err = nil
			return m, nil
		}

	case "tab", "shift+tab", "up", "down":
		// Navegar entre inputs
		s := msg.String()

		if s == "up" || s == "shift+tab" {
			m.focusIndex--
		} else {
			m.focusIndex++
		}

		if m.focusIndex > 2 {
			m.focusIndex = 0
		} else if m.focusIndex < 0 {
			m.focusIndex = 2
		}

		cmds := make([]tea.Cmd, 3)
		for i := 0; i < 3; i++ {
			if i == m.focusIndex {
				cmds[i] = m.getInput(i).Focus()
			} else {
				m.getInput(i).Blur()
			}
		}
		return m, tea.Batch(cmds...)

	case "enter":
		// Salvar mudanças
		m.selectedAsset.Type = m.typeInput.Value()
		m.selectedAsset.SubType = m.subTypeInput.Value()
		m.selectedAsset.Segment = m.segmentInput.Value()

		if err := m.wallet.Save(m.walletPath); err != nil {
			m.err = err
			return m, nil
		}

		m.saved = true
		m.mode = viewList
		return m, nil
	}

	// Atualizar input focado
	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.typeInput, cmd = m.typeInput.Update(msg)
	case 1:
		m.subTypeInput, cmd = m.subTypeInput.Update(msg)
	case 2:
		m.segmentInput, cmd = m.segmentInput.Update(msg)
	}

	return m, cmd
}

func (m model) updateConfirmMergeView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "n", "N":
		// Cancelar
		m.mode = viewEdit
		m.err = nil
		return m, nil

	case "y", "Y":
		// Confirmar merge
		result, err := m.wallet.MergeFractionalAsset(m.selectedAsset.ID)
		if err != nil {
			// Verificar se é erro de ativo não encontrado
			if len(err.Error()) > 16 && err.Error()[:16] == "TARGET_NOT_FOUND" {
				// Extrair ticker normalizado do erro
				m.normalizedName = err.Error()[17:] // Pula "TARGET_NOT_FOUND:"
				m.mode = viewConfirmCreate
				return m, nil
			}

			m.err = err
			return m, nil
		}

		// Merge bem-sucedido
		m.mergeResult = result

		// Salvar wallet
		if err := m.wallet.Save(m.walletPath); err != nil {
			m.err = err
			return m, nil
		}

		m.saved = true

		// Reconstruir lista de ativos para refletir as mudanças
		m.rebuildAssetList()

		m.mode = viewList
		return m, nil
	}

	return m, nil
}

func (m model) updateConfirmCreateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "n", "N":
		// Cancelar
		m.mode = viewEdit
		m.err = nil
		return m, nil

	case "y", "Y":
		// Criar e mesclar
		result, err := m.wallet.CreateAndMergeFractionalAsset(m.selectedAsset.ID)
		if err != nil {
			m.err = err
			return m, nil
		}

		// Merge bem-sucedido
		m.mergeResult = result

		// Salvar wallet
		if err := m.wallet.Save(m.walletPath); err != nil {
			m.err = err
			return m, nil
		}

		m.saved = true

		// Reconstruir lista de ativos para refletir as mudanças
		m.rebuildAssetList()

		m.mode = viewList
		return m, nil
	}

	return m, nil
}

func (m *model) getInput(i int) *textinput.Model {
	switch i {
	case 0:
		return &m.typeInput
	case 1:
		return &m.subTypeInput
	default:
		return &m.segmentInput
	}
}

// rebuildAssetList reconstrói a lista de ativos a partir da wallet atual
func (m *model) rebuildAssetList() {
	// Criar nova lista de items
	items := []list.Item{}
	tickers := make([]string, 0, len(m.wallet.Assets))
	for ticker := range m.wallet.Assets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	for _, ticker := range tickers {
		asset := m.wallet.Assets[ticker]
		// Incluir apenas ativos com quantity >= 1
		if asset.Quantity >= 1 {
			items = append(items, assetItem{
				ticker: ticker,
				qty:    asset.Quantity,
				pm:     asset.AveragePrice.StringFixed(4),
			})
		}
	}

	// Atualizar a lista
	m.list.SetItems(items)

	// Limpar seleção de asset
	m.selectedAsset = nil
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m model) View() string {
	switch m.mode {
	case viewList:
		return m.viewList()
	case viewEdit:
		return m.viewEdit()
	case viewConfirmMerge:
		return m.viewConfirmMerge()
	case viewConfirmCreate:
		return m.viewConfirmCreate()
	}
	return ""
}

func (m model) viewList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Gerenciar Ativos"))
	b.WriteString("\n\n")

	if m.saved {
		if m.mergeResult != nil {
			b.WriteString(selectedItemStyle.Render("✓ Ativo fracionário corrigido com sucesso!"))
			b.WriteString("\n")
			b.WriteString(helpStyle.Render(fmt.Sprintf("   %s → %s (%d transações, %d proventos)",
				m.mergeResult.SourceTicker,
				m.mergeResult.TargetTicker,
				m.mergeResult.TransactionsMoved,
				m.mergeResult.EarningsMoved)))
			b.WriteString("\n\n")
			m.mergeResult = nil // Reset
		} else {
			b.WriteString(selectedItemStyle.Render("✓ Ativo atualizado com sucesso!"))
			b.WriteString("\n\n")
		}
	}

	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: selecionar • q: sair"))

	return docStyle.Render(b.String())
}

func (m model) viewEdit() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("Editando: %s", m.selectedAsset.ID)))
	b.WriteString("\n\n")

	// Type
	if m.focusIndex == 0 {
		b.WriteString(selectedItemStyle.Render("► Type:"))
	} else {
		b.WriteString("  Type:")
	}
	b.WriteString("\n  ")
	b.WriteString(m.typeInput.View())
	b.WriteString("\n\n")

	// SubType
	if m.focusIndex == 1 {
		b.WriteString(selectedItemStyle.Render("► SubType:"))
	} else {
		b.WriteString("  SubType:")
	}
	b.WriteString("\n  ")
	b.WriteString(m.subTypeInput.View())
	b.WriteString("\n\n")

	// Segment
	if m.focusIndex == 2 {
		b.WriteString(selectedItemStyle.Render("► Segment:"))
	} else {
		b.WriteString("  Segment:")
	}
	b.WriteString("\n  ")
	b.WriteString(m.segmentInput.View())
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Erro: %s", m.err)))
		b.WriteString("\n\n")
	}

	// Mostrar botão de correção apenas para ativos fracionários
	if wallet.IsFractionalTicker(m.selectedAsset.ID) {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		b.WriteString(warningStyle.Render("⚠ Este ativo tem 'F' no final (mercado fracionário)"))
		b.WriteString("\n\n")
	}

	helpText := "tab/↑/↓: navegar • enter: salvar • esc: voltar"
	if wallet.IsFractionalTicker(m.selectedAsset.ID) {
		helpText += " • F: corrigir fracionário"
	}
	b.WriteString(helpStyle.Render(helpText))

	return docStyle.Render(b.String())
}

func (m model) viewConfirmMerge() string {
	var b strings.Builder

	normalizedTicker := parser.NormalizeTicker(m.selectedAsset.ID)

	b.WriteString(titleStyle.Render("Corrigir Ativo Fracionário"))
	b.WriteString("\n\n")

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	b.WriteString(warningStyle.Render(fmt.Sprintf("Você está prestes a corrigir: %s → %s", m.selectedAsset.ID, normalizedTicker)))
	b.WriteString("\n\n")

	b.WriteString("Esta operação irá:\n")
	b.WriteString(fmt.Sprintf("  • Remover o 'F' do ticker (%s → %s)\n", m.selectedAsset.ID, normalizedTicker))
	b.WriteString(fmt.Sprintf("  • Mesclar %d transações no ativo %s\n", len(m.selectedAsset.Negotiations), normalizedTicker))
	if len(m.selectedAsset.Earnings) > 0 {
		b.WriteString(fmt.Sprintf("  • Mesclar %d proventos no ativo %s\n", len(m.selectedAsset.Earnings), normalizedTicker))
	}
	b.WriteString(fmt.Sprintf("  • Recalcular preço médio e quantidade de %s\n", normalizedTicker))
	b.WriteString(fmt.Sprintf("  • Remover o ativo %s da carteira\n", m.selectedAsset.ID))
	b.WriteString("\n")

	// Verificar se ativo de destino existe
	if _, exists := m.wallet.Assets[normalizedTicker]; exists {
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
		b.WriteString(successStyle.Render(fmt.Sprintf("✓ O ativo %s já existe na carteira.", normalizedTicker)))
		b.WriteString("\n")
		b.WriteString("  As transações serão mescladas.\n")
	} else {
		b.WriteString(errorStyle.Render(fmt.Sprintf("⚠ O ativo %s NÃO existe na carteira.", normalizedTicker)))
		b.WriteString("\n")
		b.WriteString("  Você será perguntado se deseja criar.\n")
	}

	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Erro: %s", m.err)))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("Y: confirmar • N/Esc: cancelar"))

	return docStyle.Render(b.String())
}

func (m model) viewConfirmCreate() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Criar Ativo Original?"))
	b.WriteString("\n\n")

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	b.WriteString(warningStyle.Render(fmt.Sprintf("O ativo %s não existe na carteira.", m.normalizedName)))
	b.WriteString("\n\n")

	b.WriteString("Deseja criar o ativo e mesclar as transações?\n\n")

	b.WriteString("O novo ativo terá:\n")
	b.WriteString(fmt.Sprintf("  • Ticker: %s\n", m.normalizedName))
	b.WriteString(fmt.Sprintf("  • %d transações de %s\n", len(m.selectedAsset.Negotiations), m.selectedAsset.ID))
	if len(m.selectedAsset.Earnings) > 0 {
		b.WriteString(fmt.Sprintf("  • %d proventos de %s\n", len(m.selectedAsset.Earnings), m.selectedAsset.ID))
	}
	b.WriteString(fmt.Sprintf("  • Type: %s\n", m.selectedAsset.Type))
	b.WriteString(fmt.Sprintf("  • SubType: %s\n", m.selectedAsset.SubType))
	b.WriteString(fmt.Sprintf("  • Segment: %s\n", m.selectedAsset.Segment))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Erro: %s", m.err)))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("Y: criar e mesclar • N/Esc: cancelar"))

	return docStyle.Render(b.String())
}

// runAssetsManage executa o comando interativo de gerenciamento de assets
func runAssetsManage(cmd *cobra.Command, args []string) error {
	// Obter wallet atual
	absPath, err := config.GetCurrentWallet()
	if err != nil {
		return err
	}

	// Carregar wallet
	w, err := wallet.Load(absPath)
	if err != nil {
		return fmt.Errorf("erro ao carregar wallet: %w", err)
	}

	// Verificar se há ativos ativos (quantity >= 1)
	activeCount := 0
	for _, asset := range w.Assets {
		if asset.Quantity >= 1 {
			activeCount++
		}
	}

	if activeCount == 0 {
		fmt.Println("Nenhum ativo ativo encontrado na carteira.")
		fmt.Println("Use 'b3cli assets overview' para ver todos os ativos.")
		return nil
	}

	// Iniciar aplicação TUI
	p := tea.NewProgram(initialModel(w, absPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar interface: %w", err)
	}

	return nil
}
