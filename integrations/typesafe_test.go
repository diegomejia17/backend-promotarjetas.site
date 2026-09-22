package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"promotarjetas-backend/models"
)

func TestTypeSafeClient_Evaluate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-api-key" {
			t.Errorf("unexpected Authorization header: %q", auth)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %q", ct)
		}

		var req SystemOneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		resp := SystemOneResponse{
			Model: "jev-1.13.0",
			Answers: map[string]Answer{
				"categoria": {
					Type:       "choice",
					Choice:     "Restaurantes",
					Confidence: 0.95,
					Probabilities: map[string]float64{
						"Restaurantes": 0.95,
						"Compras":      0.05,
					},
				},
			},
			Usage: &UsageInfo{
				InputTokens:  120,
				OutputTokens: 15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewTypeSafeClient("test-api-key", server.URL, "jev-latest", 2*time.Second)
	if !client.IsEnabled() {
		t.Fatal("expected client to be enabled")
	}

	promo := models.PromocionUnificada{
		ID:               "test-1",
		NombreComercio:   "Fogo Bonito",
		Titulo:           "Almuerzo especial con tarjeta",
		DescripcionBreve: "Descuento en cortes de carne",
		BancoOrigen:      "AGRICOLA",
	}

	cat, conf, err := client.ClassifyPromotion(context.Background(), promo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat != "Restaurantes" {
		t.Errorf("expected Restaurantes, got %q", cat)
	}
	if conf != 0.95 {
		t.Errorf("expected 0.95, got %f", conf)
	}
}

func TestTypeSafeClient_Evaluate_RateLimitRetry(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
			return
		}

		resp := SystemOneResponse{
			Model: "jev-1.13.0",
			Answers: map[string]Answer{
				"test": {
					Type:   "choice",
					Choice: "Compras",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewTypeSafeClient("test-key", server.URL, "jev-latest", 5*time.Second)

	req := SystemOneRequest{
		State: "test state",
		Questions: map[string]Question{
			"test": {Type: "choice"},
		},
	}

	resp, err := client.Evaluate(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if resp.Answers["test"].Choice != "Compras" {
		t.Errorf("expected Compras, got %s", resp.Answers["test"].Choice)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestTypeSafeClient_Evaluate_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_api_key"}`))
	}))
	defer server.Close()

	client := NewTypeSafeClient("wrong-key", server.URL, "jev-latest", 2*time.Second)
	_, err := client.Evaluate(context.Background(), SystemOneRequest{State: "hello"})
	if err == nil {
		t.Fatal("expected error on 401 Unauthorized, got nil")
	}
}

func TestTypeSafeClient_Evaluate_MissingAPIKey(t *testing.T) {
	client := NewTypeSafeClient("", "http://localhost", "jev-latest", 2*time.Second)
	if client.IsEnabled() {
		t.Fatal("expected client to be disabled with empty key")
	}

	_, err := client.Evaluate(context.Background(), SystemOneRequest{})
	if err == nil {
		t.Fatal("expected error with empty key, got nil")
	}
}

func TestTypeSafeClient_Evaluate_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Canceled immediately

	client := NewTypeSafeClient("key", server.URL, "jev-latest", 2*time.Second)
	_, err := client.Evaluate(ctx, SystemOneRequest{})
	if err == nil {
		t.Fatal("expected context canceled error, got nil")
	}
}
