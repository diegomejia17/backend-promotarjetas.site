package cache

import (
	"context"
	"testing"
	"time"

	"promotarjetas-backend/models"

	"github.com/redis/go-redis/v9"
)

type mockRedisClient struct {
	redis.Cmdable
	data map[string][]byte
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string][]byte),
	}
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if ok {
			bytes = []byte(str)
		}
	}
	m.data[key] = bytes
	cmd.SetVal("OK")
	return cmd
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	val, ok := m.data[key]
	if !ok {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(string(val))
	return cmd
}

func TestCacheOperationsWithMock(t *testing.T) {
	mock := newMockRedisClient()
	Rdb = mock
	ctx := context.Background()

	// 1. Test cache miss
	raw, err := GetPromotionsRaw(ctx)
	if err != nil {
		t.Fatalf("expected no error on cache miss, got %v", err)
	}
	if raw != nil {
		t.Fatalf("expected nil raw data on cache miss, got %s", string(raw))
	}

	list, err := GetPromotionsList(ctx)
	if err != nil {
		t.Fatalf("expected no error on cache miss list, got %v", err)
	}
	if list != nil {
		t.Fatalf("expected nil list on cache miss, got %v", list)
	}

	// 2. Test SavePromotions
	promos := []models.PromocionUnificada{
		{
			ID:          "p1",
			BancoOrigen: "AGRICOLA",
			Titulo:      "Oferta Agricola",
		},
	}
	if err := SavePromotions(ctx, promos); err != nil {
		t.Fatalf("SavePromotions failed: %v", err)
	}

	// 3. Test Cache Hit
	rawHit, err := GetPromotionsRaw(ctx)
	if err != nil {
		t.Fatalf("GetPromotionsRaw failed on hit: %v", err)
	}
	if rawHit == nil {
		t.Fatal("expected raw data on hit, got nil")
	}

	listHit, err := GetPromotionsList(ctx)
	if err != nil {
		t.Fatalf("GetPromotionsList failed on hit: %v", err)
	}
	if len(listHit) != 1 || listHit[0].ID != "p1" {
		t.Fatalf("unexpected list hit content: %+v", listHit)
	}
}

func TestGetPromotionsListInvalidJSON(t *testing.T) {
	mock := newMockRedisClient()
	mock.data["promotions:all"] = []byte("invalid-json")
	Rdb = mock
	ctx := context.Background()

	_, err := GetPromotionsList(ctx)
	if err == nil {
		t.Fatal("expected unmarshal error for invalid json, got nil")
	}
}

func TestUninitializedRedis(t *testing.T) {
	Rdb = nil
	ctx := context.Background()

	raw, err := GetPromotionsRaw(ctx)
	if err != nil {
		t.Fatalf("expected nil error when Rdb is nil, got %v", err)
	}
	if raw != nil {
		t.Fatalf("expected nil raw when Rdb is nil, got %v", raw)
	}

	err = SavePromotions(ctx, []models.PromocionUnificada{})
	if err == nil {
		t.Fatal("expected error when SavePromotions called with nil Rdb")
	}
}
