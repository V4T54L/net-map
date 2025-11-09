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
	"internal-dns/internal/usecase"
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
	log.Println("Database connection successful")

	// --- Redis & Bloom Filter Setup ---
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisClient, err := cache.NewRedisClient(ctx, os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), redisDB)
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Redis connection successful")

	bloomFilterSize, _ := strconv.ParseUint(os.Getenv("BLOOM_FILTER_SIZE"), 10, 32)
	if bloomFilterSize == 0 {
		bloomFilterSize = 100000
	}
	bloomFilterHashes, _ := strconv.ParseUint(os.Getenv("BLOOM_FILTER_HASHES"), 10, 32)
	if bloomFilterHashes == 0 {
		bloomFilterHashes = 4
	}
	bloomFilter := bloomfilter.NewRedisBloomFilter(redisClient, "dns_domains_bloom", uint(bloomFilterSize), uint(bloomFilterHashes))

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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	tokenGenerator := token.NewJWTGenerator(jwtSecret)

	// --- Cache ---
	dnsCache := cache.NewDNSRecordCache(redisClient)

	// --- Services / Use Cases ---
	authService := service.NewAuthService(userRepo, tokenGenerator)
	userService := service.NewUserService(userRepo)
	dnsRecordService := service.NewDNSRecordService(dnsRecordRepo, bloomFilter, dnsCache)

	// Setup Echo HTTP server
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	httpTransport.RegisterRoutes(e, authService, userService, dnsRecordService, userRepo, tokenGenerator)

	// Start server
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}
	e.Logger.Fatal(e.Start(":" + apiPort))
}

