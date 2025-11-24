package main

import (
	"fmt"

	"github.com/john/b3-project/internal/config"
	"github.com/john/b3-project/internal/wallet"
	"github.com/spf13/cobra"
)

// currentWallet holds the unlocked wallet in memory during the session
// This is cleared when the wallet is closed or locked
var currentWallet *wallet.Wallet

var rootCmd = &cobra.Command{
	Use:   "b3cli",
	Short: "B3 Transaction Parser CLI",
	Long: `B3CLI é uma ferramenta de linha de comando para processar e analisar
transações financeiras da B3 a partir de arquivos Excel (.xlsx).

Gerencie sua carteira de investimentos, calcule preços médios ponderados,
e visualize suas transações de forma organizada.`,
}

// Execute executa o comando root
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Adicionar subcomandos aqui
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(walletCmd)
	rootCmd.AddCommand(assetsCmd)
	rootCmd.AddCommand(earningsCmd)
	rootCmd.AddCommand(eventsCmd)
}

// getOrLoadWallet returns the current wallet, loading it if necessary
// If the wallet is not already unlocked in memory, it will prompt for the password
func getOrLoadWallet() (*wallet.Wallet, error) {
	// Get current wallet path
	walletPath, err := config.GetCurrentWallet()
	if err != nil {
		return nil, err
	}

	// Check if wallet exists
	if !wallet.Exists(walletPath) {
		return nil, fmt.Errorf("wallet not found at %s", walletPath)
	}

	// If wallet is already unlocked in memory, return it
	if currentWallet != nil && !currentWallet.IsLocked() {
		return currentWallet, nil
	}

	// Wallet needs to be unlocked - prompt for password
	fmt.Println("Wallet is locked. Please enter your password to unlock.")
	password, err := readPassword("Enter master password: ")
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}

	// Load and decrypt wallet
	w, err := wallet.Load(walletPath, password)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock wallet: %w", err)
	}

	// Store unlocked wallet in memory
	currentWallet = w

	fmt.Println("✓ Wallet unlocked")
	return w, nil
}
