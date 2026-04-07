package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"promotarjetas-backend/models"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var ctx = context.Background()

func InitRedis(redisURL string, password string) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       0,
	})

	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
}

func SavePromotions(promotions []models.PromocionUnificada) error {
	data, err := json.Marshal(promotions)
	if err != nil {
		return err
	}
	// Expira en 25 horas (margen sobre el cron de 24h)
	return Rdb.Set(ctx, "promotions:all", data, 25*time.Hour).Err()
}

func GetPromotionsRaw() ([]byte, error) {
	val, err := Rdb.Get(ctx, "promotions:all").Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, err
	}

	return val, nil
}
