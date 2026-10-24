package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	Environment string

	// Email (optional). With no SMTP_HOST, verification codes are printed in the
	// server log (and shown on screen outside production) instead of being emailed.
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[CONFIG] No .env file found; using system environment variables.")
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: mustGetEnv("DATABASE_URL"),
		RedisURL:    mustGetEnv("REDIS_URL"),
		JWTSecret:   getEnv("JWT_SECRET", "default-dev-secret-key-change-me"),
		Environment: getEnv("ENV", "development"),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("[FATAL] Missing required environment variable: %s", key)
	}
	return val
}
