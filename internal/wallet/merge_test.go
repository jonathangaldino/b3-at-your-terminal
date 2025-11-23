package wallet

import (
	"testing"
	"time"

	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
)

func TestIsFractionalTicker(t *testing.T) {
	tests := []struct {
		ticker   string
		expected bool
	}{
		{"KLBN3F", true},
		{"ITSA4F", true},
		{"KLBN3", false},
		{"ITSA4", false},
		{"", false},
		{"F", true},
	}

	for _, tt := range tests {
		result := IsFractionalTicker(tt.ticker)
		if result != tt.expected {
			t.Errorf("IsFractionalTicker(%q) = %v, expected %v", tt.ticker, result, tt.expected)
		}
	}
}

func TestMergeFractionalAsset(t *testing.T) {
	// Criar transações para a wallet
	fractionalTx := parser.Transaction{
		Date:     time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		Type:     "Compra",
		Ticker:   "KLBN3F",
		Quantity: decimal.NewFromInt(100),
		Price:    decimal.NewFromFloat(10.50),
		Amount:   decimal.NewFromInt(1050),
	}
	fractionalTx.Hash = parser.CalculateHash(&fractionalTx)
	originalHash := fractionalTx.Hash

	normalTx := parser.Transaction{
		Date:     time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
		Type:     "Compra",
		Ticker:   "KLBN3",
		Quantity: decimal.NewFromInt(200),
		Price:    decimal.NewFromFloat(11.00),
		Amount:   decimal.NewFromInt(2200),
	}
	normalTx.Hash = parser.CalculateHash(&normalTx)

	// Criar wallet a partir das transações
	transactions := []parser.Transaction{fractionalTx, normalTx}
	w := NewWallet(transactions)

	// Recalcular antes do merge
	w.RecalculateAssets()

	// Fazer merge
	result, err := w.MergeFractionalAsset("KLBN3F")
	if err != nil {
		t.Fatalf("Erro ao fazer merge: %v", err)
	}

	// Verificar resultado
	if result.SourceTicker != "KLBN3F" {
		t.Errorf("SourceTicker = %q, expected KLBN3F", result.SourceTicker)
	}

	if result.TargetTicker != "KLBN3" {
		t.Errorf("TargetTicker = %q, expected KLBN3", result.TargetTicker)
	}

	if result.TransactionsMoved != 1 {
		t.Errorf("TransactionsMoved = %d, expected 1", result.TransactionsMoved)
	}

	// Verificar que ativo fracionário foi removido
	if _, exists := w.Assets["KLBN3F"]; exists {
		t.Error("Ativo KLBN3F ainda existe após merge")
	}

	// Verificar que ativo normal tem 2 transações
	klbn3 := w.Assets["KLBN3"]
	if len(klbn3.Negotiations) != 2 {
		t.Errorf("KLBN3 tem %d transações, expected 2", len(klbn3.Negotiations))
	}

	// Verificar quantidade total (100 + 200 = 300)
	if klbn3.Quantity != 300 {
		t.Errorf("KLBN3 quantity = %d, expected 300", klbn3.Quantity)
	}

	// Verificar que o hash foi recalculado na transação movida
	// A transação movida é a que tem data 2024-01-10
	var movedTx parser.Transaction
	found := false
	for _, tx := range klbn3.Negotiations {
		if tx.Date.Equal(time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)) {
			movedTx = tx
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Transação movida não encontrada")
	}

	if movedTx.Ticker != "KLBN3" {
		t.Errorf("Transação movida tem ticker %s, expected KLBN3", movedTx.Ticker)
	}

	// Hash deve ser diferente do original (porque ticker mudou)
	if movedTx.Hash == originalHash {
		t.Error("Hash não foi recalculado após mudar o ticker")
	}

	// Hash não deve estar vazio
	if movedTx.Hash == "" {
		t.Error("Hash está vazio após merge")
	}

	// Verificar que o hash está correto para o novo ticker
	expectedHash := parser.CalculateHash(&movedTx)
	if movedTx.Hash != expectedHash {
		t.Errorf("Hash incorreto. Got %s, expected %s", movedTx.Hash, expectedHash)
	}

	// Verificar que a lista global de transações também foi atualizada
	foundInGlobal := false
	for _, tx := range w.Transactions {
		if tx.Date.Equal(time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)) {
			if tx.Ticker != "KLBN3" {
				t.Errorf("Transação na lista global tem ticker %s, expected KLBN3", tx.Ticker)
			}
			if tx.Hash != movedTx.Hash {
				t.Errorf("Hash na lista global (%s) diferente do hash no asset (%s)", tx.Hash, movedTx.Hash)
			}
			foundInGlobal = true
			break
		}
	}

	if !foundInGlobal {
		t.Error("Transação movida não encontrada na lista global de transações")
	}

	// Verificar que não há mais transações com ticker KLBN3F na lista global
	for _, tx := range w.Transactions {
		if tx.Ticker == "KLBN3F" {
			t.Error("Ainda existe transação com ticker KLBN3F na lista global")
		}
	}
}

