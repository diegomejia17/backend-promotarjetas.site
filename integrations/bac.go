package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"promotarjetas-backend/models"
	"promotarjetas-backend/utils"
)

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

func FetchBAC() ([]models.PromocionUnificada, error) {
	url := "https://api.mipromo.com/api/finaluser/deals/find"

	limit := 18
	start := 0
	numFound := 1 // To enter loop

	var unificadas []models.PromocionUnificada

	for start < numFound {
		reqBody := fmt.Sprintf(`data=%%7B%%22filters%%22%%3A%%5B%%5D%%2C%%22sorters%%22%%3A%%5B%%5D%%2C%%22pagers%%22%%3A%%7B%%22start%%22%%3A%d%%2C%%22limit%%22%%3A%d%%7D%%7D`, start, limit)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(reqBody)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("country-id", "60")
		req.Header.Set("locale", "es")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var data BacResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}

		numFound = data.Data.Response.NumFound

		for _, doc := range data.Data.Response.Docs {
			urlImg := ""
			if len(doc.ChildDocuments.Image) > 0 {
				img := doc.ChildDocuments.Image[0]
				// Nueva lógica de imágenes de Geopagos CDN
				baseUrl := "https://mipromoimages.geopagoscdn.net"
				queryParams := ""
				urlImg = baseUrl + img.ImageFilepath + "/" + img.ImageFilename + queryParams
			}

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
				UrlExterna:          "https://mipromo.com/sv/deal/" + doc.Slug,
			})
		}
		start += limit
	}

	return unificadas, nil
}
