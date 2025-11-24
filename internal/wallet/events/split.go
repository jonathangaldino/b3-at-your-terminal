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

// SplitRatio represents the split ratio (e.g., 1:2 means each share becomes 2 shares)
type SplitRatio struct {
	From int // Always 1 for splits
	To   int // Number of new shares (e.g., 2, 3, 4...)
}

// SplitResult contains statistics about the split operation
type SplitResult struct {
	Ticker               string
	Ratio                SplitRatio
	EventDate            time.Time
	TransactionsAdjusted int
	QuantityBefore       int
	QuantityAfter        int
	PriceBefore          decimal.Decimal
	PriceAfter           decimal.Decimal
}

// ApplySplit applies a stock split (desdobramento) to an asset
// This increases the number of shares by splitting each share into N parts
// Example: 1:2 split → 100 shares become 200 shares, price divides by 2
func ApplySplit(w *wallet.Wallet, ticker string, ratio SplitRatio, eventDate time.Time) (*SplitResult, error) {
	// Validate that asset exists
	asset, exists := w.Assets[ticker]
	if !exists {
		return nil, fmt.Errorf("asset %s not found", ticker)
	}

	// Validate ratio (From must be 1, To must be >= 2)
	if ratio.From != 1 {
		return nil, fmt.Errorf("invalid ratio: From must be 1 (got %d)", ratio.From)
	}
	if ratio.To < 2 {
		return nil, fmt.Errorf("invalid ratio: To must be >= 2 (got %d)", ratio.To)
	}

	// Create result object
	result := &SplitResult{
		Ticker:         ticker,
		Ratio:          ratio,
		EventDate:      eventDate,
		QuantityBefore: asset.Quantity,
		PriceBefore:    asset.AveragePrice,
	}

	// Calculate multiplier (e.g., 1:2 = multiplier 2)
	multiplier := decimal.NewFromInt(int64(ratio.To)).Div(decimal.NewFromInt(int64(ratio.From)))

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

			// Adjust quantity: multiply by ratio (e.g., 100 × 2 = 200)
			tx.Quantity = tx.Quantity.Mul(multiplier)

			// Adjust price: divide by ratio (e.g., R$ 28.00 ÷ 2 = R$ 14.00)
			tx.Price = tx.Price.Div(multiplier)

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

// ParseSplitRatio parses a ratio string like "1:2" into a SplitRatio struct
func ParseSplitRatio(ratioStr string) (SplitRatio, error) {
	parts := strings.Split(ratioStr, ":")
	if len(parts) != 2 {
		return SplitRatio{}, fmt.Errorf("invalid ratio format: expected '1:N' (e.g., '1:2'), got '%s'", ratioStr)
	}

	from, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return SplitRatio{}, fmt.Errorf("invalid ratio format: expected '1:N' (e.g., '1:2'), got '%s'", ratioStr)
	}

	to, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return SplitRatio{}, fmt.Errorf("invalid ratio format: expected '1:N' (e.g., '1:2'), got '%s'", ratioStr)
	}

	return SplitRatio{From: from, To: to}, nil
}

// FormatSplitRatio formats a SplitRatio as a string (e.g., "1:2")
func FormatSplitRatio(ratio SplitRatio) string {
	return fmt.Sprintf("%d:%d", ratio.From, ratio.To)
}
