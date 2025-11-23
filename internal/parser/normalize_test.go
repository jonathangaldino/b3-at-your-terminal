package parser

import "testing"

func TestNormalizeTicker(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ticker fracionário com F",
			input:    "KLBN3F",
			expected: "KLBN3",
		},
		{
			name:     "ticker fracionário com F minúsculo",
			input:    "klbn3f",
			expected: "KLBN3",
		},
		{
			name:     "ticker normal sem F",
			input:    "ITSA4",
			expected: "ITSA4",
		},
		{
			name:     "ticker normal minúsculo",
			input:    "petr4",
			expected: "PETR4",
		},
		{
			name:     "ticker com espaços",
			input:    " VALE3F ",
			expected: "VALE3",
		},
		{
			name:     "ticker vazio",
			input:    "",
			expected: "",
		},
		{
			name:     "ticker apenas F",
			input:    "F",
			expected: "",
		},
		{
			name:     "ticker que termina naturalmente com F (não fracionário)",
			input:    "MXRF11F",
			expected: "MXRF11",
		},
		{
			name:     "FII fracionário",
			input:    "MXRF11F",
			expected: "MXRF11",
		},
		{
			name:     "ticker misto maiúscula/minúscula com F",
			input:    "BbAs3F",
			expected: "BBAS3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeTicker(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeTicker(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
