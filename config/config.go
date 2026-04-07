package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisURL        string
	RedisPassword   string
	CuscatlanAPIKey string
	Port            string
}

func LoadConfig() Config {
	// En producción no habrá .env; las variables vienen del entorno de la plataforma
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
		log.Println(os.Getenv("PORT"))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "localhost:6379"
	}

	return Config{
		RedisURL:        redisUrl,
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		CuscatlanAPIKey: os.Getenv("CUSCATLAN_API_KEY"),
		Port:            port,
	}
}
