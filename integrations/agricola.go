package integrations

import (
	"encoding/json"
	"io"
	"net/http"
	"promotarjetas-backend/models"
	"promotarjetas-backend/utils"
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

func FetchAgricola() ([]models.PromocionUnificada, error) {
	url := "https://www.bancoagricola.com/com/promociones/promociones_get?segmento=principal"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data AgricolaResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var unificadas []models.PromocionUnificada
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
