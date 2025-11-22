package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
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

O comando automaticamente deduplica transações, cria uma carteira (wallet)
e calcula os preços médios ponderados para cada ativo.`,
	Example: `  b3cli parse arquivo1.xlsx
  b3cli parse arquivo1.xlsx arquivo2.xlsx
  b3cli parse files/*.xlsx`,
	Args: cobra.MinimumNArgs(1),
	RunE: runParse,
}

func runParse(cmd *cobra.Command, args []string) error {
	filePaths := args

	// Validar que todos os arquivos existem
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("arquivo não encontrado: %s", filePath)
		}
	}

	// Parsear arquivos
	fmt.Printf("Processando %d arquivo(s)...\n\n", len(filePaths))

	transactions, err := parser.ParseFiles(filePaths)
	if err != nil {
		return fmt.Errorf("erro ao parsear arquivos: %w", err)
	}

	// Criar wallet a partir das transações
	w := wallet.NewWallet(transactions)

	// Exibir resultados
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
		fmt.Printf("  Preço Médio: R$ %.2f\n", asset.AveragePrice)

		// Calcular quantidade total (compras - vendas)
		var totalQtd float64
		for _, neg := range asset.Negotiations {
			if neg.Type == "Compra" {
				totalQtd += neg.Quantity
			} else if neg.Type == "Venda" {
				totalQtd -= neg.Quantity
			}
		}
		fmt.Printf("  Quantidade em carteira: %.0f\n", totalQtd)
	}

	fmt.Println("\n=== TRANSAÇÕES ===")
	fmt.Println("Hash                                                             | Data       | Tipo   | Ticker | Qtd    | Preço   | Valor")
	fmt.Println(strings.Repeat("-", 140))

	for _, t := range w.Transactions {
		fmt.Printf("%-64s | %s | %-6s | %-6s | %6.0f | %7.2f | %10.2f\n",
			t.Hash[:16]+"...", // Mostrar apenas parte do hash
			t.Date.Format("02/01/2006"),
			t.Type,
			t.Ticker,
			t.Quantity,
			t.Price,
			t.Amount,
		)
	}
}
