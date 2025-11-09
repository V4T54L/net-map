package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/infrastructure/database"
	httpTransport "internal-dns/internal/infrastructure/transport/http"
	"internal-dns/internal/repository"
	"internal-dns/internal/service"
	"internal-dns/internal/usecase" // Retained from original
	"internal-dns/internal/util"     // Retained from original
	"internal-dns/pkg/bloomfilter"
	"internal-dns/pkg/token"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// --- Database Setup ---
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := database.NewPostgresPool(ctx, connString)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection successful")

	// --- Redis & Bloom Filter Setup ---
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB, _ := strconv.Atoi(redisDBStr)

	redisClient, err := cache.NewRedisClient(ctx, redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Redis connection successful")

	// TODO: Make bloom filter size and hashes configurable
	bloomFilter := bloomfilter.NewRedisBloomFilter(redisClient, "dns:domains:bloom", 100000, 7)

	// --- Repositories ---
	userRepo := database.NewUserPostgresRepository(dbPool)
	dnsRecordRepo := database.NewDNSRecordPostgresRepository(dbPool)

	// --- Bloom Filter Population (on startup) ---
	go func() {
		log.Println("Populating Bloom filter with existing domain names...")
		domains, err := dnsRecordRepo.GetAllDomainNames(context.Background())
		if err != nil {
			log.Printf("Error fetching domain names for Bloom filter: %v", err)
			return
		}
		if len(domains) > 0 {
			if err := bloomFilter.AddMulti(context.Background(), domains); err != nil {
				log.Printf("Error populating Bloom filter: %v", err)
			}
		}
		log.Printf("Bloom filter populated with %d domains.", len(domains))
	}()

	// --- JWT ---
	jwtSecret := os.Getenv("JWT_SECRET") // Using JWT_SECRET from attempted
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	tokenGenerator := token.NewJWTGenerator(jwtSecret)

	// --- Services / Use Cases ---
	authService := service.NewAuthService(userRepo, tokenGenerator)
	userService := service.NewUserService(userRepo)
	dnsRecordService := service.NewDNSRecordService(dnsRecordRepo, bloomFilter)

	// Setup Echo HTTP server
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	httpTransport.RegisterRoutes(e, authService, userService, dnsRecordService, userRepo, tokenGenerator)

	// Start server
	port := os.Getenv("API_PORT") // Using API_PORT from attempted
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting API server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

