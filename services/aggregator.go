package services

import (
	"log"
	"sync"
	
	"promotarjetas-backend/cache"
	"promotarjetas-backend/config"
	"promotarjetas-backend/integrations"
	"promotarjetas-backend/models"
	"sort"
)

var (
	syncMutex sync.Mutex
	isSyncing bool
)

// Orden de categorías según el diseño (de izquierda a derecha)
var categoryPriority = map[string]int{
	"Restaurantes":    1,
	"Supermercados":   2,
	"Viajes":          3,
	"Compras":         4,
	"Salud":           5,
	"Tecnología":      6,
	"Hogar":           7,
	"Entretenimiento": 8,
}

func SyncPromotions(cfg config.Config) {
	syncMutex.Lock()
	if isSyncing {
		syncMutex.Unlock()
		log.Println("Sincronización ya está en curso, saltando ejecución redundante.")
		return
	}
	isSyncing = true
	syncMutex.Unlock()

	defer func() {
		syncMutex.Lock()
		isSyncing = false
		syncMutex.Unlock()
	}()

	log.Println("Iniciando sincronización de promociones...")
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allPromotions []models.PromocionUnificada

	wg.Add(3)

	go func() {
		defer wg.Done()
		data, err := integrations.FetchAgricola()
		if err != nil {
			log.Printf("Error fetching Agricola: %v\n", err)
			return
		}
		mu.Lock()
		allPromotions = append(allPromotions, data...)
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		data, err := integrations.FetchBAC()
		if err != nil {
			log.Printf("Error fetching BAC: %v\n", err)
			return
		}
		mu.Lock()
		allPromotions = append(allPromotions, data...)
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		data, err := integrations.FetchCuscatlan(cfg.CuscatlanAPIKey)
		if err != nil {
			log.Printf("Error fetching Cuscatlan: %v\n", err)
			return
		}
		mu.Lock()
		allPromotions = append(allPromotions, data...)
		mu.Unlock()
	}()

	wg.Wait()
	
	// 1. Unificar categorías
	allPromotions = UnifyCategories(allPromotions)

	// 2. Ordenar promociones por categoría según la prioridad del diseño
	sort.Slice(allPromotions, func(i, j int) bool {
		p1, ok1 := categoryPriority[allPromotions[i].Categoria]
		p2, ok2 := categoryPriority[allPromotions[j].Categoria]

		// Si una categoría no está en nuestro mapa, se envía al final
		if !ok1 {
			p1 = 99
		}
		if !ok2 {
			p2 = 99
		}

		if p1 != p2 {
			return p1 < p2
		}

		// En caso de empate por categoría, ordenar por título para consistencia
		return allPromotions[i].Titulo < allPromotions[j].Titulo
	})
	
	log.Printf("Sincronizacion completada. Total agregadas: %d\n", len(allPromotions))
	
	if len(allPromotions) > 0 {
		err := cache.SavePromotions(allPromotions)
		if err != nil {
			log.Printf("Error guardando en Redis: %v\n", err)
		} else {
			log.Println("Promociones actualizadas en Redis exitosamente")
		}
	}
}
