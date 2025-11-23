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
	mode          viewMode
	list          list.Model
	wallet        *wallet.Wallet
	walletPath    string
	selectedAsset *wallet.Asset
	typeInput     textinput.Model
	subTypeInput  textinput.Model
	segmentInput  textinput.Model
	focusIndex    int
	err           error
	saved         bool
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
		return m, nil

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

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m model) View() string {
	if m.mode == viewList {
		return m.viewList()
	}
	return m.viewEdit()
}

func (m model) viewList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Gerenciar Ativos"))
	b.WriteString("\n\n")

	if m.saved {
		b.WriteString(selectedItemStyle.Render("✓ Ativo atualizado com sucesso!"))
		b.WriteString("\n\n")
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

	b.WriteString(helpStyle.Render("tab/↑/↓: navegar • enter: salvar • esc: voltar • ctrl+c: sair"))

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
