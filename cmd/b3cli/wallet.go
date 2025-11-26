package main

import (
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/john/b3-project/internal/config"
	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

var walletLockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Trava a carteira (limpa chaves da memória)",
	Long: `Trava a carteira atual removendo as chaves de criptografia da memória.

A carteira permanece como "atual" mas será necessário desbloqueá-la novamente
com a senha mestra para realizar operações.`,
	Example: `  b3cli wallet lock`,
	Args: cobra.NoArgs,
	RunE: runWalletLock,
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
	walletCmd.AddCommand(walletOpenCmd)
	walletCmd.AddCommand(walletCurrentCmd)
	walletCmd.AddCommand(walletCloseCmd)
	walletCmd.AddCommand(walletLockCmd)
}

// readPassword reads a password from stdin without echoing it
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after password input
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

// readAndConfirmPassword reads a password twice to confirm
func readAndConfirmPassword() (string, error) {
	password, err := readPassword("Enter master password: ")
	if err != nil {
		return "", err
	}

	if len(password) < 12 {
		return "", fmt.Errorf("password must be at least 12 characters long")
	}

	confirm, err := readPassword("Confirm master password: ")
	if err != nil {
		return "", err
	}

	if password != confirm {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
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

	fmt.Println("Creating encrypted wallet...")
	fmt.Println("⚠️  Your master password is the ONLY way to decrypt your wallet.")
	fmt.Println("⚠️  If you lose it, your data is PERMANENTLY inaccessible (zero-knowledge encryption).")
	fmt.Println()

	// Solicitar senha mestra
	password, err := readAndConfirmPassword()
	if err != nil {
		return fmt.Errorf("erro ao ler senha: %w", err)
	}

	// Criar wallet criptografada
	w, err := wallet.Create(absPath, password)
	if err != nil {
		return fmt.Errorf("erro ao criar wallet: %w", err)
	}

	// Salvar cache descriptografado para persistência de sessão
	if err := w.SaveUnlocked(absPath); err != nil {
		return fmt.Errorf("erro ao salvar cache de sessão: %w", err)
	}

	// Definir como wallet atual
	if err := config.SetCurrentWallet(absPath); err != nil {
		return fmt.Errorf("erro ao definir wallet atual: %w", err)
	}

	// Armazenar wallet desbloqueada globalmente
	currentWallet = w

	fmt.Printf("\n✓ Encrypted wallet created successfully: %s\n", absPath)
	fmt.Printf("✓ Encryption: AES-256-GCM with Argon2id KDF\n")
	fmt.Printf("✓ Files created:\n")
	fmt.Printf("  - %s (encrypted vault)\n", filepath.Join(absPath, "vault.enc"))
	fmt.Printf("  - %s (encryption metadata)\n", filepath.Join(absPath, "salt.bin"))
	fmt.Println()
	fmt.Println("Wallet is now open for the session - no password needed for commands.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Import transactions: b3cli parse files/*.xlsx\n")
	fmt.Printf("  2. View your assets: b3cli assets overview\n")
	fmt.Println()
	fmt.Println("When done, close the wallet to secure it:")
	fmt.Println("  b3cli wallet close")

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

	// Solicitar senha mestra
	password, err := readPassword("Enter master password: ")
	if err != nil {
		return fmt.Errorf("erro ao ler senha: %w", err)
	}

	// Carregar e descriptografar wallet
	w, err := wallet.Load(absPath, password)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %w", err)
	}

	// Salvar cache descriptografado para persistência de sessão
	if err := w.SaveUnlocked(absPath); err != nil {
		return fmt.Errorf("erro ao salvar cache de sessão: %w", err)
	}

	// Definir como wallet atual
	if err := config.SetCurrentWallet(absPath); err != nil {
		return fmt.Errorf("erro ao definir wallet atual: %w", err)
	}

	// Armazenar wallet desbloqueada globalmente
	currentWallet = w

	fmt.Printf("\n✓ Wallet unlocked: %s\n", absPath)
	fmt.Printf("✓ Transactions: %d\n", len(w.Transactions))
	fmt.Printf("✓ Assets: %d\n", len(w.Assets))
	fmt.Println()
	fmt.Println("Wallet is now open for the session - no password needed for commands:")
	fmt.Println("  b3cli parse files/*.xlsx")
	fmt.Println("  b3cli assets overview")
	fmt.Println("  b3cli earnings overview")
	fmt.Println()
	fmt.Println("When done, close the wallet to secure it:")
	fmt.Println("  b3cli wallet close")

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

	// Limpar cache descriptografado (remove session persistence)
	if err := wallet.ClearUnlocked(walletPath); err != nil {
		fmt.Printf("⚠ Warning: failed to clear unlocked cache: %v\n", err)
	}

	// Limpar chaves da memória se houver wallet desbloqueada
	if currentWallet != nil {
		currentWallet.Lock()
		currentWallet = nil
	}

	if err := config.ClearCurrentWallet(); err != nil {
		return fmt.Errorf("erro ao fechar wallet: %w", err)
	}

	fmt.Printf("✓ Wallet closed and secured: %s\n", walletPath)
	fmt.Printf("✓ Session cache cleared\n")
	fmt.Printf("✓ Encrypted vault remains safe on disk\n")
	fmt.Println()
	fmt.Println("To work with a wallet again:")
	fmt.Println("  b3cli wallet open <directory>")

	return nil
}

func runWalletLock(cmd *cobra.Command, args []string) error {
	if !config.HasCurrentWallet() {
		return fmt.Errorf("no wallet is currently open")
	}

	if currentWallet == nil || currentWallet.IsLocked() {
		return fmt.Errorf("wallet is already locked")
	}

	// Save wallet before locking
	walletPath, _ := config.GetCurrentWallet()
	if err := currentWallet.Save(walletPath); err != nil {
		return fmt.Errorf("failed to save wallet before locking: %w", err)
	}

	// Clear session cache
	if err := wallet.ClearUnlocked(walletPath); err != nil {
		fmt.Printf("⚠ Warning: failed to clear unlocked cache: %v\n", err)
	}

	// Lock the wallet (clear encryption key from memory)
	currentWallet.Lock()
	currentWallet = nil

	fmt.Printf("✓ Wallet locked: %s\n", walletPath)
	fmt.Println("✓ Encryption keys cleared from memory")
	fmt.Println("✓ Session cache cleared")
	fmt.Println()
	fmt.Println("To unlock the wallet again:")
	fmt.Printf("  b3cli wallet open %s\n", walletPath)

	return nil
}
