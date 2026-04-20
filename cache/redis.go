package cache

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"promotarjetas-backend/models"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var ctx = context.Background()

func InitRedis(redisURL string, password string) {
	var options *redis.Options
	var err error

	if strings.HasPrefix(redisURL, "redis://") {
		options, err = redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Error parseando REDIS_URL: %v", err)
		}
	} else {
		options = &redis.Options{
			Addr:     redisURL,
			Password: password,
			DB:       0,
		}
	}

	Rdb = redis.NewClient(options)

	_, err = Rdb.Ping(ctx).Result()
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

func GetPromotionsList() ([]models.PromocionUnificada, error) {
	raw, err := GetPromotionsRaw()
	if err != nil || raw == nil {
		return nil, err
	}

	var promotions []models.PromocionUnificada
	if err := json.Unmarshal(raw, &promotions); err != nil {
		return nil, err
	}

	return promotions, nil
}
