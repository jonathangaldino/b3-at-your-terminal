package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// Earning representa um provento recebido (rendimento, dividendo, JCP)
type Earning struct {
	Date        time.Time       // Data do pagamento
	Type        string          // Tipo: "Rendimento" | "Dividendo" | "Juros Sobre Capital Próprio"
	Ticker      string          // Código do ativo (extraído do campo Produto)
	Quantity    decimal.Decimal // Quantidade contabilizada
	UnitPrice   decimal.Decimal // Valor por papel
	TotalAmount decimal.Decimal // Valor total a receber
	Hash        string          // Hash SHA256 único para deduplicação
}

// ParseEarningsFiles processa múltiplos arquivos .xlsx de proventos e retorna todos os earnings encontrados
// Automaticamente deduplica usando hash SHA256
func ParseEarningsFiles(filePaths []string) ([]Earning, error) {
	var allEarnings []Earning
	seenHashes := make(map[string]bool)

	for _, filePath := range filePaths {
		earnings, err := parseEarningsXLSX(filePath)
		if err != nil {
			return nil, fmt.Errorf("erro ao processar arquivo %s: %w", filePath, err)
		}

		// Deduplar usando hash
		for _, e := range earnings {
			if !seenHashes[e.Hash] {
				allEarnings = append(allEarnings, e)
				seenHashes[e.Hash] = true
			}
		}
	}

	return allEarnings, nil
}

// parseEarningsXLSX processa um único arquivo .xlsx de proventos e retorna os earnings
func parseEarningsXLSX(filePath string) ([]Earning, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer f.Close()

	// Obter a primeira sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("arquivo não contém sheets")
	}
	sheetName := sheets[0]

	// Obter todas as linhas
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler linhas: %w", err)
	}

	if len(rows) <= 1 {
		return nil, fmt.Errorf("arquivo não contém dados (apenas cabeçalho ou vazio)")
	}

	var earnings []Earning

	// Iterar sobre linhas (skip primeira linha - cabeçalho)
	for i, row := range rows[1:] {
		lineNum := i + 2 // +2 porque pulamos linha 1 e arrays começam em 0

		// Verificar se a linha tem colunas suficientes
		// Colunas esperadas: Entrada/Saída, Data, Movimentação, Produto, Instituição, Quantidade, Preço unitário, Valor da operação
		if len(row) < 8 {
			return nil, fmt.Errorf("linha %d: número insuficiente de colunas (esperado 8, encontrado %d)", lineNum, len(row))
		}

		// Coluna A (0): Entrada/Saída - IGNORAR (sempre crédito para proventos)

		// Coluna B (1): Data
		date, err := parseDate(row[1])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear data: %w", lineNum, err)
		}

		// Coluna C (2): Movimentação (tipo de provento)
		earningType := normalizeEarningType(row[2])

		// Coluna D (3): Produto (formato: "TICKER - Nome da empresa")
		produto := row[3]
		ticker, err := extractTicker(produto)
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao extrair ticker do produto '%s': %w", lineNum, produto, err)
		}

		// Coluna E (4): Instituição - IGNORAR

		// Coluna F (5): Quantidade
		quantity, err := parseFloat(row[5])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear quantidade: %w", lineNum, err)
		}

		// Coluna G (6): Preço unitário (formato: "R$ 0,50")
		unitPrice, err := parseFloatWithCurrency(row[6])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear preço unitário: %w", lineNum, err)
		}

		// Coluna H (7): Valor da operação (total) (formato: "R$ 1,50")
		totalAmount, err := parseFloatWithCurrency(row[7])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear valor total: %w", lineNum, err)
		}

		// Criar earning
		earning := Earning{
			Date:        date,
			Type:        earningType,
			Ticker:      ticker,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
			TotalAmount: totalAmount,
		}

		// Gerar hash
		earning.Hash = generateEarningHash(&earning)

		earnings = append(earnings, earning)
	}

	return earnings, nil
}

