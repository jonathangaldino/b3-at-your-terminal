package cli

import (
	"fmt"
	"path/filepath"

	"github.com/john/b3-project/internal/config"
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

var walletOpenCmd = &cobra.Command{
	Use:   "open [diretório]",
	Short: "Abre uma carteira para uso",
	Long: `Define a carteira atual para ser usada pelos comandos subsequentes.

Após abrir uma carteira, todos os comandos (parse, assets, etc.) operarão
automaticamente nesta carteira sem precisar especificar o caminho.`,
	Example: `  b3cli wallet open data
  b3cli wallet open ~/meus-investimentos
  b3cli wallet open /caminho/para/carteira`,
	Args: cobra.ExactArgs(1),
	RunE: runWalletOpen,
}

var walletCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Mostra qual carteira está aberta atualmente",
	Long:  `Exibe o caminho da carteira que está atualmente aberta e sendo usada pelos comandos.`,
	Example: `  b3cli wallet current`,
	Args: cobra.NoArgs,
	RunE: runWalletCurrent,
}

var walletCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "Fecha a carteira atual",
	Long:  `Fecha a carteira atual. Após executar este comando, será necessário abrir uma carteira novamente.`,
	Example: `  b3cli wallet close`,
	Args: cobra.NoArgs,
	RunE: runWalletClose,
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
	walletCmd.AddCommand(walletOpenCmd)
	walletCmd.AddCommand(walletCurrentCmd)
	walletCmd.AddCommand(walletCloseCmd)
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
	fmt.Printf("✓ Arquivos criados:\n")
	fmt.Printf("  - %s\n", filepath.Join(absPath, "assets.yaml"))
	fmt.Printf("  - %s\n", filepath.Join(absPath, "transactions.yaml"))
	fmt.Println()
	fmt.Println("Próximos passos:")
	fmt.Printf("  1. Abra a carteira: b3cli wallet open %s\n", absPath)
	fmt.Printf("  2. Importe transações: b3cli parse arquivos/*.xlsx\n")
	fmt.Printf("  3. Visualize seus ativos: b3cli assets overview\n")

	return nil
}

func runWalletOpen(cmd *cobra.Command, args []string) error {
	dirPath := args[0]

	// Converter para caminho absoluto
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("erro ao resolver caminho: %w", err)
	}

	// Verificar se a wallet existe
	if !wallet.Exists(absPath) {
		return fmt.Errorf("wallet não encontrada em %s\nCrie uma wallet primeiro: b3cli wallet create %s", absPath, absPath)
	}

	// Definir como wallet atual
	if err := config.SetCurrentWallet(absPath); err != nil {
		return fmt.Errorf("erro ao definir wallet atual: %w", err)
	}

	fmt.Printf("✓ Wallet aberta: %s\n", absPath)
	fmt.Printf("✓ Arquivos:\n")
	fmt.Printf("  - %s\n", filepath.Join(absPath, "assets.yaml"))
	fmt.Printf("  - %s\n", filepath.Join(absPath, "transactions.yaml"))
	fmt.Println()
	fmt.Println("Agora você pode usar os comandos sem especificar wallet:")
	fmt.Println("  b3cli parse arquivos/*.xlsx")
	fmt.Println("  b3cli assets overview")
	fmt.Println("  b3cli assets subscription TICKER subscription@PARENT")

	return nil
}

func runWalletCurrent(cmd *cobra.Command, args []string) error {
	walletPath, err := config.GetCurrentWallet()
	if err != nil {
		return err
	}

	fmt.Printf("Wallet atual: %s\n", walletPath)
	return nil
}

func runWalletClose(cmd *cobra.Command, args []string) error {
	if !config.HasCurrentWallet() {
		fmt.Println("Nenhuma wallet está aberta.")
		return nil
	}

	// Obter wallet atual antes de fechar
	walletPath, _ := config.GetCurrentWallet()

	if err := config.ClearCurrentWallet(); err != nil {
		return fmt.Errorf("erro ao fechar wallet: %w", err)
	}

	fmt.Printf("✓ Wallet fechada: %s\n", walletPath)
	fmt.Println()
	fmt.Println("Para trabalhar com uma wallet novamente:")
	fmt.Println("  b3cli wallet open <diretório>")

	return nil
}
