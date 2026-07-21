package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCuscatlanSuccess(t *testing.T) {
	mockResponse := `{
		"data": {
			"promocions": {
				"data": [
					{
						"id": "cusc-01",
						"attributes": {
							"h1_title": "Descuento Omnisport",
							"date_start": "2026-01-01",
							"date_end": "2026-12-31",
							"business": {
								"data": {
									"attributes": {
										"name": "Omnisport",
										"description": "Tienda de electrodomesticos"
									}
								}
							},
							"card": {
								"title": "Descuento Omnisport",
								"description": "15% off",
								"imagen": {
									"data": {
										"attributes": {
											"url": "https://cuscatlan.com/img.jpg"
										}
									}
								}
							}
						}
					}
				]
			},
			"coupons": {
				"data": [
					{
						"id": "cup-01",
						"attributes": {
							"title": "Cupon Cinemark",
							"publishedAt": "2026-01-01",
							"terms_cond": "2x1 en entradas",
							"imagen": {
								"data": {
									"attributes": {
										"url": "https://cuscatlan.com/cup.jpg"
									}
								}
							}
						}
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("apikey") != "test-api-key" {
			t.Errorf("expected apikey header 'test-api-key', got %q", r.Header.Get("apikey"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	oldURL := CuscatlanAPIURL
	CuscatlanAPIURL = server.URL
	defer func() { CuscatlanAPIURL = oldURL }()

	ctx := context.Background()
	promos, err := FetchCuscatlan(ctx, "test-api-key")
	if err != nil {
		t.Fatalf("FetchCuscatlan failed: %v", err)
	}

	if len(promos) != 2 {
		t.Fatalf("expected 2 items (1 promo + 1 coupon), got %d", len(promos))
	}

	p1 := promos[0]
	if p1.ID != "cusc-01" || p1.BancoOrigen != "CUSCATLAN" {
		t.Errorf("unexpected promo 1: %+v", p1)
	}

	p2 := promos[1]
	if p2.ID != "cup-01" || p2.Categoria != "Cupones" {
		t.Errorf("unexpected coupon: %+v", p2)
	}
}

func TestFetchCuscatlanHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	}))
	defer server.Close()

	oldURL := CuscatlanAPIURL
	CuscatlanAPIURL = server.URL
	defer func() { CuscatlanAPIURL = oldURL }()

	ctx := context.Background()
	_, err := FetchCuscatlan(ctx, "wrong-key")
	if err == nil {
		t.Fatal("expected error on HTTP 401, got nil")
	}
}
