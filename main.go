package main

import (
	"log"

	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/controllers"
	"promotarjetas-backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.LoadConfig()

	cache.InitRedis(cfg.RedisURL, cfg.RedisPassword)

	go services.SyncPromotions(cfg)

	c := cron.New()
	// Run every day at midnight
	c.AddFunc("0 0 * * *", func() {
		log.Println("Ejecutando cron diario para sicronizar promociones")
		services.SyncPromotions(cfg)
	})
	c.Start()

	r := gin.Default()
	r.Use(cors.Default())

	api := r.Group("/api")
	{
		api.GET("/promotions", controllers.GetPromotions(cfg))
		api.GET("/promotions/sync", controllers.ForceSyncPromotions(cfg))
	}

	log.Printf("Iniciando servidor en el puerto %s...\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
