package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var assetsDistributionCmd = &cobra.Command{
	Use:   "distribution",
	Short: "Shows portfolio distribution by asset type and segment",
	Long: `Displays the percentage distribution of your portfolio investments.

Shows two views:
1. Distribution by SubType (e.g., stocks, real estate funds)
2. Distribution by Segment (e.g., electric power, banking, insurance)

This gives a macro view of your portfolio allocation, helping identify
opportunities for rebalancing and diversification.

IMPORTANT: You must have opened a wallet before using this command.
Use 'b3cli wallet open <directory>' to open a wallet.`,
	Example: `  b3cli assets distribution`,
	Args:    cobra.NoArgs,
	RunE:    runAssetsDistribution,
}

func init() {
	assetsCmd.AddCommand(assetsDistributionCmd)
}

func runAssetsDistribution(cmd *cobra.Command, args []string) error {
	// Get or load wallet (will prompt for password if locked)
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Get active assets to check if there are any
	activeAssets := w.GetActiveAssets()
	if len(activeAssets) == 0 {
		fmt.Println("\nNo active assets found in portfolio.")
		fmt.Println("Add transactions using 'b3cli parse <file>' to get started.")
		fmt.Println()
		return nil
	}

	// Get distributions from wallet
	subTypeDistribution := w.GetDistributionBySubType()
	segmentDistribution := w.GetDistributionBySegment()

	// Start Bubble Tea UI
	p := tea.NewProgram(initialDistributionModel(subTypeDistribution, segmentDistribution), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running interface: %w", err)
	}

	return nil
}
