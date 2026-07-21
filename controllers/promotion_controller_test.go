package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/integrations"
	"promotarjetas-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type mockRedisForController struct {
	redis.Cmdable
	data map[string][]byte
}

func (m *mockRedisForController) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if bytes, ok := value.([]byte); ok {
		m.data[key] = bytes
	}
	cmd.SetVal("OK")
	return cmd
}

func (m *mockRedisForController) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	val, ok := m.data[key]
	if !ok {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(string(val))
	return cmd
}

func TestGetHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}

	expectedBody := `{"status":"ok"}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

func TestGetPromotionsCacheHit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRedis := &mockRedisForController{data: make(map[string][]byte)}
	mockRedis.data["promotions:all"] = []byte(`[{"id":"p1","bancoOrigen":"BAC","titulo":"Promo Test"}]`)
	cache.Rdb = mockRedis

	r := gin.New()
	cfg := config.Config{Port: "3000"}
	r.GET("/api/promotions", GetPromotions(cfg))

	req, _ := http.NewRequest(http.MethodGet, "/api/promotions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json Content-Type, got %q", w.Header().Get("Content-Type"))
	}
}

func TestForceSyncPromotionsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRedis := &mockRedisForController{data: make(map[string][]byte)}
	cache.Rdb = mockRedis

	// Mock external bank HTTP servers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	oldAgricolaURL := integrations.AgricolaURL
	oldBacURL := integrations.BacURL
	oldCuscatlanURL := integrations.CuscatlanAPIURL

	integrations.AgricolaURL = server.URL
	integrations.BacURL = server.URL
	integrations.CuscatlanAPIURL = server.URL

	defer func() {
		integrations.AgricolaURL = oldAgricolaURL
		integrations.BacURL = oldBacURL
		integrations.CuscatlanAPIURL = oldCuscatlanURL
	}()

	r := gin.New()
	cfg := config.Config{Port: "3000"}

	r.GET("/api/promotions/sync", ForceSyncPromotions(cfg))

	req, _ := http.NewRequest(http.MethodGet, "/api/promotions/sync", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
}

func TestGetPromotionsByCategoryEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRedis := &mockRedisForController{data: make(map[string][]byte)}
	mockRedis.data["promotions:all"] = []byte(`[
		{"id":"p1","bancoOrigen":"BAC","titulo":"Promo Burger","categoria":"Restaurantes"},
		{"id":"p2","bancoOrigen":"Cuscatlan","titulo":"Promo TV","categoria":"Tecnología"}
	]`)
	cache.Rdb = mockRedis

	r := gin.New()
	cfg := config.Config{Port: "3000"}
	api := r.Group("/api")
	{
		api.GET("/promotions", GetPromotions(cfg))
		api.GET("/promotions/category/:category", GetPromotionsByCategory(cfg))
		api.GET("/promotions/sync", ForceSyncPromotions(cfg))
	}

	t.Run("Match category case and accent insensitive", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/promotions/category/restaurantes", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", w.Code)
		}

		var result []models.PromocionUnificada
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("unmarshal response body failed: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 promotion, got %d", len(result))
		}
		if result[0].ID != "p1" {
			t.Errorf("expected promo ID p1, got %s", result[0].ID)
		}
	})

	t.Run("Match category with accent", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/promotions/category/tecnologia", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", w.Code)
		}

		var result []models.PromocionUnificada
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("unmarshal response body failed: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 promotion, got %d", len(result))
		}
		if result[0].ID != "p2" {
			t.Errorf("expected promo ID p2, got %s", result[0].ID)
		}
	})

	t.Run("No match returns empty array", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/promotions/category/Salud", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", w.Code)
		}

		var result []models.PromocionUnificada
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("unmarshal response body failed: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("expected 0 promotions, got %d", len(result))
		}
	})
}
