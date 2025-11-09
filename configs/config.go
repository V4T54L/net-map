package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	API_PORT string
	DNS_PORT string

	// Database
	DB_URL string

	// Redis
	REDIS_ADDR     string
	REDIS_PASSWORD string
	REDIS_DB       int

	// JWT
	JWT_SECRET_KEY string

	// Rate Limiter
	RATE_LIMITER_ENABLED bool
	RATE_LIMITER_RPS     float64 // requests per second
	RATE_LIMITER_BURST   int
	RATE_LIMITER_TTL     time.Duration

	// Bloom Filter
	BLOOM_FILTER_SIZE   uint
	BLOOM_FILTER_HASHES uint
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		API_PORT:             getEnv("API_PORT", "8080"),
		DNS_PORT:             getEnv("DNS_PORT", "5353"),
		DB_URL:               getEnv("DB_URL", "postgres://user:password@localhost:5432/dns_db?sslmode=disable"),
		REDIS_ADDR:           getEnv("REDIS_ADDR", "localhost:6379"),
		REDIS_PASSWORD:       getEnv("REDIS_PASSWORD", ""),
		REDIS_DB:             getEnvAsInt("REDIS_DB", 0),
		JWT_SECRET_KEY:       getEnv("JWT_SECRET_KEY", "a-very-secret-key-that-is-long-enough"),
		RATE_LIMITER_ENABLED: getEnvAsBool("RATE_LIMITER_ENABLED", true),
		RATE_LIMITER_RPS:     getEnvAsFloat64("RATE_LIMITER_RPS", 10),
		RATE_LIMITER_BURST:   getEnvAsInt("RATE_LIMITER_BURST", 20),
		RATE_LIMITER_TTL:     getEnvAsDuration("RATE_LIMITER_TTL", 1*time.Minute),
		BLOOM_FILTER_SIZE:    uint(getEnvAsInt("BLOOM_FILTER_SIZE", 100000)),
		BLOOM_FILTER_HASHES:  uint(getEnvAsInt("BLOOM_FILTER_HASHES", 4)),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return b
	}
	return fallback
}

func getEnvAsFloat64(key string, fallback float64) float64 {
	if value, ok := os.LookupEnv(key); ok {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fallback
		}
		return f
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(value)
		if err != nil {
			return fallback
		}
		return d
	}
	return fallback
}

