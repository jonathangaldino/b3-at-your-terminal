package events

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/shopspring/decimal"
)

// GroupingRatio represents the grouping ratio (e.g., 10:1 means 10 old shares become 1 new share)
type GroupingRatio struct {
	From int // Number of old shares (e.g., 10)
	To   int // Number of new shares (e.g., 1)
}

// GroupingResult contains statistics about the grouping operation
type GroupingResult struct {
	Ticker               string
	Ratio                GroupingRatio
	EventDate            time.Time
	TransactionsAdjusted int
	QuantityBefore       int
	QuantityAfter        int
	PriceBefore          decimal.Decimal
	PriceAfter           decimal.Decimal
}

// ApplyGrouping applies a reverse split (grouping) to an asset
// This reduces the number of shares by grouping N shares into 1
// Example: 10:1 grouping → 1000 shares become 100 shares, price multiplies by 10
func ApplyGrouping(w *wallet.Wallet, ticker string, ratio GroupingRatio, eventDate time.Time) (*GroupingResult, error) {
	// Validate that asset exists
	asset, exists := w.Assets[ticker]
	if !exists {
		return nil, fmt.Errorf("asset %s not found", ticker)
	}

	// Validate ratio (From must be >= 2, To must be 1)
	if ratio.From < 2 {
		return nil, fmt.Errorf("invalid ratio: From must be >= 2 (got %d)", ratio.From)
	}
	if ratio.To != 1 {
		return nil, fmt.Errorf("invalid ratio: To must be 1 (got %d)", ratio.To)
	}

	// Create result object
	result := &GroupingResult{
		Ticker:         ticker,
		Ratio:          ratio,
		EventDate:      eventDate,
		QuantityBefore: asset.Quantity,
		PriceBefore:    asset.AveragePrice,
	}

	// Calculate divisor (e.g., 10:1 = divisor 10)
	divisor := decimal.NewFromInt(int64(ratio.From)).Div(decimal.NewFromInt(int64(ratio.To)))

	// Track old hashes that need to be removed from the map
	oldHashes := make([]string, 0)
	newTransactions := make([]parser.Transaction, 0)

	// Adjust transactions in asset.Negotiations that occurred BEFORE the event date
	for i := range asset.Negotiations {
		tx := &asset.Negotiations[i]

		// Only adjust transactions before the event date
		if tx.Date.Before(eventDate) {
			// Store old hash for removal
			oldHashes = append(oldHashes, tx.Hash)

			// Adjust quantity: divide by ratio (e.g., 1000 ÷ 10 = 100)
			tx.Quantity = tx.Quantity.Div(divisor)

			// Adjust price: multiply by ratio (e.g., R$ 2.80 × 10 = R$ 28.00)
			tx.Price = tx.Price.Mul(divisor)

			// Amount stays the same (Quantity × Price should equal original Amount)
			// We recalculate to ensure precision
			tx.Amount = tx.Quantity.Mul(tx.Price)

			// Recalculate hash with new values
			tx.Hash = parser.CalculateHash(tx)

			result.TransactionsAdjusted++
		}

		// Add to new transactions list (all transactions, adjusted or not)
		newTransactions = append(newTransactions, *tx)
	}

	// Update asset.Negotiations with adjusted transactions
	asset.Negotiations = newTransactions

	// Update wallet-level Transactions list
	newWalletTransactions := make([]parser.Transaction, 0)
	for _, tx := range w.Transactions {
		// Check if this transaction's hash is in oldHashes
		wasAdjusted := false
		for _, oldHash := range oldHashes {
			if tx.Hash == oldHash {
				wasAdjusted = true
				break
			}
		}

		if wasAdjusted {
			// Find the corresponding adjusted transaction from asset.Negotiations
			for _, adjustedTx := range asset.Negotiations {
				// Match by date, type, and ticker (since hash changed)
				if adjustedTx.Date.Equal(tx.Date) &&
					adjustedTx.Type == tx.Type &&
					adjustedTx.Ticker == tx.Ticker &&
					adjustedTx.Institution == tx.Institution {
					newWalletTransactions = append(newWalletTransactions, adjustedTx)
					break
				}
			}
		} else {
			// Transaction was not adjusted, keep as is
			newWalletTransactions = append(newWalletTransactions, tx)
		}
	}
	w.Transactions = newWalletTransactions

	// Rebuild TransactionsByHash map
	w.TransactionsByHash = make(map[string]parser.Transaction)
	for _, tx := range w.Transactions {
		w.TransactionsByHash[tx.Hash] = tx
	}

	// Recalculate all asset metrics (quantity, average price, etc.)
	w.RecalculateAssets()

	// Update result with new values
	if updatedAsset, exists := w.Assets[ticker]; exists {
		result.QuantityAfter = updatedAsset.Quantity
		result.PriceAfter = updatedAsset.AveragePrice
	}

	return result, nil
}

// ParseRatio parses a ratio string like "10:1" into a GroupingRatio struct
func ParseRatio(ratioStr string) (GroupingRatio, error) {
	parts := strings.Split(ratioStr, ":")
	if len(parts) != 2 {
		return GroupingRatio{}, fmt.Errorf("invalid ratio format: expected 'N:1' (e.g., '10:1'), got '%s'", ratioStr)
	}

	from, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return GroupingRatio{}, fmt.Errorf("invalid ratio format: expected 'N:1' (e.g., '10:1'), got '%s'", ratioStr)
	}

	to, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return GroupingRatio{}, fmt.Errorf("invalid ratio format: expected 'N:1' (e.g., '10:1'), got '%s'", ratioStr)
	}

	return GroupingRatio{From: from, To: to}, nil
}

// FormatRatio formats a GroupingRatio as a string (e.g., "10:1")
func FormatRatio(ratio GroupingRatio) string {
	return fmt.Sprintf("%d:%d", ratio.From, ratio.To)
}
