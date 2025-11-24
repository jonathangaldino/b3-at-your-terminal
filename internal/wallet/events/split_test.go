package events

import (
	"testing"
	"time"

	"github.com/john/b3-project/internal/parser"
	"github.com/john/b3-project/internal/wallet"
	"github.com/shopspring/decimal"
)

func TestApplySplit(t *testing.T) {
	// Create test transactions
	// Transaction 1: Before event date (should be adjusted)
	tx1 := parser.Transaction{
		Date:        time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		Type:        "Compra",
		Institution: "XP",
		Ticker:      "ITSA4",
		Quantity:    decimal.NewFromInt(100),
		Price:       decimal.NewFromFloat(10.50),
		Amount:      decimal.NewFromFloat(1050),
	}
	tx1.Hash = parser.CalculateHash(&tx1)

	// Transaction 2: Before event date (should be adjusted)
	tx2 := parser.Transaction{
		Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
		Type:        "Compra",
		Institution: "XP",
		Ticker:      "ITSA4",
		Quantity:    decimal.NewFromInt(50),
		Price:       decimal.NewFromFloat(11.00),
		Amount:      decimal.NewFromFloat(550),
	}
	tx2.Hash = parser.CalculateHash(&tx2)

	// Transaction 3: After event date (should NOT be adjusted)
	tx3 := parser.Transaction{
		Date:        time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Type:        "Compra",
		Institution: "XP",
		Ticker:      "ITSA4",
		Quantity:    decimal.NewFromInt(100), // Already in new quantity
		Price:       decimal.NewFromFloat(5.25), // Already in new price
		Amount:      decimal.NewFromFloat(525),
	}
	tx3.Hash = parser.CalculateHash(&tx3)

	// Create wallet
	transactions := []parser.Transaction{tx1, tx2, tx3}
	w := wallet.NewWallet(transactions)

	// Event date: 2024-05-01
	eventDate := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	// Apply split 1:2
	ratio := SplitRatio{From: 1, To: 2}
	result, err := ApplySplit(w, "ITSA4", ratio, eventDate)

	if err != nil {
		t.Fatalf("Error applying split: %v", err)
	}

	// Verify result statistics
	if result.Ticker != "ITSA4" {
		t.Errorf("Result.Ticker = %s, expected ITSA4", result.Ticker)
	}

	if result.TransactionsAdjusted != 2 {
		t.Errorf("TransactionsAdjusted = %d, expected 2", result.TransactionsAdjusted)
	}

	// Verify asset exists and has correct number of transactions
	asset, exists := w.Assets["ITSA4"]
	if !exists {
		t.Fatal("Asset ITSA4 not found after split")
	}

	if len(asset.Negotiations) != 3 {
		t.Errorf("Asset has %d transactions, expected 3", len(asset.Negotiations))
	}

	// Verify first transaction (before event) was adjusted
	adjustedTx1 := asset.Negotiations[0]
	expectedQty1 := decimal.NewFromInt(200) // 100 × 2 = 200
	if !adjustedTx1.Quantity.Equal(expectedQty1) {
		t.Errorf("Transaction 1 quantity = %s, expected %s", adjustedTx1.Quantity, expectedQty1)
	}

	expectedPrice1 := decimal.NewFromFloat(5.25) // 10.50 ÷ 2 = 5.25
	if !adjustedTx1.Price.Equal(expectedPrice1) {
		t.Errorf("Transaction 1 price = %s, expected %s", adjustedTx1.Price, expectedPrice1)
	}

	// Amount should remain the same: 200 × 5.25 = 1050
	expectedAmount1 := decimal.NewFromFloat(1050)
	if !adjustedTx1.Amount.Equal(expectedAmount1) {
		t.Errorf("Transaction 1 amount = %s, expected %s", adjustedTx1.Amount, expectedAmount1)
	}

	// Verify second transaction (before event) was adjusted
	adjustedTx2 := asset.Negotiations[1]
	expectedQty2 := decimal.NewFromInt(100) // 50 × 2 = 100
	if !adjustedTx2.Quantity.Equal(expectedQty2) {
		t.Errorf("Transaction 2 quantity = %s, expected %s", adjustedTx2.Quantity, expectedQty2)
	}

	expectedPrice2 := decimal.NewFromFloat(5.50) // 11.00 ÷ 2 = 5.50
	if !adjustedTx2.Price.Equal(expectedPrice2) {
		t.Errorf("Transaction 2 price = %s, expected %s", adjustedTx2.Price, expectedPrice2)
	}

	// Verify third transaction (after event) was NOT adjusted
	unadjustedTx3 := asset.Negotiations[2]
	if !unadjustedTx3.Quantity.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Transaction 3 quantity should not change, got %s", unadjustedTx3.Quantity)
	}

	if !unadjustedTx3.Price.Equal(decimal.NewFromFloat(5.25)) {
		t.Errorf("Transaction 3 price should not change, got %s", unadjustedTx3.Price)
	}

	// Verify total quantity after split
	// (200 + 100 + 100) = 400
	expectedTotalQty := 400
	if asset.Quantity != expectedTotalQty {
		t.Errorf("Total quantity = %d, expected %d", asset.Quantity, expectedTotalQty)
	}

	// Verify hashes were recalculated for adjusted transactions
	if adjustedTx1.Hash == tx1.Hash {
		t.Error("Transaction 1 hash should have been recalculated")
	}

	if adjustedTx2.Hash == tx2.Hash {
		t.Error("Transaction 2 hash should have been recalculated")
	}

	if unadjustedTx3.Hash != tx3.Hash {
		t.Error("Transaction 3 hash should NOT have been recalculated")
	}

	// Verify wallet-level transactions were also updated
	walletTxCount := 0
	for _, tx := range w.Transactions {
		if tx.Ticker == "ITSA4" {
			walletTxCount++
		}
	}

	if walletTxCount != 3 {
		t.Errorf("Wallet has %d ITSA4 transactions, expected 3", walletTxCount)
	}
}

