package wallet

import (
	"fmt"

	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
)

// ValidateTransaction validates that a transaction has all required fields
// and that numeric values are positive.
func ValidateTransaction(tx *parser.Transaction) error {
	if tx.Ticker == "" {
		return fmt.Errorf("ticker is required")
	}

	if tx.Quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("quantity must be greater than zero")
	}

	if tx.Price.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("price must be greater than zero")
	}

	if tx.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be greater than zero")
	}

	if tx.Type != "Compra" && tx.Type != "Venda" {
		return fmt.Errorf("transaction type must be 'Compra' or 'Venda'")
	}

	return nil
}

// CanSell checks if the wallet has enough quantity of an asset to sell.
// Returns an error if the asset doesn't exist or if there's insufficient quantity.
func (w *Wallet) CanSell(ticker string, quantity decimal.Decimal) error {
	asset, exists := w.Assets[ticker]
	if !exists {
		return fmt.Errorf("asset %s not found in wallet", ticker)
	}

	quantityInt := int(quantity.IntPart())
	if asset.Quantity < quantityInt {
		return fmt.Errorf("insufficient quantity: you have %d shares, trying to sell %d",
			asset.Quantity, quantityInt)
	}

	return nil
}

// GetAssetInfo returns the asset information for a given ticker.
// Returns an error if the asset doesn't exist.
func (w *Wallet) GetAssetInfo(ticker string) (*Asset, error) {
	asset, exists := w.Assets[ticker]
	if !exists {
		return nil, fmt.Errorf("asset %s not found in wallet", ticker)
	}

	return asset, nil
}
