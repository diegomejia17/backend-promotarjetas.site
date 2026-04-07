# Promotarjetas Backend — BFF Aggregator

[![Go Version](https://img.shields.io/github/go-mod/go-version/diegomejia17/backend-promotarjetas.site)](https://go.dev/)
[![Docker](https://img.shields.io/badge/docker-%232496ED.svg?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

BFF (Backend for Frontend) robusto desarrollado en Go que agrega, limpia y categoriza promociones bancarias de múltiples instituciones en El Salvador (BAC, Cuscatlán, Agrícola).

## 🚀 Funcionalidades Principales

- **Agregación Multi-Banco**: Fetch concurrente de promociones utilizando goroutines.
- **Categorización Unificada**: Clasificación automática en 8 categorías fijas (Restaurantes, Supermercados, Viajes, Compras, Salud, Tecnología, Hogar, Entretenimiento).
- **Limpieza y Enriquecimiento**: Remoción de asteriscos (`*`), limpieza de HTML y fetch de detalles enriquecidos (h1_title, text_on_modal).
- **Caché en Redis**: Almacenamiento eficiente para asegurar tiempos de respuesta mínimos en el frontend.
- **Sincronización Inteligente**: Cron interno para actualización diaria a medianoche y soporte para sincronización forzada vía API.
- **Orden de Prioridad**: Resultados ordenados según la jerarquía visual requerida por el diseño.

## 🛠️ Stack Tecnológico

- **Lenguaje**: Go (Golang)
- **Framework Web**: Gin Gonic
- **Caché**: Redis 7
- **Despliegue**: Docker & Docker Compose
- **Sincronización**: Cron (robfig/cron)

## 📋 Requisitos Previos

- Docker y Docker Compose
- Go 1.21+ (opcional para desarrollo local)

## 🔧 Configuración y Ejecución

1.  Clona el repositorio:
    ```bash
    git clone https://github.com/diegomejia17/backend-promotarjetas.site.git
    cd backend-promotarjetas-backend
    ```

2.  Configura las variables de entorno:
    ```bash
    cp .env.example .env
    # Edita el .env con tu API Key de Cuscatlán
    ```

3.  Levanta el proyecto con Docker:
    ```bash
    docker compose up --build -d
    ```

El backend estará disponible en `http://localhost:3000`.

## 📡 API Endpoints

| Método | Ruta | Descripción |
| :--- | :--- | :--- |
| `GET` | `/api/promotions` | Listado completo de promociones unificadas |
| `GET` | `/api/promotions/sync` | Forzar sincronización manual de datos |

## 🏗️ Arquitectura del Proyecto

```text
├── cache/         # Lógica de conexión y persistencia en Redis
├── config/        # Gestión de variables de entorno y configuración
├── controllers/   # Manejadores de rutas Gin
├── integrations/  # Lógica de scraping/fetch para cada banco
├── models/        # Definiciones de structs y contrato unificado
├── services/      # Lógica de agregación, unificación y cron
├── utils/         # Utilidades de procesamiento de texto y HTML
└── main.go        # Punto de entrada de la aplicación
```

## 📄 Licencia

Este proyecto está bajo la [Licencia MIT](LICENSE).
