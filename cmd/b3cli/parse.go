package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse [arquivos...]",
	Short: "Parseia arquivos .xlsx de transações e proventos da B3",
	Long: `Parseia automaticamente arquivos .xlsx da B3, detectando se são transações ou proventos.

O comando detecta automaticamente o tipo de arquivo baseado no número de colunas:
- 9 colunas: arquivo de TRANSAÇÕES (compra/venda)
- 8 colunas: arquivo de PROVENTOS (rendimentos/dividendos/JCP/resgates)

ARQUIVOS DE TRANSAÇÕES (9 colunas):
- Data do Negócio, Tipo de Movimentação, Mercado, Prazo/Vencimento,
  Instituição, Código da Negociação, Quantidade, Preço, Valor

ARQUIVOS DE PROVENTOS (8 colunas):
- Entrada/Saída, Data, Movimentação (Rendimento/Dividendo/JCP/Resgate),
  Produto, Instituição, Quantidade, Preço unitário, Valor da Operação

O comando automaticamente deduplica registros, atualiza a carteira atual
e calcula os preços médios e totais de proventos para cada ativo.

IMPORTANTE: Você deve ter aberto uma wallet antes de usar este comando.
Use 'b3cli wallet open <diretório>' para abrir uma wallet.`,
	Example: `  b3cli parse transacoes.xlsx
  b3cli parse transacoes.xlsx proventos.xlsx
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

	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	walletPath := w.GetDirPath()

	transacoesAntes := len(w.Transactions)
	earningsBefore := countTotalEarnings(w)

	// Detectar tipo de arquivo e processar adequadamente
	fmt.Printf("Processando %d arquivo(s)...\n", len(filePaths))

	// Separar arquivos por tipo
	transactionFiles := []string{}
	earningFiles := []string{}

	for _, filePath := range filePaths {
		fileType, err := parser.DetectFileType(filePath)
		if err != nil {
			return fmt.Errorf("erro ao detectar tipo do arquivo %s: %w", filePath, err)
		}

		switch fileType {
		case parser.FileTypeTransactions:
			transactionFiles = append(transactionFiles, filePath)
			fmt.Printf("  - %s: detectado como arquivo de TRANSAÇÕES\n", filePath)
		case parser.FileTypeEarnings:
			earningFiles = append(earningFiles, filePath)
			fmt.Printf("  - %s: detectado como arquivo de PROVENTOS\n", filePath)
		default:
			return fmt.Errorf("tipo de arquivo desconhecido: %s", filePath)
		}
	}

	totalAdded := 0
	totalDuplicates := 0

	// Processar arquivos de transações
	if len(transactionFiles) > 0 {
		fmt.Printf("\nProcessando %d arquivo(s) de transações...\n", len(transactionFiles))
		newTransactions, err := parser.ParseFiles(transactionFiles)
		if err != nil {
			return fmt.Errorf("erro ao parsear arquivos de transações: %w", err)
		}

		added, duplicates, err := w.AddTransactions(newTransactions)
		if err != nil {
			return fmt.Errorf("erro ao adicionar transações: %w", err)
		}

		fmt.Printf("  ✓ Transações: %d adicionadas, %d duplicadas\n", added, duplicates)
		totalAdded += added
		totalDuplicates += duplicates
	}

	// Processar arquivos de proventos
	if len(earningFiles) > 0 {
		fmt.Printf("\nProcessando %d arquivo(s) de proventos...\n", len(earningFiles))
		newEarnings, err := parser.ParseEarningsFiles(earningFiles)
		if err != nil {
			return fmt.Errorf("erro ao parsear arquivos de proventos: %w", err)
		}

		added, duplicates, err := w.AddEarnings(newEarnings)
		if err != nil {
			return fmt.Errorf("erro ao adicionar proventos: %w", err)
		}

		fmt.Printf("  ✓ Proventos: %d adicionados, %d duplicados\n", added, duplicates)
		totalAdded += added
		totalDuplicates += duplicates
	}

	// Salvar wallet atualizada
	if err := w.Save(walletPath); err != nil {
		return fmt.Errorf("erro ao salvar wallet: %w", err)
	}

	// Exibir resultados
	fmt.Printf("\n✓ Wallet atualizada com sucesso!\n")
	if len(transactionFiles) > 0 {
		fmt.Printf("  Transações antes: %d\n", transacoesAntes)
		fmt.Printf("  Total de transações: %d\n", len(w.Transactions))
	}
	if len(earningFiles) > 0 {
		fmt.Printf("  Proventos antes: %d\n", earningsBefore)
		fmt.Printf("  Total de proventos: %d\n", countTotalEarnings(w))
	}
	fmt.Printf("\n  Total adicionado: %d\n", totalAdded)
	fmt.Printf("  Total duplicados (ignorados): %d\n\n", totalDuplicates)

	// Iniciar interface Bubble Tea
	p := tea.NewProgram(initialParseResultsModel(w), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar interface: %w", err)
	}

	return nil
}

func displayResults(w *wallet.Wallet) {
	fmt.Println("=== RESUMO ===")
	fmt.Printf("Total de transações únicas: %d\n", len(w.Transactions))
	fmt.Printf("Total de proventos: %d\n", countTotalEarnings(w))
	fmt.Printf("Total de ativos diferentes: %d\n\n", len(w.Assets))

	fmt.Println("=== ATIVOS ===")
	for ticker, asset := range w.Assets {
		fmt.Printf("\n[%s] - %s\n", ticker, asset.Type)
		fmt.Printf("  Negociações: %d\n", len(asset.Negotiations))
		fmt.Printf("  Preço Médio: R$ %s\n", asset.AveragePrice.StringFixed(4))
		fmt.Printf("  Valor Total Investido: R$ %s\n", asset.TotalInvestedValue.StringFixed(2))
		fmt.Printf("  Quantidade em carteira: %d\n", asset.Quantity)
		if len(asset.Earnings) > 0 {
			fmt.Printf("  Proventos recebidos: %d (Total: R$ %s)\n", len(asset.Earnings), asset.TotalEarnings.StringFixed(2))
		}
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
