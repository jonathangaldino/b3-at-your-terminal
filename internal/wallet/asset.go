package wallet

import (
	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
)

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
	AveragePrice decimal.Decimal

	// TotalInvestedValue é o valor total investido neste ativo
	// Calculado automaticamente baseado em todas as transações de compra
	TotalInvestedValue decimal.Decimal

	// Quantity é a quantidade atual de papéis que o investidor possui
	// Calculado como: (total de compras) - (total de vendas)
	// Arredondado para número inteiro
	Quantity int

	// IsSubscription indica se este ativo é um direito de subscrição
	IsSubscription bool

	// SubscriptionOf é o ticker do ativo ao qual este direito de subscrição pertence
	// Só é preenchido se IsSubscription for true
	SubscriptionOf string
}
