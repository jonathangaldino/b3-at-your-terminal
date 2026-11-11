package wallet

import (
	"testing"
	"time"

	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
)

func TestGetDistributionBySubType(t *testing.T) {
	// Create test transactions
	transactions := []parser.Transaction{
		{
			Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "PETR4",
			Quantity:    decimal.NewFromInt(100),
			Price:       decimal.NewFromFloat(30.00),
			Amount:      decimal.NewFromFloat(3000.00),
			Hash:        "hash1",
		},
		{
			Date:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "VALE3",
			Quantity:    decimal.NewFromInt(200),
			Price:       decimal.NewFromFloat(50.00),
			Amount:      decimal.NewFromFloat(10000.00),
			Hash:        "hash2",
		},
		{
			Date:        time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "MXRF11",
			Quantity:    decimal.NewFromInt(500),
			Price:       decimal.NewFromFloat(10.00),
			Amount:      decimal.NewFromFloat(5000.00),
			Hash:        "hash3",
		},
	}

	// Create wallet
	w := NewWallet(transactions)

	// Set SubTypes manually
	w.Assets["PETR4"].SubType = "Stocks"
	w.Assets["VALE3"].SubType = "Stocks"
	w.Assets["MXRF11"].SubType = "Real Estate Funds"

	// Get distribution
	distribution := w.GetDistributionBySubType()

	// Verify results
	if len(distribution) != 2 {
		t.Errorf("Expected 2 SubTypes, got %d", len(distribution))
	}

	// Find Stocks distribution
	var stocksDist *DistributionItem
	var fundsDist *DistributionItem
	for i := range distribution {
		if distribution[i].Category == "Stocks" {
			stocksDist = &distribution[i]
		} else if distribution[i].Category == "Real Estate Funds" {
			fundsDist = &distribution[i]
		}
	}

	if stocksDist == nil {
		t.Fatal("Stocks distribution not found")
	}
	if fundsDist == nil {
		t.Fatal("Real Estate Funds distribution not found")
	}

	// Total invested: 3000 + 10000 + 5000 = 18000
	// Stocks: 3000 + 10000 = 13000 (72.22%)
	// Funds: 5000 (27.78%)

	expectedStocksPercent := decimal.NewFromFloat(72.22)
	expectedFundsPercent := decimal.NewFromFloat(27.78)

	if !stocksDist.Percentage.Round(2).Equal(expectedStocksPercent) {
		t.Errorf("Expected Stocks percentage ~72.22%%, got %s%%", stocksDist.Percentage.StringFixed(2))
	}

	if !fundsDist.Percentage.Round(2).Equal(expectedFundsPercent) {
		t.Errorf("Expected Funds percentage ~27.78%%, got %s%%", fundsDist.Percentage.StringFixed(2))
	}

	// Verify amounts
	expectedStocksAmount := decimal.NewFromFloat(13000.00)
	expectedFundsAmount := decimal.NewFromFloat(5000.00)

	if !stocksDist.Amount.Equal(expectedStocksAmount) {
		t.Errorf("Expected Stocks amount R$ 13000.00, got R$ %s", stocksDist.Amount.StringFixed(2))
	}

	if !fundsDist.Amount.Equal(expectedFundsAmount) {
		t.Errorf("Expected Funds amount R$ 5000.00, got R$ %s", fundsDist.Amount.StringFixed(2))
	}

	// Verify sorting (highest percentage first)
	if distribution[0].Percentage.LessThan(distribution[1].Percentage) {
		t.Error("Distribution should be sorted by percentage (highest first)")
	}
}

