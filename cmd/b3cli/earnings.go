package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var earningsCmd = &cobra.Command{
	Use:   "earnings",
	Short: "Gerencia proventos (rendimentos, dividendos, JCP, resgates)",
	Long: `Gerencia proventos e resgates recebidos de ações e fundos imobiliários.

Tipos de proventos suportados:
- Rendimento
- Dividendo
- Juros Sobre Capital Próprio (JCP)
- Resgate (fechamento de capital/retirada de circulação)`,
}

var earningsParseCmd = &cobra.Command{
	Use:   "parse [arquivos...]",
	Short: "Parseia arquivos .xlsx de proventos da B3",
	Long: `Parseia um ou mais arquivos .xlsx contendo proventos recebidos da B3.

Os arquivos devem estar no formato esperado com as seguintes colunas:
- Entrada/Saída (ignorado)
- Data (DD/MM/YYYY)
- Movimentação (tipo: Rendimento/Dividendo/Juros Sobre Capital Próprio/Resgate)
- Produto (formato: TICKER - Nome da empresa)
- Instituição (ignorado)
- Quantidade
- Preço unitário
- Valor da operação (total a receber)

Tipos de movimentação aceitos:
- Rendimento: pagamentos periódicos (comum em FIIs)
- Dividendo: distribuição de lucros
- JCP / Juros Sobre Capital Próprio: distribuição com benefício fiscal
- Resgate: fechamento de capital ou retirada de circulação

O comando automaticamente deduplica proventos, atualiza a carteira atual
e calcula o total de proventos recebidos para cada ativo.

IMPORTANTE: Você deve ter aberto uma wallet antes de usar este comando.
Use 'b3cli wallet open <diretório>' para abrir uma wallet.`,
	Example: `  b3cli earnings parse proventos.xlsx
  b3cli earnings parse proventos1.xlsx proventos2.xlsx
  b3cli earnings parse files/proventos-*.xlsx`,
	Args: cobra.MinimumNArgs(1),
	RunE: runEarningsParse,
}

func runEarningsParse(cmd *cobra.Command, args []string) error {
	filePaths := args

	// Validar que todos os arquivos existem
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("arquivo não encontrado: %s", filePath)
		}
	}

	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	walletPath := w.GetDirPath()

	// Contar earnings antes
	earningsBefore := countTotalEarnings(w)

	// Parsear arquivos de proventos
	fmt.Printf("Processando %d arquivo(s) de proventos...\n", len(filePaths))

	newEarnings, err := parser.ParseEarningsFiles(filePaths)
	if err != nil {
		return fmt.Errorf("erro ao parsear arquivos: %w", err)
	}

	// Add earnings using wallet method (handles deduplication and recalculation)
	added, duplicates, err := w.AddEarnings(newEarnings)
	if err != nil {
		return fmt.Errorf("erro ao adicionar proventos: %w", err)
	}

	// Salvar wallet atualizada
	if err := w.Save(walletPath); err != nil {
		return fmt.Errorf("erro ao salvar wallet: %w", err)
	}

	// Exibir resultados
	fmt.Printf("\n✓ Wallet atualizada com sucesso!\n")
	fmt.Printf("  Proventos antes: %d\n", earningsBefore)
	fmt.Printf("  Proventos novos: %d\n", added)
	fmt.Printf("  Proventos duplicados (ignorados): %d\n", duplicates)
	fmt.Printf("  Total de proventos: %d\n\n", countTotalEarnings(w))

	displayEarningsSummary(w)

	return nil
}

func countTotalEarnings(w *wallet.Wallet) int {
	total := 0
	for _, asset := range w.Assets {
		total += len(asset.Earnings)
	}
	return total
}

func displayEarningsSummary(w *wallet.Wallet) {
	fmt.Println("=== RESUMO DE PROVENTOS POR ATIVO ===")

	// Contar quantos ativos têm earnings
	assetsWithEarnings := 0
	for _, asset := range w.Assets {
		if len(asset.Earnings) > 0 {
			assetsWithEarnings++
		}
	}

	if assetsWithEarnings == 0 {
		fmt.Println("Nenhum provento registrado ainda.")
		return
	}

	for ticker, asset := range w.Assets {
		if len(asset.Earnings) == 0 {
			continue
		}

		fmt.Printf("\n[%s]\n", ticker)
		fmt.Printf("  Total de proventos recebidos: %d\n", len(asset.Earnings))
		fmt.Printf("  Valor total recebido: R$ %s\n", asset.TotalEarnings.StringFixed(2))

		// Breakdown por tipo
		rendimentos := 0
		dividendos := 0
		jcp := 0
		resgates := 0

		for _, e := range asset.Earnings {
			switch e.Type {
			case "Rendimento":
				rendimentos++
			case "Dividendo":
				dividendos++
			case "Juros Sobre Capital Próprio":
				jcp++
			case "Resgate":
				resgates++
			}
		}

		if rendimentos > 0 {
			fmt.Printf("    - Rendimentos: %d\n", rendimentos)
		}
		if dividendos > 0 {
			fmt.Printf("    - Dividendos: %d\n", dividendos)
		}
		if jcp > 0 {
			fmt.Printf("    - JCP: %d\n", jcp)
		}
		if resgates > 0 {
			fmt.Printf("    - Resgates: %d\n", resgates)
		}
	}
}

