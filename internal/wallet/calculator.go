package wallet

import "github.com/shopspring/decimal"

// calculateAveragePrice calcula o preço médio ponderado de um ativo
// baseado em todas as transações de compra
//
// Fórmula: Preço Médio = Σ(preço × quantidade) / Σ(quantidade)
//
// Apenas transações do tipo "Compra" são consideradas no cálculo
func calculateAveragePrice(asset *Asset) decimal.Decimal {
	totalCost := decimal.Zero
	totalQuantity := decimal.Zero

	for _, negotiation := range asset.Negotiations {
		// Considerar apenas compras para o cálculo do preço médio
		if negotiation.Type == "Compra" {
			totalCost = totalCost.Add(negotiation.Amount)
			totalQuantity = totalQuantity.Add(negotiation.Quantity)
		}
	}

	// Evitar divisão por zero
	if totalQuantity.IsZero() {
		return decimal.Zero
	}

	// Dividir e arredondar para 4 casas decimais
	return totalCost.Div(totalQuantity).Round(4)
}

// calculateTotalInvestedValue calcula o valor total investido em um ativo
// Soma apenas os valores das transações de compra
func calculateTotalInvestedValue(asset *Asset) decimal.Decimal {
	total := decimal.Zero

	for _, negotiation := range asset.Negotiations {
		if negotiation.Type == "Compra" {
			total = total.Add(negotiation.Amount)
		}
	}

	return total.Round(4)
}

// calculateQuantity calcula a quantidade atual de papéis do ativo
// Fórmula: (Σ compras) - (Σ vendas)
// Retorna um inteiro (arredondado)
func calculateQuantity(asset *Asset) int {
	quantity := decimal.Zero

	for _, negotiation := range asset.Negotiations {
		if negotiation.Type == "Compra" {
			quantity = quantity.Add(negotiation.Quantity)
		} else if negotiation.Type == "Venda" {
			quantity = quantity.Sub(negotiation.Quantity)
		}
	}

	// Converter para int (arredondando)
	return int(quantity.Round(0).IntPart())
}
