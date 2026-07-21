package main

import (
	"log"

	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/controllers"
	_ "promotarjetas-backend/docs"
	"promotarjetas-backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           PromoTarjetas Backend API
// @version         1.0
// @description     API para consultar y sincronizar promociones bancarias unificadas.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Soporte PromoTarjetas

// @host      localhost:3000
// @BasePath  /
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
	
	// Configurar CORS de manera segura para producción
	corsConfig := cors.DefaultConfig()
	// En lugar de AllowAll, puedes restringir a tus dominios específicos:
	// corsConfig.AllowOrigins = []string{"https://tu-dominio.com", "http://localhost:4200"}
	corsConfig.AllowAllOrigins = true // Cambiar a false y usar AllowOrigins en producción real
	r.Use(cors.New(corsConfig))

	r.GET("/health", controllers.HealthCheck)

	// Endpoint para Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
