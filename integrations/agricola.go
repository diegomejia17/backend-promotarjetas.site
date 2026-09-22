package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"promotarjetas-backend/models"
	"promotarjetas-backend/utils"
)

const (
	DefaultAgricolaURL                = "https://www.bancoagricola.com/com/promociones/promociones_get?segmento=principal"
	DefaultAgricolaPromocionesPageURL = "https://www.bancoagricola.com/promociones"
)

var (
	AgricolaURL                = DefaultAgricolaURL
	AgricolaPromocionesPageURL = DefaultAgricolaPromocionesPageURL
)

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

func FetchAgricola(ctx context.Context, apiURL string, cookie string) ([]models.PromocionUnificada, error) {
	targetURL := apiURL
	if targetURL == "" {
		targetURL = AgricolaURL
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create Agricola cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
	}

	// Auto-negociar cookies de sesión visitando la página de promociones si no se especificó cookie manual
	if cookie == "" && (strings.Contains(targetURL, "bancoagricola.com") || (AgricolaPromocionesPageURL != DefaultAgricolaPromocionesPageURL && AgricolaPromocionesPageURL != "")) {
		warmReq, err := http.NewRequestWithContext(ctx, http.MethodGet, AgricolaPromocionesPageURL, nil)
		if err == nil {
			warmReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:155.0) Gecko/20100101 Firefox/155.0")
			warmReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			warmReq.Header.Set("Accept-Language", "es-SV,en-US;q=0.9,en;q=0.8")
			warmResp, err := client.Do(warmReq)
			if err == nil {
				io.Copy(io.Discard, warmResp.Body)
				warmResp.Body.Close()
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create Agricola request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:155.0) Gecko/20100101 Firefox/155.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "es-SV,en-US;q=0.9,en;q=0.8")
	req.Header.Set("Referer", AgricolaPromocionesPageURL)
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
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
