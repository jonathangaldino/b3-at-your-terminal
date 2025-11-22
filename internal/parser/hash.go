package parser

import (
	"crypto/sha256"
	"fmt"
)

// generateHash gera um hash SHA256 único para uma transação
// O hash é baseado em todos os campos da transação para garantir unicidade
func generateHash(t *Transaction) string {
	data := fmt.Sprintf(
		"%s|%s|%s|%s|%.8f|%.8f|%.8f",
		t.Date.Format("2006-01-02"),
		t.Type,
		t.Institution,
		t.Ticker,
		t.Quantity,
		t.Price,
		t.Amount,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
