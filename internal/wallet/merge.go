package wallet

import (
	"fmt"

	"github.com/john/b3-project/internal/parser"
)

// MergeResult contém os resultados de uma operação de merge
type MergeResult struct {
	SourceTicker         string
	TargetTicker         string
	TransactionsMoved    int
	EarningsMoved        int
	TargetCreated        bool
	TargetQuantityBefore int
	TargetQuantityAfter  int
}

// MergeFractionalAsset mescla um ativo fracionário (com F) no ativo original
// Remove o "F" do ticker, procura o ativo original e mescla as transações
// Se o ativo original não existir, retorna erro indicando que deve criar
func (w *Wallet) MergeFractionalAsset(fractionalTicker string) (*MergeResult, error) {
	// Verificar se ticker termina com F
	if len(fractionalTicker) == 0 || fractionalTicker[len(fractionalTicker)-1] != 'F' {
		return nil, fmt.Errorf("ticker %s não é um ativo fracionário (não termina com F)", fractionalTicker)
	}

	// Verificar se ativo fracionário existe
	sourceAsset, exists := w.Assets[fractionalTicker]
	if !exists {
		return nil, fmt.Errorf("ativo fracionário %s não encontrado", fractionalTicker)
	}

	// Obter ticker normalizado (sem F)
	normalizedTicker := parser.NormalizeTicker(fractionalTicker)

	// Verificar se já é o mesmo (edge case)
	if normalizedTicker == fractionalTicker {
		return nil, fmt.Errorf("falha ao normalizar ticker %s", fractionalTicker)
	}

	result := &MergeResult{
		SourceTicker: fractionalTicker,
		TargetTicker: normalizedTicker,
	}

	// Verificar se ativo original existe
	targetAsset, targetExists := w.Assets[normalizedTicker]

	if !targetExists {
		// Ativo original não existe - retornar erro especial
		return nil, fmt.Errorf("TARGET_NOT_FOUND:%s", normalizedTicker)
	}

	// Guardar quantidade antes do merge
	result.TargetQuantityBefore = targetAsset.Quantity

	// Mover transações
	for i := range sourceAsset.Negotiations {
		tx := sourceAsset.Negotiations[i]
		oldHash := tx.Hash

		// Atualizar ticker da transação
		tx.Ticker = normalizedTicker
		// Recalcular hash com novo ticker
		tx.Hash = parser.CalculateHash(&tx)

		// Adicionar ao ativo de destino
		targetAsset.Negotiations = append(targetAsset.Negotiations, tx)

		// Atualizar também na lista global de transações
		for j := range w.Transactions {
			if w.Transactions[j].Hash == oldHash {
				w.Transactions[j] = tx
				// Atualizar no mapa de hash também
				delete(w.TransactionsByHash, oldHash)
				w.TransactionsByHash[tx.Hash] = tx
				break
			}
		}

		result.TransactionsMoved++
	}

	// Mover proventos (earnings ficam em cada asset, não há lista global)
	for i := range sourceAsset.Earnings {
		earning := sourceAsset.Earnings[i]

		// Atualizar ticker do provento
		earning.Ticker = normalizedTicker
		// Recalcular hash com novo ticker
		earning.Hash = parser.CalculateEarningHash(&earning)

		// Adicionar ao ativo de destino
		targetAsset.Earnings = append(targetAsset.Earnings, earning)

		result.EarningsMoved++
	}

	// Remover ativo fracionário ANTES de recalcular
	// para evitar que ele seja recalculado com as transações antigas
	delete(w.Assets, fractionalTicker)

	// Recalcular todos os ativos
	w.RecalculateAssets()

	// Obter quantidade atualizada após recálculo
	result.TargetQuantityAfter = w.Assets[normalizedTicker].Quantity

	return result, nil
}

// CreateAndMergeFractionalAsset cria o ativo original e mescla o fracionário
// Usado quando o usuário confirma que quer criar o ativo original
func (w *Wallet) CreateAndMergeFractionalAsset(fractionalTicker string) (*MergeResult, error) {
	// Verificar se ticker termina com F
	if len(fractionalTicker) == 0 || fractionalTicker[len(fractionalTicker)-1] != 'F' {
		return nil, fmt.Errorf("ticker %s não é um ativo fracionário", fractionalTicker)
	}

	// Verificar se ativo fracionário existe
	sourceAsset, exists := w.Assets[fractionalTicker]
	if !exists {
		return nil, fmt.Errorf("ativo fracionário %s não encontrado", fractionalTicker)
	}

	// Obter ticker normalizado
	normalizedTicker := parser.NormalizeTicker(fractionalTicker)

	// Criar ativo de destino com metadados do source
	targetAsset := &Asset{
		ID:           normalizedTicker,
		Negotiations: make([]parser.Transaction, 0),
		Earnings:     make([]parser.Earning, 0),
		Type:         sourceAsset.Type,
		SubType:      sourceAsset.SubType,
		Segment:      sourceAsset.Segment,
	}

	w.Assets[normalizedTicker] = targetAsset

	result := &MergeResult{
		SourceTicker:         fractionalTicker,
		TargetTicker:         normalizedTicker,
		TargetCreated:        true,
		TargetQuantityBefore: 0,
	}

	// Mover transações
	for i := range sourceAsset.Negotiations {
		tx := sourceAsset.Negotiations[i]
		oldHash := tx.Hash

		// Atualizar ticker da transação
		tx.Ticker = normalizedTicker
		// Recalcular hash com novo ticker
		tx.Hash = parser.CalculateHash(&tx)

		// Adicionar ao ativo de destino
		targetAsset.Negotiations = append(targetAsset.Negotiations, tx)

		// Atualizar também na lista global de transações
		for j := range w.Transactions {
			if w.Transactions[j].Hash == oldHash {
				w.Transactions[j] = tx
				// Atualizar no mapa de hash também
				delete(w.TransactionsByHash, oldHash)
				w.TransactionsByHash[tx.Hash] = tx
				break
			}
		}

		result.TransactionsMoved++
	}

	// Mover proventos (earnings ficam em cada asset, não há lista global)
	for i := range sourceAsset.Earnings {
		earning := sourceAsset.Earnings[i]

		// Atualizar ticker do provento
		earning.Ticker = normalizedTicker
		// Recalcular hash com novo ticker
		earning.Hash = parser.CalculateEarningHash(&earning)

		// Adicionar ao ativo de destino
		targetAsset.Earnings = append(targetAsset.Earnings, earning)

		result.EarningsMoved++
	}

	// Remover ativo fracionário ANTES de recalcular
	// para evitar que ele seja recalculado com as transações antigas
	delete(w.Assets, fractionalTicker)

	// Recalcular todos os ativos
	w.RecalculateAssets()

	// Obter quantidade atualizada após recálculo
	result.TargetQuantityAfter = w.Assets[normalizedTicker].Quantity

	return result, nil
}

// IsFractionalTicker verifica se um ticker é fracionário (termina com F)
func IsFractionalTicker(ticker string) bool {
	return len(ticker) > 0 && ticker[len(ticker)-1] == 'F'
}
