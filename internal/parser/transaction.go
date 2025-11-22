package parser

import (
	"time"

	"github.com/shopspring/decimal"
)

// Transaction representa uma transação financeira da B3
type Transaction struct {
	Date        time.Time       // Data do Negócio
	Type        string          // Tipo de Movimentação (Compra/Venda)
	Institution string          // Instituição
	Ticker      string          // Código de Negociação
	Quantity    decimal.Decimal // Quantidade
	Price       decimal.Decimal // Preço unitário
	Amount      decimal.Decimal // Valor total
	Hash        string          // Hash SHA256 único
}
