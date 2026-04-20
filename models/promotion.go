package models

type PromocionUnificada struct {
	ID                  string  `json:"id"`
	BancoOrigen         string  `json:"bancoOrigen"`
	Titulo              string  `json:"titulo"`
	DescripcionBreve    string  `json:"descripcionBreve"`
	UrlImagen           string  `json:"urlImagen"`
	NombreComercio      string  `json:"nombreComercio"`
	RestriccionesHtml   string  `json:"restriccionesHtml,omitempty"`
	Categoria           string  `json:"categoria,omitempty"`
	FechaInicio         string  `json:"fechaInicio,omitempty"`
	FechaFin            string  `json:"fechaFin,omitempty"`
	PorcentajeDescuento float64 `json:"porcentajeDescuento,omitempty"`
	UrlExterna          string  `json:"urlExterna,omitempty"`
	CreatedAt           int64   `json:"createdAt"`
}
