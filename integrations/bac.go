package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"promotarjetas-backend/models"
	"promotarjetas-backend/utils"
	"time"
)

var BacURL = "https://api.mipromo.com/api/finaluser/deals/find"

type BacRequest struct {
	Filters []interface{} `json:"filters"`
	Sorters []interface{} `json:"sorters"`
	Pagers  BacPagers     `json:"pagers"`
}

type BacPagers struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
}

type BacResponse struct {
	Data struct {
		Response struct {
			NumFound int      `json:"numFound"`
			Docs     []BacDoc `json:"docs"`
		} `json:"response"`
	} `json:"data"`
}

type BacDoc struct {
	Id                   string       `json:"id"`
	Title                string       `json:"title"`
	Description          string       `json:"description"`
	Restrictions         string       `json:"restrictions"`
	ValidityFrom         string       `json:"validity_from"`
	ValidityTo           string       `json:"validity_to"`
	DiscountPercentValue float64      `json:"discount_percent_value"`
	CategoryTranslation  string       `json:"category_translation"`
	MerchantName         string       `json:"merchant_name"`
	Slug                 string       `json:"slug"`
	ChildDocuments       BacChildDocs `json:"_childDocuments_"`
}

type BacChildDocs struct {
	Image []BacImage `json:"IMAGE"`
}

type BacImage struct {
	ImageFilepath string `json:"image_filepath"`
	ImageFilename string `json:"image_filename"`
}

func FetchBAC(ctx context.Context) ([]models.PromocionUnificada, error) {
	seen := make(map[string]bool)
	limit := 18
	start := 0
	numFound := 1 // To enter loop

	var unificadas []models.PromocionUnificada
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	for start < numFound {
		reqBody := fmt.Sprintf(`data=%%7B%%22filters%%22%%3A%%5B%%5D%%2C%%22sorters%%22%%3A%%5B%%5D%%2C%%22pagers%%22%%3A%%7B%%22start%%22%%3A%d%%2C%%22limit%%22%%3A%d%%7D%%7D`, start, limit)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, BacURL, bytes.NewBuffer([]byte(reqBody)))
		if err != nil {
			return nil, fmt.Errorf("create BAC request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("country-id", "60")
		req.Header.Set("locale", "es")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request BAC deals: %w", err)
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			resp.Body.Close()
			return nil, fmt.Errorf("BAC API returned HTTP %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read BAC response body: %w", err)
		}

		var data BacResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("decode BAC JSON: %w", err)
		}

		numFound = data.Data.Response.NumFound
		if numFound == 0 {
			break
		}

		for _, doc := range data.Data.Response.Docs {
			urlImg := ""
			if len(doc.ChildDocuments.Image) > 0 {
				img := doc.ChildDocuments.Image[0]
				baseUrl := "https://mipromoimages.geopagoscdn.net"
				queryParams := ""
				urlImg = baseUrl + img.ImageFilepath + "/" + img.ImageFilename + queryParams
			}

			if !seen[doc.Id] {
				seen[doc.Id] = true
				unificadas = append(unificadas, models.PromocionUnificada{
					ID:                  doc.Id,
					BancoOrigen:         "BAC",
					Titulo:              utils.CleanText(doc.Title),
					DescripcionBreve:    utils.StripTags(doc.Description),
					UrlImagen:           urlImg,
					NombreComercio:      doc.MerchantName,
					RestriccionesHtml:   utils.DecodeHtml(doc.Description + "<br/>" + doc.Restrictions),
					Categoria:           doc.CategoryTranslation,
					FechaInicio:         doc.ValidityFrom,
					FechaFin:            doc.ValidityTo,
					PorcentajeDescuento: doc.DiscountPercentValue,
					UrlExterna:          "https://mipromo.com/sv/deal/" + doc.Id,
				})
			}

		}
		start += limit
	}

	return unificadas, nil
}
