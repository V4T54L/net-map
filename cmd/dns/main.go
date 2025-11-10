package main

import (
	"context"
	"fmt"
	"log"

	"internal-dns/configs"
	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/infrastructure/database"
	dnsTransport "internal-dns/internal/infrastructure/transport/dns"
	"internal-dns/internal/service"
	"internal-dns/pkg/bloomfilter"
)

func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	// Initialize database
	dbPool, err := database.NewPostgresPool(ctx, cfg.DB_URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established")

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(ctx, cfg.REDIS_ADDR, cfg.REDIS_PASSWORD, cfg.REDIS_DB)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connection established")

	// Initialize repositories
	dnsRecordRepo := database.NewDNSRecordPostgresRepository(dbPool)
	// dnsRecordRepo := database.NewDNSRecordInMemoryRepository()
	auditLogRepo := database.NewAuditLogPostgresRepository(dbPool) // Added auditLogRepo

	// Initialize Bloom Filter (needed for service, though not directly used by DNS server logic)
	bf := bloomfilter.NewRedisBloomFilter(redisClient, "dns_domains_bloom", cfg.BLOOM_FILTER_SIZE, cfg.BLOOM_FILTER_HASHES)

	// Initialize Cache
	dnsCache := cache.NewDNSRecordCache(redisClient)
	
	// Initialize Service
	dnsRecordService := service.NewDNSRecordService(dnsRecordRepo, bf, dnsCache, auditLogRepo) // Passed auditLogRepo

	// Initialize and start DNS server
	dnsServerAddr := fmt.Sprintf(":%s", cfg.DNS_PORT)
	server := dnsTransport.NewServer(dnsServerAddr, dnsRecordService, dnsCache)

	log.Printf("Starting DNS server on %s", dnsServerAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start DNS server: %v", err)
	}
}
