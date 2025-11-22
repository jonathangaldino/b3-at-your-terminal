package wallet

import (
	"testing"
	"time"

	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
)

// TestCalculateAveragePrice testa o cálculo do preço médio ponderado
func TestCalculateAveragePrice(t *testing.T) {
	tests := []struct {
		name     string
		asset    *Asset
		expected string // Usar string para comparação exata
	}{
		{
			name: "Uma única compra",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(100),
						Price:    decimal.NewFromFloat(10.50),
						Amount:   decimal.NewFromFloat(1050.00),
					},
				},
			},
			expected: "10.5000",
		},
		{
			name: "Múltiplas compras - preço médio ponderado",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(100),
						Price:    decimal.NewFromFloat(10.00),
						Amount:   decimal.NewFromFloat(1000.00),
					},
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(200),
						Price:    decimal.NewFromFloat(15.00),
						Amount:   decimal.NewFromFloat(3000.00),
					},
				},
			},
			// (1000 + 3000) / (100 + 200) = 4000 / 300 = 13.3333...
			expected: "13.3333",
		},
		{
			name: "Compras e vendas (vendas não devem afetar o preço médio)",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(100),
						Price:    decimal.NewFromFloat(10.00),
						Amount:   decimal.NewFromFloat(1000.00),
					},
					{
						Type:     "Venda",
						Quantity: decimal.NewFromInt(50),
						Price:    decimal.NewFromFloat(20.00),
						Amount:   decimal.NewFromFloat(1000.00),
					},
				},
			},
			// Apenas a compra conta: 1000 / 100 = 10.00
			expected: "10.0000",
		},
		{
			name: "Sem transações",
			asset: &Asset{
				Negotiations: []parser.Transaction{},
			},
			expected: "0.0000",
		},
		{
			name: "Apenas vendas (não deve calcular preço médio)",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Venda",
						Quantity: decimal.NewFromInt(100),
						Price:    decimal.NewFromFloat(10.00),
						Amount:   decimal.NewFromFloat(1000.00),
					},
				},
			},
			expected: "0.0000",
		},
		{
			name: "Teste de precisão decimal - evitar erros de float",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromFloat(1.5),
						Price:    decimal.NewFromFloat(10.33),
						Amount:   decimal.NewFromFloat(15.495),
					},
					{
						Type:     "Compra",
						Quantity: decimal.NewFromFloat(2.7),
						Price:    decimal.NewFromFloat(12.47),
						Amount:   decimal.NewFromFloat(33.669),
					},
				},
			},
			// (15.495 + 33.669) / (1.5 + 2.7) = 49.164 / 4.2 = 11.7057...
			expected: "11.7057",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAveragePrice(tt.asset)
			resultStr := result.StringFixed(4)
			if resultStr != tt.expected {
				t.Errorf("calculateAveragePrice() = %v, expected %v", resultStr, tt.expected)
			}
		})
	}
}

// TestCalculateTotalInvestedValue testa o cálculo do valor total investido
func TestCalculateTotalInvestedValue(t *testing.T) {
	tests := []struct {
		name     string
		asset    *Asset
		expected string
	}{
		{
			name: "Uma única compra",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:   "Compra",
						Amount: decimal.NewFromFloat(1000.00),
					},
				},
			},
			expected: "1000.0000",
		},
		{
			name: "Múltiplas compras",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:   "Compra",
						Amount: decimal.NewFromFloat(1000.00),
					},
					{
						Type:   "Compra",
						Amount: decimal.NewFromFloat(2000.00),
					},
					{
						Type:   "Compra",
						Amount: decimal.NewFromFloat(500.50),
					},
				},
			},
			expected: "3500.5000",
		},
		{
			name: "Compras e vendas (vendas não contam)",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:   "Compra",
						Amount: decimal.NewFromFloat(1000.00),
					},
					{
						Type:   "Venda",
						Amount: decimal.NewFromFloat(500.00),
					},
				},
			},
			expected: "1000.0000",
		},
		{
			name: "Sem transações",
			asset: &Asset{
				Negotiations: []parser.Transaction{},
			},
			expected: "0.0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTotalInvestedValue(tt.asset)
			resultStr := result.StringFixed(4)
			if resultStr != tt.expected {
				t.Errorf("calculateTotalInvestedValue() = %v, expected %v", resultStr, tt.expected)
			}
		})
	}
}

