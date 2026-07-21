package models

import (
	"encoding/json"
	"testing"
)

func TestPromocionUnificadaJSON(t *testing.T) {
	promo := PromocionUnificada{
		ID:                  "promo-123",
		BancoOrigen:         "BAC",
		Titulo:              "2x1 en Pizza",
		DescripcionBreve:    "Aplica los martes",
		UrlImagen:           "https://example.com/img.jpg",
		NombreComercio:      "Papa John's",
		Categoria:           "Restaurantes",
		FechaInicio:         "2026-01-01",
		FechaFin:            "2026-12-31",
		PorcentajeDescuento: 50.0,
		UrlExterna:          "https://example.com/deal",
		CreatedAt:           1700000000,
	}

	bytes, err := json.Marshal(promo)
	if err != nil {
		t.Fatalf("Failed to marshal PromocionUnificada: %v", err)
	}

	var unmarshaled PromocionUnificada
	if err := json.Unmarshal(bytes, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal PromocionUnificada: %v", err)
	}

	if unmarshaled.ID != promo.ID {
		t.Errorf("expected ID %q, got %q", promo.ID, unmarshaled.ID)
	}
	if unmarshaled.BancoOrigen != promo.BancoOrigen {
		t.Errorf("expected BancoOrigen %q, got %q", promo.BancoOrigen, unmarshaled.BancoOrigen)
	}
	if unmarshaled.PorcentajeDescuento != promo.PorcentajeDescuento {
		t.Errorf("expected PorcentajeDescuento %f, got %f", promo.PorcentajeDescuento, unmarshaled.PorcentajeDescuento)
	}
	if unmarshaled.CreatedAt != promo.CreatedAt {
		t.Errorf("expected CreatedAt %d, got %d", promo.CreatedAt, unmarshaled.CreatedAt)
	}
}
