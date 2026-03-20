package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	NeoHubBaseURL        string
	NeoHubAPIKey         string
	NeoHubWabaID         string
	WatsonXBaseURL       string
	WatsonXAPIKey        string
	WatsonXAssistantID   string
	WatsonXEnvironmentID string
	WatsonXVersion       string
	RedisAddr            string
	RedisPassword        string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		NeoHubBaseURL:        getEnv("NEOHUB_BASE_URL", ""),
		NeoHubAPIKey:         getEnv("NEOHUB_API_KEY", ""),
		NeoHubWabaID:         getEnv("NEOHUB_WABA_ID", ""),
		WatsonXBaseURL:       getEnv("WATSONX_BASE_URL", ""),
		WatsonXAPIKey:        getEnv("WATSONX_API_KEY", ""),
		WatsonXAssistantID:   getEnv("WATSONX_ASSISTANT_ID", ""),
		WatsonXEnvironmentID: getEnv("WATSONX_ENVIRONMENT_ID", ""),
		WatsonXVersion:       getEnv("WATSONX_VERSION", "2021-11-27"),
		RedisAddr:            getEnv("REDIS_ADDR", ""),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
