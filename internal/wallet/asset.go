package wallet

import "github.com/john/b3-project/internal/parser"

// Asset representa um ativo financeiro (ação ou fundo imobiliário)
type Asset struct {
	// ID é o código da negociação (ticker) do ativo
	ID string

	// Negotiations são todas as negociações (compra/venda) feitas com esse ativo
	Negotiations []parser.Transaction

	// Type representa o tipo de ativo - sempre será "renda variável"
	Type string

	// SubType define se é "ações" ou "fundos imobiliários"
	// Campo definido manualmente pelo usuário
	SubType string

	// Segment significa o segmento que essa empresa atua
	// Campo para categorização livre pelo usuário
	Segment string

	// AveragePrice é o preço médio ponderado pago pelo ativo
	// Calculado automaticamente baseado nas transações de compra
	AveragePrice float64
}
