package wallet

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestDecimalPrecisionVsFloat64 demonstra as diferenças entre decimal e float64
func TestDecimalPrecisionVsFloat64(t *testing.T) {
	t.Run("Problema clássico: 0.1 + 0.2 = 0.3", func(t *testing.T) {
		// Com float64 (ERRADO)
		var f1 float64 = 0.1
		var f2 float64 = 0.2
		floatSum := f1 + f2
		// floatSum será 0.30000000000000004 ❌

		// Com decimal.Decimal (CORRETO)
		d1 := decimal.NewFromFloat(0.1)
		d2 := decimal.NewFromFloat(0.2)
		decimalSum := d1.Add(d2)

		// Verificar que decimal é exato
		expected := decimal.NewFromFloat(0.3)
		if !decimalSum.Equal(expected) {
			t.Errorf("Decimal: 0.1 + 0.2 = %v, expected %v", decimalSum, expected)
		}

		// Verificar que float64 tem erro
		if floatSum == 0.3 {
			t.Error("Float64 não deveria ser exatamente 0.3 (mas pode passar dependendo da implementação)")
		}

		t.Logf("Float64: 0.1 + 0.2 = %.20f (impreciso)", floatSum)
		t.Logf("Decimal: 0.1 + 0.2 = %s (preciso)", decimalSum.String())
	})

	t.Run("Multiplicação de valores monetários", func(t *testing.T) {
		// Cenário: 3 ações a R$ 10.33 cada
		// Com float64
		var floatPrice float64 = 10.33
		var floatQty float64 = 3
		floatTotal := floatPrice * floatQty
		// floatTotal = 30.990000000000002 ❌

		// Com decimal
		decimalPrice := decimal.NewFromFloat(10.33)
		decimalQty := decimal.NewFromInt(3)
		decimalTotal := decimalPrice.Mul(decimalQty)

		expected := decimal.NewFromFloat(30.99)
		if !decimalTotal.Equal(expected) {
			t.Errorf("Decimal: 10.33 × 3 = %v, expected %v", decimalTotal, expected)
		}

		t.Logf("Float64: 10.33 × 3 = %.20f", floatTotal)
		t.Logf("Decimal: 10.33 × 3 = %s", decimalTotal.String())
	})

	t.Run("Divisão e arredondamento", func(t *testing.T) {
		// Cenário: R$ 100 dividido por 3
		// Com float64
		var floatAmount float64 = 100.0
		var floatDivisor float64 = 3.0
		floatResult := floatAmount / floatDivisor
		// floatResult = 33.333333333333336 (muitas casas decimais)

		// Com decimal (arredondado para 4 casas)
		decimalAmount := decimal.NewFromInt(100)
		decimalDivisor := decimal.NewFromInt(3)
		decimalResult := decimalAmount.Div(decimalDivisor).Round(4)

		expected := "33.3333"
		if decimalResult.StringFixed(4) != expected {
			t.Errorf("Decimal: 100 ÷ 3 = %v, expected %v", decimalResult.StringFixed(4), expected)
		}

		t.Logf("Float64: 100 ÷ 3 = %.20f", floatResult)
		t.Logf("Decimal: 100 ÷ 3 = %s (4 casas)", decimalResult.StringFixed(4))
	})
}

// TestRealWorldFinancialScenarios testa cenários reais do mercado financeiro
func TestRealWorldFinancialScenarios(t *testing.T) {
	t.Run("Cálculo de preço médio com valores reais da B3", func(t *testing.T) {
		// Compra 1: 100 ações a R$ 28.47
		qty1 := decimal.NewFromInt(100)
		price1 := decimal.NewFromFloat(28.47)
		amount1 := qty1.Mul(price1)

		// Compra 2: 50 ações a R$ 29.13
		qty2 := decimal.NewFromInt(50)
		price2 := decimal.NewFromFloat(29.13)
		amount2 := qty2.Mul(price2)

		// Preço médio ponderado
		totalQty := qty1.Add(qty2)
		totalAmount := amount1.Add(amount2)
		avgPrice := totalAmount.Div(totalQty).Round(4)

		// Cálculo manual: (2847 + 1456.5) / 150 = 4303.5 / 150 = 28.69
		expected := "28.6900"
		if avgPrice.StringFixed(4) != expected {
			t.Errorf("Preço médio = %v, expected %v", avgPrice.StringFixed(4), expected)
		}
	})

	t.Run("Precisão com 4 casas decimais (padrão B3)", func(t *testing.T) {
		// Valores com 4 casas decimais são comuns na B3
		price, _ := decimal.NewFromString("15.2374")
		qty := decimal.NewFromInt(100)
		total := price.Mul(qty).Round(2) // Total em R$ com 2 casas

		expected := "1523.74"
		if total.StringFixed(2) != expected {
			t.Errorf("Total = %v, expected %v", total.StringFixed(2), expected)
		}
	})

	t.Run("Soma de pequenos valores sem perda de precisão", func(t *testing.T) {
		// Cenário: múltiplas operações pequenas
		values := []string{"0.01", "0.02", "0.03", "0.04", "0.05"}
		sum := decimal.Zero

		for _, v := range values {
			val, _ := decimal.NewFromString(v)
			sum = sum.Add(val)
		}

		// Com float64, poderia haver erro de arredondamento
		// Com decimal, sempre exato
		expected := "0.15"
		if sum.StringFixed(2) != expected {
			t.Errorf("Sum = %v, expected %v", sum.StringFixed(2), expected)
		}
	})
}

