package controllers

import (
	"net/http"

	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/models"
	"promotarjetas-backend/services"

	"github.com/gin-gonic/gin"
)

// GetPromotions godoc
// @Summary      Obtener promociones unificadas
// @Description  Obtiene la lista completa de promociones de tarjetas de crédito/débito unificadas y almacenadas en caché.
// @Tags         promotions
// @Produce      json
// @Success      200  {array}   models.PromocionUnificada
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/promotions [get]
func GetPromotions(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		rawPromotions, err := cache.GetPromotionsRaw(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno resolviendo el caché"})
			return
		}

		if rawPromotions == nil { // Cache miss
			services.SyncPromotions(ctx, cfg)
			rawPromotions, _ = cache.GetPromotionsRaw(ctx)
		}

		// Enviar directamente los bytes pre-procesados como application/json
		// Evita deserializar y serializar en cada petición.
		c.Data(http.StatusOK, "application/json", rawPromotions)
	}
}

// ForceSyncPromotions godoc
// @Summary      Forzar sincronización de promociones
// @Description  Fuerza la sincronización inmediata de promociones desde las APIs externas de los bancos.
// @Tags         promotions
// @Produce      json
// @Success      200  {object}  models.MessageResponse
// @Router       /api/promotions/sync [get]
func ForceSyncPromotions(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		services.SyncPromotions(c.Request.Context(), cfg)
		c.JSON(http.StatusOK, gin.H{"message": "Sincronizacion ejecutada exitosamente"})
	}
}

// HealthCheck godoc
// @Summary      Estado del servicio
// @Description  Verifica el estado de salud del servidor backend.
// @Tags         health
// @Produce      json
// @Success      200  {object}  models.HealthResponse
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{Status: "ok"})
}
