package main

import (
	"context"
	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/infrastructure/database"
	dnsTransport "internal-dns/internal/infrastructure/transport/dns"
	"internal-dns/internal/service"
	"internal-dns/pkg/bloomfilter"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	// --- Database ---
	dbPool, err := database.NewPostgresPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()

	// --- Redis ---
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisClient, err := cache.NewRedisClient(ctx, os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), redisDB)
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	// --- Repositories ---
	dnsRecordRepo := database.NewDNSRecordPostgresRepository(dbPool)

	// --- Bloom Filter (not used by DNS server directly, but needed for service) ---
	bloomFilterSize, _ := strconv.ParseUint(os.Getenv("BLOOM_FILTER_SIZE"), 10, 32)
	if bloomFilterSize == 0 {
		bloomFilterSize = 100000
	}
	bloomFilterHashes, _ := strconv.ParseUint(os.Getenv("BLOOM_FILTER_HASHES"), 10, 32)
	if bloomFilterHashes == 0 {
		bloomFilterHashes = 4
	}
	bf := bloomfilter.NewRedisBloomFilter(redisClient, "dns_domains_bloom", uint(bloomFilterSize), uint(bloomFilterHashes))

	// --- Cache ---
	dnsCache := cache.NewDNSRecordCache(redisClient)

	// --- Use Cases ---
	dnsRecordUC := service.NewDNSRecordService(dnsRecordRepo, bf, dnsCache)

	// --- DNS Server ---
	dnsPort := os.Getenv("DNS_PORT")
	if dnsPort == "" {
		dnsPort = "53"
	}
	dnsAddr := ":" + dnsPort
	server := dnsTransport.NewServer(dnsAddr, dnsRecordUC, dnsCache)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start DNS server: %v", err)
	}
}

