package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchAgricolaSuccess(t *testing.T) {
	mockResponse := `{
		"promociones": [
			{
				"id_promocion": "agr-01",
				"nombre_promocion": "*Descuento en Restaurante*",
				"descripcion": "20% de descuento",
				"restricciones": "Aplica tarjetas de crédito",
				"imagen_preview": "https://agricola.com/img.png",
				"slug": "https://agricola.com/promo1",
				"nombre_comercio": "La Pampa"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	oldURL := AgricolaURL
	AgricolaURL = server.URL
	defer func() { AgricolaURL = oldURL }()

	ctx := context.Background()
	promos, err := FetchAgricola(ctx)
	if err != nil {
		t.Fatalf("FetchAgricola failed: %v", err)
	}

	if len(promos) != 1 {
		t.Fatalf("expected 1 promo, got %d", len(promos))
	}

	p := promos[0]
	if p.ID != "agr-01" {
		t.Errorf("expected ID 'agr-01', got %q", p.ID)
	}
	if p.BancoOrigen != "AGRICOLA" {
		t.Errorf("expected BancoOrigen 'AGRICOLA', got %q", p.BancoOrigen)
	}
	if p.Titulo != "Descuento en Restaurante" {
		t.Errorf("expected cleaned title 'Descuento en Restaurante', got %q", p.Titulo)
	}
}

func TestFetchAgricolaHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldURL := AgricolaURL
	AgricolaURL = server.URL
	defer func() { AgricolaURL = oldURL }()

	ctx := context.Background()
	_, err := FetchAgricola(ctx)
	if err == nil {
		t.Fatal("expected error on HTTP 500, got nil")
	}
}

func TestFetchAgricolaInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid-json"))
	}))
	defer server.Close()

	oldURL := AgricolaURL
	AgricolaURL = server.URL
	defer func() { AgricolaURL = oldURL }()

	ctx := context.Background()
	_, err := FetchAgricola(ctx)
	if err == nil {
		t.Fatal("expected error on invalid json, got nil")
	}
}

func TestFetchAgricolaContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := AgricolaURL
	AgricolaURL = server.URL
	defer func() { AgricolaURL = oldURL }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := FetchAgricola(ctx)
	if err == nil {
		t.Fatal("expected error on canceled context, got nil")
	}
}
