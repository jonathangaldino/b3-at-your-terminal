package wallet

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