func TestGetDistributionBySegment(t *testing.T) {
	// Create test transactions
	transactions := []parser.Transaction{
		{
			Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "PETR4",
			Quantity:    decimal.NewFromInt(100),
			Price:       decimal.NewFromFloat(30.00),
			Amount:      decimal.NewFromFloat(3000.00),
			Hash:        "hash1",
		},
		{
			Date:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "VALE3",
			Quantity:    decimal.NewFromInt(200),
			Price:       decimal.NewFromFloat(50.00),
			Amount:      decimal.NewFromFloat(10000.00),
			Hash:        "hash2",
		},
		{
			Date:        time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "BBDC4",
			Quantity:    decimal.NewFromInt(300),
			Price:       decimal.NewFromFloat(20.00),
			Amount:      decimal.NewFromFloat(6000.00),
			Hash:        "hash3",
		},
	}

	// Create wallet
	w := NewWallet(transactions)

	// Set Segments manually
	w.Assets["PETR4"].Segment = "Oil & Gas"
	w.Assets["VALE3"].Segment = "Mining"
	w.Assets["BBDC4"].Segment = "Banking"

	// Get distribution
	distribution := w.GetDistributionBySegment()

	// Verify results
	if len(distribution) != 3 {
		t.Errorf("Expected 3 Segments, got %d", len(distribution))
	}

	// Total invested: 3000 + 10000 + 6000 = 19000
	// Mining: 10000 (52.63%) - should be first
	// Banking: 6000 (31.58%) - should be second
	// Oil & Gas: 3000 (15.79%) - should be third

	if distribution[0].Category != "Mining" {
		t.Errorf("Expected first item to be Mining, got %s", distribution[0].Category)
	}

	if distribution[1].Category != "Banking" {
		t.Errorf("Expected second item to be Banking, got %s", distribution[1].Category)
	}

	if distribution[2].Category != "Oil & Gas" {
		t.Errorf("Expected third item to be Oil & Gas, got %s", distribution[2].Category)
	}
}

func TestGetDistributionWithSoldAssets(t *testing.T) {
	// Create test transactions including a sale
	transactions := []parser.Transaction{
		{
			Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "PETR4",
			Quantity:    decimal.NewFromInt(100),
			Price:       decimal.NewFromFloat(30.00),
			Amount:      decimal.NewFromFloat(3000.00),
			Hash:        "hash1",
		},
		{
			Date:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:        "Compra",
			Institution: "XP",
			Ticker:      "VALE3",
			Quantity:    decimal.NewFromInt(100),
			Price:       decimal.NewFromFloat(50.00),
			Amount:      decimal.NewFromFloat(5000.00),
			Hash:        "hash2",
		},
		{
			Date:        time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			Type:        "Venda",
			Institution: "XP",
			Ticker:      "VALE3",
			Quantity:    decimal.NewFromInt(100),
			Price:       decimal.NewFromFloat(55.00),
			Amount:      decimal.NewFromFloat(5500.00),
			Hash:        "hash3",
		},
	}

	// Create wallet
	w := NewWallet(transactions)

	// Set SubTypes
	w.Assets["PETR4"].SubType = "Stocks"
	w.Assets["VALE3"].SubType = "Stocks"

	// Get distribution
	distribution := w.GetDistributionBySubType()

	// Should only count PETR4 (VALE3 has quantity 0)
	if len(distribution) != 1 {
		t.Errorf("Expected 1 SubType (only active assets), got %d", len(distribution))
	}

	if distribution[0].Category != "Stocks" {
		t.Errorf("Expected Stocks, got %s", distribution[0].Category)
	}

	// Should be 100% since only PETR4 is active
	expectedPercent := decimal.NewFromFloat(100.0)
	if !distribution[0].Percentage.Equal(expectedPercent) {
		t.Errorf("Expected 100%%, got %s%%", distribution[0].Percentage.StringFixed(2))
	}

	// Amount should only be PETR4's investment
	expectedAmount := decimal.NewFromFloat(3000.00)
	if !distribution[0].Amount.Equal(expectedAmount) {
		t.Errorf("Expected R$ 3000.00, got R$ %s", distribution[0].Amount.StringFixed(2))
	}
}

func TestGetDistributionWithEmptyWallet(t *testing.T) {
	// Create empty wallet
	w := NewWallet([]parser.Transaction{})

	// Get distributions
	subTypeDistribution := w.GetDistributionBySubType()
	segmentDistribution := w.GetDistributionBySegment()

	// Should be empty
	if len(subTypeDistribution) != 0 {
		t.Errorf("Expected empty SubType distribution, got %d items", len(subTypeDistribution))
	}

	if len(segmentDistribution) != 0 {
		t.Errorf("Expected empty Segment distribution, got %d items", len(segmentDistribution))
	}
}
