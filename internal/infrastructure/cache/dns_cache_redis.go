package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"internal-dns/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	dnsCacheKeyPrefix = "dns_cache:"
	dnsCacheTTL       = 5 * time.Minute
)

var ErrCacheMiss = errors.New("cache: key not found")

// DNSRecordCache defines the interface for a DNS record cache.
type DNSRecordCache interface {
	Get(ctx context.Context, domainName string) (*domain.DNSRecord, error)
	Set(ctx context.Context, record *domain.DNSRecord) error
	Delete(ctx context.Context, domainName string) error
}

type dnsCacheRedis struct {
	client *redis.Client
}

// NewDNSRecordCache creates a new Redis-backed DNS record cache.
func NewDNSRecordCache(client *redis.Client) DNSRecordCache {
	return &dnsCacheRedis{client: client}
}

func (c *dnsCacheRedis) Get(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	key := dnsCacheKeyPrefix + domainName
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}

	var record domain.DNSRecord
	if err := json.Unmarshal([]byte(val), &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record from cache: %w", err)
	}

	return &record, nil
}

func (c *dnsCacheRedis) Set(ctx context.Context, record *domain.DNSRecord) error {
	key := dnsCacheKeyPrefix + record.DomainName
	val, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record for cache: %w", err)
	}

	if err := c.client.Set(ctx, key, val, dnsCacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set to redis: %w", err)
	}

	return nil
}

func (c *dnsCacheRedis) Delete(ctx context.Context, domainName string) error {
	key := dnsCacheKeyPrefix + domainName
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from redis: %w", err)
	}
	return nil
}