// TestDecimalComparison testa comparações exatas com decimal
func TestDecimalComparison(t *testing.T) {
	t.Run("Comparação exata de valores", func(t *testing.T) {
		val1 := decimal.NewFromFloat(10.50)
		val2 := decimal.NewFromFloat(10.50)
		val3 := decimal.NewFromFloat(10.51)

		if !val1.Equal(val2) {
			t.Error("10.50 deveria ser igual a 10.50")
		}

		if val1.Equal(val3) {
			t.Error("10.50 não deveria ser igual a 10.51")
		}

		if !val1.LessThan(val3) {
			t.Error("10.50 deveria ser menor que 10.51")
		}

		if !val3.GreaterThan(val1) {
			t.Error("10.51 deveria ser maior que 10.50")
		}
	})

	t.Run("Comparação com zero", func(t *testing.T) {
		zero := decimal.Zero
		positive := decimal.NewFromInt(1)
		negative := decimal.NewFromInt(-1)

		if !zero.IsZero() {
			t.Error("Zero deveria ser zero")
		}

		if positive.IsZero() {
			t.Error("1 não deveria ser zero")
		}

		if !positive.IsPositive() {
			t.Error("1 deveria ser positivo")
		}

		if !negative.IsNegative() {
			t.Error("-1 deveria ser negativo")
		}
	})
}

// TestDecimalRounding testa diferentes estratégias de arredondamento
func TestDecimalRounding(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		decimals int32
		expected string
	}{
		{
			name:     "Arredondar para 2 casas (para cima)",
			value:    "10.555",
			decimals: 2,
			expected: "10.56",
		},
		{
			name:     "Arredondar para 2 casas (para baixo)",
			value:    "10.554",
			decimals: 2,
			expected: "10.55",
		},
		{
			name:     "Arredondar para 4 casas",
			value:    "15.23749",
			decimals: 4,
			expected: "15.2375",
		},
		{
			name:     "Arredondar para 0 casas",
			value:    "99.5",
			decimals: 0,
			expected: "100",
		},
		{
			name:     "Já tem menos casas que o solicitado",
			value:    "10.5",
			decimals: 4,
			expected: "10.5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, _ := decimal.NewFromString(tt.value)
			rounded := val.Round(tt.decimals)
			result := rounded.StringFixed(tt.decimals)

			if result != tt.expected {
				t.Errorf("Round(%s, %d) = %v, expected %v", tt.value, tt.decimals, result, tt.expected)
			}
		})
	}
}

// TestDecimalStringConversion testa conversões entre decimal e string
func TestDecimalStringConversion(t *testing.T) {
	t.Run("Parsing de strings com vírgula (formato BR)", func(t *testing.T) {
		// Nota: decimal.NewFromString não aceita vírgula nativamente
		// Mas nosso parseFloat converte vírgula para ponto primeiro
		val, err := decimal.NewFromString("10.99") // Já convertido
		if err != nil {
			t.Fatalf("Erro ao parsear: %v", err)
		}

		if val.StringFixed(2) != "10.99" {
			t.Errorf("Valor parseado = %v, expected 10.99", val.StringFixed(2))
		}
	})

	t.Run("Conversão para string com formato fixo", func(t *testing.T) {
		val := decimal.NewFromFloat(10.5)

		// StringFixed adiciona zeros à direita
		if val.StringFixed(2) != "10.50" {
			t.Errorf("StringFixed(2) = %v, expected 10.50", val.StringFixed(2))
		}

		if val.StringFixed(4) != "10.5000" {
			t.Errorf("StringFixed(4) = %v, expected 10.5000", val.StringFixed(4))
		}

		// String() não adiciona zeros
		if val.String() != "10.5" {
			t.Errorf("String() = %v, expected 10.5", val.String())
		}
	})
}

// BenchmarkDecimalVsFloat64 compara performance entre decimal e float64
func BenchmarkDecimalVsFloat64(b *testing.B) {
	b.Run("Float64 Addition", func(b *testing.B) {
		var result float64
		for i := 0; i < b.N; i++ {
			result = 10.50 + 20.75
		}
		_ = result
	})

	b.Run("Decimal Addition", func(b *testing.B) {
		d1 := decimal.NewFromFloat(10.50)
		d2 := decimal.NewFromFloat(20.75)
		var result decimal.Decimal
		for i := 0; i < b.N; i++ {
			result = d1.Add(d2)
		}
		_ = result
	})

	b.Run("Float64 Multiplication", func(b *testing.B) {
		var result float64
		for i := 0; i < b.N; i++ {
			result = 10.50 * 100
		}
		_ = result
	})

	b.Run("Decimal Multiplication", func(b *testing.B) {
		d1 := decimal.NewFromFloat(10.50)
		d2 := decimal.NewFromInt(100)
		var result decimal.Decimal
		for i := 0; i < b.N; i++ {
			result = d1.Mul(d2)
		}
		_ = result
	})
}