func TestApplySplit_DifferentRatios(t *testing.T) {
	tests := []struct {
		name          string
		ratio         SplitRatio
		initialQty    int64
		initialPrice  float64
		expectedQty   int64
		expectedPrice float64
	}{
		{
			name:          "1:2 ratio",
			ratio:         SplitRatio{From: 1, To: 2},
			initialQty:    100,
			initialPrice:  10.00,
			expectedQty:   200,
			expectedPrice: 5.00,
		},
		{
			name:          "1:3 ratio",
			ratio:         SplitRatio{From: 1, To: 3},
			initialQty:    100,
			initialPrice:  30.00,
			expectedQty:   300,
			expectedPrice: 10.00,
		},
		{
			name:          "1:4 ratio",
			ratio:         SplitRatio{From: 1, To: 4},
			initialQty:    100,
			initialPrice:  40.00,
			expectedQty:   400,
			expectedPrice: 10.00,
		},
		{
			name:          "1:5 ratio",
			ratio:         SplitRatio{From: 1, To: 5},
			initialQty:    100,
			initialPrice:  50.00,
			expectedQty:   500,
			expectedPrice: 10.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create transaction
			tx := parser.Transaction{
				Date:        time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
				Type:        "Compra",
				Institution: "XP",
				Ticker:      "TEST",
				Quantity:    decimal.NewFromInt(tt.initialQty),
				Price:       decimal.NewFromFloat(tt.initialPrice),
				Amount:      decimal.NewFromInt(tt.initialQty).Mul(decimal.NewFromFloat(tt.initialPrice)),
			}
			tx.Hash = parser.CalculateHash(&tx)

			// Create wallet
			w := wallet.NewWallet([]parser.Transaction{tx})

			// Apply split
			eventDate := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
			_, err := ApplySplit(w, "TEST", tt.ratio, eventDate)

			if err != nil {
				t.Fatalf("Error applying split: %v", err)
			}

			// Verify quantity
			asset := w.Assets["TEST"]
			if asset.Quantity != int(tt.expectedQty) {
				t.Errorf("Quantity = %d, expected %d", asset.Quantity, tt.expectedQty)
			}

			// Verify adjusted transaction
			adjustedTx := asset.Negotiations[0]
			if !adjustedTx.Quantity.Equal(decimal.NewFromInt(tt.expectedQty)) {
				t.Errorf("Adjusted quantity = %s, expected %d", adjustedTx.Quantity, tt.expectedQty)
			}

			if !adjustedTx.Price.Equal(decimal.NewFromFloat(tt.expectedPrice)) {
				t.Errorf("Adjusted price = %s, expected %.2f", adjustedTx.Price, tt.expectedPrice)
			}

			// Verify amount stayed the same
			expectedAmount := decimal.NewFromInt(tt.initialQty).Mul(decimal.NewFromFloat(tt.initialPrice))
			if !adjustedTx.Amount.Equal(expectedAmount) {
				t.Errorf("Amount changed! Got %s, expected %s", adjustedTx.Amount, expectedAmount)
			}
		})
	}
}

