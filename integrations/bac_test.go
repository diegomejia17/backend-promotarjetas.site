package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchBACSuccess(t *testing.T) {
	mockResponse := `{
		"data": {
			"response": {
				"numFound": 1,
				"docs": [
					{
						"id": "bac-101",
						"title": "*Promo Pizza*",
						"description": "<p>50% off</p>",
						"restrictions": "Solo efectivo",
						"validity_from": "2026-01-01",
						"validity_to": "2026-01-31",
						"discount_percent_value": 50.0,
						"category_translation": "Restaurantes",
						"merchant_name": "Pizza Hut",
						"_childDocuments_": {
							"IMAGE": [
								{
									"image_filepath": "/deals",
									"image_filename": "pizza.jpg"
								}
							]
						}
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("country-id") != "60" {
			t.Errorf("expected country-id header '60', got %q", r.Header.Get("country-id"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	oldURL := BacURL
	BacURL = server.URL
	defer func() { BacURL = oldURL }()

	ctx := context.Background()
	promos, err := FetchBAC(ctx)
	if err != nil {
		t.Fatalf("FetchBAC failed: %v", err)
	}

	if len(promos) != 1 {
		t.Fatalf("expected 1 promo, got %d", len(promos))
	}

	p := promos[0]
	if p.ID != "bac-101" {
		t.Errorf("expected ID 'bac-101', got %q", p.ID)
	}
	if p.BancoOrigen != "BAC" {
		t.Errorf("expected BancoOrigen 'BAC', got %q", p.BancoOrigen)
	}
	if p.Titulo != "Promo Pizza" {
		t.Errorf("expected title 'Promo Pizza', got %q", p.Titulo)
	}
	if p.UrlImagen != "https://mipromoimages.geopagoscdn.net/deals/pizza.jpg" {
		t.Errorf("expected full image URL, got %q", p.UrlImagen)
	}
}

func TestFetchBACHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	oldURL := BacURL
	BacURL = server.URL
	defer func() { BacURL = oldURL }()

	ctx := context.Background()
	_, err := FetchBAC(ctx)
	if err == nil {
		t.Fatal("expected error on HTTP 502, got nil")
	}
}
