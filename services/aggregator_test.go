package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/integrations"
	"promotarjetas-backend/models"

	"github.com/redis/go-redis/v9"
)

type mockSyncRedisClient struct {
	redis.Cmdable
	data map[string][]byte
}

func (m *mockSyncRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if bytes, ok := value.([]byte); ok {
		m.data[key] = bytes
	}
	cmd.SetVal("OK")
	return cmd
}

func (m *mockSyncRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	val, ok := m.data[key]
	if !ok {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(string(val))
	return cmd
}

func TestSyncPromotionsConcurrent(t *testing.T) {
	mockRedis := &mockSyncRedisClient{data: make(map[string][]byte)}
	cache.Rdb = mockRedis

	agricolaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"promociones":[{"id_promocion":"a1","nombre_promocion":"Promo Agricola","nombre_comercio":"Papa John's"}]}`))
	}))
	defer agricolaServer.Close()

	bacServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"response":{"numFound":1,"docs":[{"id":"b1","title":"Promo BAC","merchant_name":"Cinemark"}]}}}`))
	}))
	defer bacServer.Close()

	cuscatlanServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"promocions":{"data":[{"id":"c1","attributes":{"h1_title":"Promo Cuscatlan","business":{"data":{"attributes":{"name":"Omnisport"}}}}}]},"coupons":{"data":[]}}}`))
	}))
	defer cuscatlanServer.Close()

	oldAgricolaURL := integrations.AgricolaURL
	oldBacURL := integrations.BacURL
	oldCuscatlanURL := integrations.CuscatlanAPIURL

	integrations.AgricolaURL = agricolaServer.URL
	integrations.BacURL = bacServer.URL
	integrations.CuscatlanAPIURL = cuscatlanServer.URL

	defer func() {
		integrations.AgricolaURL = oldAgricolaURL
		integrations.BacURL = oldBacURL
		integrations.CuscatlanAPIURL = oldCuscatlanURL
	}()

	cfg := config.Config{CuscatlanAPIKey: "test-key"}
	ctx := context.Background()

	// Execute SyncPromotions
	SyncPromotions(ctx, cfg)

	// Verify sync single flight lock state reset
	syncMutex.Lock()
	syncing := isSyncing
	syncMutex.Unlock()

	if syncing {
		t.Error("expected isSyncing to be false after completion")
	}

	// Verify mock redis stored items
	list, err := cache.GetPromotionsList(ctx)
	if err != nil {
		t.Fatalf("failed to retrieve stored promotions from mock redis: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 promotions aggregated, got %d", len(list))
	}
}

func TestFilterPromotionsByCategory(t *testing.T) {
	promotions := []models.PromocionUnificada{
		{ID: "1", Titulo: "Pizza Gratis", Categoria: "Restaurantes"},
		{ID: "2", Titulo: "Laptop Descuento", Categoria: "Tecnología"},
		{ID: "3", Titulo: "Hamburguesa 2x1", Categoria: "Restaurantes"},
		{ID: "4", Titulo: "Super Oferta", Categoria: "Supermercados"},
	}

	tests := []struct {
		name          string
		category      string
		expectedCount int
	}{
		{
			name:          "Exact match case-sensitive",
			category:      "Restaurantes",
			expectedCount: 2,
		},
		{
			name:          "Case-insensitive match",
			category:      "restaurantes",
			expectedCount: 2,
		},
		{
			name:          "Accent-insensitive match",
			category:      "tecnologia",
			expectedCount: 1,
		},
		{
			name:          "Category with accent match",
			category:      "Tecnología",
			expectedCount: 1,
		},
		{
			name:          "No matching category",
			category:      "Viajes",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterPromotionsByCategory(promotions, tt.category)
			if len(got) != tt.expectedCount {
				t.Errorf("FilterPromotionsByCategory(%q) count = %d; want %d", tt.category, len(got), tt.expectedCount)
			}
		})
	}
}
