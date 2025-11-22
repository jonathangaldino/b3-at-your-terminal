package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
)

var (
	assetsWalletPath string
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
	Short: "Exibe um resumo de todos os ativos da carteira",
	Long: `Exibe uma visão geral de todos os ativos na carteira, mostrando:
- Código de negociação (ticker)
- Quantidade de ativos em carteira
- Valor total investido (soma de todas as compras)
- Preço médio ponderado

A lista é ordenada alfabeticamente por ticker.`,
	Example: `  # Visualizar todos os ativos
  b3cli assets overview --wallet data

  # Visualizar ativos no diretório atual
  b3cli assets overview`,
	Args: cobra.NoArgs,
	RunE: runAssetsOverview,
}

func init() {
	assetsCmd.AddCommand(assetsSubscriptionCmd)
	assetsCmd.AddCommand(assetsOverviewCmd)

	assetsSubscriptionCmd.Flags().StringVarP(&assetsWalletPath, "wallet", "w", ".", "Diretório da wallet")
	assetsOverviewCmd.Flags().StringVarP(&assetsWalletPath, "wallet", "w", ".", "Diretório da wallet")
}

func runAssetsSubscription(cmd *cobra.Command, args []string) error {
	ticker := strings.ToUpper(args[0])
	subscriptionArg := args[1]

	// Parse subscription@parent
	parts := strings.Split(subscriptionArg, "@")
	if len(parts) != 2 || parts[0] != "subscription" {
		return fmt.Errorf("formato inválido. Use: subscription@<ticker-pai>")
	}

	parentTicker := strings.ToUpper(parts[1])

	// Validar que os tickers são diferentes
	if ticker == parentTicker {
		return fmt.Errorf("o ativo não pode ser subscrição de si mesmo")
	}

	// Obter diretório da wallet (usar flag)
	absPath, err := filepath.Abs(assetsWalletPath)
	if err != nil {
		return fmt.Errorf("erro ao resolver caminho: %w", err)
	}

	// Verificar se existe uma wallet
	if !wallet.Exists(absPath) {
		return fmt.Errorf("não foi encontrada uma wallet em %s. Use 'b3cli wallet create' primeiro", absPath)
	}

	// Carregar wallet
	w, err := wallet.Load(absPath)
	if err != nil {
		return fmt.Errorf("erro ao carregar wallet: %w", err)
	}

	// Converter subscrição para ativo pai
	fmt.Printf("Processando subscrição %s → %s...\n", ticker, parentTicker)
	result, err := w.ConvertSubscriptionToParent(ticker, parentTicker)
	if err != nil {
		return fmt.Errorf("erro ao converter subscrição: %w", err)
	}

	// Salvar wallet
	if err := w.Save(absPath); err != nil {
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
	fmt.Printf("✓ Wallet atualizada em: %s\n", filepath.Join(absPath, "wallet.yaml"))

	return nil
}

func runAssetsOverview(cmd *cobra.Command, args []string) error {
	// Obter diretório da wallet
	absPath, err := filepath.Abs(assetsWalletPath)
	if err != nil {
		return fmt.Errorf("erro ao resolver caminho: %w", err)
	}

	// Verificar se existe uma wallet
	if !wallet.Exists(absPath) {
		return fmt.Errorf("não foi encontrada uma wallet em %s. Use 'b3cli wallet create' primeiro", absPath)
	}

	// Carregar wallet
	w, err := wallet.Load(absPath)
	if err != nil {
		return fmt.Errorf("erro ao carregar wallet: %w", err)
	}

	// Verificar se há ativos
	if len(w.Assets) == 0 {
		fmt.Println("Nenhum ativo encontrado na carteira.")
		return nil
	}

	// Coletar e ordenar tickers
	tickers := make([]string, 0, len(w.Assets))
	for ticker := range w.Assets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	// Exibir cabeçalho
	fmt.Printf("\n=== RESUMO DE ATIVOS ===\n")
	fmt.Printf("Total de ativos: %d\n\n", len(w.Assets))

	// Exibir cada ativo
	for _, ticker := range tickers {
		asset := w.Assets[ticker]
		fmt.Printf("%s - %d ativos - R$ %s investido - PM: R$ %s\n",
			ticker,
			asset.Quantity,
			asset.TotalInvestedValue.StringFixed(2),
			asset.AveragePrice.StringFixed(4),
		)
	}

	fmt.Println()
	return nil
}
