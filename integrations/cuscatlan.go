package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"promotarjetas-backend/models"
	"promotarjetas-backend/utils"
)

var CuscatlanAPIURL = "https://apigw.bancocuscatlan.com/webapi/"

const (
	cuscatlanBaseURL = "https://www.bancocuscatlan.com/tarjetas/promociones/promocion"
	cuscatlanBank    = "CUSCATLAN"
)

type CuscatlanRequest struct {
	Query string `json:"query"`
}

type CuscatlanResponse struct {
	Data struct {
		Promocions struct {
			Data []CuscatlanDoc `json:"data"`
		} `json:"promocions"`
		Coupons struct {
			Data []CuscatlanCoupon `json:"data"`
		} `json:"coupons"`
	} `json:"data"`
}

type CuscatlanCoupon struct {
	Id         string `json:"id"`
	Attributes struct {
		Title       string `json:"title"`
		PublishedAt string `json:"publishedAt"`
		TermsCond   string `json:"terms_cond"`
		Imagen      struct {
			Data struct {
				Attributes struct {
					Url string `json:"url"`
				} `json:"attributes"`
			} `json:"data"`
		} `json:"imagen"`
	} `json:"attributes"`
}

type CuscatlanDoc struct {
	Id         string `json:"id"`
	Attributes struct {
		PublishedAt string `json:"publishedAt"`
		H1Title     string `json:"h1_title"`
		DateStart   string `json:"date_start"`
		DateEnd     string `json:"date_end"`
		Business    struct {
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

func FetchCuscatlan(ctx context.Context, apiKey string) ([]models.PromocionUnificada, error) {
	currentDate := time.Now().Format("2006-01-02")
	reqBody := CuscatlanRequest{Query: cuscatlanQuery(currentDate)}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal Cuscatlan request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, CuscatlanAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create Cuscatlan request: %w", err)
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
		return nil, fmt.Errorf("request Cuscatlan promotions: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Cuscatlan response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Cuscatlan API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var data CuscatlanResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode Cuscatlan response: %w", err)
	}

	unificadas := make([]models.PromocionUnificada, 0, len(data.Data.Promocions.Data)+len(data.Data.Coupons.Data))
	for _, doc := range data.Data.Promocions.Data {
		unificadas = append(unificadas, cuscatlanPromotion(doc))
	}

	for _, cup := range data.Data.Coupons.Data {
		unificadas = append(unificadas, cuscatlanCoupon(cup))
	}

	return unificadas, nil
}

func cuscatlanQuery(currentDate string) string {
	return fmt.Sprintf(`query PromocionesYCupones { promocions (pagination:{limit:100} sort:"priority" filters: {or:[{hide:{eq: null}} {hide:{eq: false}}] and:[{date_start:{lte:"%s"} date_end:{gte:"%s"}}] business:{id:{not:null}}}) { data{ id attributes { h1_title tags{data{attributes{description}}} business {data{attributes{name description logo{data{attributes{url}}}}}} card { title description imagen {data{attributes{url}}}} detail_promotion{title subtitle list_bullets{text} action{text_on_modal}} date_start date_end priority open_graph{og_title og_description og_image{data{attributes{url}}}}  }}} coupons(pagination:{limit:100} sort:"priority"){data{id attributes{title publishedAt priority terms_cond imagen{data{attributes{url}}}}}} }`, currentDate, currentDate)
}

func cuscatlanPromotion(doc CuscatlanDoc) models.PromocionUnificada {
	attributes := doc.Attributes
	title := attributes.H1Title
	if title == "" {
		title = attributes.Card.Title
	}

	briefDescription := utils.StripTags(attributes.DetailPromotion.Action.TextOnModal)
	if briefDescription == "" {
		briefDescription = attributes.Card.Description
	}

	business := attributes.Business.Data.Attributes
	return models.PromocionUnificada{
		ID:                doc.Id,
		BancoOrigen:       cuscatlanBank,
		Titulo:            utils.CleanText(title),
		DescripcionBreve:  utils.CleanText(briefDescription),
		UrlImagen:         attributes.Card.Imagen.Data.Attributes.Url,
		NombreComercio:    business.Name,
		Categoria:         cuscatlanCategory(doc),
		FechaInicio:       attributes.DateStart,
		FechaFin:          attributes.DateEnd,
		RestriccionesHtml: cuscatlanRestrictions(doc),
		UrlExterna:        fmt.Sprintf("%s/%s/%s", cuscatlanBaseURL, neturl.PathEscape(utils.CleanText(business.Name)), doc.Id),
	}
}

func cuscatlanCoupon(coupon CuscatlanCoupon) models.PromocionUnificada {
	attributes := coupon.Attributes
	return models.PromocionUnificada{
		ID:                coupon.Id,
		BancoOrigen:       cuscatlanBank,
		Titulo:            utils.CleanText(attributes.Title),
		DescripcionBreve:  utils.CleanText(utils.StripTags(attributes.TermsCond)),
		UrlImagen:         attributes.Imagen.Data.Attributes.Url,
		NombreComercio:    attributes.Title,
		Categoria:         "Cupones",
		FechaInicio:       attributes.PublishedAt,
		RestriccionesHtml: attributes.TermsCond,
	}
}

func cuscatlanCategory(doc CuscatlanDoc) string {
	if len(doc.Attributes.Tags.Data) == 0 {
		return ""
	}
	return doc.Attributes.Tags.Data[0].Attributes.Description
}

func cuscatlanRestrictions(doc CuscatlanDoc) string {
	detail := doc.Attributes.DetailPromotion
	var restrictions strings.Builder

	if detail.Title != "" || detail.Subtitle != "" || len(detail.ListBullets) > 0 {
		fmt.Fprintf(&restrictions, "<h3>%s</h3><h4>%s</h4><ul>", detail.Title, detail.Subtitle)
		for _, bullet := range detail.ListBullets {
			fmt.Fprintf(&restrictions, "<li>%s</li>", bullet.Text)
		}
		restrictions.WriteString("</ul>")
	}

	if detail.Action.TextOnModal != "" {
		fmt.Fprintf(&restrictions, "<div class='modal-text'>%s</div>", detail.Action.TextOnModal)
	}

	if description := doc.Attributes.Business.Data.Attributes.Description; description != "" {
		return fmt.Sprintf("<p><strong>Sobre el comercio:</strong> %s</p>%s", description, restrictions.String())
	}
	return restrictions.String()
}
