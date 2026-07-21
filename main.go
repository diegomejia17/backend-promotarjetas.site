package main

import (
	"context"
	"log"
	"net/http"

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

	if err := cache.InitRedis(cfg.RedisURL, cfg.RedisPassword); err != nil {
		log.Printf("Advertencia: No se pudo conectar a Redis durante el inicio: %v. El servicio continuará intentándolo al recibir peticiones.\n", err)
	}

	go services.SyncPromotions(context.Background(), cfg)

	c := cron.New()
	// Run every day at midnight
	c.AddFunc("0 0 * * *", func() {
		log.Println("Ejecutando cron diario para sincronizar promociones")
		services.SyncPromotions(context.Background(), cfg)
	})
	c.Start()

	r := gin.Default()

	// Configurar CORS de manera segura para producción
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	r.Use(cors.New(corsConfig))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

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
