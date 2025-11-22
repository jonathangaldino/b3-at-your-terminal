package parser

import "time"

// Transaction representa uma transação financeira da B3
type Transaction struct {
	Date        time.Time // Data do Negócio
	Type        string    // Tipo de Movimentação (Compra/Venda)
	Institution string    // Instituição
	Ticker      string    // Código de Negociação
	Quantity    float64   // Quantidade
	Price       float64   // Preço unitário
	Amount      float64   // Valor total
	Hash        string    // Hash SHA256 único
}
