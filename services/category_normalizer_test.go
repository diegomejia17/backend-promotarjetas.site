package services

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"promotarjetas-backend/models"
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

type mockClassifier struct {
	classifyFunc func(ctx context.Context, promo models.PromocionUnificada) (string, float64, error)
	calls        int32
}

func (m *mockClassifier) ClassifyPromotion(ctx context.Context, promo models.PromocionUnificada) (string, float64, error) {
	atomic.AddInt32(&m.calls, 1)
	if m.classifyFunc != nil {
		return m.classifyFunc(ctx, promo)
	}
	return "Otros", 0.0, nil
}

func TestUnifyCategoriesWithClassifier_MockJevSuccess(t *testing.T) {
	classifier := &mockClassifier{
		classifyFunc: func(ctx context.Context, promo models.PromocionUnificada) (string, float64, error) {
			if promo.NombreComercio == "Fogo Bonito" {
				return "Restaurantes", 0.95, nil
			}
			if promo.NombreComercio == "Converse" {
				return "Compras", 0.88, nil
			}
			return "Otros", 0.0, nil
		},
	}

	promos := []models.PromocionUnificada{
		{
			ID:             "p1",
			NombreComercio: "Papa John's", // Override exacto
			Titulo:         "50% Descuento",
		},
		{
			ID:               "p2",
			NombreComercio:   "Fogo Bonito", // Desconocido en overrides y sin palabras clave en texto -> "Otros"
			Titulo:           "Ese momento especial ya tiene una razón más para disfrutarse",
			DescripcionBreve: "Pasa por el local y aprovecha este beneficio con tus tarjetas",
		},
		{
			ID:               "p3",
			NombreComercio:   "Converse", // Desconocido en overrides y sin palabras clave en texto -> "Otros"
			Titulo:           "Detalles únicos para inspirar",
			DescripcionBreve: "Visita la sucursal y disfruta este beneficio exclusivo",
		},
	}

	res := UnifyCategoriesWithClassifier(context.Background(), promos, classifier, false)

	// p1 se resuelve por override exacto sin invocar el clasificador
	if res[0].Categoria != "Restaurantes" {
		t.Errorf("expected Papa John's to be Restaurantes via override, got %q", res[0].Categoria)
	}
	// p2 y p3 son resueltas por el clasificador Jev mock
	if res[1].Categoria != "Restaurantes" {
		t.Errorf("expected Fogo Bonito to be Restaurantes via Jev, got %q", res[1].Categoria)
	}
	if res[2].Categoria != "Compras" {
		t.Errorf("expected Converse to be Compras via Jev, got %q", res[2].Categoria)
	}

	// El clasificador solo debe haber sido llamado para Fogo Bonito y Converse (2 veces), no para Papa John's
	if calls := atomic.LoadInt32(&classifier.calls); calls != 2 {
		t.Errorf("expected classifier to be called 2 times, got %d", calls)
	}
}

func TestUnifyCategoriesWithClassifier_LowConfidenceFallback(t *testing.T) {
	classifier := &mockClassifier{
		classifyFunc: func(ctx context.Context, promo models.PromocionUnificada) (string, float64, error) {
			// Retorna baja certeza (< 0.40)
			return "Tecnología", 0.25, nil
		},
	}

	promos := []models.PromocionUnificada{
		{
			ID:               "p-low",
			NombreComercio:   "Comercio Incierto",
			Titulo:           "Promocion",
			DescripcionBreve: "Detalles",
		},
	}

	res := UnifyCategoriesWithClassifier(context.Background(), promos, classifier, false)

	if res[0].Categoria != "Otros" {
		t.Errorf("expected fallback to 'Otros' due to low confidence, got %q", res[0].Categoria)
	}
}

func TestUnifyCategoriesWithClassifier_ErrorFallback(t *testing.T) {
	classifier := &mockClassifier{
		classifyFunc: func(ctx context.Context, promo models.PromocionUnificada) (string, float64, error) {
			return "", 0, errors.New("timeout connecting to typesafe")
		},
	}

	promos := []models.PromocionUnificada{
		{
			ID:               "p-err",
			NombreComercio:   "Comercio Error",
			Titulo:           "Promocion",
			DescripcionBreve: "Detalles",
		},
	}

	res := UnifyCategoriesWithClassifier(context.Background(), promos, classifier, false)

	if res[0].Categoria != "Otros" {
		t.Errorf("expected fallback to 'Otros' due to classifier error, got %q", res[0].Categoria)
	}
}

func TestUnifyCategoriesWithClassifier_Concurrent(t *testing.T) {
	classifier := &mockClassifier{
		classifyFunc: func(ctx context.Context, promo models.PromocionUnificada) (string, float64, error) {
			return "Salud", 0.90, nil
		},
	}

	var promos []models.PromocionUnificada
	for i := 0; i < 30; i++ {
		promos = append(promos, models.PromocionUnificada{
			ID:             fmt.Sprintf("promo-%d", i),
			NombreComercio: fmt.Sprintf("Comercio %d", i),
			Titulo:         "Promocion especial",
		})
	}

	res := UnifyCategoriesWithClassifier(context.Background(), promos, classifier, false)
	if len(res) != 30 {
		t.Fatalf("expected 30 results, got %d", len(res))
	}
	for i, p := range res {
		if p.Categoria != "Salud" {
			t.Errorf("item %d expected Salud, got %q", i, p.Categoria)
		}
	}
}
