package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// FileType representa o tipo de arquivo Excel da B3
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeTransactions
	FileTypeEarnings
)

// DetectFileType detecta automaticamente se um arquivo é de transações ou proventos
// baseado no número de colunas da primeira linha de dados
func DetectFileType(filePath string) (FileType, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return FileTypeUnknown, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer f.Close()

	// Obter a primeira sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return FileTypeUnknown, fmt.Errorf("arquivo não contém sheets")
	}
	sheetName := sheets[0]

	// Obter todas as linhas
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return FileTypeUnknown, fmt.Errorf("erro ao ler linhas: %w", err)
	}

	if len(rows) <= 1 {
		return FileTypeUnknown, fmt.Errorf("arquivo não contém dados (apenas cabeçalho ou vazio)")
	}

	// Verificar número de colunas da primeira linha de dados (linha 2)
	firstDataRow := rows[1]
	numCols := len(firstDataRow)

	// 9 colunas = transações
	// 8 colunas = proventos
	if numCols >= 9 {
		return FileTypeTransactions, nil
	} else if numCols >= 8 {
		return FileTypeEarnings, nil
	}

	return FileTypeUnknown, fmt.Errorf("número de colunas não reconhecido: %d (esperado 8 ou 9)", numCols)
}

// ParseFiles processa múltiplos arquivos .xlsx e retorna todas as transações encontradas
// Automaticamente deduplica transações usando hash SHA256
func ParseFiles(filePaths []string) ([]Transaction, error) {
	var allTransactions []Transaction
	seenHashes := make(map[string]bool)

	for _, filePath := range filePaths {
		transactions, err := parseXLSX(filePath)
		if err != nil {
			return nil, fmt.Errorf("erro ao processar arquivo %s: %w", filePath, err)
		}

		// Deduplar usando hash
		for _, t := range transactions {
			if !seenHashes[t.Hash] {
				allTransactions = append(allTransactions, t)
				seenHashes[t.Hash] = true
			}
		}
	}

	return allTransactions, nil
}

// parseXLSX processa um único arquivo .xlsx e retorna as transações
func parseXLSX(filePath string) ([]Transaction, error) {
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

	var transactions []Transaction

	// Iterar sobre linhas (skip primeira linha - cabeçalho)
	for i, row := range rows[1:] {
		lineNum := i + 2 // +2 porque pulamos linha 1 e arrays começam em 0

		// Verificar se a linha tem colunas suficientes
		if len(row) < 9 {
			return nil, fmt.Errorf("linha %d: número insuficiente de colunas (esperado 9, encontrado %d)", lineNum, len(row))
		}

		// Parsear cada campo
		// Coluna A (0): Data do Negócio
		dataNegocio, err := parseDate(row[0])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear data: %w", lineNum, err)
		}

		// Coluna B (1): Tipo de Movimentação
		tipoMovimentacao := row[1]

		// Coluna C (2): Mercado
		mercado := row[2]

		// Coluna D (3): Prazo/Vencimento - IGNORAR

		// Coluna E (4): Instituição
		instituicao := row[4]

		// Coluna F (5): Código da Negociação
		codigoNegociacao := row[5]

		// Normalizar código do mercado fracionário
		// Se mercado é "Mercado Fracionário" e código termina com "F", remover o "F"
		codigoNegociacao = normalizeFractionalCode(mercado, codigoNegociacao)

		// Coluna G (6): Quantidade
		quantidade, err := parseFloat(row[6])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear quantidade: %w", lineNum, err)
		}

		// Coluna H (7): Preço
		preco, err := parseFloat(row[7])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear preço: %w", lineNum, err)
		}

		// Coluna I (8): Valor
		valor, err := parseFloat(row[8])
		if err != nil {
			return nil, fmt.Errorf("linha %d: erro ao parsear valor: %w", lineNum, err)
		}

		// Criar transação
		transaction := Transaction{
			Date:        dataNegocio,
			Type:        tipoMovimentacao,
			Institution: instituicao,
			Ticker:      codigoNegociacao,
			Quantity:    quantidade,
			Price:       preco,
			Amount:      valor,
		}

		// Gerar hash
		transaction.Hash = generateHash(&transaction)

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// parseDate converte string DD/MM/YYYY para time.Time
func parseDate(dateStr string) (time.Time, error) {
	// Formato: DD/MM/YYYY
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("formato de data inválido (esperado DD/MM/YYYY): %w", err)
	}
	return t, nil
}

// parseFloat converte string para decimal.Decimal
// Aceita tanto ponto quanto vírgula como separador decimal
// Mantém precisão de 4 casas decimais (padrão B3)
func parseFloat(str string) (decimal.Decimal, error) {
	// Remover espaços em branco
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

// replaceCommaWithDot substitui vírgula por ponto em strings numéricas
func replaceCommaWithDot(s string) string {
	result := ""
	for _, ch := range s {
		if ch == ',' {
			result += "."
		} else {
			result += string(ch)
		}
	}
	return result
}

// trimSpaces remove espaços em branco do início e fim da string
func trimSpaces(s string) string {
	result := s
	// Remove espaços do início
	for len(result) > 0 && result[0] == ' ' {
		result = result[1:]
	}
	// Remove espaços do fim
	for len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}
	return result
}

// normalizeFractionalCode normaliza códigos do mercado fracionário
// Remove o "F" do final do código quando o mercado é "Mercado Fracionário"
// Exemplo: BOVA11F -> BOVA11 (quando mercado é fracionário)
func normalizeFractionalCode(mercado, codigo string) string {
	// Verificar se é mercado fracionário
	if mercado != "Mercado Fracionário" {
		return codigo
	}

	// Usar normalização genérica
	return NormalizeTicker(codigo)
}

// NormalizeTicker normaliza um ticker removendo o "F" do final se presente
// O "F" indica mercado fracionário, mas representa a mesma empresa
// Exemplo: KLBN3F -> KLBN3, ITSA4F -> ITSA4
// Esta função é pública para ser usada em transações manuais e imports
func NormalizeTicker(ticker string) string {
	// Converter para maiúsculas
	ticker = strings.ToUpper(ticker)

	// Remover espaços
	ticker = trimSpaces(ticker)

	// Verificar se o código termina com "F"
	if len(ticker) > 0 && ticker[len(ticker)-1] == 'F' {
		// Remover o "F" do final
		return ticker[:len(ticker)-1]
	}

	return ticker
}
