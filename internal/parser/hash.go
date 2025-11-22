package parser

import (
	"crypto/sha256"
	"fmt"
)

// generateHash gera um hash SHA256 único para uma transação
// O hash é baseado em todos os campos da transação para garantir unicidade
func generateHash(t *Transaction) string {
	data := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%s",
		t.Date.Format("2006-01-02"),
		t.Type,
		t.Institution,
		t.Ticker,
		t.Quantity.StringFixed(8),
		t.Price.StringFixed(8),
		t.Amount.StringFixed(8),
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// CalculateHash é uma função pública que calcula o hash de uma transação
// Útil para recalcular hashes quando campos de uma transação são modificados
func CalculateHash(t *Transaction) string {
	return generateHash(t)
}
