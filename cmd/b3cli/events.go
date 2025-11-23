package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/john/b3-project/internal/config"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage corporate events (grouping, split, bonus, etc.)",
	Long: `Commands to manage corporate events that affect your assets:
- Grouping (reverse split): reduces number of shares
- Split: increases number of shares (coming soon)
- Bonus shares: free share distribution (coming soon)
- Mergers & acquisitions: asset conversion (coming soon)`,
}

var eventsGroupingCmd = &cobra.Command{
	Use:   "grouping",
	Short: "Apply a grouping (reverse split) to an asset",
	Long: `Apply a grouping (reverse split) to an asset in your portfolio.

A grouping reduces the number of shares by combining N old shares into 1 new share.
For example, in a 10:1 grouping:
- 1,000 shares become 100 shares
- Price of R$ 2.80 becomes R$ 28.00
- Total invested value remains the same

This command launches an interactive interface where you can:
1. Select which asset to apply the grouping to
2. Enter the grouping ratio (e.g., "10:1")
3. Enter the event date
4. See a preview of the changes
5. Confirm and apply the grouping`,
	Example: `  # Launch interactive grouping interface
  b3cli events grouping`,
	RunE: runEventsGrouping,
}

func init() {
	// Add subcommands to events
	eventsCmd.AddCommand(eventsGroupingCmd)
}

func runEventsGrouping(cmd *cobra.Command, args []string) error {
	walletPath, err := config.GetCurrentWallet()
	if err != nil {
		return err
	}

	// Launch interactive TUI
	p := tea.NewProgram(
		newGroupingModel(walletPath),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
