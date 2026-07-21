package utils

import "testing"

func TestStripTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML paragraph and bold tags",
			input:    "<p>Descuento del <b>20%</b> en restaurantes</p>",
			expected: "Descuento del  20%  en restaurantes",
		},
		{
			name:     "HTML entities inside tags",
			input:    "<div>V&aacute;lido hasta el 31 de diciembre</div>",
			expected: "Válido hasta el 31 de diciembre",
		},
		{
			name:     "Asterisks and extra space cleanup",
			input:    "*** <p>Promoci&oacute;n especial*</p> ***",
			expected: "Promoción especial",
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripTags(tt.input)
			if got != tt.expected {
				t.Errorf("StripTags(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDecodeHtml(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Encoded spanish accents",
			input:    "Promoci&oacute;n v&aacute;lida en El Salvador",
			expected: "Promoción válida en El Salvador",
		},
		{
			name:     "Plain text without entities",
			input:    "Sin entidades HTML",
			expected: "Sin entidades HTML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeHtml(tt.input)
			if got != tt.expected {
				t.Errorf("DecodeHtml(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Trim leading/trailing asterisks and spaces",
			input:    " **  * Descuento exclusivo * ** ",
			expected: "Descuento exclusivo",
		},
		{
			name:     "No change for clean string",
			input:    "Texto limpio",
			expected: "Texto limpio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanText(tt.input)
			if got != tt.expected {
				t.Errorf("CleanText(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
