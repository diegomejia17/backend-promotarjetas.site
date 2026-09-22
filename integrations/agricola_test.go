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

	var receivedUA, receivedReferer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		receivedReferer = r.Header.Get("Referer")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Test fallback to AgricolaURL when apiURL is empty
	oldURL := AgricolaURL
	AgricolaURL = server.URL
	defer func() { AgricolaURL = oldURL }()

	ctx := context.Background()
	promos, err := FetchAgricola(ctx, "", "")
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
	if receivedUA == "" {
		t.Error("expected User-Agent header to be set")
	}
	if receivedReferer != "https://www.bancoagricola.com/promociones" {
		t.Errorf("expected Referer header 'https://www.bancoagricola.com/promociones', got %q", receivedReferer)
	}
}

func TestFetchAgricolaWithCustomURLAndCookie(t *testing.T) {
	mockResponse := `{"promociones":[]}`
	var receivedCookie string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCookie = r.Header.Get("Cookie")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	promos, err := FetchAgricola(ctx, server.URL, "session_id=xyz; visid_incap=123")
	if err != nil {
		t.Fatalf("FetchAgricola failed: %v", err)
	}

	if len(promos) != 0 {
		t.Errorf("expected 0 promos, got %d", len(promos))
	}
	if receivedCookie != "session_id=xyz; visid_incap=123" {
		t.Errorf("expected Cookie header 'session_id=xyz; visid_incap=123', got %q", receivedCookie)
	}
}

func TestFetchAgricolaAutoNegotiateCookies(t *testing.T) {
	var receivedWarmup bool
	var receivedAPICookie string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/promociones" {
			receivedWarmup = true
			http.SetCookie(w, &http.Cookie{
				Name:  "incap_ses",
				Value: "auto-session-123",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><body>Warmup OK</body></html>`))
			return
		}

		if r.URL.Path == "/api" {
			receivedAPICookie = r.Header.Get("Cookie")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"promociones":[{"id_promocion":"auto-1","nombre_promocion":"Auto Promo"}]}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oldAgricolaURL := AgricolaURL
	oldPageURL := AgricolaPromocionesPageURL
	AgricolaURL = server.URL + "/api"
	AgricolaPromocionesPageURL = server.URL + "/promociones"
	defer func() {
		AgricolaURL = oldAgricolaURL
		AgricolaPromocionesPageURL = oldPageURL
	}()

	ctx := context.Background()
	promos, err := FetchAgricola(ctx, "", "")
	if err != nil {
		t.Fatalf("FetchAgricola auto-negotiation failed: %v", err)
	}

	if !receivedWarmup {
		t.Error("expected warmup request to /promociones to be made")
	}
	if receivedAPICookie != "incap_ses=auto-session-123" {
		t.Errorf("expected API to receive negotiated cookie, got %q", receivedAPICookie)
	}
	if len(promos) != 1 || promos[0].ID != "auto-1" {
		t.Errorf("unexpected promos returned: %+v", promos)
	}
}

func TestFetchAgricolaHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	_, err := FetchAgricola(ctx, server.URL, "")
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

	ctx := context.Background()
	_, err := FetchAgricola(ctx, server.URL, "")
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := FetchAgricola(ctx, server.URL, "")
	if err == nil {
		t.Fatal("expected error on canceled context, got nil")
	}
}
