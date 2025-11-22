package parser

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestParseFloat testa a conversão de strings para decimal.Decimal
func TestParseFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Usar string para comparação exata
		wantErr  bool
	}{
		{
			name:     "Valor com ponto decimal (formato padrão)",
			input:    "10.99",
			expected: "10.9900",
			wantErr:  false,
		},
		{
			name:     "Valor com vírgula decimal (formato brasileiro)",
			input:    "10,99",
			expected: "10.9900",
			wantErr:  false,
		},
		{
			name:     "Valor inteiro",
			input:    "100",
			expected: "100.0000",
			wantErr:  false,
		},
		{
			name:     "Valor com 4 casas decimais",
			input:    "15.2374",
			expected: "15.2374",
			wantErr:  false,
		},
		{
			name:     "Valor com mais de 4 casas decimais (deve arredondar)",
			input:    "15.237456",
			expected: "15.2375", // Arredonda para 4 casas
			wantErr:  false,
		},
		{
			name:     "Valor com espaços",
			input:    "  10.50  ",
			expected: "10.5000",
			wantErr:  false,
		},
		{
			name:     "Valor zero",
			input:    "0",
			expected: "0.0000",
			wantErr:  false,
		},
		{
			name:     "Valor muito pequeno",
			input:    "0.0001",
			expected: "0.0001",
			wantErr:  false,
		},
		{
			name:     "Valor inválido (letras)",
			input:    "abc",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "String vazia",
			input:    "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFloat(tt.input)

			// Verificar se o erro está correto
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Se não esperamos erro, verificar o valor
			if !tt.wantErr {
				resultStr := result.StringFixed(4)
				if resultStr != tt.expected {
					t.Errorf("parseFloat() = %v, expected %v", resultStr, tt.expected)
				}
			}
		})
	}
}

// TestParseFloatPrecision testa a precisão decimal em operações matemáticas
func TestParseFloatPrecision(t *testing.T) {
	// Testar que 0.1 + 0.2 = 0.3 (problema clássico de float)
	val1, _ := parseFloat("0.1")
	val2, _ := parseFloat("0.2")
	sum := val1.Add(val2)

	expected := decimal.NewFromFloat(0.3).Round(4)
	if !sum.Equal(expected) {
		t.Errorf("0.1 + 0.2 = %s, expected %s", sum.StringFixed(4), expected.StringFixed(4))
	}

	// Verificar que o resultado é exatamente 0.3
	if sum.StringFixed(4) != "0.3000" {
		t.Errorf("0.1 + 0.2 = %s, expected 0.3000", sum.StringFixed(4))
	}
}

// TestReplaceCommaWithDot testa a substituição de vírgula por ponto
func TestReplaceCommaWithDot(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "String com vírgula",
			input:    "10,99",
			expected: "10.99",
		},
		{
			name:     "String com ponto",
			input:    "10.99",
			expected: "10.99",
		},
		{
			name:     "String sem separador",
			input:    "1099",
			expected: "1099",
		},
		{
			name:     "String com múltiplas vírgulas",
			input:    "1,234,567.89",
			expected: "1.234.567.89",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceCommaWithDot(tt.input)
			if result != tt.expected {
				t.Errorf("replaceCommaWithDot() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestTrimSpaces testa a remoção de espaços em branco
func TestTrimSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "String com espaços no início e fim",
			input:    "  hello  ",
			expected: "hello",
		},
		{
			name:     "String sem espaços",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "String com apenas espaços",
			input:    "   ",
			expected: "",
		},
		{
			name:     "String vazia",
			input:    "",
			expected: "",
		},
		{
			name:     "String com espaços internos (devem ser mantidos)",
			input:    "  hello world  ",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("trimSpaces() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestNormalizeFractionalCode testa a normalização de códigos do mercado fracionário
func TestNormalizeFractionalCode(t *testing.T) {
	tests := []struct {
		name     string
		mercado  string
		codigo   string
		expected string
	}{
		{
			name:     "Mercado fracionário com F no final",
			mercado:  "Mercado Fracionário",
			codigo:   "BOVA11F",
			expected: "BOVA11",
		},
		{
			name:     "Mercado fracionário sem F no final",
			mercado:  "Mercado Fracionário",
			codigo:   "BOVA11",
			expected: "BOVA11",
		},
		{
			name:     "Mercado à vista com F no final (não deve remover)",
			mercado:  "Mercado à Vista",
			codigo:   "SOMEF",
			expected: "SOMEF",
		},
		{
			name:     "Mercado à vista sem F no final",
			mercado:  "Mercado à Vista",
			codigo:   "PETR4",
			expected: "PETR4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeFractionalCode(tt.mercado, tt.codigo)
			if result != tt.expected {
				t.Errorf("normalizeFractionalCode() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
