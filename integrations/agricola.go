package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"promotarjetas-backend/models"
	"promotarjetas-backend/utils"
	"time"
)

var AgricolaURL = "https://www.bancoagricola.com/com/promociones/promociones_get?segmento=principal"

type AgricolaResponse struct {
	Promociones []AgricolaPromo `json:"promociones"`
}
type AgricolaPromo struct {
	IdPromocion     string `json:"id_promocion"`
	NombrePromocion string `json:"nombre_promocion"`
	Descripcion     string `json:"descripcion"`
	Restricciones   string `json:"restricciones"`
	ImagenBanner    string `json:"imagen_banner"`
	ImagenPreview   string `json:"imagen_preview"`
	Slug            string `json:"slug"`
	NombreComercio  string `json:"nombre_comercio"`
}

func FetchAgricola(ctx context.Context) ([]models.PromocionUnificada, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, AgricolaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create Agricola request: %w", err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Agricola promotions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Agricola API returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Agricola response body: %w", err)
	}

	var data AgricolaResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode Agricola JSON: %w", err)
	}

	unificadas := make([]models.PromocionUnificada, 0, len(data.Promociones))
	for _, p := range data.Promociones {
		urlImg := p.ImagenPreview
		if urlImg == "" {
			urlImg = p.ImagenBanner
		}

		unificadas = append(unificadas, models.PromocionUnificada{
			ID:                p.IdPromocion,
			BancoOrigen:       "AGRICOLA",
			Titulo:            utils.CleanText(p.NombrePromocion),
			DescripcionBreve:  utils.CleanText(p.Descripcion),
			UrlImagen:         urlImg,
			NombreComercio:    p.NombreComercio,
			RestriccionesHtml: p.Restricciones,
			UrlExterna:        p.Slug,
		})
	}
	return unificadas, nil
}
