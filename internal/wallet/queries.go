package wallet

import (
	"github.com/shopspring/decimal"
)

// GetActiveAssets returns all assets that have a non-zero quantity.
// These are assets currently held in the portfolio.
func (w *Wallet) GetActiveAssets() map[string]*Asset {
	activeAssets := make(map[string]*Asset)

	for ticker, asset := range w.Assets {
		if asset.Quantity != 0 {
			activeAssets[ticker] = asset
		}
	}

	return activeAssets
}

// GetSoldAssets returns all assets that have been completely sold (quantity == 0).
// These assets have transaction history but are no longer held.
func (w *Wallet) GetSoldAssets() map[string]*Asset {
	soldAssets := make(map[string]*Asset)

	for ticker, asset := range w.Assets {
		if asset.Quantity == 0 {
			soldAssets[ticker] = asset
		}
	}

	return soldAssets
}

// GroupKey represents a group in the asset hierarchy
type GroupKey struct {
	Type    string
	Segment string
}

// GroupAssetsByTypeAndSegment groups assets by their Type and Segment fields.
// Returns a nested map: GroupKey -> []*Asset
// This is useful for displaying assets organized by category.
func (w *Wallet) GroupAssetsByTypeAndSegment() map[GroupKey][]*Asset {
	groups := make(map[GroupKey][]*Asset)

	for _, asset := range w.Assets {
		key := GroupKey{
			Type:    asset.Type,
			Segment: asset.Segment,
		}

		groups[key] = append(groups[key], asset)
	}

	return groups
}

// GroupActiveAssetsByTypeAndSegment groups only active assets (quantity != 0)
// by their Type and Segment fields.
func (w *Wallet) GroupActiveAssetsByTypeAndSegment() map[GroupKey][]*Asset {
	groups := make(map[GroupKey][]*Asset)

	for _, asset := range w.Assets {
		if asset.Quantity == 0 {
			continue
		}

		key := GroupKey{
			Type:    asset.Type,
			Segment: asset.Segment,
		}

		groups[key] = append(groups[key], asset)
	}

	return groups
}

// DistributionItem represents a distribution percentage for a category
type DistributionItem struct {
	Category   string
	Percentage decimal.Decimal
	Amount     decimal.Decimal
}

// GetDistributionBySubType calculates the portfolio distribution by SubType.
// Returns a slice of DistributionItems sorted by percentage (highest first).
// Only considers active assets (quantity != 0).
func (w *Wallet) GetDistributionBySubType() []DistributionItem {
	// Calculate total invested value across all active assets
	totalValue := decimal.Zero
	subTypeValues := make(map[string]decimal.Decimal)

	for _, asset := range w.Assets {
		if asset.Quantity == 0 {
			continue
		}

		totalValue = totalValue.Add(asset.TotalInvestedValue)

		// Accumulate value per SubType
		current := subTypeValues[asset.SubType]
		subTypeValues[asset.SubType] = current.Add(asset.TotalInvestedValue)
	}

	// Convert to distribution items
	var items []DistributionItem
	for subType, amount := range subTypeValues {
		percentage := decimal.Zero
		if !totalValue.IsZero() {
			percentage = amount.Div(totalValue).Mul(decimal.NewFromInt(100))
		}

		items = append(items, DistributionItem{
			Category:   subType,
			Percentage: percentage,
			Amount:     amount,
		})
	}

	// Sort by percentage (highest first)
	sortDistributionItems(items)

	return items
}

// GetDistributionBySegment calculates the portfolio distribution by Segment.
// Returns a slice of DistributionItems sorted by percentage (highest first).
// Only considers active assets (quantity != 0).
func (w *Wallet) GetDistributionBySegment() []DistributionItem {
	// Calculate total invested value across all active assets
	totalValue := decimal.Zero
	segmentValues := make(map[string]decimal.Decimal)

	for _, asset := range w.Assets {
		if asset.Quantity == 0 {
			continue
		}

		totalValue = totalValue.Add(asset.TotalInvestedValue)

		// Accumulate value per Segment
		current := segmentValues[asset.Segment]
		segmentValues[asset.Segment] = current.Add(asset.TotalInvestedValue)
	}

	// Convert to distribution items
	var items []DistributionItem
	for segment, amount := range segmentValues {
		percentage := decimal.Zero
		if !totalValue.IsZero() {
			percentage = amount.Div(totalValue).Mul(decimal.NewFromInt(100))
		}

		items = append(items, DistributionItem{
			Category:   segment,
			Percentage: percentage,
			Amount:     amount,
		})
	}

	// Sort by percentage (highest first)
	sortDistributionItems(items)

	return items
}

// sortDistributionItems sorts items by percentage in descending order
func sortDistributionItems(items []DistributionItem) {
	// Simple bubble sort (good enough for small lists)
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if items[j].Percentage.LessThan(items[j+1].Percentage) {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}
}
