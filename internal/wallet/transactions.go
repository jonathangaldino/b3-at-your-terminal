package wallet

import (
	"fmt"

	"github.com/john/b3-project/internal/parser"
)

// AddTransaction adds a single transaction to the wallet.
// It validates the transaction, checks for duplicates, creates the asset if needed,
// and recalculates all asset values.
// Returns an error if the transaction is invalid or already exists.
func (w *Wallet) AddTransaction(tx parser.Transaction) error {
	// Validate transaction
	if err := ValidateTransaction(&tx); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Calculate hash if not already set
	if tx.Hash == "" {
		tx.Hash = parser.CalculateHash(&tx)
	}

	// Check for duplicate
	if _, exists := w.TransactionsByHash[tx.Hash]; exists {
		return fmt.Errorf("duplicate transaction detected")
	}

	// Add to wallet transactions
	w.Transactions = append(w.Transactions, tx)
	w.TransactionsByHash[tx.Hash] = tx

	// Create or update asset
	asset, exists := w.Assets[tx.Ticker]
	if !exists {
		asset = &Asset{
			ID:           tx.Ticker,
			Negotiations: make([]parser.Transaction, 0),
			Type:         "renda variável",
			SubType:      "",
			Segment:      "",
		}
		w.Assets[tx.Ticker] = asset
	}

	// Add transaction to asset
	asset.Negotiations = append(asset.Negotiations, tx)

	// Recalculate all asset values
	w.RecalculateAssets()

	return nil
}

// AddTransactions adds multiple transactions to the wallet in batch.
// It returns the number of transactions added, the number of duplicates skipped,
// and any error that occurred during validation.
// If a transaction is invalid, the entire operation is aborted and an error is returned.
func (w *Wallet) AddTransactions(transactions []parser.Transaction) (added int, duplicates int, err error) {
	for _, tx := range transactions {
		// Calculate hash if not already set
		if tx.Hash == "" {
			tx.Hash = parser.CalculateHash(&tx)
		}

		// Check for duplicate
		if _, exists := w.TransactionsByHash[tx.Hash]; exists {
			duplicates++
			continue
		}

		// Validate transaction
		if err := ValidateTransaction(&tx); err != nil {
			return added, duplicates, fmt.Errorf("invalid transaction for %s: %w", tx.Ticker, err)
		}

		// Add to wallet transactions
		w.Transactions = append(w.Transactions, tx)
		w.TransactionsByHash[tx.Hash] = tx

		// Create or update asset
		asset, exists := w.Assets[tx.Ticker]
		if !exists {
			asset = &Asset{
				ID:           tx.Ticker,
				Negotiations: make([]parser.Transaction, 0),
				Type:         "renda variável",
				SubType:      "",
				Segment:      "",
			}
			w.Assets[tx.Ticker] = asset
		}

		// Add transaction to asset
		asset.Negotiations = append(asset.Negotiations, tx)
		added++
	}

	// Recalculate all asset values after adding all transactions
	if added > 0 {
		w.RecalculateAssets()
	}

	return added, duplicates, nil
}
