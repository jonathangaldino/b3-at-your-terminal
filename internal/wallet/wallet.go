package wallet

import (
	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
)

// Wallet representa a carteira de investimentos completa
type Wallet struct {
	// Transactions é a lista de todas as transações
	Transactions []parser.Transaction

	// TransactionsByHash permite acesso rápido às transações por hash
	TransactionsByHash map[string]parser.Transaction

	// Assets mapeia ticker -> Asset para acesso rápido aos ativos
	Assets map[string]*Asset
}

// NewWallet cria uma nova Wallet a partir de uma lista de transações
// Automaticamente deduplica transações, cria Assets e calcula preços médios
func NewWallet(transactions []parser.Transaction) *Wallet {
	w := &Wallet{
		Transactions:       make([]parser.Transaction, 0),
		TransactionsByHash: make(map[string]parser.Transaction),
		Assets:             make(map[string]*Asset),
	}

	for _, t := range transactions {
		// Verificar se já existe (deduplicação por hash)
		if _, exists := w.TransactionsByHash[t.Hash]; exists {
			continue
		}

		// Adicionar transação à wallet
		w.Transactions = append(w.Transactions, t)
		w.TransactionsByHash[t.Hash] = t

		// Verificar se Asset já existe para este ticker
		asset, exists := w.Assets[t.Ticker]
		if !exists {
			// Criar novo Asset
			asset = &Asset{
				ID:           t.Ticker,
				Negotiations: make([]parser.Transaction, 0),
				Type:         "renda variável",
				SubType:      "", // Usuário define manualmente
				Segment:      "", // Usuário define manualmente
				AveragePrice: decimal.Zero,
			}
			w.Assets[t.Ticker] = asset
		}

		// Adicionar transação às negociações do asset
		asset.Negotiations = append(asset.Negotiations, t)

		// Recalcular campos derivados
		asset.AveragePrice = calculateAveragePrice(asset)
		asset.TotalInvestedValue = calculateTotalInvestedValue(asset)
		asset.Quantity = calculateQuantity(asset)
	}

	return w
}

// RecalculateAssets recalcula todos os campos derivados de todos os Assets
func (w *Wallet) RecalculateAssets() {
	for _, asset := range w.Assets {
		asset.AveragePrice = calculateAveragePrice(asset)
		asset.TotalInvestedValue = calculateTotalInvestedValue(asset)
		asset.Quantity = calculateQuantity(asset)
	}
}