func TestApplySplit_Errors(t *testing.T) {
	// Create a simple wallet
	tx := parser.Transaction{
		Date:     time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		Type:     "Compra",
		Ticker:   "PETR4",
		Quantity: decimal.NewFromInt(100),
		Price:    decimal.NewFromFloat(30.00),
		Amount:   decimal.NewFromFloat(3000),
	}
	tx.Hash = parser.CalculateHash(&tx)

	w := wallet.NewWallet([]parser.Transaction{tx})
	eventDate := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	t.Run("Asset not found", func(t *testing.T) {
		ratio := SplitRatio{From: 1, To: 2}
		_, err := ApplySplit(w, "INVALID", ratio, eventDate)

		if err == nil {
			t.Error("Should return error for non-existent asset")
		}
	})

	t.Run("Invalid ratio - From not 1", func(t *testing.T) {
		ratio := SplitRatio{From: 2, To: 4}
		_, err := ApplySplit(w, "PETR4", ratio, eventDate)

		if err == nil {
			t.Error("Should return error for From != 1")
		}
	})

	t.Run("Invalid ratio - To too small", func(t *testing.T) {
		ratio := SplitRatio{From: 1, To: 1}
		_, err := ApplySplit(w, "PETR4", ratio, eventDate)

		if err == nil {
			t.Error("Should return error for To < 2")
		}
	})

	t.Run("Zero ratio", func(t *testing.T) {
		ratio := SplitRatio{From: 1, To: 0}
		_, err := ApplySplit(w, "PETR4", ratio, eventDate)

		if err == nil {
			t.Error("Should return error for To = 0")
		}
	})
}

func TestApplySplit_NoTransactionsBeforeEvent(t *testing.T) {
	// Create transaction AFTER event date
	tx := parser.Transaction{
		Date:     time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
		Type:     "Compra",
		Ticker:   "VALE3",
		Quantity: decimal.NewFromInt(100),
		Price:    decimal.NewFromFloat(60.00),
		Amount:   decimal.NewFromFloat(6000),
	}
	tx.Hash = parser.CalculateHash(&tx)
	originalHash := tx.Hash

	w := wallet.NewWallet([]parser.Transaction{tx})

	// Event date is BEFORE the transaction
	eventDate := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	ratio := SplitRatio{From: 1, To: 2}

	result, err := ApplySplit(w, "VALE3", ratio, eventDate)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// No transactions should be adjusted
	if result.TransactionsAdjusted != 0 {
		t.Errorf("TransactionsAdjusted = %d, expected 0", result.TransactionsAdjusted)
	}

	// Transaction should remain unchanged
	asset := w.Assets["VALE3"]
	tx1 := asset.Negotiations[0]

	if !tx1.Quantity.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Quantity should not change, got %s", tx1.Quantity)
	}

	if !tx1.Price.Equal(decimal.NewFromFloat(60.00)) {
		t.Errorf("Price should not change, got %s", tx1.Price)
	}

	// Hash should not change
	if tx1.Hash != originalHash {
		t.Error("Hash should not change for transaction after event date")
	}
}

func TestParseSplitRatio(t *testing.T) {
	tests := []struct {
		input        string
		expectedOk   bool
		expectedFrom int
		expectedTo   int
	}{
		{"1:2", true, 1, 2},
		{"1:3", true, 1, 3},
		{"1:4", true, 1, 4},
		{"1:5", true, 1, 5},
		{"1:10", true, 1, 10},
		{"invalid", false, 0, 0},
		{"1", false, 0, 0},
		{"1:2:3", false, 0, 0},
		{"", false, 0, 0},
		{"abc:def", false, 0, 0},
		{"2:4", true, 2, 4}, // Valid parse but will fail validation in ApplySplit
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ratio, err := ParseSplitRatio(tt.input)

			if tt.expectedOk {
				if err != nil {
					t.Errorf("ParseSplitRatio(%q) returned error: %v", tt.input, err)
				}
				if ratio.From != tt.expectedFrom {
					t.Errorf("From = %d, expected %d", ratio.From, tt.expectedFrom)
				}
				if ratio.To != tt.expectedTo {
					t.Errorf("To = %d, expected %d", ratio.To, tt.expectedTo)
				}
			} else {
				if err == nil {
					t.Errorf("ParseSplitRatio(%q) should return error", tt.input)
				}
			}
		})
	}
}

func TestFormatSplitRatio(t *testing.T) {
	tests := []struct {
		ratio    SplitRatio
		expected string
	}{
		{SplitRatio{From: 1, To: 2}, "1:2"},
		{SplitRatio{From: 1, To: 3}, "1:3"},
		{SplitRatio{From: 1, To: 4}, "1:4"},
		{SplitRatio{From: 1, To: 10}, "1:10"},
	}

	for _, tt := range tests {
		result := FormatSplitRatio(tt.ratio)
		if result != tt.expected {
			t.Errorf("FormatSplitRatio(%+v) = %q, expected %q", tt.ratio, result, tt.expected)
		}
	}
}
