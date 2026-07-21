package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"promotarjetas-backend/models"

	"github.com/redis/go-redis/v9"
)

var Rdb redis.Cmdable

func InitRedis(redisURL string, password string) error {
	var options *redis.Options
	var err error

	if strings.HasPrefix(redisURL, "redis://") {
		options, err = redis.ParseURL(redisURL)
		if err != nil {
			return fmt.Errorf("parse REDIS_URL: %w", err)
		}
	} else {
		options = &redis.Options{
			Addr:     redisURL,
			Password: password,
			DB:       0,
		}
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("ping Redis: %w", err)
	}
	Rdb = client
	log.Println("Connected to Redis")
	return nil
}

func SavePromotions(ctx context.Context, promotions []models.PromocionUnificada) error {
	if Rdb == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	data, err := json.Marshal(promotions)
	if err != nil {
		return fmt.Errorf("marshal promotions: %w", err)
	}
	if err := Rdb.Set(ctx, "promotions:all", data, 25*time.Hour).Err(); err != nil {
		return fmt.Errorf("save promotions to redis: %w", err)
	}
	return nil
}

func GetPromotionsRaw(ctx context.Context) ([]byte, error) {
	if Rdb == nil {
		return nil, nil // Cache miss if redis is not connected
	}
	val, err := Rdb.Get(ctx, "promotions:all").Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, fmt.Errorf("get promotions from redis: %w", err)
	}

	return val, nil
}

func GetPromotionsList(ctx context.Context) ([]models.PromocionUnificada, error) {
	raw, err := GetPromotionsRaw(ctx)
	if err != nil || raw == nil {
		return nil, err
	}

	var promotions []models.PromocionUnificada
	if err := json.Unmarshal(raw, &promotions); err != nil {
		return nil, fmt.Errorf("unmarshal promotions: %w", err)
	}

	return promotions, nil
}
