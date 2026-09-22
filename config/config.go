package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisURL            string
	RedisPassword       string
	CuscatlanAPIKey     string
	Port                string
	AgricolaURL         string
	AgricolaCookie      string
	TypeSafeAPIKey      string
	TypeSafeBaseURL     string
	TypeSafeModel       string
	TypeSafeClassifyAll bool
}

func LoadConfig() Config {
	// En producción no habrá .env; las variables vienen del entorno de la plataforma
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "localhost:6379"
	}

	agricolaUrl := os.Getenv("AGRICOLA_URL")
	if agricolaUrl == "" {
		agricolaUrl = "https://www.bancoagricola.com/com/promociones/promociones_get?segmento=principal"
	}

	typeSafeBaseUrl := os.Getenv("TYPESAFE_BASE_URL")
	if typeSafeBaseUrl == "" {
		typeSafeBaseUrl = "https://api.typesafe.ai/v1/systemone"
	}

	typeSafeModel := os.Getenv("TYPESAFE_MODEL")
	if typeSafeModel == "" {
		typeSafeModel = "jev-latest"
	}

	classifyAll := os.Getenv("TYPESAFE_CLASSIFY_ALL") == "true" || os.Getenv("TYPESAFE_CLASSIFY_ALL") == "1"

	return Config{
		RedisURL:            redisUrl,
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		CuscatlanAPIKey:     os.Getenv("CUSCATLAN_API_KEY"),
		Port:                port,
		AgricolaURL:         agricolaUrl,
		AgricolaCookie:      os.Getenv("AGRICOLA_COOKIE"),
		TypeSafeAPIKey:      os.Getenv("TYPESAFE_API_KEY"),
		TypeSafeBaseURL:     typeSafeBaseUrl,
		TypeSafeModel:       typeSafeModel,
		TypeSafeClassifyAll: classifyAll,
	}
}
