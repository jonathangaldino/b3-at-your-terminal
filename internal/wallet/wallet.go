package wallet

import (
	"fmt"

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

// ConversionResult contém estatísticas sobre a conversão de subscrição
type ConversionResult struct {
	PurchasesFound      int
	SalesFound          int
	TransactionsAdded   int
	ParentQuantityBefore int
	ParentQuantityAfter  int
	ParentAveragePrice   decimal.Decimal
}

// ConvertSubscriptionToParent converte um ativo de subscrição para o ativo pai
// Remove o ativo de subscrição, ignora vendas e converte compras em compras do ativo pai
func (w *Wallet) ConvertSubscriptionToParent(subscriptionTicker, parentTicker string) (*ConversionResult, error) {
	// Verificar se o ativo de subscrição existe
	subscriptionAsset, exists := w.Assets[subscriptionTicker]
	if !exists {
		return nil, fmt.Errorf("ativo de subscrição %s não encontrado", subscriptionTicker)
	}

	// Verificar se o ativo pai existe, se não criar
	parentAsset, exists := w.Assets[parentTicker]
	if !exists {
		parentAsset = &Asset{
			ID:           parentTicker,
			Negotiations: make([]parser.Transaction, 0),
			Type:         "renda variável",
			SubType:      "",
			Segment:      "",
			AveragePrice: decimal.Zero,
		}
		w.Assets[parentTicker] = parentAsset
	}

	result := &ConversionResult{
		ParentQuantityBefore: parentAsset.Quantity,
		ParentAveragePrice:   parentAsset.AveragePrice,
	}

	// Coletar transações a serem removidas e transformadas
	var transactionsToRemove []string // hashes
	var transactionsToAdd []parser.Transaction

	for _, transaction := range subscriptionAsset.Negotiations {
		if transaction.Type == "Venda" {
			// Ignorar vendas de subscrição
			result.SalesFound++
			transactionsToRemove = append(transactionsToRemove, transaction.Hash)
		} else if transaction.Type == "Compra" {
			// Transformar compra para o ativo pai
			result.PurchasesFound++

			// Criar nova transação com ticker do pai
			newTransaction := parser.Transaction{
				Date:        transaction.Date,
				Type:        transaction.Type,
				Institution: transaction.Institution,
				Ticker:      parentTicker, // Mudança principal!
				Quantity:    transaction.Quantity,
				Price:       transaction.Price,
				Amount:      transaction.Amount,
				Hash:        "", // Será calculado
			}

			// Calcular novo hash
			newTransaction.Hash = parser.CalculateHash(&newTransaction)

			// Verificar se já existe (evitar colisão)
			if _, exists := w.TransactionsByHash[newTransaction.Hash]; !exists {
				transactionsToAdd = append(transactionsToAdd, newTransaction)
				result.TransactionsAdded++
			}

			// Adicionar hash antigo à lista de remoção
			transactionsToRemove = append(transactionsToRemove, transaction.Hash)
		}
	}

	// Remover transações antigas do wallet
	for _, hash := range transactionsToRemove {
		delete(w.TransactionsByHash, hash)
	}

	// Remover transações antigas do array Transactions
	newTransactions := make([]parser.Transaction, 0)
	for _, t := range w.Transactions {
		// Manter apenas se não for transação de subscrição
		shouldKeep := true
		for _, hash := range transactionsToRemove {
			if t.Hash == hash {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newTransactions = append(newTransactions, t)
		}
	}
	w.Transactions = newTransactions

	// Adicionar novas transações ao wallet
	for _, t := range transactionsToAdd {
		w.Transactions = append(w.Transactions, t)
		w.TransactionsByHash[t.Hash] = t

		// Adicionar à lista de negociações do ativo pai
		parentAsset.Negotiations = append(parentAsset.Negotiations, t)
	}

	// Remover ativo de subscrição
	delete(w.Assets, subscriptionTicker)

	// Recalcular campos derivados
	w.RecalculateAssets()

	// Atualizar resultado
	if parentAsset, exists := w.Assets[parentTicker]; exists {
		result.ParentQuantityAfter = parentAsset.Quantity
		result.ParentAveragePrice = parentAsset.AveragePrice
	}

	return result, nil
}
