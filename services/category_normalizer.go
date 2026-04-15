package services

import (
	"promotarjetas-backend/models"
	"strings"
)

type UnifiedCategory string

const (
	CatHogar           UnifiedCategory = "Hogar"
	CatEntretenimiento UnifiedCategory = "Entretenimiento"
	CatTecnologia      UnifiedCategory = "Tecnología"
	CatSalud           UnifiedCategory = "Salud"
	CatCompras         UnifiedCategory = "Compras"
	CatSuperMercados   UnifiedCategory = "Supermercados"
	CatRestaurantes    UnifiedCategory = "Restaurantes"
	CatViajes          UnifiedCategory = "Viajes"
)

func UnifyCategories(promotions []models.PromocionUnificada) []models.PromocionUnificada {
	for i := range promotions {
		p := &promotions[i]
		p.Categoria = GetUnifiedCategory(p)
	}
	return promotions
}

type categoryMapping struct {
	category UnifiedCategory
	keywords []string
}

var mappings = []categoryMapping{
	{CatRestaurantes, []string{"restaurante", "café", "gastronom", "pizza", "hamburguesa", "comida", "sushi", "bistro", "cocina", "dining", "steak"}},
	{CatViajes, []string{"hotel", "vuelo", "viaje", "vacación", "turismo", "airline", "aerolínea", "boletos", "travel", "playa", "resort"}},
	{CatEntretenimiento, []string{"cine", "película", "teatro", "concierto", "diversión", "entretenimiento", "cinépolis", "museo", "parque"}},
	{CatSalud, []string{"salud", "médico", "clínica", "farmacia", "belleza", "spa", "salón", "dental", "laboratorio", "óptica", "health", "bienestar", "hospital"}},
	{CatTecnologia, []string{"televisor", "celular", "tecnolog", "gadget", "iphone", "apple", "samsung", "computadora", "audio", "gamer"}},
	{CatSuperMercados, []string{"supermercado", "despensa", "la colonia", "market", "gasolinera", "pantry", "selectos", "walmart", "pricesmart"}},
	{CatHogar, []string{"hogar", "mueble", "electro", "construcción", "pintura", "ferretería", "decoración", "colchón", "remodelación", "cama"}},
	{CatCompras, []string{"ropa", "zapato", "moda", "almacén", "tienda", "boutique", "joyería", "shoe", "store", "vestuario", "shopping", "regalo", "mall"}},
}

func GetUnifiedCategory(p *models.PromocionUnificada) string {
	raw := strings.ToLower(p.Categoria)
	title := strings.ToLower(p.Titulo)
	desc := strings.ToLower(p.DescripcionBreve)
	comercio := strings.ToLower(p.NombreComercio)

	fullText := title + " " + desc + " " + comercio + " " + raw

	for _, m := range mappings {
		for _, k := range m.keywords {
			if strings.Contains(fullText, k) {
				return string(m.category)
			}
		}
	}

	// Default for everything else
	return string(CatCompras)
}
