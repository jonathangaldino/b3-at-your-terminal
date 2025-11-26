package main

import (
	"fmt"

	"github.com/john/b3-project/internal/config"
	"github.com/john/b3-project/internal/wallet"
	wcrypto "github.com/john/b3-project/internal/wallet/crypto"
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
// First tries to load from unlocked session cache (no password needed)
// If cache doesn't exist, prompts for password to decrypt
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

	// Try to load from unlocked cache first (session persistence)
	if wallet.IsUnlocked(walletPath) {
		w, err := wallet.LoadUnlocked(walletPath)
		if err == nil {
			// Successfully loaded from cache
			// However, unlocked cache doesn't include encryption key
			// Need to unlock vault to get the key for saving
			fmt.Println("Loading wallet from session cache...")
			fmt.Println("Please enter your password to enable saving.")
			password, err := readPassword("Enter master password: ")
			if err != nil {
				return nil, fmt.Errorf("failed to read password: %w", err)
			}

			// Unlock vault to get encryption key
			encryptionKey, err := wcrypto.UnlockVault(walletPath, password)
			if err != nil {
				return nil, fmt.Errorf("failed to unlock wallet: %w", err)
			}

			// Set encryption key on the cached wallet
			w.SetEncryptionKey(encryptionKey)
			currentWallet = w
			fmt.Println("✓ Wallet ready")
			return w, nil
		}
		// If cache is corrupted, fall through to password prompt
		fmt.Printf("⚠ Unlocked cache corrupted, will prompt for password\n")
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

	// Save unlocked cache for future commands in this session
	if err := w.SaveUnlocked(walletPath); err != nil {
		// Non-fatal - just warn
		fmt.Printf("⚠ Warning: failed to save session cache: %v\n", err)
	}

	// Store unlocked wallet in memory
	currentWallet = w

	fmt.Println("✓ Wallet unlocked")
	return w, nil
}
