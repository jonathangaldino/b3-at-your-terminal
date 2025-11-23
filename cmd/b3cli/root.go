package main

import (
	"github.com/spf13/cobra"
)

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