func displayEarningsOverview(w *wallet.Wallet) {
	// Estrutura para agrupar earnings por tipo
	type EarningsByType struct {
		Count       int
		TotalAmount decimal.Decimal
		Assets      map[string]decimal.Decimal // ticker -> total amount
	}

	categories := map[string]*EarningsByType{
		"Rendimento":                   {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
		"Dividendo":                    {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
		"Juros Sobre Capital Próprio": {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
		"Resgate":                      {Count: 0, TotalAmount: decimal.Zero, Assets: make(map[string]decimal.Decimal)},
	}

	totalGeneral := decimal.Zero
	totalCount := 0

	// Agrupar earnings por tipo
	for ticker, asset := range w.Assets {
		for _, earning := range asset.Earnings {
			if cat, exists := categories[earning.Type]; exists {
				cat.Count++
				cat.TotalAmount = cat.TotalAmount.Add(earning.TotalAmount)

				// Adicionar ao total do ativo
				if current, ok := cat.Assets[ticker]; ok {
					cat.Assets[ticker] = current.Add(earning.TotalAmount)
				} else {
					cat.Assets[ticker] = earning.TotalAmount
				}

				totalGeneral = totalGeneral.Add(earning.TotalAmount)
				totalCount++
			}
		}
	}

	// Exibir resumo
	fmt.Println("=== RESUMO GERAL DE PROVENTOS ===")
	fmt.Printf("Total de pagamentos recebidos: %d\n", totalCount)
	fmt.Printf("Valor total recebido: R$ %s\n\n", totalGeneral.StringFixed(2))

	// Exibir por categoria
	types := []string{"Rendimento", "Dividendo", "Juros Sobre Capital Próprio", "Resgate"}
	typeLabels := map[string]string{
		"Rendimento":                   "RENDIMENTOS",
		"Dividendo":                    "DIVIDENDOS",
		"Juros Sobre Capital Próprio": "JUROS SOBRE CAPITAL PRÓPRIO (JCP)",
		"Resgate":                      "RESGATES",
	}

	for _, earningType := range types {
		cat := categories[earningType]
		if cat.Count == 0 {
			continue
		}

		fmt.Printf("=== %s ===\n", typeLabels[earningType])
		fmt.Printf("Quantidade de pagamentos: %d\n", cat.Count)
		fmt.Printf("Valor total: R$ %s\n", cat.TotalAmount.StringFixed(2))

		// Calcular percentual do total
		if !totalGeneral.IsZero() {
			percentage := cat.TotalAmount.Div(totalGeneral).Mul(decimal.NewFromInt(100))
			fmt.Printf("Percentual do total: %.2f%%\n", percentage.InexactFloat64())
		}

		// Listar ativos
		fmt.Println("\nAtivos que pagaram:")

		// Criar slice para ordenar
		type AssetEarning struct {
			Ticker string
			Amount decimal.Decimal
		}
		var assetList []AssetEarning
		for ticker, amount := range cat.Assets {
			assetList = append(assetList, AssetEarning{ticker, amount})
		}

		// Ordenar por valor (maior primeiro)
		for i := 0; i < len(assetList); i++ {
			for j := i + 1; j < len(assetList); j++ {
				if assetList[j].Amount.GreaterThan(assetList[i].Amount) {
					assetList[i], assetList[j] = assetList[j], assetList[i]
				}
			}
		}

		// Exibir lista de ativos
		for _, item := range assetList {
			fmt.Printf("  %-8s  R$ %10s\n", item.Ticker, item.Amount.StringFixed(2))
		}

		fmt.Println()
	}
}

var earningsOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Exibe resumo de proventos agrupados por tipo",
	Long: `Exibe um resumo completo de todos os proventos recebidos, agrupados por tipo.

Mostra para cada categoria (Rendimento, Dividendo, JCP, Resgate):
- Quantidade total de pagamentos
- Valor total recebido
- Lista de ativos que pagaram este tipo de provento

Útil para entender a composição dos seus ganhos passivos.`,
	Example: `  b3cli earnings overview`,
	Args:    cobra.NoArgs,
	RunE:    runEarningsOverview,
}

func runEarningsOverview(cmd *cobra.Command, args []string) error {
	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Verificar se há proventos
	totalEarnings := countTotalEarnings(w)
	if totalEarnings == 0 {
		fmt.Println("Nenhum provento registrado ainda.")
		fmt.Println("\nUse 'b3cli earnings parse <arquivo>' para adicionar proventos.")
		return nil
	}

	// Iniciar interface Bubble Tea
	p := tea.NewProgram(initialOverviewModel(w), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar interface: %w", err)
	}

	return nil
}

var earningsReportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Exibe relatórios de proventos por período",
	Long: `Exibe relatórios detalhados de proventos recebidos.

Permite visualizar proventos de forma:
- Anual: resumo por ano (todos os anos disponíveis)
- Mensal: resumo por mês (com seleção de ano se houver múltiplos anos)

Útil para analisar a evolução dos ganhos passivos ao longo do tempo.`,
	Example: `  b3cli earnings reports`,
	Args:    cobra.NoArgs,
	RunE:    runEarningsReports,
}

func runEarningsReports(cmd *cobra.Command, args []string) error {
	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Verificar se há proventos
	totalEarnings := countTotalEarnings(w)
	if totalEarnings == 0 {
		fmt.Println("Nenhum provento registrado ainda.")
		fmt.Println("\nUse 'b3cli earnings parse <arquivo>' para adicionar proventos.")
		return nil
	}

	// Iniciar interface Bubble Tea
	p := tea.NewProgram(initialReportsModel(w), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar interface: %w", err)
	}

	return nil
}

func init() {
	// Adicionar subcomandos ao earnings
	earningsCmd.AddCommand(earningsParseCmd)
	earningsCmd.AddCommand(earningsOverviewCmd)
	earningsCmd.AddCommand(earningsReportsCmd)
}
