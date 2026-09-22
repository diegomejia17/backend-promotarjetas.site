package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"promotarjetas-backend/models"
)

// DefaultCategoryCriteria define la rúbrica de opciones para clasificar promociones con Jev
var DefaultCategoryCriteria = map[string]string{
	"Restaurantes":    "Restaurantes, cafeterías, comida rápida, panaderías, pastelerías, bares, carnes, pizzerías, sushi y delivery de alimentos.",
	"Supermercados":   "Supermercados, abarrotes, tiendas de conveniencia, clubes de compra (como PriceSmart) y productos de la canasta básica.",
	"Viajes":          "Aerolíneas, boletos aéreos, hoteles, resorts, agencias de viaje, alquiler de autos, turismo, millas y casilleros o envíos internacionales.",
	"Compras":         "Ropa, calzado, zapaterías, tiendas por departamento, boutiques, accesorios, joyerías, librerías, papelerías y artículos personales.",
	"Salud":           "Farmacias, clínicas, hospitales, laboratorios médicos, dentistas, odontología, ópticas, spas, salones de belleza y cuidado personal.",
	"Tecnología":      "Electrónica, telefonía celular, smartphones, computadoras, laptops, audio, electrodomésticos, internet, software y telecomunicaciones.",
	"Hogar":           "Muebles, artículos para el hogar, decoración, ferretería, construcción, pintura, jardín, colchones y remodelación de espacios.",
	"Entretenimiento": "Cines, películas, teatros, conciertos, espectáculos, parques de atracciones, museos y plataformas de streaming.",
	"Otros":           "Servicios financieros generales, trámites o promociones que no pertenecen a ninguna de las categorías anteriores.",
}

// SystemOneRequest representa el payload para el endpoint /v1/systemone de TypeSafe AI
type SystemOneRequest struct {
	State     any                 `json:"state"`
	Model     string              `json:"model"`
	Questions map[string]Question `json:"questions"`
}

// Question define una pregunta tipada para el modelo Jev
type Question struct {
	Type         string            `json:"type"`
	Instructions any               `json:"instructions"`
	Criteria     map[string]string `json:"criteria,omitempty"`
}

// SystemOneResponse representa la respuesta estructurada devuelta por TypeSafe AI
type SystemOneResponse struct {
	Model   string            `json:"model"`
	Answers map[string]Answer `json:"answers"`
	Usage   *UsageInfo        `json:"usage,omitempty"`
}

// Answer representa el resultado estructurado de una pregunta evaluada por Jev
type Answer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
}

// UsageInfo reporta el consumo de tokens en la petición
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// TypeSafeClient es el cliente HTTP para interactuar con la API System One (Jev) de TypeSafe
type TypeSafeClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	maxRetries int
}

// NewTypeSafeClient inicializa una nueva instancia de TypeSafeClient
func NewTypeSafeClient(apiKey, baseURL, model string, timeout time.Duration) *TypeSafeClient {
	if baseURL == "" {
		baseURL = "https://api.typesafe.ai/v1/systemone"
	}
	if model == "" {
		model = "jev-latest"
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &TypeSafeClient{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: 2,
	}
}

// IsEnabled indica si el cliente cuenta con API Key para realizar evaluaciones
func (c *TypeSafeClient) IsEnabled() bool {
	return c != nil && c.apiKey != ""
}

// Evaluate envía una solicitud estructurada al endpoint de evaluación System One con reintentos para límites de tasa
func (c *TypeSafeClient) Evaluate(ctx context.Context, req SystemOneRequest) (*SystemOneResponse, error) {
	if c == nil || c.apiKey == "" {
		return nil, errors.New("typesafe evaluate: apiKey no configurada")
	}

	if req.Model == "" {
		req.Model = c.model
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("typesafe evaluate: serializar petición: %w", err)
	}

	var lastErr error
	var backoff = 300 * time.Millisecond

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("typesafe evaluate: cancelado durante reintento: %w", ctx.Err())
			case <-time.After(backoff):
				backoff *= 2
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("typesafe evaluate: crear solicitud HTTP: %w", err)
		}

		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", "promotarjetas-backend/1.0")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("typesafe evaluate: enviar petición: %w", err)
			continue
		}

		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Límite de 1MB
		_ = resp.Body.Close()

		if readErr != nil {
			lastErr = fmt.Errorf("typesafe evaluate: leer respuesta: %w", readErr)
			continue
		}

		// Reintentar si la API indica saturación o límite de tasa (429 o 529)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == 529 {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 && seconds <= 5 {
					backoff = time.Duration(seconds) * time.Second
				}
			}
			lastErr = fmt.Errorf("typesafe evaluate: recibido status %d: reintentando", resp.StatusCode)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("typesafe evaluate: código HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		}

		var sysResp SystemOneResponse
		if err := json.Unmarshal(respBody, &sysResp); err != nil {
			return nil, fmt.Errorf("typesafe evaluate: deserializar respuesta: %w", err)
		}

		return &sysResp, nil
	}

	return nil, fmt.Errorf("typesafe evaluate: excedido número de reintentos: %w", lastErr)
}

// ClassifyPromotion evalúa una promoción unificada mediante Jev Choice y retorna la categoría y la certeza
func (c *TypeSafeClient) ClassifyPromotion(ctx context.Context, promo models.PromocionUnificada) (string, float64, error) {
	state := map[string]string{
		"comercio":    promo.NombreComercio,
		"titulo":      promo.Titulo,
		"descripcion": promo.DescripcionBreve,
		"banco":       promo.BancoOrigen,
	}

	req := SystemOneRequest{
		State: state,
		Model: c.model,
		Questions: map[string]Question{
			"categoria": {
				Type:         "choice",
				Instructions: "Clasifica esta promoción bancaria en la categoría más adecuada basándote en el comercio y el beneficio.",
				Criteria:     DefaultCategoryCriteria,
			},
		},
	}

	resp, err := c.Evaluate(ctx, req)
	if err != nil {
		return "", 0, fmt.Errorf("typesafe clasificar promocion %s: %w", promo.ID, err)
	}

	ans, ok := resp.Answers["categoria"]
	if !ok {
		return "", 0, fmt.Errorf("typesafe clasificar promocion %s: respuesta sin campo 'categoria'", promo.ID)
	}

	return ans.Choice, ans.Confidence, nil
}
