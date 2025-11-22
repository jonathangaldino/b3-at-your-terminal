package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
)

var (
	walletPath string
)

var parseCmd = &cobra.Command{
	Use:   "parse [arquivos...]",
	Short: "Parseia arquivos .xlsx de transações da B3",
	Long: `Parseia um ou mais arquivos .xlsx contendo transações financeiras da B3.

Os arquivos devem estar no formato esperado com as seguintes colunas:
- Data do Negócio (DD/MM/YYYY)
- Tipo de Movimentação (Compra/Venda)
- Prazo/Vencimento (ignorado)
- Instituição
- Código da Negociação (ticker)
- Quantidade
- Preço
- Valor

O comando automaticamente deduplica transações, atualiza a carteira (wallet)
e calcula os preços médios ponderados para cada ativo.

IMPORTANTE: Você deve ter criado uma wallet antes de usar este comando.
Use 'b3cli wallet create <diretório>' para criar uma nova wallet.`,
	Example: `  b3cli parse --wallet . arquivo1.xlsx
  b3cli parse --wallet ~/investimentos arquivo1.xlsx arquivo2.xlsx
  b3cli parse --wallet . files/*.xlsx`,
	Args: cobra.MinimumNArgs(1),
	RunE: runParse,
}

func init() {
	parseCmd.Flags().StringVarP(&walletPath, "wallet", "w", ".", "Diretório da wallet")
}

func runParse(cmd *cobra.Command, args []string) error {
	filePaths := args

	// Verificar se existe uma wallet no diretório especificado
	if !wallet.Exists(walletPath) {
		return fmt.Errorf("wallet não encontrada em %s\nCrie uma wallet primeiro: b3cli wallet create %s", walletPath, walletPath)
	}

	// Validar que todos os arquivos existem
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("arquivo não encontrado: %s", filePath)
		}
	}

	// Carregar wallet existente
	fmt.Printf("Carregando wallet de: %s\n", walletPath)
	w, err := wallet.Load(walletPath)
	if err != nil {
		return fmt.Errorf("erro ao carregar wallet: %w", err)
	}

	transacoesAntes := len(w.Transactions)

	// Parsear arquivos
	fmt.Printf("Processando %d arquivo(s)...\n", len(filePaths))

	newTransactions, err := parser.ParseFiles(filePaths)
	if err != nil {
		return fmt.Errorf("erro ao parsear arquivos: %w", err)
	}

	// Mesclar transações (deduplicar por hash)
	added := 0
	for _, t := range newTransactions {
		if _, exists := w.TransactionsByHash[t.Hash]; !exists {
			w.Transactions = append(w.Transactions, t)
			w.TransactionsByHash[t.Hash] = t
			added++

			// Atualizar Asset
			asset, exists := w.Assets[t.Ticker]
			if !exists {
				asset = &wallet.Asset{
					ID:           t.Ticker,
					Negotiations: make([]parser.Transaction, 0),
					Type:         "renda variável",
					SubType:      "",
					Segment:      "",
				}
				w.Assets[t.Ticker] = asset
			}

			asset.Negotiations = append(asset.Negotiations, t)
		}
	}

	// Recalcular todos os campos derivados dos Assets
	w.RecalculateAssets()

	// Salvar wallet atualizada
	if err := w.Save(walletPath); err != nil {
		return fmt.Errorf("erro ao salvar wallet: %w", err)
	}

	// Exibir resultados
	fmt.Printf("\n✓ Wallet atualizada com sucesso!\n")
	fmt.Printf("  Transações antes: %d\n", transacoesAntes)
	fmt.Printf("  Transações novas: %d\n", added)
	fmt.Printf("  Transações duplicadas (ignoradas): %d\n", len(newTransactions)-added)
	fmt.Printf("  Total de transações: %d\n\n", len(w.Transactions))

	displayResults(w)

	return nil
}

func displayResults(w *wallet.Wallet) {
	fmt.Println("=== RESUMO ===")
	fmt.Printf("Total de transações únicas: %d\n", len(w.Transactions))
	fmt.Printf("Total de ativos diferentes: %d\n\n", len(w.Assets))

	fmt.Println("=== ATIVOS ===")
	for ticker, asset := range w.Assets {
		fmt.Printf("\n[%s] - %s\n", ticker, asset.Type)
		fmt.Printf("  Negociações: %d\n", len(asset.Negotiations))
		fmt.Printf("  Preço Médio: R$ %s\n", asset.AveragePrice.StringFixed(4))
		fmt.Printf("  Valor Total Investido: R$ %s\n", asset.TotalInvestedValue.StringFixed(2))
		fmt.Printf("  Quantidade em carteira: %s\n", asset.Quantity.StringFixed(0))
	}

	fmt.Println("\n=== TRANSAÇÕES ===")
	fmt.Println("Hash                                                             | Data       | Tipo   | Ticker | Qtd    | Preço   | Valor")
	fmt.Println(strings.Repeat("-", 140))

	for _, t := range w.Transactions {
		fmt.Printf("%-64s | %s | %-6s | %-6s | %6s | %7s | %10s\n",
			t.Hash[:16]+"...", // Mostrar apenas parte do hash
			t.Date.Format("02/01/2006"),
			t.Type,
			t.Ticker,
			t.Quantity.StringFixed(0),
			t.Price.StringFixed(4),
			t.Amount.StringFixed(2),
		)
	}
}
