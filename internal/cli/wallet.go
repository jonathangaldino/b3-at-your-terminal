package cli

import (
	"fmt"
	"path/filepath"

	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
)

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Gerencia sua carteira de investimentos",
	Long:  `Comandos para criar, visualizar e gerenciar sua carteira de investimentos da B3.`,
}

var walletCreateCmd = &cobra.Command{
	Use:   "create [diretório]",
	Short: "Cria uma nova carteira vazia",
	Long: `Cria uma nova carteira de investimentos vazia no diretório especificado.

O arquivo wallet.yaml será criado para armazenar suas transações e ativos.
Use o comando 'parse' posteriormente para importar transações de arquivos .xlsx.`,
	Example: `  b3cli wallet create .
  b3cli wallet create ~/meus-investimentos
  b3cli wallet create /caminho/para/carteira`,
	Args: cobra.ExactArgs(1),
	RunE: runWalletCreate,
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
}

func runWalletCreate(cmd *cobra.Command, args []string) error {
	dirPath := args[0]

	// Converter para caminho absoluto
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("erro ao resolver caminho: %w", err)
	}

	// Verificar se já existe uma wallet
	if wallet.Exists(absPath) {
		return fmt.Errorf("já existe uma wallet em %s", absPath)
	}

	// Criar wallet vazia
	w := wallet.NewWallet([]parser.Transaction{})

	// Salvar wallet
	if err := w.Save(absPath); err != nil {
		return fmt.Errorf("erro ao criar wallet: %w", err)
	}

	fmt.Printf("✓ Carteira criada com sucesso em: %s\n", absPath)
	fmt.Printf("✓ Arquivo: %s\n", filepath.Join(absPath, "wallet.yaml"))
	fmt.Println()
	fmt.Println("Próximos passos:")
	fmt.Printf("  1. Importe transações: b3cli parse --wallet %s arquivos/*.xlsx\n", absPath)
	fmt.Println("  2. Visualize sua carteira: b3cli wallet show")

	return nil
}
