package controllers

import (
	"net/http"

	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/services"

	"github.com/gin-gonic/gin"
)

func GetPromotions(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPromotions, err := cache.GetPromotionsRaw()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno resolviendo el caché"})
			return
		}

		if rawPromotions == nil { // Cache miss
			services.SyncPromotions(cfg)
			rawPromotions, _ = cache.GetPromotionsRaw()
		}

		// Enviar directamente los bytes pre-procesados como application/json
		// Evita deserializar y serializar en cada petición.
		c.Data(http.StatusOK, "application/json", rawPromotions)
	}
}

func ForceSyncPromotions(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		services.SyncPromotions(cfg)
		c.JSON(http.StatusOK, gin.H{"message": "Sincronizacion ejecutada exitosamente"})
	}
}
