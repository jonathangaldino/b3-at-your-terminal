package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/john/b3-project/internal/config"
	"github.com/john/b3-project/internal/wallet"
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
	Args: cobra.NoArgs,
	RunE: runAssetsOverview,
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
	Args: cobra.NoArgs,
	RunE: runAssetsSold,
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
	Args: cobra.NoArgs,
	RunE: runAssetsManage,
}

func init() {
	assetsCmd.AddCommand(assetsSubscriptionCmd)
	assetsCmd.AddCommand(assetsOverviewCmd)
	assetsCmd.AddCommand(assetsSoldCmd)
	assetsCmd.AddCommand(assetsManageCmd)
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

	// Obter wallet atual
	absPath, err := config.GetCurrentWallet()
	if err != nil {
		return err
	}

	// Verificar se a wallet existe
	if !wallet.Exists(absPath) {
		return fmt.Errorf("wallet não encontrada em %s", absPath)
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

// assetGroup representa um grupo de ativos por SubType e Segment
type assetGroup struct {
	subType string
	segment string
	assets  []*wallet.Asset
	tickers []string
}

func runAssetsOverview(cmd *cobra.Command, args []string) error {
	// Obter wallet atual
	absPath, err := config.GetCurrentWallet()
	if err != nil {
		return err
	}

	// Verificar se a wallet existe
	if !wallet.Exists(absPath) {
		return fmt.Errorf("wallet não encontrada em %s", absPath)
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

	// Agrupar ativos por (SubType, Segment)
	groups := make(map[string]*assetGroup)
	activeCount := 0

	for ticker, asset := range w.Assets {
		// Apenas ativos ativos (quantity != 0)
		if asset.Quantity == 0 {
			continue
		}

		activeCount++

		// Usar valores padrão se não estiverem definidos
		subType := asset.SubType
		if subType == "" {
			subType = "(sem classificação)"
		}
		segment := asset.Segment
		if segment == "" {
			segment = "(sem segmento)"
		}

		// Criar chave única para o grupo
		key := fmt.Sprintf("%s|%s", subType, segment)

		// Criar grupo se não existir
		if _, exists := groups[key]; !exists {
			groups[key] = &assetGroup{
				subType: subType,
				segment: segment,
				assets:  make([]*wallet.Asset, 0),
				tickers: make([]string, 0),
			}
		}

		// Adicionar ativo ao grupo
		groups[key].assets = append(groups[key].assets, asset)
		groups[key].tickers = append(groups[key].tickers, ticker)
	}

	// Verificar se há ativos ativos
	if activeCount == 0 {
		fmt.Println("\nNenhum ativo ativo encontrado na carteira.")
		soldCount := len(w.Assets)
		if soldCount > 0 {
			fmt.Printf("Você possui %d ativo(s) vendido(s) completamente.\n", soldCount)
			fmt.Printf("Use 'b3cli assets sold' para visualizá-los.\n")
		}
		fmt.Println()
		return nil
	}

	// Ordenar grupos por SubType e depois por Segment
	sortedKeys := make([]string, 0, len(groups))
	for key := range groups {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		gi := groups[sortedKeys[i]]
		gj := groups[sortedKeys[j]]
		if gi.subType != gj.subType {
			return gi.subType < gj.subType
		}
		return gi.segment < gj.segment
	})

	// Exibir cabeçalho
	fmt.Printf("\n=== RESUMO DE ATIVOS ===\n")
	fmt.Printf("Ativos em carteira: %d\n\n", activeCount)

	// Exibir cada grupo
	for _, key := range sortedKeys {
		group := groups[key]

		// Ordenar tickers dentro do grupo
		sort.Strings(group.tickers)

		// Exibir cabeçalho do grupo
		fmt.Printf("[%s / %s]\n", group.subType, group.segment)

		// Exibir cada ativo do grupo
		for i, ticker := range group.tickers {
			asset := group.assets[i]
			fmt.Printf("  %s - %d ativos - R$ %s investido - PM: R$ %s\n",
				ticker,
				asset.Quantity,
				asset.TotalInvestedValue.StringFixed(2),
				asset.AveragePrice.StringFixed(4),
			)
		}

		fmt.Println() // Linha em branco entre grupos
	}

	// Mostrar dica sobre assets vendidos se houver
	soldCount := len(w.Assets) - activeCount
	if soldCount > 0 {
		fmt.Printf("ℹ  Você possui %d ativo(s) vendido(s) completamente.\n", soldCount)
		fmt.Printf("   Use 'b3cli assets sold' para visualizá-los.\n\n")
	}

	return nil
}

func runAssetsSold(cmd *cobra.Command, args []string) error {
	// Obter wallet atual
	absPath, err := config.GetCurrentWallet()
	if err != nil {
		return err
	}

	// Verificar se a wallet existe
	if !wallet.Exists(absPath) {
		return fmt.Errorf("wallet não encontrada em %s", absPath)
	}

	// Carregar wallet
	w, err := wallet.Load(absPath)
	if err != nil {
		return fmt.Errorf("erro ao carregar wallet: %w", err)
	}

	// Coletar ativos vendidos (quantity == 0)
	soldTickers := make([]string, 0)
	for ticker, asset := range w.Assets {
		if asset.Quantity == 0 {
			soldTickers = append(soldTickers, ticker)
		}
	}

	// Verificar se há ativos vendidos
	if len(soldTickers) == 0 {
		fmt.Println("\nNenhum ativo vendido completamente encontrado.")
		fmt.Println("Todos os ativos que você comprou ainda estão em carteira.")
		fmt.Println()
		return nil
	}

	// Ordenar por ticker
	sort.Strings(soldTickers)

	// Exibir cabeçalho
	fmt.Printf("\n=== ATIVOS VENDIDOS COMPLETAMENTE ===\n")
	fmt.Printf("Total: %d\n\n", len(soldTickers))

	// Exibir cada ativo vendido
	for _, ticker := range soldTickers {
		asset := w.Assets[ticker]
		fmt.Printf("%s - Vendido - R$ %s investido (PM: R$ %s)\n",
			ticker,
			asset.TotalInvestedValue.StringFixed(2),
			asset.AveragePrice.StringFixed(4),
		)
	}

	fmt.Println()
	fmt.Println("ℹ  Estes ativos foram vendidos completamente mas seu histórico")
	fmt.Println("   de transações ainda está disponível em transactions.yaml")
	fmt.Println()
	return nil
}
