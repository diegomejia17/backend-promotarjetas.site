package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"promotarjetas-backend/models"
	"promotarjetas-backend/utils"
	"time"
)

type CuscatlanRequest struct {
	Query string `json:"query"`
}

type CuscatlanResponse struct {
	Data struct {
		Promocions struct {
			Data []CuscatlanDoc `json:"data"`
		} `json:"promocions"`
	} `json:"data"`
}

type CuscatlanDoc struct {
	Id         string `json:"id"`
	Attributes struct {
		PublishedAt string `json:"publishedAt"`
		H1Title     string `json:"h1_title"`
		DateStart   string `json:"date_start"`
		DateEnd   string `json:"date_end"`
		Business  struct {
			Data struct {
				Attributes struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"attributes"`
			} `json:"data"`
		} `json:"business"`
		Tags struct {
			Data []struct {
				Attributes struct {
					Description string `json:"description"`
				} `json:"attributes"`
			} `json:"data"`
		} `json:"tags"`
		Card struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Imagen      struct {
				Data struct {
					Attributes struct {
						Url string `json:"url"`
					} `json:"attributes"`
				} `json:"data"`
			} `json:"imagen"`
		} `json:"card"`
		DetailPromotion struct {
			Title       string `json:"title"`
			Subtitle    string `json:"subtitle"`
			ListBullets []struct {
				Text string `json:"text"`
			} `json:"list_bullets"`
			Action struct {
				TextOnModal string `json:"text_on_modal"`
			} `json:"action"`
		} `json:"detail_promotion"`
	} `json:"attributes"`
}

func FetchCuscatlan(apiKey string) ([]models.PromocionUnificada, error) {
	url := "https://apigw.bancocuscatlan.com/webapi/"

	currentDate := time.Now().Format("2006-01-02")

	// Expandimos la query para traer h1_title, detail_promotion con text_on_modal usando la fecha actual
	queryStr := fmt.Sprintf(`query Promociones { promocions (pagination:{limit:100} sort:"priority" filters: {or:[{hide:{eq: null}} {hide:{eq: false}}] and:[{date_start:{lte:"%s"} date_end:{gte:"%s"}}] business:{id:{not:null}}}) { data{ id attributes { h1_title tags{data{attributes{description}}} business {data{attributes{name description logo{data{attributes{url}}}}}} card { title description imagen {data{attributes{url}}}} detail_promotion{title subtitle list_bullets{text} action{text_on_modal}} date_start date_end priority open_graph{og_title og_description og_image{data{attributes{url}}}}  }}}}`, currentDate, currentDate)

	reqBody := CuscatlanRequest{Query: queryStr}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("apikey", apiKey)
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data CuscatlanResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var unificadas []models.PromocionUnificada
	for _, doc := range data.Data.Promocions.Data {
		categoria := ""
		if len(doc.Attributes.Tags.Data) > 0 {
			categoria = doc.Attributes.Tags.Data[0].Attributes.Description
		}

		restricciones := ""
		dp := doc.Attributes.DetailPromotion
		if dp.Title != "" || dp.Subtitle != "" || len(dp.ListBullets) > 0 {
			restricciones = fmt.Sprintf("<h3>%s</h3><h4>%s</h4><ul>", dp.Title, dp.Subtitle)
			for _, bullet := range dp.ListBullets {
				restricciones += fmt.Sprintf("<li>%s</li>", bullet.Text)
			}
			restricciones += "</ul>"
		}

		if dp.Action.TextOnModal != "" {
			restricciones += fmt.Sprintf("<div class='modal-text'>%s</div>", dp.Action.TextOnModal)
		}

		if doc.Attributes.Business.Data.Attributes.Description != "" {
			restricciones = fmt.Sprintf("<p><strong>Sobre el comercio:</strong> %s</p>%s", 
				doc.Attributes.Business.Data.Attributes.Description, restricciones)
		}

		titulo := doc.Attributes.H1Title
		if titulo == "" {
			titulo = doc.Attributes.Card.Title
		}

		resBrief := utils.StripTags(doc.Attributes.DetailPromotion.Action.TextOnModal)
		if resBrief == "" {
			resBrief = doc.Attributes.Card.Description
		}

		unificadas = append(unificadas, models.PromocionUnificada{
			ID:                doc.Id,
			BancoOrigen:       "CUSCATLAN",
			Titulo:            utils.CleanText(titulo),
			DescripcionBreve:  utils.CleanText(resBrief),
			UrlImagen:         doc.Attributes.Card.Imagen.Data.Attributes.Url,
			NombreComercio:    doc.Attributes.Business.Data.Attributes.Name,
			Categoria:         categoria,
			FechaInicio:       doc.Attributes.DateStart,
			FechaFin:          doc.Attributes.DateEnd,
			RestriccionesHtml: restricciones,
		})
	}
	return unificadas, nil
}
