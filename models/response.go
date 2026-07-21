package models

// HealthResponse representa la respuesta del endpoint de salud.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// MessageResponse representa una respuesta con un mensaje descriptivo.
type MessageResponse struct {
	Message string `json:"message" example:"Sincronizacion ejecutada exitosamente"`
}

// ErrorResponse representa una respuesta de error.
type ErrorResponse struct {
	Error string `json:"error" example:"Error interno resolviendo el caché"`
}
