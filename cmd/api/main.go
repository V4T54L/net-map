package main

import (
	"context"
	"fmt"
	"log"
	"time" // Keep time for context timeout if needed, but attempted removes it. Sticking with attempted's context.Background()

	"internal-dns/configs"
	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/infrastructure/database"
	"internal-dns/internal/infrastructure/transport/http"
	"internal-dns/internal/service"
	"internal-dns/internal/usecase" // Keep usecase import for service interfaces
	"internal-dns/pkg/bloomfilter"
	"internal-dns/pkg/token"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// @title Internal DNS Server API
// @version 1.0
// @description This is the API for the Internal DNS management service.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background() // Using context.Background() as per attempted content

	// --- Database Setup ---
	dbPool, err := database.NewPostgresPool(ctx, cfg.DB_URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established")

	// --- Redis & Bloom Filter Setup ---
	redisClient, err := cache.NewRedisClient(ctx, cfg.REDIS_ADDR, cfg.REDIS_PASSWORD, cfg.REDIS_DB)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connection established")

	// Initialize Bloom Filter
	bf := bloomfilter.NewRedisBloomFilter(redisClient, "dns_domains_bloom", cfg.BLOOM_FILTER_SIZE, cfg.BLOOM_FILTER_HASHES)

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
	tokenGenerator := token.NewJWTGenerator(cfg.JWT_SECRET_KEY)

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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.POST, echo.GET, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Register routes
	http.RegisterRoutes(e, cfg, authService, userService, dnsRecordService, userRepo, tokenGenerator)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.API_PORT)
	log.Printf("Starting API server on %s", serverAddr)
	if err := e.Start(serverAddr); err != nil {
		e.Logger.Fatal(err)
	}
}

