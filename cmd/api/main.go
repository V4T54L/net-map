package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/infrastructure/database"
	"internal-dns/internal/infrastructure/transport/http/routes"
	"internal-dns/internal/service"
	"internal-dns/internal/usecase" // Keep usecase import for service interfaces
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- Database Setup ---
	dbPool, err := database.NewPostgresPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established")

	// --- Redis & Bloom Filter Setup ---
	redisClient, err := cache.NewRedisClient(ctx, os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), 0) // Using 0 for default DB
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connection established")

	// Hardcoding bloom filter parameters as per attempted content
	bf := bloomfilter.NewRedisBloomFilter(redisClient, "dns_domains_bloom", 100000, 7)

	// --- Repositories ---
	userRepo := database.NewUserPostgresRepository(dbPool)
	dnsRecordRepo := database.NewDNSRecordPostgresRepository(dbPool)
	auditLogRepo := database.NewAuditLogPostgresRepository(dbPool)

	// --- Bloom Filter Population (on startup) ---
	go func() {
		log.Println("Populating Bloom filter with existing domain names...")
		domains, err := dnsRecordRepo.GetAllDomainNames(context.Background())
		if err != nil {
			log.Printf("Error fetching domain names for Bloom filter: %v", err)
			return
		}
		if len(domains) > 0 {
			if err := bf.AddMulti(context.Background(), domains); err != nil {
				log.Printf("Error populating Bloom filter: %v", err)
			}
		}
		log.Printf("Bloom filter populated with %d domains.", len(domains))
	}()

	// --- JWT ---
	jwtSecret := os.Getenv("JWT_SECRET_KEY") // Changed env var name
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET_KEY environment variable is not set")
	}
	tokenGenerator := token.NewJWTGenerator(jwtSecret)

	// --- Cache ---
	dnsCache := cache.NewDNSRecordCache(redisClient)

	// --- Services / Use Cases ---
	authService := service.NewAuthService(userRepo, tokenGenerator, auditLogRepo)
	userService := service.NewUserService(userRepo, auditLogRepo)
	dnsRecordService := service.NewDNSRecordService(dnsRecordRepo, bf, dnsCache, auditLogRepo)

	// Setup Echo HTTP server
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	routes.RegisterRoutes(e, authService, userService, dnsRecordService, userRepo, tokenGenerator)

	// Start server
	apiPort := os.Getenv("HTTP_PORT") // Changed env var name
	if apiPort == "" {
		apiPort = "8080"
	}
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", apiPort)))
}

