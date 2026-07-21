package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/integrations"

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
