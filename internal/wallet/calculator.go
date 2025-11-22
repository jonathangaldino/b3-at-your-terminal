package wallet

// calculateAveragePrice calcula o preço médio ponderado de um ativo
// baseado em todas as transações de compra
//
// Fórmula: Preço Médio = Σ(preço × quantidade) / Σ(quantidade)
//
// Apenas transações do tipo "Compra" são consideradas no cálculo
func calculateAveragePrice(asset *Asset) float64 {
	var totalCost float64
	var totalQuantity float64

	for _, negotiation := range asset.Negotiations {
		// Considerar apenas compras para o cálculo do preço médio
		if negotiation.Type == "Compra" {
			totalCost += negotiation.Amount
			totalQuantity += negotiation.Quantity
		}
	}

	// Evitar divisão por zero
	if totalQuantity == 0 {
		return 0
	}

	return totalCost / totalQuantity
}