// extractTicker extrai o ticker do campo Produto
// Formato esperado: "TICKER - Nome da empresa"
// Exemplo: "BCFF11 - FDO INV IMOB - FII BTG PACTUAL FUNDO DE FUNDOS" -> "BCFF11"
func extractTicker(produto string) (string, error) {
	// Remover espaços em branco no início e fim
	produto = trimSpaces(produto)

	// Encontrar a primeira ocorrência de " - "
	parts := strings.SplitN(produto, " - ", 2)
	if len(parts) < 1 {
		return "", fmt.Errorf("formato inválido do campo Produto (esperado 'TICKER - Nome')")
	}

	ticker := trimSpaces(parts[0])
	if ticker == "" {
		return "", fmt.Errorf("ticker vazio no campo Produto")
	}

	return ticker, nil
}

// generateEarningHash gera um hash SHA256 único para um earning
// O hash é baseado nos campos principais para garantir unicidade
func generateEarningHash(e *Earning) string {
	data := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s",
		e.Date.Format("2006-01-02"),
		e.Type,
		e.Ticker,
		e.Quantity.StringFixed(8),
		e.UnitPrice.StringFixed(8),
		e.TotalAmount.StringFixed(8),
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// CalculateEarningHash é uma função pública que calcula o hash de um earning
// Útil para recalcular hashes quando campos de um earning são modificados
func CalculateEarningHash(e *Earning) string {
	return generateEarningHash(e)
}

// parseFloatWithCurrency converte string com formato monetário brasileiro para decimal.Decimal
// Remove "R$" e espaços, depois substitui vírgula por ponto
// Formato aceito: "R$ 1,50" ou "R$ 0,50" ou "1,50"
func parseFloatWithCurrency(str string) (decimal.Decimal, error) {
	// Remover espaços em branco
	str = trimSpaces(str)

	// Remover "R$" se presente
	str = removeCurrencySymbol(str)

	// Remover espaços novamente após remover R$
	str = trimSpaces(str)

	// Substituir vírgula por ponto (formato brasileiro → formato padrão)
	str = replaceCommaWithDot(str)

	val, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.Zero, fmt.Errorf("valor numérico inválido: %w", err)
	}

	// Arredondar para 4 casas decimais (precisão padrão B3)
	return val.Round(4), nil
}

// removeCurrencySymbol remove símbolos de moeda (R$, US$, etc) de uma string
func removeCurrencySymbol(s string) string {
	result := ""
	for _, ch := range s {
		// Manter apenas dígitos, vírgula, ponto e sinal de menos
		if (ch >= '0' && ch <= '9') || ch == ',' || ch == '.' || ch == '-' {
			result += string(ch)
		}
	}
	return result
}

// normalizeEarningType normaliza o tipo de provento para um dos três valores padrões
// Aceita variações comuns de texto encontradas nos arquivos da B3
func normalizeEarningType(rawType string) string {
	// Remover espaços extras e converter para lowercase para comparação
	normalized := trimSpaces(rawType)
	lower := toLower(normalized)

	// Detectar e normalizar baseado em palavras-chave
	if contains(lower, "rendimento") {
		return "Rendimento"
	}
	if contains(lower, "dividendo") {
		return "Dividendo"
	}
	if contains(lower, "juros") && contains(lower, "capital") {
		return "Juros Sobre Capital Próprio"
	}
	if lower == "jcp" {
		return "Juros Sobre Capital Próprio"
	}
	if contains(lower, "resgate") {
		return "Resgate"
	}

	// Se não reconhecer, retornar o valor original (para que a validação pegue)
	return normalized
}

// toLower converte uma string para minúsculas (implementação simples)
func toLower(s string) string {
	result := ""
	for _, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			result += string(ch + 32)
		} else {
			result += string(ch)
		}
	}
	return result
}

// contains verifica se uma string contém uma substring (case-insensitive já aplicado)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