func TestCreateAndMergeFractionalAsset(t *testing.T) {
	// Criar wallet apenas com ativo fracionário
	w := &Wallet{
		Assets: make(map[string]*Asset),
	}

	// Criar ativo fracionário ITSA4F
	w.Assets["ITSA4F"] = &Asset{
		ID:      "ITSA4F",
		Type:    "renda variável",
		SubType: "ações",
		Segment: "bancos",
		Negotiations: []parser.Transaction{
			{
				Date:     time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
				Type:     "Compra",
				Ticker:   "ITSA4F",
				Quantity: decimal.NewFromInt(150),
				Price:    decimal.NewFromFloat(9.50),
				Amount:   decimal.NewFromInt(1425),
			},
		},
		Earnings: []parser.Earning{},
	}

	// Recalcular
	w.RecalculateAssets()

	// Criar e mesclar
	result, err := w.CreateAndMergeFractionalAsset("ITSA4F")
	if err != nil {
		t.Fatalf("Erro ao criar e mesclar: %v", err)
	}

	// Verificar resultado
	if result.TargetCreated != true {
		t.Error("TargetCreated deve ser true")
	}

	if result.TargetQuantityBefore != 0 {
		t.Errorf("TargetQuantityBefore = %d, expected 0", result.TargetQuantityBefore)
	}

	// Verificar que ativo fracionário foi removido
	if _, exists := w.Assets["ITSA4F"]; exists {
		t.Error("Ativo ITSA4F ainda existe após merge")
	}

	// Verificar que ativo normal foi criado
	itsa4, exists := w.Assets["ITSA4"]
	if !exists {
		t.Fatal("Ativo ITSA4 não foi criado")
	}

	// Verificar metadados copiados
	if itsa4.Type != "renda variável" {
		t.Errorf("Type = %q, expected 'renda variável'", itsa4.Type)
	}

	if itsa4.SubType != "ações" {
		t.Errorf("SubType = %q, expected 'ações'", itsa4.SubType)
	}

	if itsa4.Segment != "bancos" {
		t.Errorf("Segment = %q, expected 'bancos'", itsa4.Segment)
	}

	// Verificar transações
	if len(itsa4.Negotiations) != 1 {
		t.Errorf("ITSA4 tem %d transações, expected 1", len(itsa4.Negotiations))
	}

	// Verificar quantidade
	if itsa4.Quantity != 150 {
		t.Errorf("ITSA4 quantity = %d, expected 150", itsa4.Quantity)
	}
}

func TestMergeFractionalAssetErrors(t *testing.T) {
	w := &Wallet{
		Assets: make(map[string]*Asset),
	}

	// Tentar mesclar ticker que não termina com F
	_, err := w.MergeFractionalAsset("KLBN3")
	if err == nil {
		t.Error("Deveria retornar erro para ticker sem F")
	}

	// Tentar mesclar ticker que não existe
	_, err = w.MergeFractionalAsset("XYZF")
	if err == nil {
		t.Error("Deveria retornar erro para ticker inexistente")
	}

	// Criar ativo fracionário sem correspondente
	w.Assets["PETR4F"] = &Asset{
		ID:           "PETR4F",
		Negotiations: []parser.Transaction{},
	}

	// Tentar mesclar - deve retornar erro TARGET_NOT_FOUND
	_, err = w.MergeFractionalAsset("PETR4F")
	if err == nil {
		t.Error("Deveria retornar erro TARGET_NOT_FOUND")
	}

	if len(err.Error()) < 16 || err.Error()[:16] != "TARGET_NOT_FOUND" {
		t.Errorf("Erro deveria começar com TARGET_NOT_FOUND, got: %v", err)
	}
}
