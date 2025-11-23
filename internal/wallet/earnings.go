package wallet

import (
	"fmt"

	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
)

// AddEarning adds a single earning to the wallet.
// It validates the earning, checks for duplicates, creates the asset if needed,
// and recalculates all asset values.
// Returns an error if the earning is invalid or already exists.
func (w *Wallet) AddEarning(earning parser.Earning) error {
	// Validate earning
	if err := ValidateEarning(&earning); err != nil {
		return fmt.Errorf("invalid earning: %w", err)
	}

	// Calculate hash if not already set
	if earning.Hash == "" {
		earning.Hash = parser.CalculateEarningHash(&earning)
	}

	// Create or update asset
	asset, exists := w.Assets[earning.Ticker]
	if !exists {
		// Create new asset for this ticker
		asset = &Asset{
			ID:           earning.Ticker,
			Negotiations: make([]parser.Transaction, 0),
			Earnings:     make([]parser.Earning, 0),
			Type:         "renda variável",
			SubType:      "",
			Segment:      "",
		}
		w.Assets[earning.Ticker] = asset
	}

	// Check for duplicate within asset's earnings
	for _, e := range asset.Earnings {
		if e.Hash == earning.Hash {
			return fmt.Errorf("duplicate earning detected")
		}
	}

	// Add earning to asset
	asset.Earnings = append(asset.Earnings, earning)

	// Recalculate all asset values
	w.RecalculateAssets()

	return nil
}

// AddEarnings adds multiple earnings to the wallet in batch.
// It returns the number of earnings added, the number of duplicates skipped,
// and any error that occurred during validation.
// If an earning is invalid, the entire operation is aborted and an error is returned.
func (w *Wallet) AddEarnings(earnings []parser.Earning) (added int, duplicates int, err error) {
	// Track seen hashes across all assets
	seenHashes := make(map[string]bool)

	// Populate seenHashes with existing earnings
	for _, asset := range w.Assets {
		for _, e := range asset.Earnings {
			seenHashes[e.Hash] = true
		}
	}

	for _, earning := range earnings {
		// Calculate hash if not already set
		if earning.Hash == "" {
			earning.Hash = parser.CalculateEarningHash(&earning)
		}

		// Check for duplicate
		if seenHashes[earning.Hash] {
			duplicates++
			continue
		}

		// Validate earning
		if err := ValidateEarning(&earning); err != nil {
			return added, duplicates, fmt.Errorf("invalid earning for %s: %w", earning.Ticker, err)
		}

		// Create or update asset
		asset, exists := w.Assets[earning.Ticker]
		if !exists {
			// Create new asset for this ticker
			asset = &Asset{
				ID:           earning.Ticker,
				Negotiations: make([]parser.Transaction, 0),
				Earnings:     make([]parser.Earning, 0),
				Type:         "renda variável",
				SubType:      "",
				Segment:      "",
			}
			w.Assets[earning.Ticker] = asset
		}

		// Add earning to asset
		asset.Earnings = append(asset.Earnings, earning)
		seenHashes[earning.Hash] = true
		added++
	}

	// Recalculate all asset values after adding all earnings
	if added > 0 {
		w.RecalculateAssets()
	}

	return added, duplicates, nil
}

// ValidateEarning validates that an earning has all required fields
// and that numeric values are valid.
func ValidateEarning(e *parser.Earning) error {
	if e.Ticker == "" {
		return fmt.Errorf("ticker is required")
	}

	if e.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Validate type is one of the three expected values
	validTypes := map[string]bool{
		"Rendimento":                   true,
		"Dividendo":                    true,
		"Juros Sobre Capital Próprio": true,
	}

	if !validTypes[e.Type] {
		return fmt.Errorf("type must be 'Rendimento', 'Dividendo', or 'Juros Sobre Capital Próprio' (received: '%s')", e.Type)
	}

	if e.Quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("quantity must be greater than zero")
	}

	if e.UnitPrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("unit price must be greater than zero")
	}

	if e.TotalAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("total amount must be greater than zero")
	}

	return nil
}

// calculateTotalEarnings calcula o valor total de proventos recebidos de um ativo
// Soma todos os valores de earnings
func calculateTotalEarnings(asset *Asset) decimal.Decimal {
	total := decimal.Zero

	for _, earning := range asset.Earnings {
		total = total.Add(earning.TotalAmount)
	}

	return total.Round(4)
}
