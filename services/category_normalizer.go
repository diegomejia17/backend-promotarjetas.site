package services

import (
	"context"
	"log"
	"regexp"
	"strings"
	"sync"

	"promotarjetas-backend/models"
)

// CategoryClassifier define la interfaz para clasificar promociones con modelos de IA como Jev
type CategoryClassifier interface {
	ClassifyPromotion(ctx context.Context, promo models.PromocionUnificada) (string, float64, error)
}

// ValidCategories contiene las categorías unificadas válidas del sistema
var ValidCategories = map[string]UnifiedCategory{
	string(CatHogar):           CatHogar,
	string(CatEntretenimiento): CatEntretenimiento,
	string(CatTecnologia):      CatTecnologia,
	string(CatSalud):           CatSalud,
	string(CatCompras):         CatCompras,
	string(CatSuperMercados):   CatSuperMercados,
	string(CatRestaurantes):    CatRestaurantes,
	string(CatViajes):          CatViajes,
	string(CatOtros):           CatOtros,
}

// UnifyCategories unifica las categorías de una lista de promociones utilizando el motor de reglas estándar
func UnifyCategories(promotions []models.PromocionUnificada) []models.PromocionUnificada {
	return UnifyCategoriesWithClassifier(context.Background(), promotions, nil, false)
}

// UnifyCategoriesWithClassifier unifica categorías aplicando primero reglas deterministas (overrides y regex)
// y utilizando el clasificador (Jev) para resolver promociones no categorizadas ("Otros") o todas las que no tengan override.
func UnifyCategoriesWithClassifier(
	ctx context.Context,
	promotions []models.PromocionUnificada,
	classifier CategoryClassifier,
	classifyAll bool,
) []models.PromocionUnificada {
	var candidateIndices []int

	for i := range promotions {
		p := &promotions[i]
		p.Categoria = GetUnifiedCategory(p)

		if classifier == nil {
			continue
		}

		comercioLower := strings.ToLower(strings.TrimSpace(p.NombreComercio))
		_, hasOverride := merchantCategoryOverrides[comercioLower]
		if hasOverride {
			continue // Los overrides de comercios conocidos se respetan siempre
		}

		// Candidata a clasificación con Jev si quedó en Otros o si classifyAll está activado
		if p.Categoria == string(CatOtros) || classifyAll {
			candidateIndices = append(candidateIndices, i)
		}
	}

	if len(candidateIndices) == 0 || classifier == nil {
		return promotions
	}

	const maxWorkers = 5
	workerCount := maxWorkers
	if len(candidateIndices) < workerCount {
		workerCount = len(candidateIndices)
	}

	type job struct {
		idx int
	}
	jobs := make(chan job, len(candidateIndices))
	for _, idx := range candidateIndices {
		jobs <- job{idx: idx}
	}
	close(jobs)

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				promo := promotions[j.idx]
				cat, conf, err := classifier.ClassifyPromotion(ctx, promo)
				if err != nil {
					log.Printf("Aviso: No se pudo clasificar promocion %s (%s) con Jev: %v", promo.ID, promo.NombreComercio, err)
					continue
				}

				// Si la confianza es suficiente y la categoría es válida, la asignamos
				if conf >= 0.40 {
					if validCat, ok := ValidCategories[cat]; ok && validCat != CatOtros {
						promotions[j.idx].Categoria = string(validCat)
					}
				}
			}
		}()
	}
	wg.Wait()

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
