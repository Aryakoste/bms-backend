package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment string
}

func Load() *Config {
	// Load .env file (ignore error if not exists)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "mongodb://localhost:27017/building_management_society"),
		JWTSecret:   getEnv("JWT_SECRET", "jwt-secret"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	log.Printf("🔧 Configuration loaded:")
	log.Printf("   Port: %s", cfg.Port)
	log.Printf("   Database: %s", cfg.DatabaseURL)
	log.Printf("   Environment: %s", cfg.Environment)

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