// TestCalculateQuantity testa o cálculo da quantidade em carteira
func TestCalculateQuantity(t *testing.T) {
	tests := []struct {
		name     string
		asset    *Asset
		expected string
	}{
		{
			name: "Uma única compra",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(100),
					},
				},
			},
			expected: "100.0000",
		},
		{
			name: "Múltiplas compras",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(100),
					},
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(50),
					},
				},
			},
			expected: "150.0000",
		},
		{
			name: "Compras e vendas",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(100),
					},
					{
						Type:     "Venda",
						Quantity: decimal.NewFromInt(30),
					},
				},
			},
			expected: "70.0000",
		},
		{
			name: "Vendas maiores que compras (posição zerada ou vendida)",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromInt(100),
					},
					{
						Type:     "Venda",
						Quantity: decimal.NewFromInt(100),
					},
				},
			},
			expected: "0.0000",
		},
		{
			name: "Sem transações",
			asset: &Asset{
				Negotiations: []parser.Transaction{},
			},
			expected: "0.0000",
		},
		{
			name: "Quantidades decimais (fracionário)",
			asset: &Asset{
				Negotiations: []parser.Transaction{
					{
						Type:     "Compra",
						Quantity: decimal.NewFromFloat(10.5),
					},
					{
						Type:     "Compra",
						Quantity: decimal.NewFromFloat(5.25),
					},
					{
						Type:     "Venda",
						Quantity: decimal.NewFromFloat(3.75),
					},
				},
			},
			expected: "12.0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateQuantity(tt.asset)
			resultStr := result.StringFixed(4)
			if resultStr != tt.expected {
				t.Errorf("calculateQuantity() = %v, expected %v", resultStr, tt.expected)
			}
		})
	}
}

// TestCalculatorsIntegration testa a integração entre os calculadores
func TestCalculatorsIntegration(t *testing.T) {
	// Criar um cenário realista
	asset := &Asset{
		ID:   "PETR4",
		Type: "renda variável",
		Negotiations: []parser.Transaction{
			{
				Date:        time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC),
				Type:        "Compra",
				Ticker:      "PETR4",
				Quantity:    decimal.NewFromInt(100),
				Price:       decimal.NewFromFloat(28.50),
				Amount:      decimal.NewFromFloat(2850.00),
			},
			{
				Date:        time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC),
				Type:        "Compra",
				Ticker:      "PETR4",
				Quantity:    decimal.NewFromInt(50),
				Price:       decimal.NewFromFloat(30.00),
				Amount:      decimal.NewFromFloat(1500.00),
			},
			{
				Date:        time.Date(2023, 3, 20, 0, 0, 0, 0, time.UTC),
				Type:        "Venda",
				Ticker:      "PETR4",
				Quantity:    decimal.NewFromInt(30),
				Price:       decimal.NewFromFloat(32.00),
				Amount:      decimal.NewFromFloat(960.00),
			},
		},
	}

	// Calcular todos os campos
	avgPrice := calculateAveragePrice(asset)
	totalInvested := calculateTotalInvestedValue(asset)
	quantity := calculateQuantity(asset)

	// Verificações
	// Preço médio: (2850 + 1500) / (100 + 50) = 4350 / 150 = 29.00
	if avgPrice.StringFixed(4) != "29.0000" {
		t.Errorf("AveragePrice = %v, expected 29.0000", avgPrice.StringFixed(4))
	}

	// Total investido: 2850 + 1500 = 4350 (vendas não contam)
	if totalInvested.StringFixed(4) != "4350.0000" {
		t.Errorf("TotalInvestedValue = %v, expected 4350.0000", totalInvested.StringFixed(4))
	}

	// Quantidade: 100 + 50 - 30 = 120
	if quantity.StringFixed(4) != "120.0000" {
		t.Errorf("Quantity = %v, expected 120.0000", quantity.StringFixed(4))
	}
}
