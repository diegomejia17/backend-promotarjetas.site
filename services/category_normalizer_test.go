package services

import (
	"promotarjetas-backend/models"
	"testing"
)

func TestGetUnifiedCategory(t *testing.T) {
	tests := []struct {
		name     string
		promo    models.PromocionUnificada
		expected string
	}{
		{
			name: "Merchant override Papa John's",
			promo: models.PromocionUnificada{
				NombreComercio: "Papa John's",
				Titulo:         "50% Descuento",
			},
			expected: "Restaurantes",
		},
		{
			name: "Merchant override Super Selectos",
			promo: models.PromocionUnificada{
				NombreComercio: "Super Selectos",
				Titulo:         "Compras de la semana",
			},
			expected: "Supermercados",
		},
		{
			name: "Keyword match in Title: Clinica dental",
			promo: models.PromocionUnificada{
				NombreComercio: "Comercio Desconocido",
				Titulo:         "Consulta gratis en Clinica Dental",
			},
			expected: "Salud",
		},
		{
			name: "Keyword match in Description: Hotel y resort",
			promo: models.PromocionUnificada{
				NombreComercio:   "Desconocido",
				Titulo:           "Descuento exclusivo",
				DescripcionBreve: "Resort en la playa todo incluido",
			},
			expected: "Viajes",
		},
		{
			name: "Fallback to Otros",
			promo: models.PromocionUnificada{
				NombreComercio:   "Comercio ABC",
				Titulo:           "Promocion especial",
				DescripcionBreve: "Aplica restricciones",
			},
			expected: "Otros",
		},
		{
			name: "Generic BAC merchant with PriceSmart title",
			promo: models.PromocionUnificada{
				NombreComercio:   "BAC",
				Titulo:           "Compra hoy en PriceSmart y paga en cuotas",
				DescripcionBreve: "Descuento en tu compra",
			},
			expected: "Supermercados",
		},
		{
			name: "Generic Agricola merchant with Súper Selectos description",
			promo: models.PromocionUnificada{
				NombreComercio:   "Actualiza tus datos",
				Titulo:           "Actualiza tus datos",
				DescripcionBreve: "Gana certificado de regalo de Súper Selectos",
			},
			expected: "Supermercados",
		},
		{
			name: "Generic Cuscatlan merchant with Copa Airlines title",
			promo: models.PromocionUnificada{
				NombreComercio:   "Banco CUSCATLAN",
				Titulo:           "Copa Airlines",
				DescripcionBreve: "Convierte tus MultiPuntos a ConnectMiles",
			},
			expected: "Viajes",
		},
		{
			name: "Merchant Sazón de Mar",
			promo: models.PromocionUnificada{
				NombreComercio:   "Sazón de mar",
				Titulo:           "Sazón de Mar en restaurante",
				DescripcionBreve: "30% de descuento",
			},
			expected: "Restaurantes",
		},
		{
			name: "Merchant Tartaleta",
			promo: models.PromocionUnificada{
				NombreComercio:   "Tartaleta",
				Titulo:           "Pastel perfecto en Tartaleta",
				DescripcionBreve: "10% de descuento en pasteles",
			},
			expected: "Restaurantes",
		},
		{
			name: "Merchant Naturalizer",
			promo: models.PromocionUnificada{
				NombreComercio:   "Naturalizer",
				Titulo:           "Beneficio especial",
				DescripcionBreve: "Descuento en calzado y closet",
			},
			expected: "Compras",
		},
		{
			name: "Merchant NAWI Beach House",
			promo: models.PromocionUnificada{
				NombreComercio:   "NAWI Beach House",
				Titulo:           "Tu escapada lista",
				DescripcionBreve: "Regular Day Pass en la playa",
			},
			expected: "Viajes",
		},
		{
			name: "Merchant Rincón Argentino",
			promo: models.PromocionUnificada{
				NombreComercio:   "Rincón Argentino",
				Titulo:           "25% OFF",
				DescripcionBreve: "Descuento al instante",
			},
			expected: "Restaurantes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUnifiedCategory(&tt.promo)
			if got != tt.expected {
				t.Errorf("GetUnifiedCategory() = %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestUnifyCategoriesSlice(t *testing.T) {
	promos := []models.PromocionUnificada{
		{NombreComercio: "Cinemark", Titulo: "2x1 peliculas"},
		{NombreComercio: "Claro", Titulo: "Plan prepago"},
	}

	unified := UnifyCategories(promos)
	if len(unified) != 2 {
		t.Fatalf("expected 2 promos, got %d", len(unified))
	}
	if unified[0].Categoria != "Entretenimiento" {
		t.Errorf("expected Cinemark to be Entretenimiento, got %q", unified[0].Categoria)
	}
	if unified[1].Categoria != "Tecnología" {
		t.Errorf("expected Claro to be Tecnología, got %q", unified[1].Categoria)
	}
}
