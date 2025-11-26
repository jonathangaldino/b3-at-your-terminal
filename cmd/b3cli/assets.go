package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/john/b3-project/internal/parser"
	"github.com/spf13/cobra"
)

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Gerencia ativos individuais da carteira",
	Long:  `Comandos para visualizar e gerenciar ativos individuais da sua carteira de investimentos.`,
}

var assetsSubscriptionCmd = &cobra.Command{
	Use:   "subscription [ticker] [subscription@parent]",
	Short: "Marca um ativo como direito de subscrição de outro ativo",
	Long: `Marca um ativo como sendo um direito de subscrição de outro ativo.

Direitos de subscrição são direitos de compra de novas ações/cotas emitidas por uma empresa
ou fundo. Quando você recebe ou vende direitos de subscrição, eles aparecem como um ticker
separado (geralmente terminando em 11, 12, etc.).

Este comando permite vincular o direito de subscrição ao ativo original.`,
	Example: `  # Marcar MXRF12 como subscrição de MXRF11
  b3cli assets subscription MXRF12 subscription@MXRF11

  # Marcar PETR12 como subscrição de PETR4
  b3cli assets subscription PETR12 subscription@PETR4`,
	Args: cobra.ExactArgs(2),
	RunE: runAssetsSubscription,
}

var assetsOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Exibe um resumo dos ativos ativos da carteira",
	Long: `Exibe uma visão geral dos ativos que você possui atualmente (quantity != 0), mostrando:
- Código de negociação (ticker)
- Quantidade de ativos em carteira
- Valor total investido (soma de todas as compras)
- Preço médio ponderado

A lista é ordenada alfabeticamente por ticker.

IMPORTANTE: Você deve ter aberto uma wallet antes de usar este comando.
Use 'b3cli wallet open <diretório>' para abrir uma wallet.`,
	Example: `  b3cli assets overview`,
	Args:    cobra.NoArgs,
	RunE:    runAssetsOverview,
}

var assetsSoldCmd = &cobra.Command{
	Use:   "sold",
	Short: "Exibe ativos que foram vendidos completamente",
	Long: `Exibe uma lista de ativos que foram vendidos completamente (quantity == 0).

Estes ativos não aparecem mais na carteira principal, mas seu histórico
de transações e informações são mantidos para referência futura.

A lista mostra:
- Código de negociação (ticker)
- Valor total que foi investido
- Preço médio que foi pago
- Quantidade vendida

IMPORTANTE: Você deve ter aberto uma wallet antes de usar este comando.
Use 'b3cli wallet open <diretório>' para abrir uma wallet.`,
	Example: `  b3cli assets sold`,
	Args:    cobra.NoArgs,
	RunE:    runAssetsSold,
}

var assetsManageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Gerencia ativos interativamente",
	Long: `Interface interativa para gerenciar metadados dos ativos da carteira.

Permite navegar pela lista de ativos e editar:
- Type: tipo do ativo (ex: renda variável, renda fixa)
- SubType: subtipo (ex: ações, fundos imobiliários)
- Segment: segmento (ex: tecnologia, energia)

Navegação:
- ↑/↓ ou j/k: navegar pela lista
- Enter: selecionar ativo para editar
- Tab/↑/↓: navegar entre campos de edição
- Enter: salvar alterações
- Esc: voltar para lista
- q ou Ctrl+C: sair

IMPORTANTE: Você deve ter aberto uma wallet antes de usar este comando.
Use 'b3cli wallet open <diretório>' para abrir uma wallet.`,
	Example: `  b3cli assets manage`,
	Args:    cobra.NoArgs,
	RunE:    runAssetsManage,
}

func init() {
	assetsCmd.AddCommand(assetsSubscriptionCmd)
	assetsCmd.AddCommand(assetsOverviewCmd)
	assetsCmd.AddCommand(assetsSoldCmd)
	assetsCmd.AddCommand(assetsManageCmd)
}

func runAssetsSubscription(cmd *cobra.Command, args []string) error {
	ticker := parser.NormalizeTicker(args[0])
	subscriptionArg := args[1]

	// Parse subscription@parent
	parts := strings.Split(subscriptionArg, "@")
	if len(parts) != 2 || parts[0] != "subscription" {
		return fmt.Errorf("formato inválido. Use: subscription@<ticker-pai>")
	}

	parentTicker := parser.NormalizeTicker(parts[1])

	// Validar que os tickers são diferentes
	if ticker == parentTicker {
		return fmt.Errorf("o ativo não pode ser subscrição de si mesmo")
	}

	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Converter subscrição para ativo pai
	fmt.Printf("Processando subscrição %s → %s...\n", ticker, parentTicker)
	result, err := w.ConvertSubscriptionToParent(ticker, parentTicker)
	if err != nil {
		return fmt.Errorf("erro ao converter subscrição: %w", err)
	}

	// Salvar wallet
	if err := w.Save(w.GetDirPath()); err != nil {
		return fmt.Errorf("erro ao salvar wallet: %w", err)
	}

	// Exibir resultados
	fmt.Println()
	fmt.Printf("✓ Processamento concluído:\n")
	fmt.Printf("  - Compras encontradas: %d\n", result.PurchasesFound)
	fmt.Printf("  - Vendas encontradas: %d (ignoradas)\n", result.SalesFound)
	fmt.Printf("  - Transações transferidas: %d\n", result.TransactionsAdded)
	fmt.Println()
	fmt.Printf("✓ Ativo %s removido da carteira\n", ticker)
	fmt.Printf("✓ Ativo %s atualizado:\n", parentTicker)
	fmt.Printf("  - Quantidade antes: %d\n", result.ParentQuantityBefore)
	fmt.Printf("  - Quantidade depois: %d\n", result.ParentQuantityAfter)
	fmt.Printf("  - Preço médio: R$ %s\n", result.ParentAveragePrice.StringFixed(4))
	fmt.Println()
	fmt.Printf("✓ Wallet atualizada em: %s\n", w.GetDirPath())

	return nil
}

func runAssetsOverview(cmd *cobra.Command, args []string) error {
	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Get active assets using wallet method
	activeAssets := w.GetActiveAssets()

	// Verificar se há ativos ativos
	if len(activeAssets) == 0 {
		fmt.Println("\nNenhum ativo ativo encontrado na carteira.")
		soldAssets := w.GetSoldAssets()
		if len(soldAssets) > 0 {
			fmt.Printf("Você possui %d ativo(s) vendido(s) completamente.\n", len(soldAssets))
			fmt.Printf("Use 'b3cli assets sold' para visualizá-los.\n")
		}
		fmt.Println()
		return nil
	}

	// Iniciar interface Bubble Tea
	p := tea.NewProgram(initialAssetsOverviewModel(w), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar interface: %w", err)
	}

	return nil
}

func runAssetsSold(cmd *cobra.Command, args []string) error {
	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Get sold assets using wallet method
	soldAssets := w.GetSoldAssets()

	// Verificar se há ativos vendidos
	if len(soldAssets) == 0 {
		fmt.Println("\nNenhum ativo vendido completamente encontrado.")
		fmt.Println("Todos os ativos que você comprou ainda estão em carteira.")
		fmt.Println()
		return nil
	}

	// Iniciar interface Bubble Tea
	p := tea.NewProgram(initialAssetsSoldModel(soldAssets), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar interface: %w", err)
	}

	return nil
}
