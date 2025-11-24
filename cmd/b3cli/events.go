package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage corporate events (grouping, split, bonus, etc.)",
	Long: `Commands to manage corporate events that affect your assets:
- Grouping (reverse split): reduces number of shares
- Split (stock split): increases number of shares
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

var eventsSplitCmd = &cobra.Command{
	Use:   "split",
	Short: "Apply a split (stock split) to an asset",
	Long: `Apply a split (stock split / desdobramento) to an asset in your portfolio.

A split increases the number of shares by dividing each share into N parts.
For example, in a 1:2 split:
- 100 shares become 200 shares
- Price of R$ 10.50 becomes R$ 5.25
- Total invested value remains the same

This command launches an interactive interface where you can:
1. Select which asset to apply the split to
2. Enter the split ratio (e.g., "1:2", "1:3")
3. Enter the event date
4. See a preview of the changes
5. Confirm and apply the split`,
	Example: `  # Launch interactive split interface
  b3cli events split`,
	RunE: runEventsSplit,
}

func init() {
	// Add subcommands to events
	eventsCmd.AddCommand(eventsGroupingCmd)
	eventsCmd.AddCommand(eventsSplitCmd)
}

func runEventsGrouping(cmd *cobra.Command, args []string) error {
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Launch interactive TUI
	p := tea.NewProgram(
		newGroupingModel(w, w.GetDirPath()),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

func runEventsSplit(cmd *cobra.Command, args []string) error {
	w, err := getOrLoadWallet()
	if err != nil {
		return err
	}

	// Launch interactive TUI
	p := tea.NewProgram(
		newSplitModel(w, w.GetDirPath()),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
