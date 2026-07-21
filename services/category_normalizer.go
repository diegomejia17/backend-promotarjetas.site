package services

import (
	"promotarjetas-backend/models"
	"regexp"
	"strings"
)

func UnifyCategories(promotions []models.PromocionUnificada) []models.PromocionUnificada {
	for i := range promotions {
		p := &promotions[i]
		p.Categoria = GetUnifiedCategory(p)
	}
	return promotions
}

var categoryRegexes map[UnifiedCategory]*regexp.Regexp

func init() {
	categoryRegexes = make(map[UnifiedCategory]*regexp.Regexp)
	for _, m := range mappings {
		var escaped []string
		for _, k := range m.keywords {
			escaped = append(escaped, regexp.QuoteMeta(k))
		}
		// Límite de palabra compatible con unicode (tildes, eñes, números)
		pattern := `(?i)(^|[^\p{L}\p{N}])(` + strings.Join(escaped, "|") + `)([^\p{L}\p{N}]|$)`
		categoryRegexes[m.category] = regexp.MustCompile(pattern)
	}
}

func GetUnifiedCategory(p *models.PromocionUnificada) string {
	comercioLower := strings.ToLower(strings.TrimSpace(p.NombreComercio))

	// Fast path: Exact match from our overrides map
	if cat, ok := merchantCategoryOverrides[comercioLower]; ok {
		return string(cat)
	}

	comercio := p.NombreComercio
	raw := p.Categoria
	title := p.Titulo
	desc := p.DescripcionBreve

	scores := make(map[UnifiedCategory]int)

	for cat, re := range categoryRegexes {
		if re.MatchString(comercio) {
			scores[cat] += 10
		}
		if re.MatchString(raw) {
			scores[cat] += 5
		}
		if re.MatchString(title) {
			scores[cat] += 3
		}
		if re.MatchString(desc) {
			scores[cat] += 1
		}
	}

	bestCategory := CatOtros
	maxScore := 0

	// Iteramos sobre mappings para desempatar con un orden determinista
	for _, m := range mappings {
		score := scores[m.category]
		if score > maxScore {
			maxScore = score
			bestCategory = m.category
		}
	}

	return string(bestCategory)
}
